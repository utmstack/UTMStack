//go:build !windows && !linux

package responder

// liveDescendants has no OS snapshot on non-Windows, non-Linux dev hosts; the
// caller falls back to the proctable (which the unit tests populate).
func liveDescendants(rootPID int) ([]int, bool) { return nil, false }

// ReleaseAllFreezes is a no-op off Linux — there is no cgroup freezer here.
func ReleaseAllFreezes() {}

// Stubs so the package builds on non-Windows dev hosts.
type OSKiller struct{}

func (OSKiller) KillPID(pid int) error { return nil }

type OSSuspender struct{}

func (OSSuspender) SuspendPID(pid int) error { return nil }
func (OSSuspender) ResumePID(pid int) error  { return nil }
