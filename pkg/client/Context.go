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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func DefaultConfig(rootDir string) *Config {
	config := &Config{
		RootDir:     filepath.Join(rootDir),
		APITimeout:  30 * time.Second,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
		UseInsecure: false,
		InMemory:    false,
	}

	if rootDir == "" {
		config.InMemory = true
	}

	return config
}

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

func (c *ClientContext) GetClient() *http.Client {
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

func LoadByName(name string, cfg *Config) (*ClientContext, error) {
	if name == "" {
		return nil, errors.New("context name cannot be empty")
	}

	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	ctx, err := NewContext(cfg)
	if err != nil {
		return nil, err
	}

	ctx = ctx.WithName(name)

	if err := ctx.Load(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func LoadActive(cfg *Config) (*ClientContext, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	ctx, err := NewContext(cfg)
	if err != nil {
		return nil, err
	}

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

	client := c.GetClient()
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

	compressed := helpers.Compress(data)

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
	}

	encrypted, err := helpers.Encrypt(string(compressed.Bytes()), key)
	if err != nil {
		return "", "", fmt.Errorf("encryption failed: %w", err)
	}

	return encrypted, key, nil
}

func Import(cfg *Config, encrypted, key string) (*ClientContext, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	if encrypted == "" {
		return nil, errors.New("encrypted data is empty")
	}

	if key == "" {
		return nil, errors.New("key is empty")
	}

	decrypted, err := helpers.Decrypt(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	decompressed := helpers.Decompress([]byte(decrypted))

	ctx, err := NewContext(cfg)
	if err != nil {
		return nil, err
	}

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

func NewManager(cfg *Config) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	manager := &Manager{
		config: cfg,
	}

	if cfg.InMemory {
		manager.store = NewMemoryStorage()
	} else {
		contextDir := filepath.Join(cfg.RootDir, static.CONTEXTDIR)
		if err := os.MkdirAll(contextDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create context directory: %w", err)
		}
		manager.store = NewFileStorage(contextDir)
	}

	return manager, nil
}

func NewFileStorage(contextDir string) *FileStorage {
	return &FileStorage{
		contextDir: contextDir,
	}
}

func (fs *FileStorage) Save(ctx *ClientContext) error {
	if ctx.Name == "" {
		return errors.New("context name cannot be empty")
	}

	contextPath := filepath.Join(fs.contextDir, ctx.Name)
	data, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := os.WriteFile(contextPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

func (fs *FileStorage) Load(name string) (*ClientContext, error) {
	var err error

	if name == "" {
		name, err = fs.GetActive()
		if err != nil {
			return nil, err
		}
	}

	contextPath := filepath.Join(fs.contextDir, name)
	data, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file '%s': %w", contextPath, err)
	}

	ctx := &ClientContext{}
	if err := json.Unmarshal(data, ctx); err != nil {
		return nil, fmt.Errorf("invalid context file format: %w", err)
	}

	return ctx, nil
}

func (fs *FileStorage) GetActive() (string, error) {
	activeContextPath := filepath.Join(fs.contextDir, ".active")
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

func (fs *FileStorage) SetActive(name string) error {
	if name == "" {
		return errors.New("context name cannot be empty")
	}

	contextPath := filepath.Join(fs.contextDir, name)
	if _, err := os.Stat(contextPath); err != nil {
		return fmt.Errorf("context '%s' not found: %w", name, err)
	}

	activeContextPath := filepath.Join(fs.contextDir, ".active")
	if err := os.WriteFile(activeContextPath, []byte(name), 0600); err != nil {
		return fmt.Errorf("failed to write active context file: %w", err)
	}

	return nil
}

func (fs *FileStorage) Delete(name string) error {
	if name == "" {
		return errors.New("context name cannot be empty")
	}

	contextPath := filepath.Join(fs.contextDir, name)
	if _, err := os.Stat(contextPath); err != nil {
		return fmt.Errorf("context '%s' not found: %w", name, err)
	}

	activeContext, err := fs.GetActive()
	if err == nil && activeContext == name {
		return fmt.Errorf("cannot delete active context '%s'", name)
	}

	if err := os.Remove(contextPath); err != nil {
		return fmt.Errorf("failed to delete context file: %w", err)
	}

	keyPath := filepath.Join(fs.contextDir, name+".key")
	if _, err := os.Stat(keyPath); err == nil {
		if err := os.Remove(keyPath); err != nil {
			logger.Log.Warn("failed to delete context key file", zap.String("context", name), zap.Error(err))
		}
	}

	return nil
}

func (fs *FileStorage) List() ([]string, error) {
	files, err := os.ReadDir(fs.contextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read context directory: %w", err)
	}

	contexts := make([]string, 0, len(files))
	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && name != ".active" && strings.HasSuffix(name, ".key") {
			contexts = append(contexts, name)
		}
	}

	return contexts, nil
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		contexts:      make(map[string]*ClientContext),
		activeContext: "default",
	}
}

func (ms *MemoryStorage) Save(ctx *ClientContext) error {
	if ctx.Name == "" {
		return errors.New("context name cannot be empty")
	}

	ctxCopy := *ctx
	ms.contexts[ctx.Name] = &ctxCopy

	if len(ms.contexts) == 1 || ctx.Name == ms.activeContext {
		ms.activeContext = ctx.Name
	}

	return nil
}

func (ms *MemoryStorage) Load(name string) (*ClientContext, error) {
	if name == "" {
		name = ms.activeContext
	}

	ctx, exists := ms.contexts[name]
	if !exists {
		return nil, fmt.Errorf("context '%s' not found", name)
	}

	ctxCopy := *ctx
	return &ctxCopy, nil
}

func (ms *MemoryStorage) GetActive() (string, error) {
	return ms.activeContext, nil
}

func (ms *MemoryStorage) SetActive(name string) error {
	if _, exists := ms.contexts[name]; !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	ms.activeContext = name
	return nil
}

func (ms *MemoryStorage) Delete(name string) error {
	if name == "" {
		return errors.New("context name cannot be empty")
	}

	if name == ms.activeContext {
		return fmt.Errorf("cannot delete active context '%s'", name)
	}

	if _, exists := ms.contexts[name]; !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	delete(ms.contexts, name)
	return nil
}

func (ms *MemoryStorage) List() ([]string, error) {
	contexts := make([]string, 0, len(ms.contexts))
	for name := range ms.contexts {
		contexts = append(contexts, name)
	}
	return contexts, nil
}

func (m *Manager) CreateContext(name, apiURL string, creds *Credentials) (*ClientContext, error) {
	if name == "" {
		return nil, errors.New("context name cannot be empty")
	}

	if apiURL == "" {
		return nil, errors.New("API URL cannot be empty")
	}

	if creds == nil {
		return nil, errors.New("credentials cannot be nil")
	}

	ctx, err := NewContext(m.config)
	if err != nil {
		return nil, err
	}

	ctx.WithName(name).WithAPIURL(apiURL).WithCredentials(creds)

	if err := m.store.Save(ctx); err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	err = ctx.setupTLSClient()

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (m *Manager) GetContext(name string) (*ClientContext, error) {
	return m.store.Load(name)
}

func (m *Manager) GetActive() (*ClientContext, error) {
	name, err := m.store.GetActive()
	if err != nil {
		return nil, err
	}

	return m.store.Load(name)
}

func (m *Manager) SetActive(name string) error {
	return m.store.SetActive(name)
}

func (m *Manager) DeleteContext(name string) error {
	return m.store.Delete(name)
}

func (m *Manager) ListContexts() ([]string, error) {
	return m.store.List()
}

func (m *Manager) ImportContext(encrypted, key string) (*ClientContext, error) {
	ctx, err := Import(m.config, encrypted, key)
	if err != nil {
		return nil, err
	}

	if err := m.store.Save(ctx); err != nil {
		return nil, fmt.Errorf("failed to save imported context: %w", err)
	}

	return ctx, nil
}

func (m *Manager) ExportContext(name string) (string, string, error) {
	var ctx *ClientContext
	var err error

	if name == "" {
		ctx, err = m.GetActive()
	} else {
		ctx, err = m.GetContext(name)
	}

	if err != nil {
		return "", "", err
	}

	if ctx == nil {
		return "", "", errors.New("context not found")
	}

	ctx.config = m.config

	return ctx.Export()
}

func BundleToCredentials(bundle interface{}) (*Credentials, error) {
	var bundleBytes []byte

	switch b := bundle.(type) {
	case string:
		bundleBytes = []byte(b)
	case []byte:
		bundleBytes = b
	case *bytes.Buffer:
		bundleBytes = b.Bytes()
	default:
		return nil, fmt.Errorf("unsupported bundle type: %T", bundle)
	}

	if len(bundleBytes) == 0 {
		return nil, errors.New("empty certificate bundle")
	}

	creds := &Credentials{
		PrivateKey: &bytes.Buffer{},
		Cert:       &bytes.Buffer{},
		Ca:         &bytes.Buffer{},
		CertBundle: string(bundleBytes),
	}

	var foundCert, foundKey, foundCA bool

	for block, rest := pem.Decode(bundleBytes); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}

			if cert.IsCA {
				if err := pem.Encode(creds.Ca, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}); err != nil {
					return nil, fmt.Errorf("failed to encode CA certificate: %w", err)
				}
				foundCA = true
			} else {
				if err := pem.Encode(creds.Cert, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}); err != nil {
					return nil, fmt.Errorf("failed to encode client certificate: %w", err)
				}
				foundCert = true
			}

		case "EC PRIVATE KEY", "PRIVATE KEY", "RSA PRIVATE KEY":
			if err := pem.Encode(creds.PrivateKey, block); err != nil {
				return nil, fmt.Errorf("failed to encode private key: %w", err)
			}
			foundKey = true

		default:
			logger.Log.Debug("unknown PEM type in the cert bundle", zap.String("type", block.Type))
		}
	}

	if !foundCert {
		return nil, errors.New("no client certificate found in bundle")
	}
	if !foundKey {
		return nil, errors.New("no private key found in bundle")
	}
	if !foundCA {
		return nil, errors.New("no CA certificate found in bundle")
	}

	return creds, nil
}
