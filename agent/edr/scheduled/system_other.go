//go:build !linux

package scheduled

func skipSystemPath(string) bool { return false }
