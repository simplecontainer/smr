package contexts

import (
	"fmt"
	"github.com/pkg/errors"
)

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
