package contexts

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

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

	if helpers.IsRunningAsSudo() {
		user, err := helpers.GetRealUser()

		if err != nil {
			return err
		}

		if err := helpers.Chown(contextPath, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context directory: %w", err)
		}
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

	if helpers.IsRunningAsSudo() {
		user, err := helpers.GetRealUser()

		if err != nil {
			return err
		}

		if err := helpers.Chown(contextPath, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context directory: %w", err)
		}

		if err := helpers.Chown(activeContextPath, user.Uid, user.Gid); err != nil {
			return fmt.Errorf("failed to change owner of context directory: %w", err)
		}
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
		if !file.IsDir() && name != ".active" && !strings.HasSuffix(name, ".key") {
			contexts = append(contexts, name)
		}
	}

	return contexts, nil
}
