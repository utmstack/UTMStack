package executor

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// Executor turns a single node instance into a side effect (kind=executor) or a
// data lookup (kind=enrichment). It returns raw JSON `output` on success — the
// dispatcher stores it only when the node kind is enrichment; for executor
// kind it is discarded.
type Executor interface {
	Type() string
	Execute(ctx context.Context, exec *domain.SoarExecution) (output json.RawMessage, err error)
}

// Registry is a plain map — executors are wired at module start and never
// mutated afterwards, so no locking is needed.
// ponytail: map, not sync.Map — reads only after Init.
type Registry map[string]Executor

// ErrExecutorNotFound is returned by the dispatcher when a node references a
// type that was not registered. FlowStore should catch this at parse time, so
// hitting it at dispatch means a mis-wired module.
var ErrExecutorNotFound = errors.New("soar: executor not registered")

// Lookup fetches an executor by type or returns ErrExecutorNotFound.
func (r Registry) Lookup(t string) (Executor, error) {
	if e, ok := r[t]; ok {
		return e, nil
	}
	return nil, ErrExecutorNotFound
}

// Types returns the sorted set of registered executor names — used by the
// FlowStore validator to reject flows referencing unknown types.
func (r Registry) Has(t string) bool { _, ok := r[t]; return ok }
