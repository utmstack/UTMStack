//go:build !windows && !linux

package service

// defaultFixedVolumes returns nil off Windows (the USN watcher is a stub there).
func defaultFixedVolumes() []string { return nil }
