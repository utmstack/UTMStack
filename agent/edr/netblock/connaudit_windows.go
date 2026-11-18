// edr/netblock/connaudit_windows.go
//go:build windows

package netblock

import (
	"context"
	"encoding/binary"
	"net/netip"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procNetEventSub2  = modFwpuclnt.NewProc("FwpmNetEventSubscribe2")
	procNetEventUnsub = modFwpuclnt.NewProc("FwpmNetEventUnsubscribe0")
)

// fwpByteBlob mirrors FWP_BYTE_BLOB. golang.org/x/sys/windows (v0.46.0) does not
// export this type, so it is defined locally: { size; *data }.
type fwpByteBlob struct {
	Size uint32
	Data *uint8
}

// FWPM_NET_EVENT_SUBSCRIPTION0 (trimmed) — nil enumTemplate subscribes to all
// events. The callback accepts only classify drops by our own filter ID.
type netEventSubscription struct {
	enumTemplate uintptr
	flags        uint32
	sessionKey   windows.GUID
}

type winConnAudit struct {
	mu     sync.Mutex
	engine windows.Handle
	handle windows.Handle
}

func NewConnAudit() ConnAudit { return &winConnAudit{} }

func (w *winConnAudit) Run(ctx context.Context, out chan<- BlockEvent) {
	var eng windows.Handle
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
	// NON-dynamic session: live net-event subscriptions do not deliver on a dynamic
	// session (VM-observed — the callback never fires). We unsubscribe explicitly on
	// ctx.Done. WFP net-event collection defaults ON (outbound drops collected), so
	// no FwpmEngineSetOption0 is needed. netsh proves WFP records our drops.
	sess := session0{}
	r, _, _ := procEngineOpen.Call(0, rpcCAuthnWinNT, 0,
		uintptr(unsafe.Pointer(&sess)), uintptr(unsafe.Pointer(&eng)))
	if r != 0 {
		logWarn("connaudit engine open: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.mu.Lock()
	w.engine = eng
	w.mu.Unlock()
	defer procEngineClose.Call(uintptr(eng))

	// Decode the FWPM_NET_EVENT3 header and classify-drop record before forwarding.
	// The manager checks both current policy and the originating filter ID.
	cb := windows.NewCallback(func(ctxPtr uintptr, event *netEvent) uintptr {
		if event != nil {
			if be, ok := decodeNetEvent(event); ok {
				select {
				case out <- be:
				default: // never block the WFP callback
				}
			}
		}
		return 0
	})

	sub := netEventSubscription{}
	var subHandle windows.Handle
	r, _, _ = procNetEventSub2.Call(uintptr(eng), uintptr(unsafe.Pointer(&sub)),
		cb, 0, uintptr(unsafe.Pointer(&subHandle)))
	if r != 0 {
		logWarn("connaudit net-event subscribe: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.mu.Lock()
	w.handle = subHandle
	w.mu.Unlock()
	<-ctx.Done()
	procNetEventUnsub.Call(uintptr(eng), uintptr(subHandle))
}

// FwpmNetEventSubscribe2 calls FWPM_NET_EVENT_CALLBACK2 with FWPM_NET_EVENT3.
// The header ends at byte 136 on 64-bit Windows; type is at 136 and the
// classify-drop union pointer at 144. Decode only the documented drop record.
type netEvent struct {
	header  netEventHeader
	typ     uint32
	_       uint32
	dropPtr *classifyDrop2
}

// netEventHeader mirrors FWPM_NET_EVENT_HEADER3. Offsets validated on Windows 11
// (build 26200) from a live classify-drop dump: remoteAddr@36, ports@52/54,
// appId@64. localAddr/remoteAddr are 16-byte unions (v4 in the first 4 bytes,
// host-order uint32; v6 = raw network-order bytes).
type netEventHeader struct {
	timestamp     windows.Filetime // [0:8]
	flags         uint32           // [8:12]  bitmask of valid fields
	ipVersion     uint32           // [12:16] 0=v4, 1=v6
	ipProtocol    uint32           // [16:20] UINT8 + 3 pad natively
	localAddr     [16]byte         // [20:36] union {UINT32 v4 | UINT8 v6[16]}
	remoteAddr    [16]byte         // [36:52] union {UINT32 v4 | UINT8 v6[16]}
	localPort     uint16           // [52:54]
	remotePort    uint16           // [54:56]
	scopeID       uint32           // [56:60]
	_             [4]byte          // [60:64] pad (blob is 8-aligned)
	appID         fwpByteBlob      // [64:80] {size; pad; *data}
	userID        uintptr          // [80:88]
	addressFamily uint32           // [88:92]
	_             uint32           // [92:96]
	packageSID    uintptr          // [96:104]
	enterpriseID  uintptr          // [104:112]
	policyFlags   uint64           // [112:120]
	effectiveName fwpByteBlob      // [120:136]
}

const (
	netEventTypeClassifyDrop = 3
	fwpDirectionIn           = 0x3900
	fwpDirectionOut          = 0x3901
)

// Prefix of FWPM_NET_EVENT_CLASSIFY_DROP2. Its first fields are stable for
// the callback version; msFwpDirection distinguishes inbound from outbound.
type classifyDrop2 struct {
	filterID        uint64
	layerID         uint16
	_               uint16
	reauthReason    uint32
	originalProfile uint32
	currentProfile  uint32
	direction       uint32
}

// Make a layout drift a cross-build failure, not a silent misdecode on an
// endpoint. Both supported Windows targets are 64-bit.
var (
	_ [136 - int(unsafe.Sizeof(netEventHeader{}))]byte
	_ [int(unsafe.Sizeof(netEventHeader{})) - 136]byte
	_ [136 - int(unsafe.Offsetof(netEvent{}.typ))]byte
	_ [int(unsafe.Offsetof(netEvent{}.typ)) - 136]byte
	_ [144 - int(unsafe.Offsetof(netEvent{}.dropPtr))]byte
	_ [int(unsafe.Offsetof(netEvent{}.dropPtr)) - 144]byte
	_ [24 - int(unsafe.Offsetof(classifyDrop2{}.direction))]byte
	_ [int(unsafe.Offsetof(classifyDrop2{}.direction)) - 24]byte
)

func decodeNetEvent(e *netEvent) (BlockEvent, bool) {
	if e == nil || e.typ != netEventTypeClassifyDrop || e.dropPtr == nil {
		return BlockEvent{}, false
	}
	drop := e.dropPtr
	if drop.filterID == 0 {
		return BlockEvent{}, false
	}
	direction := ""
	switch drop.direction {
	case fwpDirectionIn:
		direction = "inbound"
	case fwpDirectionOut:
		direction = "outbound"
	default:
		return BlockEvent{}, false
	}
	h := &e.header
	var ip netip.Addr
	if h.ipVersion == 0 { // FWP_IP_VERSION_V4: host-order uint32 in the low 4 bytes
		v := binary.LittleEndian.Uint32(h.remoteAddr[0:4])
		ip = netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	} else if h.ipVersion == 1 { // v6: raw network-order bytes
		var a16 [16]byte
		copy(a16[:], h.remoteAddr[:])
		ip = netip.AddrFrom16(a16)
	} else {
		return BlockEvent{}, false
	}
	if !ip.IsValid() || ip.IsUnspecified() {
		return BlockEvent{}, false
	}
	return BlockEvent{
		RemoteIP:    ip,
		RemotePort:  int(h.remotePort),
		FilterID:    drop.filterID,
		Direction:   direction,
		ProcessPath: appIDPath(h.appID),
	}, true
}

// appIDPath converts the FWP_BYTE_BLOB app id (device-path UTF16) to a string.
func appIDPath(blob fwpByteBlob) string {
	if blob.Size == 0 || blob.Data == nil {
		return ""
	}
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(blob.Data)), blob.Size/2)
	return NormalizeDevicePath(windows.UTF16ToString(u16))
}
