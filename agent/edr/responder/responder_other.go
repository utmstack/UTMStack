//go:build !windows && !linux

package responder

// liveProcessTree has no OS snapshot on unsupported platforms; the caller falls back to the
// proctable (which the unit tests populate).
func liveProcessTree(rootPID int) ([]processParent, bool) { return nil, false }

// Stubs so the package builds on non-Windows dev hosts.
type OSKiller struct{}

func (OSKiller) KillPID(pid int) error { return ErrUnsupported }

type OSSuspender struct{}

func (OSSuspender) SuspendPID(pid int) error { return ErrUnsupported }
func (OSSuspender) ResumePID(pid int) error  { return ErrUnsupported }
