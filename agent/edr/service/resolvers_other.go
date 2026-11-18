//go:build !linux

package service

import "net/netip"

func ResolverAddresses() []netip.Addr                 { return nil }
func ResolverAddressesChecked() ([]netip.Addr, error) { return nil, nil }
