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

// FWPM_NET_EVENT_SUBSCRIPTION0 (trimmed) — nil enumTemplate = subscribe to all
// events. We filter to our blocklist in the manager (store.Match) rather than by
// classify-drop type, so we never decode the version-sensitive event tail/union.
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

	// Real callback: decode the header (validated offsets) and forward. Filtering to
	// our blocklist happens in the manager, so a system-wide drop we don't care
	// about resolves to no indicator and is dropped there.
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

// netEvent is the head of FWPM_NET_EVENT{1..5} — only the header, which is stable
// across versions for the fields we read. The type + union tail is intentionally
// not modeled (version-sensitive); we filter by blocklist membership instead.
type netEvent struct {
	header netEventHeader
}

// netEventHeader mirrors FWPM_NET_EVENT_HEADER. Offsets validated on Windows 11
// (build 26200) from a live classify-drop dump: remoteAddr@36, ports@52/54,
// appId@64. localAddr/remoteAddr are 16-byte unions (v4 in the first 4 bytes,
// host-order uint32; v6 = raw network-order bytes).
type netEventHeader struct {
	timestamp  windows.Filetime // [0:8]
	flags      uint32           // [8:12]  bitmask of valid fields
	ipVersion  uint32           // [12:16] 0=v4, 1=v6
	ipProtocol uint32           // [16:20] UINT8 + 3 pad natively
	localAddr  [16]byte         // [20:36] union {UINT32 v4 | UINT8 v6[16]}
	remoteAddr [16]byte         // [36:52] union {UINT32 v4 | UINT8 v6[16]}
	localPort  uint16           // [52:54]
	remotePort uint16           // [54:56]
	scopeID    uint32           // [56:60]
	_          [4]byte          // [60:64] pad (blob is 8-aligned)
	appID      fwpByteBlob      // [64:80] {size; pad; *data}
	userID     uintptr          // [80:88]
}

func decodeNetEvent(e *netEvent) (BlockEvent, bool) {
	if e == nil {
		return BlockEvent{}, false
	}
	h := &e.header
	var ip netip.Addr
	if h.ipVersion == 0 { // FWP_IP_VERSION_V4: host-order uint32 in the low 4 bytes
		v := binary.LittleEndian.Uint32(h.remoteAddr[0:4])
		ip = netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
	} else { // v6: raw network-order bytes
		var a16 [16]byte
		copy(a16[:], h.remoteAddr[:])
		ip = netip.AddrFrom16(a16)
	}
	if !ip.IsValid() || ip.IsUnspecified() {
		return BlockEvent{}, false
	}
	// Direction is not decoded from the version-sensitive classify-drop union;
	// blocklist blocks are overwhelmingly outbound (C2/exfil). Refining this to read
	// FWP_DIRECTION is a follow-up.
	return BlockEvent{
		RemoteIP:    ip,
		RemotePort:  int(h.remotePort),
		Direction:   "outbound",
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
