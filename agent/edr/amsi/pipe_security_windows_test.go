//go:build windows

package amsi

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// This parses the production descriptor with the Windows security API. It is
// safe under an ordinary user and does not create/install an AMSI provider.
func TestPipeSecurityDescriptorRestrictsClientRights(t *testing.T) {
	sd, err := windows.SecurityDescriptorFromString(pipeSecuritySDDL)
	if err != nil || !sd.IsValid() {
		t.Fatalf("descriptor: %v", err)
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !owner.IsWellKnown(windows.WinLocalSystemSid) {
		t.Fatalf("pipe owner must be LocalSystem: %v %v", owner, err)
	}
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 3 {
		t.Fatalf("explicit DACL: %v %v", dacl, err)
	}
	want := map[string]uint32{
		"S-1-5-18":     windows.GENERIC_ALL,
		"S-1-5-32-544": windows.GENERIC_ALL,
		"S-1-5-11":     pipeClientRights,
	}
	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, i, &ace); err != nil {
			t.Fatal(err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 {
			t.Fatalf("unexpected ACE header: %+v", ace.Header)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		expected, ok := want[sid.String()]
		if !ok || uint32(ace.Mask) != expected {
			t.Fatalf("unexpected rights for %s: %#x", sid.String(), ace.Mask)
		}
		delete(want, sid.String())
	}
	if len(want) != 0 {
		t.Fatalf("missing principals: %v", want)
	}
}

func TestPipeOwnerSIDPolicyRejectsOrdinaryUsers(t *testing.T) {
	for sid, trusted := range map[string]bool{
		"S-1-5-18": true, "S-1-5-11": false, "S-1-5-7": false,
		"S-1-5-32-544": false, "S-1-5-21-1-2-3-1001": false,
	} {
		parsed, err := windows.StringToSid(sid)
		if err != nil {
			t.Fatal(err)
		}
		if got := parsed.IsWellKnown(windows.WinLocalSystemSid); got != trusted {
			t.Fatalf("SID %s trusted=%v want=%v", sid, got, trusted)
		}
	}
}
