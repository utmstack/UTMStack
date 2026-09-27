//go:build !linux

package watcher

// PermDecider is declared here (fanotify_linux.go has the live version) so
// the service can call w.SetPermDecider on every platform. Off Linux the
// watcher is USN (Windows) or a no-op (other) and is notification-only, so
// the decider is never invoked.
type PermDecider interface {
	Scan(path string) (malicious bool, err error)
}

// SetMode is a no-op off Linux: permission mode is a fanotify capability.
// Kept so the service wiring is platform-agnostic.
func (w *Watcher) SetMode(string) {}

// SetPermDecider is a no-op off Linux.
func (w *Watcher) SetPermDecider(PermDecider) {}
