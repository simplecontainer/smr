package contexts

import (
	"path/filepath"
	"time"
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
