//go:build !linux
// +build !linux

package auditd

import (
	"context"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/utils"
)

// AuditdCollector is a no-op stub for non-Linux platforms
type AuditdCollector struct{}

// New creates a new AuditdCollector (no-op on non-Linux)
func New() *AuditdCollector {
	return &AuditdCollector{}
}

// Name returns the collector name
func (a *AuditdCollector) Name() string {
	return "auditd"
}

// Start is a no-op on non-Linux platforms
func (a *AuditdCollector) Start(ctx context.Context, queue chan *plugins.Log) {
	utils.Logger.Info("auditd collector not supported on this platform, skipping")
}

// Stop is a no-op on non-Linux platforms
func (a *AuditdCollector) Stop() {}
