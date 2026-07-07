package ransomware

import "context"

type FileOp int

const (
	OpCreate FileOp = iota
	OpWrite
	OpRename
	OpDelete
	OpSetInfo
)

// FileEvent is a per-process file operation. PID is the process that performed
// the op — the attribution USN cannot provide, which is why v1 uses an ETW
// feed for this stream.
type FileEvent struct {
	PID  int
	Path string
	Op   FileOp
}

// FileActivityFeed delivers per-process file operations to sink until ctx is
// cancelled. v1 implementation: ETW Microsoft-Windows-Kernel-File (Task 11).
// The interface makes the feed swappable (e.g. a USN-fallback impl) without
// touching the guard.
type FileActivityFeed interface {
	Run(ctx context.Context, sink func(FileEvent)) error
}

// Consider is the portable pre-filter applied before evidence is generated:
// keep only mutating ops (write/rename/delete/setinfo), drop the EDR's own PID
// and excluded paths, and always keep a canary tamper (highest signal). This
// keeps the guard's evidence stream small and self-noise-free.
func Consider(ev FileEvent, isCanary func(string) bool, excluded func(string) bool, selfPID int) bool {
	if ev.PID == selfPID {
		return false
	}
	// A canary tamper is always kept, even under an operator-excluded directory —
	// it is the highest-confidence signal and canary dirs can overlap exclusions.
	if isCanary != nil && isCanary(ev.Path) {
		return true
	}
	if excluded != nil && excluded(ev.Path) {
		return false
	}
	switch ev.Op {
	case OpWrite, OpRename, OpDelete, OpSetInfo:
		return true
	default:
		return false
	}
}
