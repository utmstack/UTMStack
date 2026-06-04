package usecase

import (
	"sync"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
)

type moduleFactory struct {
	mu    sync.RWMutex
	kinds map[string]connectors.ModuleKind
}

func NewModuleFactory() connectors.ModuleFactory {
	return &moduleFactory{kinds: make(map[string]connectors.ModuleKind)}
}

func (f *moduleFactory) Register(kind connectors.ModuleKind) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.kinds[kind.Name()] = kind
}

func (f *moduleFactory) Get(name string) (connectors.ModuleKind, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	k, ok := f.kinds[name]
	return k, ok
}

func (f *moduleFactory) Has(name string) bool {
	_, ok := f.Get(name)
	return ok
}
