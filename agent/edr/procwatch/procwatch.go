package procwatch

type ProcStart struct {
	PID     int
	PPID    int
	Image   string
	Cmdline string
	// EventKind distinguishes fork context and current image snapshots from an
	// identity-bound executable transition. Neither implies an original command.
	EventKind string
	// OSIdentity is the observed process creation identity, not the event time.
	// Linux uses procfs start ticks; zero means the source did not capture one.
	OSIdentity       int64
	IdentityRequired bool
}

type Handler interface {
	OnProcStart(ProcStart)
}
