//go:build !windows

package responder

// liveDescendants has no OS snapshot off Windows; the caller falls back to the
// proctable (which the unit tests populate).
func liveDescendants(rootPID int) ([]int, bool) { return nil, false }

// Stubs so the package builds on non-Windows dev hosts.
type OSKiller struct{}

func (OSKiller) KillPID(pid int) error { return nil }

type OSSuspender struct{}

func (OSSuspender) SuspendPID(pid int) error { return nil }
func (OSSuspender) ResumePID(pid int) error  { return nil }
