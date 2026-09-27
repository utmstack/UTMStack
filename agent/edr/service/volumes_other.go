//go:build !windows

package service

// defaultFixedVolumes returns the filesystem root on all non-Windows
// platforms. On Linux the fanotify watcher marks "/" (mount-level) so the
// primary local filesystem is covered; the isSupportedVolume filter skips
// network/virtual mounts below it. On other targets (e.g. macOS) the watcher
// is a no-op and this value is unused, but "/" is the harmless default. This
// mirrors the Windows default of "all fixed drives".
func defaultFixedVolumes() []string { return []string{"/"} }
