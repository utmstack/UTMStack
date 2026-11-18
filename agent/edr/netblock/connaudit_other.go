// edr/netblock/connaudit_other.go
//go:build !windows

package netblock

import "context"

type noopConnAudit struct{}

// NewConnAudit returns a no-op ConnAudit on non-Windows hosts so the package
// builds and unit-tests everywhere. The real WFP net-event reader lives in
// connaudit_windows.go.
func NewConnAudit() ConnAudit { return noopConnAudit{} }

func (noopConnAudit) Run(ctx context.Context, out chan<- BlockEvent) { <-ctx.Done() }
