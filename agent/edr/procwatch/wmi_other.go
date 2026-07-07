//go:build !windows

package procwatch

import "context"

type Watcher struct{}

func New(h Handler) *Watcher { return &Watcher{} }

// Run is a no-op off Windows (WMI is Windows-only).
func (w *Watcher) Run(ctx context.Context) { <-ctx.Done() }
