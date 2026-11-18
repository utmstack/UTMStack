//go:build linux

package service

import "github.com/utmstack/UTMStack/agent/edr/orchestrator"

// Linux monitoring retains bounded pending work while admission is full.
// Kernel loss still requires a visible gap and directory reconciliation.
func (s sinkAdapter) TryEnqueue(path, op string) bool {
	return s.o.Enqueue(orchestrator.FileEvent{Path: path, Op: op})
}
