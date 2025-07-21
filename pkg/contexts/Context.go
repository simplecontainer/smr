package contexts

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/encrypt"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io"
	"net"
	http "net/http"
	"os"
	"path/filepath"
	"time"
)

func NewContext(cfg *Config) (*ClientContext, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	ctx := &ClientContext{
		config: cfg,
		Credentials: &Credentials{
			PrivateKey: &bytes.Buffer{},
			Cert:       &bytes.Buffer{},
			Ca:         &bytes.Buffer{},
		},
	}

	if !cfg.InMemory && cfg.RootDir != "" {
		contextDir := filepath.Join(cfg.RootDir, static.CONTEXTDIR)
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create context directory: %w", err)
		}
		ctx.Directory = contextDir

		if helpers.IsRunningAsSudo() {
			user, err := helpers.GetRealUser()

			if err != nil {
				return nil, err
			}

			if err := helpers.Chown(contextDir, user.Uid, user.Gid); err != nil {
				return nil, fmt.Errorf("failed to change owner of context directory: %w", err)
			}
		}
	}

	return ctx, nil
}

func (c *ClientContext) WithName(name string) *ClientContext {
	if name == "" {
		logger.Log.Warn("attempted to set empty context name")
		return c
	}
	c.Name = name
	return c
}

func (c *ClientContext) WithAPIURL(url string) *ClientContext {
	if url == "" {
		logger.Log.Warn("attempted to set empty API URL")
		return c
	}
	c.APIURL = url
	return c
}

func (c *ClientContext) WithCredentials(creds *Credentials) *ClientContext {
	if creds == nil {
		logger.Log.Warn("attempted to set nil credentials")
		return c
	}
	c.Credentials = creds
	return c
}

func (c *ClientContext) GetHTTPClient() *http.Client {
	return c.client
}

func (c *ClientContext) GetClientNoTimeout() *http.Client {
	c.client.Timeout = 0
	return c.client
}

func (c *ClientContext) setupTLSClient() error {
	if c.Credentials == nil {
		return errors.New("credentials are nil")
	}

	if len(c.Credentials.CertBundle) == 0 {
		return errors.New("certificate bundle is empty")
	}

	if err := c.parseCertBundle([]byte(c.Credentials.CertBundle)); err != nil {
		return fmt.Errorf("failed to parse certificate bundle: %w", err)
	}

	if c.Credentials.Cert.Len() == 0 {
		return errors.New("client certificate is missing")
	}

	if c.Credentials.PrivateKey.Len() == 0 {
		return errors.New("private key is missing")
	}

	if c.Credentials.Ca.Len() == 0 {
		return errors.New("CA certificate is missing")
	}

	cert, err := tls.X509KeyPair(c.Credentials.Cert.Bytes(), c.Credentials.PrivateKey.Bytes())
	if err != nil {
		return fmt.Errorf("failed to create X509 key pair: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(c.Credentials.Ca.Bytes()) {
		return errors.New("failed to append CA certificate to cert pool")
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if c.config.UseInsecure {
		tlsConfig.InsecureSkipVerify = true
		logger.Log.Warn("TLS certificate verification disabled")
	}

	dialer := &net.Dialer{
		Timeout:   5 * c.config.APITimeout,
		KeepAlive: 1 * c.config.APITimeout,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   true,
			Idle:     0,
			Interval: 0,
			Count:    0,
		},
	}

	c.client = &http.Client{
		Transport: &http.Transport{
			DialContext:     dialer.DialContext,
			TLSClientConfig: tlsConfig,
		},
	}

	return nil
}

func (c *ClientContext) parseCertBundle(certBundle []byte) error {
	if c.Credentials == nil {
		return errors.New("credentials are nil")
	}

	c.Credentials.PrivateKey.Reset()
	c.Credentials.Cert.Reset()
	c.Credentials.Ca.Reset()

	foundCert := false
	foundKey := false
	foundCA := false

	for block, rest := pem.Decode(certBundle); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %w", err)
			}

			if cert.IsCA {
				if err := pem.Encode(c.Credentials.Ca, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}); err != nil {
					return fmt.Errorf("failed to encode CA certificate: %w", err)
				}
				foundCA = true
			} else {
				if err := pem.Encode(c.Credentials.Cert, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}); err != nil {
					return fmt.Errorf("failed to encode client certificate: %w", err)
				}
				foundCert = true
			}

		case "EC PRIVATE KEY":
			if err := pem.Encode(c.Credentials.PrivateKey, &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: block.Bytes,
			}); err != nil {
				return fmt.Errorf("failed to encode private key: %w", err)
			}
			foundKey = true

		default:
			logger.Log.Warn("unknown PEM type in the cert bundle", zap.String("type", block.Type))
		}
	}

	if !foundCert {
		return errors.New("no client certificate found in bundle")
	}
	if !foundKey {
		return errors.New("no private key found in bundle")
	}
	if !foundCA {
		return errors.New("no CA certificate found in bundle")
	}

	c.Credentials.CertBundle = string(certBundle)
	return nil
}

func (c *ClientContext) Connect(ctx context.Context, retry bool) error {
	if c.APIURL == "" {
		return errors.New("API URL is not set")
	}

	client := c.GetHTTPClient()
	connectURL := fmt.Sprintf("%s/connect", c.APIURL)

	if !retry {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, connectURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			return fmt.Errorf("connection failed with status: %d, response: %s", resp.StatusCode, body)
		}

		return nil
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = time.Duration(c.config.MaxRetries) * c.config.RetryDelay

	err := backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, connectURL, nil)
		if err != nil {
			return backoff.Permanent(fmt.Errorf("failed to create request: %w", err))
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Debug("connection attempt failed, will retry", zap.Error(err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Log.Debug("connection failed with status, will retry", zap.Int("status", resp.StatusCode))
			return fmt.Errorf("connection failed with status: %d", resp.StatusCode)
		}

		return nil
	}, expBackoff)

	if err != nil {
		return fmt.Errorf("connection failed after retries: %w", err)
	}

	return nil
}

func (c *ClientContext) Save() error {
	if c.Name == "" {
		c.Name = viper.GetString("context")

		if c.Name == "" {
			return errors.New("context name cannot be empty")
		}
	}

	if c.config.InMemory {
		return errors.New("cannot save in in-memory mode")
	}

	if c.Directory == "" {
		return errors.New("context directory not set")
	}

	if err := os.MkdirAll(c.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	contextPath := filepath.Join(c.Directory, c.Name)
	activeContextPath := filepath.Join(c.Directory, ".active")

	if _, err := os.Stat(contextPath); err == nil {
		if !viper.GetBool("y") &&
			!helpers.Confirm(fmt.Sprintf("Context with the same name: %s already exists. Do you want to overwrite it?", c.Name)) {
			return errors.New("action aborted by user")
		}
	}

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := os.WriteFile(contextPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	if err := os.WriteFile(activeContextPath, []byte(c.Name), 0600); err != nil {
		return fmt.Errorf("failed to write active context file: %w", err)
	}

	if helpers.IsRunningAsSudo() {
		user, err := helpers.GetRealUser()

		if err != nil {
			return err
		}

		if err := helpers.Chown(c.Directory, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context directory: %w", err)
		}

		if err := helpers.Chown(contextPath, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context file: %w", err)
		}

		if err := helpers.Chown(activeContextPath, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of active context file: %w", err)
		}
	}

	return nil
}

func (c *ClientContext) Load() error {
	if c.config.InMemory {
		return errors.New("cannot load in-memory context from disk")
	}

	if c.Directory == "" {
		return errors.New("context directory not set")
	}

	if c.Name == "" {
		activeContextPath := filepath.Join(c.Directory, ".active")

		activeContext, err := os.ReadFile(activeContextPath)
		if err != nil {
			return fmt.Errorf("failed to read active context file: %w", err)
		}

		contextName := string(activeContext)
		if contextName == "" {
			contextName = "default"
		}

		c.Name = contextName
	}

	contextPath := filepath.Join(c.Directory, c.Name)
	data, err := os.ReadFile(contextPath)

	if err != nil {
		return fmt.Errorf("failed to read context file '%s': %w", contextPath, err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("invalid context file format: %w", err)
	}

	if c.Credentials.CertBundle != "" {
		if err := c.setupTLSClient(); err != nil {
			return fmt.Errorf("failed to set up TLS client: %w", err)
		}
	} else {
		return errors.New("missing certificate bundle")
	}

	return nil
}

func (c *ClientContext) ImportCertificates(ctx context.Context, sshDir string) error {
	key, err := c.deriveKeyFromPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	client := c.GetHTTPClient()
	response := network.Send(client, fmt.Sprintf("%s/fetch/certs", c.APIURL), http.MethodGet, nil)

	if !response.Success {
		return errors.New(response.ErrorExplanation)
	}

	keysEncrypted := keys.Encrypted{}
	bytes, err := response.Data.MarshalJSON()

	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err = json.Unmarshal(bytes, &keysEncrypted); err != nil {
		return fmt.Errorf("failed to unmarshal encrypted keys: %w", err)
	}

	decrypted, err := encrypt.Decrypt(keysEncrypted.Keys, key)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	var importedKeys keys.Keys
	if err = json.Unmarshal([]byte(decrypted), &importedKeys); err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	if err = os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	if helpers.IsRunningAsSudo() {
		user, err := helpers.GetRealUser()

		if err != nil {
			return err
		}

		if err := helpers.Chown(sshDir, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context directory: %w", err)
		}
	}

	if err = importedKeys.CA.Write(sshDir); err != nil {
		return fmt.Errorf("failed to write CA: %w", err)
	}

	for user, client := range importedKeys.Clients {
		if err = client.Write(sshDir, user); err != nil {
			return fmt.Errorf("failed to write client certificate for %s: %w", user, err)
		}

		if err = importedKeys.GeneratePemBundle(sshDir, user, importedKeys.Clients[user]); err != nil {
			return fmt.Errorf("failed to generate PEM bundle for %s: %w", user, err)
		}
	}

	return nil
}

func (c *ClientContext) deriveKeyFromPrivateKey() (string, error) {
	if c.Credentials.PrivateKey == nil || c.Credentials.PrivateKey.Len() == 0 {
		return "", errors.New("private key is empty")
	}

	block, _ := pem.Decode(c.Credentials.PrivateKey.Bytes())
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	var privateKey *ecdsa.PrivateKey
	var err error

	switch block.Type {
	case "EC PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse EC private key: %w", err)
		}
		privateKey = key.(*ecdsa.PrivateKey)
	case "PRIVATE KEY":
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		var ok bool
		privateKey, ok = keyInterface.(*ecdsa.PrivateKey)
		if !ok {
			return "", errors.New("not an ECDSA private key")
		}
	default:
		return "", fmt.Errorf("unsupported key type: %s", block.Type)
	}

	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if len(bytes) < 32 {
		return "", errors.New("private key too short for key derivation")
	}

	return hex.EncodeToString(bytes[:32]), nil
}

func (c *ClientContext) Export() (string, string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal context: %w", err)
	}

	compressed := encrypt.Compress(data)

	randbytes := make([]byte, 32)
	if _, err = rand.Read(randbytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random key: %w", err)
	}

	key := hex.EncodeToString(randbytes)

	if !c.config.InMemory && c.Directory != "" && c.Name != "" {
		keyPath := filepath.Join(c.Directory, c.Name+".key")

		if err = os.WriteFile(keyPath, []byte(key), 0600); err != nil {
			return "", "", fmt.Errorf("failed to write key file: %w", err)
		}

		if helpers.IsRunningAsSudo() {
			user, err := helpers.GetRealUser()

			if err != nil {
				return "", "", err
			}

			if err := helpers.Chown(keyPath, user.Uid, user.Gid); err != nil {
				return "", "", fmt.Errorf("failed to change owner of context directory: %w", err)
			}
		}
	}

	encrypted, err := encrypt.Encrypt(string(compressed.Bytes()), key)
	if err != nil {
		return "", "", fmt.Errorf("encryption failed: %w", err)
	}

	return encrypted, key, nil
}
