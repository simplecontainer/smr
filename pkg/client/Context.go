package client

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func NewContext(clientRootDir string) *Context {
	return &Context{
		ApiURL:        "",
		Name:          "",
		Directory:     fmt.Sprintf("%s/%s", clientRootDir, static.CONTEXTDIR),
		CertBundle:    "",
		ActiveContext: "",
		PrivateKey:    &bytes.Buffer{},
		Cert:          &bytes.Buffer{},
		Ca:            &bytes.Buffer{},
	}
}

func (c *Context) LoadContext() (*Context, error) {
	k := keys.NewKeys()
	err := k.LoadClient(c.Name, c.CertBundle)

	if err != nil {
		return nil, err
	}

	hostport, err := configuration.NewHostPort(c.ApiURL)

	if err != nil {
		return nil, err
	}

	if c.GetActiveContext() {
		if c.ReadFromFile() {
			client, err := clients.GenerateHttpClients(k, *hostport, nil)

			if err != nil {
				return nil, err
			}

			c.Client = client.Get(c.Name).Http

			return c, nil
		}
	}

	return nil, errors.New("context not found")
}

func (c *Context) GetActiveContext() bool {
	activeContextPath := fmt.Sprintf("%s/%s", c.Directory, ".active")

	activeContext, err := os.ReadFile(activeContextPath)
	if err != nil {
		return false
	}

	if string(activeContext) == "" {
		activeContext = []byte("default")
	}

	c.ActiveContext = fmt.Sprintf("%s/%s", c.Directory, string(activeContext))
	return true
}
func (c *Context) SetActiveContext(contextName string) bool {
	activeContextPath := fmt.Sprintf("%s/%s", c.Directory, ".active")

	err := os.WriteFile(activeContextPath, []byte(contextName), 0755)
	if err != nil {
		logger.Log.Error("active context file not saved", zap.Error(err))
	}

	return true
}

func (c *Context) ReadFromFile() bool {
	activeContext, err := os.ReadFile(c.ActiveContext)
	if err != nil {
		logger.Log.Info("active context file not found", zap.Error(err))
		return false
	}

	if err = json.Unmarshal(activeContext, &c); err != nil {
		logger.Log.Info("active context file not valid json", zap.Error(err))
		return false
	}

	return true
}
func (c *Context) SaveToFile() error {
	if c.Name == "" {
		c.Name = viper.GetString("context")
	}

	if c.Name == "" {
		return errors.New("context name cannot be empty")
	}

	jsonData, err := json.Marshal(c)

	if err != nil {
		return err
	}

	contextPath := fmt.Sprintf("%s/%s", c.Directory, c.Name)

	if _, err = os.Stat(contextPath); err == nil {
		if viper.GetBool("y") || helpers.Confirm("Context with the same name already exists. Do you want to overwrite it?") {
			err = os.WriteFile(contextPath, jsonData, 0600)
			if err != nil {
				logger.Log.Error("active context file not saved", zap.Error(err))
			}

			activeContextPath := fmt.Sprintf("%s/%s", c.Directory, ".active")

			err = os.WriteFile(activeContextPath, []byte(c.Name), 0755)
			if err != nil {
				logger.Log.Error("active context file not saved", zap.Error(err))
			}

			return nil
		} else {
			return errors.New("action aborted")
		}
	} else {
		err = os.WriteFile(contextPath, jsonData, 0600)
		if err != nil {
			return err
		}

		activeContextPath := fmt.Sprintf("%s/%s", c.Directory, ".active")

		err = os.WriteFile(activeContextPath, []byte(c.Name), 0755)
		if err != nil {
			return err
		}

		return nil
	}
}

func (c *Context) Export(API string) (string, string, error) {
	var err error

	c.ApiURL = API

	bytes, err := json.Marshal(c)

	if err != nil {
		panic(err)
	}

	compressed := helpers.Compress(bytes)

	randbytes := make([]byte, 32)
	if _, err = rand.Read(randbytes); err != nil {
		return "", "", err
	}

	key := hex.EncodeToString(randbytes)

	contextPath := fmt.Sprintf("%s/%s.key", c.Directory, c.Name)
	err = os.WriteFile(contextPath, []byte(key), 0600)
	if err != nil {
		return "", "", err
	}

	encrypted, err := helpers.Encrypt(string(compressed.Bytes()), key)

	if err != nil {
		return "", "", err
	}

	return encrypted, key, nil
}
func (c *Context) Import(encrypted string, key string) error {
	ctx := NewContext(c.Directory)

	if key == "" {
		return errors.New("key is empty")
	}

	decrypted, err := helpers.Decrypt(encrypted, key)

	if err != nil {
		return err
	}

	decompressed := helpers.Decompress([]byte(decrypted))

	err = json.Unmarshal([]byte(decompressed), ctx)

	if err != nil {
		return err
	}

	err = ctx.GenerateHttpClient([]byte(ctx.CertBundle))

	if err != nil {
		return err
	}

	err = ctx.Connect(true)

	if err != nil {
		return err
	}

	viper.Set("y", true)
	return ctx.SaveToFile()
}
func (c *Context) ImportCertificates(key string, rootDir string) error {
	response := network.Send(c.Client, fmt.Sprintf("%s/fetch/certs", c.ApiURL), http.MethodGet, nil)

	if response.Success {
		keysEncrypted := keys.Encrypted{}
		bytes, err := response.Data.MarshalJSON()

		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &keysEncrypted)

		if err != nil {
			return err
		}

		decrypted, _ := helpers.Decrypt(keysEncrypted.Keys, key)

		var importedKeys keys.Keys
		err = json.Unmarshal([]byte(decrypted), &importedKeys)

		if err != nil {
			return err
		}

		err = importedKeys.CA.Write(fmt.Sprintf("%s/.ssh", rootDir))

		if err != nil {
			return err
		}

		for user, client := range importedKeys.Clients {
			err = client.Write(fmt.Sprintf("%s/.ssh", rootDir), user)
			if err != nil {
				return err
			}

			err = importedKeys.GeneratePemBundle(fmt.Sprintf("%s/.ssh", rootDir), user, importedKeys.Clients[user])

			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (c *Context) GenerateHttpClient(CertBundle []byte) error {
	for block, rest := pem.Decode(CertBundle); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return err
			}

			if cert.IsCA {
				pem.Encode(c.Ca, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				})
			} else {
				pem.Encode(c.Cert, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				})
			}

		case "EC PRIVATE KEY":
			pem.Encode(c.PrivateKey, &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: block.Bytes,
			})

		default:
			return errors.New("unknown pem type in the cert bundle")
		}
	}

	c.CertBundle = string(CertBundle)

	cert, err := tls.X509KeyPair(c.Cert.Bytes(), c.PrivateKey.Bytes())
	if err != nil {
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(c.Ca.Bytes())

	c.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	return nil
}
func (c *Context) Connect(retry bool) error {
	if retry {
		err := backoff.Retry(func() error {
			response := network.Send(c.Client, fmt.Sprintf("%s/connect", c.ApiURL), http.MethodGet, nil)

			if response.HttpStatus == http.StatusOK {
				return nil
			} else {
				return errors.New("failed to connect to the node")
			}
		}, backoff.NewExponentialBackOff())

		if err != nil {
			return err
		}

		return nil
	} else {
		response := network.Send(c.Client, fmt.Sprintf("%s/connect", c.ApiURL), http.MethodGet, nil)

		if response.HttpStatus == http.StatusOK {
			return nil
		} else {
			return errors.New("failed to connect to the node")
		}
	}
}
