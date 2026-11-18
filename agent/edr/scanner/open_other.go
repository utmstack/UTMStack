//go:build !linux && !darwin

package scanner

import "os"

func openScannerTarget(path string) (*os.File, error) { return os.Open(path) }
