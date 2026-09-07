//go:build windows
// +build windows

package platform

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const bookmarkFlushInterval = 2 * time.Second

var windowsBookmarkFile = filepath.Join(fs.GetExecutablePath(), "wineventlog_bookmarks.json")

type bookmarkStore struct {
	mu      sync.Mutex
	current map[string]string
	written map[string]string
}

func loadBookmarkStore() *bookmarkStore {
	store := &bookmarkStore{current: map[string]string{}, written: map[string]string{}}

	data, err := os.ReadFile(windowsBookmarkFile)
	if err != nil {
		return store
	}

	var saved map[string]string
	if err := json.Unmarshal(data, &saved); err != nil {
		utils.Logger.ErrorF("error reading windows event log bookmarks: %v", err)
		return store
	}
	for channel, bookmarkXML := range saved {
		store.current[channel] = bookmarkXML
		store.written[channel] = bookmarkXML
	}
	return store
}

func (b *bookmarkStore) get(channel string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.current[channel]
}

func (b *bookmarkStore) set(channel, bookmarkXML string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current[channel] = bookmarkXML
}

func (b *bookmarkStore) flush() {
	b.mu.Lock()
	changed := false
	for channel, bookmarkXML := range b.current {
		if b.written[channel] != bookmarkXML {
			changed = true
			break
		}
	}
	if !changed {
		b.mu.Unlock()
		return
	}
	snapshot := make(map[string]string, len(b.current))
	for channel, bookmarkXML := range b.current {
		snapshot[channel] = bookmarkXML
	}
	b.mu.Unlock()

	data, err := json.Marshal(snapshot)
	if err != nil {
		utils.Logger.ErrorF("error encoding windows event log bookmarks: %v", err)
		return
	}
	if err := os.WriteFile(windowsBookmarkFile, data, 0o600); err != nil {
		utils.Logger.ErrorF("error persisting windows event log bookmarks: %v", err)
		return
	}

	b.mu.Lock()
	for channel, bookmarkXML := range snapshot {
		b.written[channel] = bookmarkXML
	}
	b.mu.Unlock()
}

func (b *bookmarkStore) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(bookmarkFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			b.flush()
			return
		case <-ticker.C:
			b.flush()
		}
	}
}
