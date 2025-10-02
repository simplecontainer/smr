package ocicredentials

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func NewManager(home string) *Manager {
	return &Manager{
		authFilePath: filepath.Join(home, DefaultConfigFile),
		Credentials:  make(map[string]*Credentials),
	}
}

func (m *Manager) Save() error {
	configDir := filepath.Dir(m.authFilePath)
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		return err
	}

	data, err := yaml.Marshal(m.Credentials)
	if err != nil {
		return err
	}

	if err = os.WriteFile(m.authFilePath, data, ConfigFilePermissions); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Load() error {
	if _, err := os.Stat(m.authFilePath); os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(m.authFilePath)
	if err != nil {
		return err
	}

	var credentials map[string]*Credentials
	if err = yaml.Unmarshal(data, &credentials); err != nil {
		return err
	}

	for _, credential := range credentials {
		if err := credential.Validate(); err != nil {
			return ErrInvalidCredentials
		}
	}

	m.Credentials = credentials
	return nil
}

func (m *Manager) Find(registry string) (*Credentials, bool) {
	val, ok := m.Credentials[registry]
	return val, ok
}

func (m *Manager) Exists() bool {
	_, err := os.Stat(m.authFilePath)
	return err == nil
}

func (m *Manager) Delete() error {
	if err := os.Remove(m.authFilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (m *Manager) GetPath() string {
	return m.authFilePath
}

func (m *Manager) Input(registry string) error {
	input := NewInputReader()
	credential, err := input.PromptForCredentials(registry)
	if err != nil {
		return err
	}

	m.Credentials[registry] = credential
	return nil
}

func (m *Manager) Default(registry string) {
	m.Credentials[registry] = Default(registry)
}
