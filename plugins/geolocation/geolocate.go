package main

import (
	"net"
	"net/netip"
	"sort"
	"sync/atomic"

	"github.com/utmstack/go-sdk/plugins"
)

// Compact, pointer-free range entries so Go's GC marks the multi-million-element
// IPv4/IPv6 backing slices as noscan (~8x lower RSS than map[string][]*block).
type asnBlockV4 struct {
	start  uint32
	end    uint32
	asn    uint32
	asoIdx uint32
}

type asnBlockV6 struct {
	startHi uint64
	startLo uint64
	endHi   uint64
	endLo   uint64
	asn     uint32
	asoIdx  uint32
}

type cityBlockV4 struct {
	start  uint32
	end    uint32
	locIdx uint32
	latE6  int32
	lonE6  int32
}

type cityBlockV6 struct {
	startHi uint64
	startLo uint64
	endHi   uint64
	endLo   uint64
	locIdx  uint32
	latE6   int32
	lonE6   int32
}

type cityLocation struct {
	countryISOCode string
	countryName    string
	cityName       string
}

type geoSnapshot struct {
	asnV4         []asnBlockV4
	asnV4Overlap  bool
	asnV6         []asnBlockV6
	asnV6Overlap  bool
	cityV4        []cityBlockV4
	cityV4Overlap bool
	cityV6        []cityBlockV6
	cityV6Overlap bool
	locations     []cityLocation
	asos          []string
}

var geoData atomic.Pointer[geoSnapshot]

// IsLocal reports whether a falls in any of the 6 local/non-routable CIDR ranges
// (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 169.254.0.0/16,
// 224.0.0.0/24) or their IPv6 equivalents (::1, fc00::/7, fe80::/10, ff02::/16).
func IsLocal(a net.IP) bool {
	if a == nil {
		return false
	}
	addr, ok := netip.AddrFromSlice(a)
	if !ok {
		return false
	}
	return isLocalAddr(addr.Unmap())
}

func isLocalAddr(addr netip.Addr) bool {
	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast()
}

func cmpU128(aHi, aLo, bHi, bLo uint64) int {
	if aHi != bHi {
		if aHi < bHi {
			return -1
		}
		return 1
	}
	if aLo != bLo {
		if aLo < bLo {
			return -1
		}
		return 1
	}
	return 0
}

func addrV4ToU32(addr netip.Addr) uint32 {
	b := addr.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func addrV6ToU128(addr netip.Addr) (uint64, uint64) {
	b := addr.As16()
	hi := uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	lo := uint64(b[8])<<56 | uint64(b[9])<<48 | uint64(b[10])<<40 | uint64(b[11])<<32 |
		uint64(b[12])<<24 | uint64(b[13])<<16 | uint64(b[14])<<8 | uint64(b[15])
	return hi, lo
}

func (s *geoSnapshot) lookupASNV4(ip uint32) (asnBlockV4, bool) {
	idx := sort.Search(len(s.asnV4), func(i int) bool {
		return s.asnV4[i].start > ip
	}) - 1
	if idx < 0 {
		return asnBlockV4{}, false
	}
	if !s.asnV4Overlap {
		if ip <= s.asnV4[idx].end {
			return s.asnV4[idx], true
		}
		return asnBlockV4{}, false
	}
	for j := idx; j >= 0; j-- {
		if ip <= s.asnV4[j].end {
			return s.asnV4[j], true
		}
	}
	return asnBlockV4{}, false
}

func (s *geoSnapshot) lookupCityV4(ip uint32) (cityBlockV4, bool) {
	idx := sort.Search(len(s.cityV4), func(i int) bool {
		return s.cityV4[i].start > ip
	}) - 1
	if idx < 0 {
		return cityBlockV4{}, false
	}
	if !s.cityV4Overlap {
		if ip <= s.cityV4[idx].end {
			return s.cityV4[idx], true
		}
		return cityBlockV4{}, false
	}
	for j := idx; j >= 0; j-- {
		if ip <= s.cityV4[j].end {
			return s.cityV4[j], true
		}
	}
	return cityBlockV4{}, false
}

func (s *geoSnapshot) lookupASNV6(hi, lo uint64) (asnBlockV6, bool) {
	idx := sort.Search(len(s.asnV6), func(i int) bool {
		return cmpU128(s.asnV6[i].startHi, s.asnV6[i].startLo, hi, lo) > 0
	}) - 1
	if idx < 0 {
		return asnBlockV6{}, false
	}
	if !s.asnV6Overlap {
		if cmpU128(hi, lo, s.asnV6[idx].endHi, s.asnV6[idx].endLo) <= 0 {
			return s.asnV6[idx], true
		}
		return asnBlockV6{}, false
	}
	for j := idx; j >= 0; j-- {
		if cmpU128(hi, lo, s.asnV6[j].endHi, s.asnV6[j].endLo) <= 0 {
			return s.asnV6[j], true
		}
	}
	return asnBlockV6{}, false
}

func (s *geoSnapshot) lookupCityV6(hi, lo uint64) (cityBlockV6, bool) {
	idx := sort.Search(len(s.cityV6), func(i int) bool {
		return cmpU128(s.cityV6[i].startHi, s.cityV6[i].startLo, hi, lo) > 0
	}) - 1
	if idx < 0 {
		return cityBlockV6{}, false
	}
	if !s.cityV6Overlap {
		if cmpU128(hi, lo, s.cityV6[idx].endHi, s.cityV6[idx].endLo) <= 0 {
			return s.cityV6[idx], true
		}
		return cityBlockV6{}, false
	}
	for j := idx; j >= 0; j-- {
		if cmpU128(hi, lo, s.cityV6[j].endHi, s.cityV6[j].endLo) <= 0 {
			return s.cityV6[j], true
		}
	}
	return cityBlockV6{}, false
}

func geolocate(ip string) *plugins.Geolocation {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil
	}
	addr = addr.Unmap()
	if isLocalAddr(addr) {
		return nil
	}

	snap := geoData.Load()
	if snap == nil {
		return nil
	}

	var (
		hasASN  bool
		asn     uint64
		aso     string
		hasCity bool
		locIdx  uint32
		latE6   int32
		lonE6   int32
	)

	if addr.Is4() {
		v4 := addrV4ToU32(addr)
		if b, ok := snap.lookupASNV4(v4); ok {
			hasASN = true
			asn = uint64(b.asn)
			if int(b.asoIdx) < len(snap.asos) {
				aso = snap.asos[b.asoIdx]
			}
		}
		if b, ok := snap.lookupCityV4(v4); ok {
			hasCity = true
			locIdx = b.locIdx
			latE6 = b.latE6
			lonE6 = b.lonE6
		}
	} else {
		hi, lo := addrV6ToU128(addr)
		if b, ok := snap.lookupASNV6(hi, lo); ok {
			hasASN = true
			asn = uint64(b.asn)
			if int(b.asoIdx) < len(snap.asos) {
				aso = snap.asos[b.asoIdx]
			}
		}
		if b, ok := snap.lookupCityV6(hi, lo); ok {
			hasCity = true
			locIdx = b.locIdx
			latE6 = b.latE6
			lonE6 = b.lonE6
		}
	}

	if !hasASN && !hasCity {
		return nil
	}

	geo := &plugins.Geolocation{}
	if hasASN {
		geo.Asn = asn
		geo.Aso = aso
	}
	if hasCity {
		if locIdx > 0 && int(locIdx) < len(snap.locations) {
			loc := snap.locations[locIdx]
			geo.City = loc.cityName
			geo.Country = loc.countryName
			geo.CountryCode = loc.countryISOCode
		}
		geo.Latitude = float64(latE6) / coordScale
		geo.Longitude = float64(lonE6) / coordScale
	}
	return geo
}
