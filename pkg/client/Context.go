// Package client provides functionality for managing API contexts with secure connections.
package client

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
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func DefaultConfig(rootDir string) *Config {
	return &Config{
		RootDir:     filepath.Join(rootDir),
		APITimeout:  30 * time.Second,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
		UseInsecure: false,
	}
}

func NewContext(cfg *Config) *ClientContext {
	if cfg == nil {
		panic(errors.New("provide config to the NewContext()"))
	}

	contextDir := filepath.Join(cfg.RootDir, static.CONTEXTDIR)
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		logger.Log.Error("failed to create context directory", zap.Error(err))
	}

	return &ClientContext{
		Directory: contextDir,
		Credentials: &Credentials{
			PrivateKey: &bytes.Buffer{},
			Cert:       &bytes.Buffer{},
			Ca:         &bytes.Buffer{},
		},
		config: cfg,
	}
}

func (c *ClientContext) WithName(name string) *ClientContext {
	c.Name = name
	return c
}

func (c *ClientContext) WithAPIURL(url string) *ClientContext {
	c.APIURL = url
	return c
}

func (c *ClientContext) WithCredentials(creds *Credentials) *ClientContext {
	c.Credentials = creds
	return c
}

func (c *ClientContext) GetClient() *http.Client {
	if c.client != nil {
		return c.client
	}

	if c.client != nil {
		return c.client
	}

	if c.Credentials.CertBundle != "" {
		if err := c.setupTLSClient(); err != nil {
			logger.Log.Error("failed to set up TLS client", zap.Error(err))
		}
	} else {
		logger.Log.Error("failed to set up TLS client: empty CertBundle")
	}

	return c.client
}

func (c *ClientContext) setupTLSClient() error {
	if len(c.Credentials.CertBundle) == 0 {
		return errors.New("certificate bundle is empty")
	}

	if err := c.parseCertBundle([]byte(c.Credentials.CertBundle)); err != nil {
		return err
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

	c.client = &http.Client{
		Timeout: c.config.APITimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	return nil
}

func (c *ClientContext) parseCertBundle(certBundle []byte) error {
	c.Credentials.PrivateKey.Reset()
	c.Credentials.Cert.Reset()
	c.Credentials.Ca.Reset()

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
			} else {
				if err := pem.Encode(c.Credentials.Cert, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}); err != nil {
					return fmt.Errorf("failed to encode client certificate: %w", err)
				}
			}

		case "EC PRIVATE KEY":
			if err := pem.Encode(c.Credentials.PrivateKey, &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: block.Bytes,
			}); err != nil {
				return fmt.Errorf("failed to encode private key: %w", err)
			}

		default:
			return fmt.Errorf("unknown PEM type in the cert bundle: %s", block.Type)
		}
	}

	c.Credentials.CertBundle = string(certBundle)
	return nil
}

func (c *ClientContext) Connect(ctx context.Context, retry bool) error {
	if c.APIURL == "" {
		return errors.New("API URL is not set")
	}

	client := c.GetClient()
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
			return fmt.Errorf("connection failed with status: %d", resp.StatusCode)
		}

		return nil
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = time.Duration(c.config.MaxRetries) * c.config.RetryDelay

	err := backoff.Retry(func() error {
		// Check if context was canceled
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

	if err := os.MkdirAll(c.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	contextPath := filepath.Join(c.Directory, c.Name)
	activeContextPath := filepath.Join(c.Directory, ".active")

	if _, err := os.Stat(contextPath); err == nil {
		if !viper.GetBool("y") && !helpers.Confirm("Context with the same name already exists. Do you want to overwrite it?") {
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

	return nil
}

func (c *ClientContext) Load() error {
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
		return fmt.Errorf("missing cert bundle")
	}

	return nil
}

func LoadByName(name string, cfg *Config) (*ClientContext, error) {
	ctx := NewContext(cfg).WithName(name)

	if err := ctx.Load(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func LoadActive(cfg *Config) (*ClientContext, error) {
	ctx := NewContext(cfg)

	if err := ctx.Load(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (c *ClientContext) ImportCertificates(ctx context.Context, sshDir string) error {
	key, err := c.deriveKeyFromPrivateKey()

	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	response := network.Send(c.client, fmt.Sprintf("%s/fetch/certs", c.APIURL), http.MethodGet, nil)

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

	decrypted, err := helpers.Decrypt(keysEncrypted.Keys, key)

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
	if c.Credentials.PrivateKey.Len() == 0 {
		return "", errors.New("private key is empty")
	}

	block, _ := pem.Decode(c.Credentials.PrivateKey.Bytes())
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	privateKeyTmp, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := privateKeyTmp.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("not an ECDSA private key")
	}

	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if len(bytes) < 32 {
		return "", errors.New("private key too short")
	}

	return hex.EncodeToString(bytes[:32]), nil
}

func (c *ClientContext) Export() (string, string, error) {
	data, err := json.Marshal(c)

	if err != nil {
		return "", "", fmt.Errorf("failed to marshal context: %w", err)
	}

	compressed := helpers.Compress(data)

	randbytes := make([]byte, 32)
	if _, err = rand.Read(randbytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random key: %w", err)
	}

	key := hex.EncodeToString(randbytes)

	if c.Name != "" {
		keyPath := filepath.Join(c.Directory, c.Name+".key")
		if err = os.WriteFile(keyPath, []byte(key), 0600); err != nil {
			return "", "", fmt.Errorf("failed to write key file: %w", err)
		}
	}

	encrypted, err := helpers.Encrypt(string(compressed.Bytes()), key)
	if err != nil {
		return "", "", fmt.Errorf("encryption failed: %w", err)
	}

	return encrypted, key, nil
}

func Import(cfg *Config, encrypted, key string) (*ClientContext, error) {
	if key == "" {
		return nil, errors.New("key is empty")
	}

	decrypted, err := helpers.Decrypt(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	decompressed := helpers.Decompress([]byte(decrypted))

	ctx := NewContext(cfg)

	if err = json.Unmarshal([]byte(decompressed), ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	if ctx.APIURL == "" {
		return nil, errors.New("imported context has no API URL")
	}

	if ctx.Credentials.CertBundle == "" {
		return nil, errors.New("imported context has no certificate bundle")
	}

	if err = ctx.setupTLSClient(); err != nil {
		return nil, fmt.Errorf("failed to set up TLS client: %w", err)
	}

	return ctx, nil
}

func NewManager(cfg *Config) *Manager {
	if cfg == nil {
		panic(errors.New("provide config to the NewManager()"))
	}

	return &Manager{config: cfg}
}

func (m *Manager) ListContexts() ([]string, error) {
	contextDir := filepath.Join(m.config.RootDir, static.CONTEXTDIR)

	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	files, err := os.ReadDir(contextDir)

	if err != nil {
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}

	contexts := make([]string, 0, len(files))

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && name != ".active" && filepath.Ext(name) == ".key" {
			contexts = append(contexts, name)
		}
	}

	return contexts, nil
}

func (m *Manager) GetActiveContext() (string, error) {
	activeContextPath := filepath.Join(m.config.RootDir, static.CONTEXTDIR, ".active")

	data, err := os.ReadFile(activeContextPath)

	if err != nil {
		if os.IsNotExist(err) {
			return "default", nil
		}
		return "", fmt.Errorf("failed to read active context file: %w", err)
	}

	contextName := string(data)

	if contextName == "" {
		return "default", nil
	}

	return contextName, nil
}

func (m *Manager) SetActiveContext(name string) error {
	contextDir := filepath.Join(m.config.RootDir, static.CONTEXTDIR)

	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	contextPath := filepath.Join(contextDir, name)

	if _, err := os.Stat(contextPath); err != nil {
		return fmt.Errorf("context '%s' not found: %w", name, err)
	}

	activeContextPath := filepath.Join(contextDir, ".active")

	if err := os.WriteFile(activeContextPath, []byte(name), 0600); err != nil {
		return fmt.Errorf("failed to write active context file: %w", err)
	}

	return nil
}

func (m *Manager) DeleteContext(name string) error {
	contextDir := filepath.Join(m.config.RootDir, static.CONTEXTDIR)
	contextPath := filepath.Join(contextDir, name)

	if _, err := os.Stat(contextPath); err != nil {
		return fmt.Errorf("context '%s' not found: %w", name, err)
	}

	activeContextPath := filepath.Join(contextDir, ".active")
	activeContext, err := os.ReadFile(activeContextPath)

	if err == nil && string(activeContext) == name {
		return fmt.Errorf("cannot delete active context '%s'", name)
	}

	if err := os.Remove(contextPath); err != nil {
		return fmt.Errorf("failed to delete context file: %w", err)
	}

	keyPath := filepath.Join(contextDir, name+".key")

	if _, err := os.Stat(keyPath); err == nil {
		if err := os.Remove(keyPath); err != nil {
			logger.Log.Warn("failed to delete context key file", zap.String("context", name), zap.Error(err))
		}
	}

	return nil
}

func (m *Manager) CreateContext(name string, apiURL string, certBundle []byte) (*ClientContext, error) {
	ctx := NewContext(m.config).WithName(name).WithAPIURL(apiURL)

	if err := ctx.parseCertBundle(certBundle); err != nil {
		return nil, fmt.Errorf("failed to parse certificate bundle: %w", err)
	}

	if err := ctx.setupTLSClient(); err != nil {
		return nil, fmt.Errorf("failed to set up TLS client: %w", err)
	}

	connCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ctx.Connect(connCtx, true); err != nil {
		return nil, fmt.Errorf("failed to connect to API: %w", err)
	}

	if err := ctx.Save(); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	return ctx, nil
}
