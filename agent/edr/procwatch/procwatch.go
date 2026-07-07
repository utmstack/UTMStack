package procwatch

type ProcStart struct {
	PID     int
	PPID    int
	Image   string
	Cmdline string
}

type Handler interface {
	OnProcStart(ProcStart)
}
