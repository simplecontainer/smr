package contexts

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
	"path/filepath"
)

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

		if helpers.IsRunningAsSudo() {
			user, err := helpers.GetRealUser()

			if err != nil {
				return nil, err
			}

			if err := helpers.Chown(cfg.RootDir, user.Uid, user.Gid); err != nil {
				return nil, fmt.Errorf("failed to change owner of context directory: %w", err)
			}

			if err := helpers.Chown(contextDir, user.Uid, user.Gid); err != nil {
				return nil, fmt.Errorf("failed to change owner of context directory: %w", err)
			}
		}
	}

	return manager, nil
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
	return Import(m.config, encrypted, key)
}

func (m *Manager) ExportContext(name string, api string) (string, string, error) {
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

	url, err := helpers.EnforceHTTPS(api)

	if err != nil {
		return "", "", err
	}

	ctx.APIURL = url.String()
	ctx.config = m.config

	return ctx.Export()
}

func (m *Manager) Upload(token string, registry string, data string) (string, error) {
	return Upload(token, registry, data)
}

func (m *Manager) Download(token string, registry string) (string, error) {
	return Download(token, registry)
}
