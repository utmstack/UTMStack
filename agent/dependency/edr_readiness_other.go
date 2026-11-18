//go:build !linux

package dependency

func checkEDRPlatformReadiness(string, []byte) error { return nil }
