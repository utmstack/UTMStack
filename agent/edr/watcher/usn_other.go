//go:build !windows

package watcher

import "context"

type Sink interface{ Enqueue(path, op string) }

type CursorStore interface {
	LoadUSN(volume string) (uint64, int64)
	SaveUSN(volume string, journalID uint64, nextUSN int64)
}

type Watcher struct{}

func New(volumes []string, sink Sink, cursors CursorStore, excluded func(string) bool) *Watcher {
	return &Watcher{}
}

// Run is a no-op on non-Windows platforms (USN is NTFS-only).
func (w *Watcher) Run(ctx context.Context) { <-ctx.Done() }
