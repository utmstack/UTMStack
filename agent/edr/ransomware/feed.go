package ransomware

import "context"

type FileOp int

const (
	OpCreate FileOp = iota // an open; never scored (an open is not a modification)
	OpWrite
	OpRename
	OpDelete
	OpSetInfo
)

// Modifying reports whether op changes a file's content, name or existence.
// Only modifying ops on a decoy count as tampering; opening or reading one does not.
func (op FileOp) Modifying() bool {
	switch op {
	case OpWrite, OpRename, OpDelete, OpSetInfo:
		return true
	default:
		return false
	}
}

// FileEvent is a per-process file operation. PID is the process that performed
// the op — the attribution USN cannot provide, which is why v1 uses an ETW
// feed for this stream.
type FileEvent struct {
	PID                  int
	Path                 string
	Op                   FileOp
	Device, Inode        uint64
	FileIdentityRequired bool
	OSIdentity           int64
	IdentityRequired     bool
	PPID                 int
	Image, Cmdline       string
}

// FileActivityFeed delivers per-process file operations to sink until ctx is
// cancelled. v1 implementation: ETW Microsoft-Windows-Kernel-File (Task 11).
// The interface makes the feed swappable (e.g. a USN-fallback impl) without
// touching the guard.
type FileActivityFeed interface {
	Run(ctx context.Context, sink func(FileEvent)) error
}

// decoyAware is implemented by a feed that must know the decoy paths to
// attribute events that carry no file name (the ETW feed). The guard calls it
// once, before Run.
type decoyAware interface {
	setInterest(isDecoy func(path string) bool)
}

// Highest PID that is never a culprit: 0 is Idle and 4 is the System process,
// which issues the cache manager's lazy-writer flushes of a write another
// process already made (and SMB-server I/O). Neither can be suspended or killed.
const maxKernelPID = 4

// Consider is the portable pre-filter applied before evidence is generated:
// keep only modifying ops (write/rename/delete/setinfo), drop the EDR's own PID,
// Idle/System and excluded paths, and always keep a canary tamper (highest
// signal) even under an excluded directory. Opening or reading a canary is never
// kept. This keeps the guard's evidence stream small and self-noise-free.
func Consider(ev FileEvent, isCanary func(string) bool, excluded func(string) bool, selfPID int) bool {
	if ev.PID == selfPID || ev.PID <= maxKernelPID || !ev.Op.Modifying() {
		return false
	}
	// A canary tamper is always kept, even under an operator-excluded directory —
	// it is the highest-confidence signal and canary dirs can overlap exclusions.
	if isCanary != nil && isCanary(ev.Path) {
		return true
	}
	return excluded == nil || !excluded(ev.Path)
}
