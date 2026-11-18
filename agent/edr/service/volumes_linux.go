//go:build linux

package service

// Linux real-time monitoring is scoped; it is not a filesystem-wide journal.
func defaultFixedVolumes() []string { return []string{"/home", "/tmp", "/var/tmp"} }
