//go:build !linux
// +build !linux

package dependency

// AuditdVersion is the UTMStack audit rules version.
// This is a stub for non-Linux systems.
const AuditdVersion = "1.0.0"

// configureAuditd is a no-op on non-Linux systems.
func configureAuditd() error {
	return nil
}

// updateAuditdRules is a no-op on non-Linux systems.
func updateAuditdRules() error {
	return nil
}
