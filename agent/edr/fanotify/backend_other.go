//go:build !linux

package fanotify

func openBackend(Config) (backend, error) { return nil, ErrUnsupported }
