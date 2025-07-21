package contexts

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/pkg/encrypt"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"go.uber.org/zap"
	"io"
	"net/http"
)

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

func Upload(token string, registry string, data string) (string, error) {
	resp, err := network.Http(
		context.Background(),
		http.DefaultClient,
		fmt.Sprintf("%s/api/v1/store/upload", registry),
		"POST",
		map[string]string{"context": data},
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
	)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func Download(token string, registry string) (string, error) {
	resp, err := network.Http(
		context.Background(),
		http.DefaultClient,
		fmt.Sprintf("%s/api/v1/store/download", registry),
		"POST",
		nil,
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
	)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
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

	decrypted, err := encrypt.Decrypt(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	decompressed := encrypt.Decompress([]byte(decrypted))

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
