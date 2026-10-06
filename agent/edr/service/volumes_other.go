//go:build !windows && !linux

package service

// defaultFixedVolumes returns the filesystem root on non-Linux unix targets
// (e.g. macOS). The Linux watcher is a no-op there and this value is unused,
// but "/" is the harmless default. It mirrors the Windows default of "all
// fixed drives"; Linux gets real mount-point enumeration in volumes_linux.go.
func defaultFixedVolumes() []string { return []string{"/"} }
