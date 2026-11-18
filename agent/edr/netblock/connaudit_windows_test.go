//go:build windows

package netblock

import (
	"net/netip"
	"testing"
	"unsafe"
)

func TestWFPEvent3LayoutAndDropDecoding(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 || unsafe.Sizeof(netEventHeader{}) != 136 ||
		unsafe.Offsetof(netEvent{}.typ) != 136 || unsafe.Offsetof(netEvent{}.dropPtr) != 144 ||
		unsafe.Offsetof(classifyDrop2{}.direction) != 24 {
		t.Fatal("FWPM_NET_EVENT3 layout no longer matches the Windows callback ABI")
	}
	drop := classifyDrop2{filterID: 42, direction: fwpDirectionOut}
	e := netEvent{typ: netEventTypeClassifyDrop, dropPtr: &drop}
	e.header.ipVersion = 0
	copy(e.header.remoteAddr[:], []byte{8, 2, 0, 192})
	e.header.remotePort = 443
	be, ok := decodeNetEvent(&e)
	if !ok || be.RemoteIP != netip.MustParseAddr("192.0.2.8") || be.FilterID != 42 ||
		be.Direction != "outbound" || be.RemotePort != 443 {
		t.Fatalf("valid classify drop: %+v, %t", be, ok)
	}
	e.typ = 6 // classify allow, with the same IP and union bytes
	if _, ok := decodeNetEvent(&e); ok {
		t.Fatal("classify allow was reported as a drop")
	}
	e.typ = netEventTypeClassifyDrop
	drop.filterID = 0
	if _, ok := decodeNetEvent(&e); ok {
		t.Fatal("drop without an originating filter ID was accepted")
	}
	drop.filterID, drop.direction = 42, fwpDirectionIn
	be, ok = decodeNetEvent(&e)
	if !ok || be.Direction != "inbound" {
		t.Fatalf("inbound drop direction: %+v, %t", be, ok)
	}
}
