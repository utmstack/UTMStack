//go:build !windows

package netblock

func NormalizeDevicePath(p string) string { return stripQ(p) }
