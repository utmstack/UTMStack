//go:build windows

package netblock

import (
	"fmt"
	"net/netip"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// --- fwpuclnt.dll bindings -------------------------------------------------

var (
	modFwpuclnt = windows.NewLazySystemDLL("fwpuclnt.dll")

	procEngineOpen        = modFwpuclnt.NewProc("FwpmEngineOpen0")
	procEngineClose       = modFwpuclnt.NewProc("FwpmEngineClose0")
	procSubLayerAdd       = modFwpuclnt.NewProc("FwpmSubLayerAdd0")
	procFilterAdd         = modFwpuclnt.NewProc("FwpmFilterAdd0")
	procFilterDeleteById  = modFwpuclnt.NewProc("FwpmFilterDeleteById0")
	procTransactionBegin  = modFwpuclnt.NewProc("FwpmTransactionBegin0")
	procTransactionCommit = modFwpuclnt.NewProc("FwpmTransactionCommit0")
	procTransactionAbort  = modFwpuclnt.NewProc("FwpmTransactionAbort0")
)

const (
	rpcCAuthnWinNT     = 10
	fwpmSessionFlagDyn = 0x00000001            // FWPM_SESSION_FLAG_DYNAMIC
	fwpActionBlock     = 0x00000001 | 0x00001000 // FWP_ACTION_BLOCK
	fwpMatchEqual      = 0
	fwpUint32          = 8  // FWP_UINT32
	fwpV6AddrMask      = 20 // FWP_V6_ADDR_MASK
	fwpByteArray16Type = 11 // FWP_BYTE_ARRAY16_TYPE
	fwpV4AddrMask      = 19 // FWP_V4_ADDR_MASK
)

// Well-known layer + condition GUIDs (windows headers: fwpmu.h / fwpmtypes.h).
var (
	layerConnectV4    = mustGUID("c38d57d1-05a7-4c33-904f-7fbceee60e82")
	layerConnectV6    = mustGUID("4a72393b-319f-44bc-84c3-ba54dcb3b6b4")
	layerRecvAcceptV4 = mustGUID("e1cd9fe7-f4b5-4273-96c0-592e487b8650")
	layerRecvAcceptV6 = mustGUID("a3b42c97-9f04-4672-b87b-cee8514723b0")
	condRemoteAddress = mustGUID("b235ae9a-1d64-49b8-a44c-5ff3d9095045")
	// Private, fixed sublayer key for the UTMStack EDR blocklist. Any valid GUID
	// works as long as it stays stable across runs.
	edrSublayerKey = mustGUID("bd7a0753-1077-4c55-a3ff-c2a93d24f5bf")
)

func mustGUID(s string) windows.GUID {
	g, err := windows.GUIDFromString("{" + s + "}")
	if err != nil {
		// Panic on a malformed literal — all layer/condition/sublayer GUIDs must
		// be valid at init time.
		panic("netblock: bad GUID literal " + s + ": " + err.Error())
	}
	return g
}

// FWPM_DISPLAY_DATA0
type displayData struct {
	name        *uint16
	description *uint16
}

// FWPM_SUBLAYER0 (trimmed to fields we set)
type sublayer0 struct {
	subLayerKey  windows.GUID
	displayData  displayData
	flags        uint32
	providerKey  uintptr
	providerData windows.GUID // reused as blob placeholder; zeroed
	weight       uint16
	_            [6]byte
}

// FWP_VALUE0 / FWP_CONDITION_VALUE0
type fwpValue struct {
	typ  uint32
	_    uint32
	data uintptr // pointer or inline uint32
}

// FWPM_FILTER_CONDITION0
type filterCondition struct {
	fieldKey       windows.GUID
	matchType      uint32
	conditionValue fwpValue
}

// FWP_V4_ADDR_AND_MASK
type v4AddrAndMask struct {
	addr uint32
	mask uint32
}

// FWP_V6_ADDR_AND_MASK
type v6AddrAndMask struct {
	addr         [16]byte
	prefixLength byte
	_            [3]byte
}

// FWPM_FILTER0 (trimmed) — action + single remote-address condition.
type filter0 struct {
	filterKey           windows.GUID
	displayData         displayData
	flags               uint32
	providerKey         uintptr
	providerData        [16]byte
	layerKey            windows.GUID
	subLayerKey         windows.GUID
	weight              fwpValue
	numFilterConditions uint32
	filterConditions    *filterCondition
	action              filterAction
	context             uint64
	reserved            uintptr
	filterID            uint64
	effectiveWeight     fwpValue
}

type filterAction struct {
	actionType uint32
	filterType windows.GUID // union: calloutKey/filterType — zeroed for BLOCK
}

// --- Blocker impl ----------------------------------------------------------

type wfpBlocker struct {
	mu      sync.Mutex
	engine  windows.Handle
	filters map[netip.Addr][]uint64 // addr -> filter IDs (v4/v6 × layers)
}

func NewOSBlocker(sublayerName string) (Blocker, error) {
	b := &wfpBlocker{filters: map[netip.Addr][]uint64{}}
	if err := b.open(sublayerName); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *wfpBlocker) open(name string) error {
	// Dynamic session → all filters auto-removed when this handle closes (fail-open).
	type session0 struct {
		sessionKey       windows.GUID
		displayData      displayData
		flags            uint32
		txnWaitTimeoutMs uint32
		processID        uint32
		sid              uintptr
		username         *uint16
		kernelMode       int32
	}
	sess := session0{flags: fwpmSessionFlagDyn}
	r, _, _ := procEngineOpen.Call(0, rpcCAuthnWinNT, 0,
		uintptr(unsafe.Pointer(&sess)), uintptr(unsafe.Pointer(&b.engine)))
	if r != 0 {
		return fmt.Errorf("FwpmEngineOpen0: 0x%x", r)
	}
	np, _ := windows.UTF16PtrFromString(name)
	sl := sublayer0{subLayerKey: edrSublayerKey, displayData: displayData{name: np}, weight: 0x8000}
	r, _, _ = procSubLayerAdd.Call(uintptr(b.engine), uintptr(unsafe.Pointer(&sl)), 0)
	if r != 0 && r != 0x80320009 /* FWP_E_ALREADY_EXISTS */ {
		return fmt.Errorf("FwpmSubLayerAdd0: 0x%x", r)
	}
	return nil
}

func layersFor(dir string, v6 bool) []windows.GUID {
	var out []windows.GUID
	// Fail safe: any direction that is not exactly "in" gets the outbound
	// (connect) layer, and any direction that is not exactly "out" gets the
	// inbound (recv-accept) layer. So "both" and any garbage/empty/typo value
	// block BOTH directions rather than silently enforcing nothing.
	if dir != "in" {
		if v6 {
			out = append(out, layerConnectV6)
		} else {
			out = append(out, layerConnectV4)
		}
	}
	if dir != "out" {
		if v6 {
			out = append(out, layerRecvAcceptV6)
		} else {
			out = append(out, layerRecvAcceptV4)
		}
	}
	return out
}

func (b *wfpBlocker) AddIP(addr netip.Addr, dir string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.filters[addr]; ok {
		return nil
	}
	var ids []uint64
	for _, layer := range layersFor(dir, addr.Is6()) {
		id, err := b.addFilter(addr, layer)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	b.filters[addr] = ids
	return nil
}

func (b *wfpBlocker) addFilter(addr netip.Addr, layer windows.GUID) (uint64, error) {
	cond := filterCondition{fieldKey: condRemoteAddress, matchType: fwpMatchEqual}
	if addr.Is4() {
		v4 := v4AddrAndMask{addr: be32(addr.As4()), mask: 0xffffffff}
		cond.conditionValue = fwpValue{typ: fwpV4AddrMask, data: uintptr(unsafe.Pointer(&v4))}
		return b.commitFilter(layer, &cond, unsafe.Pointer(&v4))
	}
	a16 := addr.As16()
	v6 := v6AddrAndMask{addr: a16, prefixLength: 128}
	cond.conditionValue = fwpValue{typ: fwpV6AddrMask, data: uintptr(unsafe.Pointer(&v6))}
	return b.commitFilter(layer, &cond, unsafe.Pointer(&v6))
}

func (b *wfpBlocker) commitFilter(layer windows.GUID, cond *filterCondition, keep unsafe.Pointer) (uint64, error) {
	np, _ := windows.UTF16PtrFromString("UTMStack EDR blocklist")
	f := filter0{
		displayData:         displayData{name: np},
		layerKey:            layer,
		subLayerKey:         edrSublayerKey,
		weight:              fwpValue{typ: fwpUint32, data: 0x8000},
		numFilterConditions: 1,
		filterConditions:    cond,
		action:              filterAction{actionType: fwpActionBlock},
	}
	var id uint64
	r, _, _ := procFilterAdd.Call(uintptr(b.engine), uintptr(unsafe.Pointer(&f)), 0,
		uintptr(unsafe.Pointer(&id)))
	// The condition buffer's address is stashed as a uintptr inside the filter
	// condition; keep it provably alive across the entire FwpmFilterAdd0 syscall
	// so escape analysis / the GC cannot reclaim it mid-call.
	runtime.KeepAlive(keep)
	if r != 0 {
		return 0, fmt.Errorf("FwpmFilterAdd0: 0x%x", r)
	}
	return id, nil
}

func (b *wfpBlocker) RemoveIP(addr netip.Addr) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, id := range b.filters[addr] {
		procFilterDeleteById.Call(uintptr(b.engine), uintptr(id))
	}
	delete(b.filters, addr)
	return nil
}

func (b *wfpBlocker) Reset() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for addr, ids := range b.filters {
		for _, id := range ids {
			procFilterDeleteById.Call(uintptr(b.engine), uintptr(id))
		}
		delete(b.filters, addr)
	}
	return nil
}

func (b *wfpBlocker) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.filters)
}

func be32(b [4]byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
