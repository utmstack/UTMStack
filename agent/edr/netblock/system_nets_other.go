//go:build !linux

package netblock

func (*Manager) initialEnforcementHealthLocked() error { return nil }
