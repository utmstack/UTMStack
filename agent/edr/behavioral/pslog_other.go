//go:build !windows

package behavioral

import (
	"context"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

type PSLogReader struct{}

func NewPSLogReader(sp *event.Spool) *PSLogReader { return &PSLogReader{} }
func (r *PSLogReader) Run(ctx context.Context)    { <-ctx.Done() }
