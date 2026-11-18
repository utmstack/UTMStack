package netblock

import (
	"errors"
	"net/netip"
)

var ErrUnsupported = errors.New("UTMStack EDR: network enforcement is unavailable on this platform")

type unavailableBlocker struct{}

func (unavailableBlocker) AddIP(netip.Addr, string) error       { return ErrUnsupported }
func (unavailableBlocker) RemoveIP(netip.Addr) error            { return ErrUnsupported }
func (unavailableBlocker) AddPrefix(netip.Prefix, string) error { return ErrUnsupported }
func (unavailableBlocker) RemovePrefix(netip.Prefix) error      { return ErrUnsupported }
func (unavailableBlocker) Reset() error                         { return ErrUnsupported }
func (unavailableBlocker) Count() int                           { return 0 }
