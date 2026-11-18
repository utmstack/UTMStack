//go:build !linux

package dependency

func prepareEDRExecutable(string) error { return nil }
