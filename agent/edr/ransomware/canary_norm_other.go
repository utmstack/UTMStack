//go:build !windows

package ransomware

// canaryNorm preserves case: POSIX filesystems are case-sensitive, so a
// path that differs in case is a DIFFERENT file and must not match the
// planted canary (which would make the scorer credit an unrelated write).
func canaryNorm(s string) string { return s }
