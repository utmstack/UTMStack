//go:build !windows && !linux

package behavioral

import (
	"context"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

// PSLogReader is a no-op on non-Windows, non-Linux targets: there is no
// PowerShell script-block log to poll, and no netlink process connector.
type PSLogReader struct{}

func NewPSLogReader(sp *event.Spool) *PSLogReader { return &PSLogReader{} }
func (r *PSLogReader) Run(ctx context.Context)    { <-ctx.Done() }

// IsInterpreter is a no-op off Linux: shell_activity telemetry only exists
// where the netlink watcher feeds it. The basename match is intentionally
// absent so dispatch never emits shell_activity on this platform.
func IsInterpreter(image string) (string, bool) { return "", false }
