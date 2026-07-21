package registry

import "sync"

var (
	datasets   = make(map[string]*Dataset)
	registryMu sync.RWMutex
)

func Get(id string) (*Dataset, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ds, ok := datasets[id]
	return ds, ok
}

func Swap(id string, ds *Dataset) {
	registryMu.Lock()
	defer registryMu.Unlock()
	datasets[id] = ds
}

func Delete(id string) bool {
	registryMu.Lock()
	defer registryMu.Unlock()
	_, ok := datasets[id]
	if ok {
		delete(datasets, id)
	}
	return ok
}

func Exists(id string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := datasets[id]
	return ok
}

func IDs() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ids := make([]string, 0, len(datasets))
	for id := range datasets {
		ids = append(ids, id)
	}
	return ids
}
