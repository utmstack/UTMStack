//go:build linux
// +build linux

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

const journaldCursorFlushInterval = 2 * time.Second

var journaldCursorFile = filepath.Join(fs.GetExecutablePath(), "journald_cursor.txt")

type journaldCursor struct {
	mu          sync.Mutex
	current     string
	lastWritten string
}

func loadJournaldCursor() string {
	data, err := os.ReadFile(journaldCursorFile)
	if err != nil {
		return ""
	}
	return string(data)
}

func newJournaldCursor() *journaldCursor {
	initial := loadJournaldCursor()
	return &journaldCursor{current: initial, lastWritten: initial}
}

func (c *journaldCursor) resumePoint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *journaldCursor) set(cursor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = cursor
}

func (c *journaldCursor) flush() {
	c.mu.Lock()
	cursor := c.current
	unchanged := cursor == c.lastWritten
	c.mu.Unlock()

	if cursor == "" || unchanged {
		return
	}
	if err := os.WriteFile(journaldCursorFile, []byte(cursor), 0o600); err != nil {
		utils.Logger.ErrorF("error persisting journald cursor: %v", err)
		return
	}

	c.mu.Lock()
	c.lastWritten = cursor
	c.mu.Unlock()
}

func (c *journaldCursor) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(journaldCursorFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.flush()
			return
		case <-ticker.C:
			c.flush()
		}
	}
}

func extractJournaldCursor(line string) string {
	var entry struct {
		Cursor string `json:"__CURSOR"`
	}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return ""
	}
	return entry.Cursor
}
