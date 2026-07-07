// edr/netblock/connaudit_windows.go
//go:build windows

package netblock

import (
	"context"
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

// FWPM_NET_EVENT_SUBSCRIPTION0 (trimmed) — no template = all events; we filter in
// the callback for classify-drop within our sublayer.
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
	// Reuse a dedicated dynamic engine handle for the subscription lifetime.
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
	sess := session0{flags: fwpmSessionFlagDyn}
	r, _, _ := procEngineOpen.Call(0, rpcCAuthnWinNT, 0,
		uintptr(unsafe.Pointer(&sess)), uintptr(unsafe.Pointer(&eng)))
	if r != 0 {
		logWarn("conn audit engine open: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.mu.Lock()
	w.engine = eng
	w.mu.Unlock()
	defer procEngineClose.Call(uintptr(eng))

	cb := windows.NewCallback(func(ctxPtr uintptr, event *netEvent1) uintptr {
		be, ok := decodeNetEvent(event)
		if ok {
			select {
			case out <- be:
			default: // never block the ETW/WFP callback
			}
		}
		return 0
	})

	sub := netEventSubscription{}
	var subHandle windows.Handle
	r, _, _ = procNetEventSub2.Call(uintptr(eng), uintptr(unsafe.Pointer(&sub)),
		cb, 0, uintptr(unsafe.Pointer(&subHandle)))
	if r != 0 {
		logWarn("net event subscribe: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.mu.Lock()
	w.handle = subHandle
	w.mu.Unlock()
	<-ctx.Done()
	procNetEventUnsub.Call(uintptr(eng), uintptr(subHandle))
}

// netEvent1 mirrors FWPM_NET_EVENT1 header fields we read (timestamp/flags/ip
// version/protocol/local+remote addr+port) plus the classify-drop union pointer.
type netEvent1 struct {
	header  netEventHeader
	typ     uint32
	dropPtr *classifyDrop1 // set (non-nil) when typ == classify-drop; pointer-sized union member
}

type netEventHeader struct {
	timestamp    windows.Filetime
	flags        uint32
	ipVersion    uint32
	ipProtocol   uint8
	_            [3]byte
	localAddrV4  uint32
	remoteAddrV4 uint32
	localAddrV6  [16]byte
	remoteAddrV6 [16]byte
	localPort    uint16
	remotePort   uint16
	scopeID      uint32
	appID        fwpByteBlob
	userID       uintptr
}

type classifyDrop1 struct {
	filterID     uint64
	layerID      uint16
	_            [6]byte
	reauthReason uint32
	origDir      uint32
	msFwpDir     uint32
	isLoopback   int32
}

const netEventTypeClassifyDrop = 3 // FWPM_NET_EVENT_TYPE_CLASSIFY_DROP

func decodeNetEvent(e *netEvent1) (BlockEvent, bool) {
	if e == nil || e.typ != netEventTypeClassifyDrop {
		return BlockEvent{}, false
	}
	h := e.header
	var ip netip.Addr
	if h.ipVersion == 0 { // FWP_IP_VERSION_V4
		ip = netip.AddrFrom4([4]byte{
			byte(h.remoteAddrV4 >> 24), byte(h.remoteAddrV4 >> 16),
			byte(h.remoteAddrV4 >> 8), byte(h.remoteAddrV4)})
	} else {
		ip = netip.AddrFrom16(h.remoteAddrV6)
	}
	dir := "outbound"
	if e.dropPtr != nil {
		if e.dropPtr.msFwpDir == 1 { // FWP_DIRECTION_INBOUND
			dir = "inbound"
		}
	}
	be := BlockEvent{
		RemoteIP:    ip,
		RemotePort:  int(h.remotePort),
		Direction:   dir,
		ProcessPath: appIDPath(h.appID),
	}
	return be, true
}

// appIDPath converts the FWP_BYTE_BLOB app id (device-path UTF16) to a string.
func appIDPath(blob fwpByteBlob) string {
	if blob.Size == 0 || blob.Data == nil {
		return ""
	}
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(blob.Data)), blob.Size/2)
	return windows.UTF16ToString(u16)
}
