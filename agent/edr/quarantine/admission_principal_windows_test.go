//go:build windows

package quarantine

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestAdmissionWindowsPrivatePrincipalDescriptors(t *testing.T) {
	for _, c := range []struct {
		name, principal, owner string
		administrative         bool
		aces                   uint16
	}{
		{"system", "S-1-5-18", "S-1-5-18", false, 2},
		{"elevated-administrator", "S-1-5-21-1-2-3-1001", "S-1-5-32-544", true, 2},
		{"custom-service", "S-1-5-21-1-2-3-1002", "S-1-5-21-1-2-3-1002", false, 3},
		{"filtered-administrator", "S-1-5-21-1-2-3-1001", "S-1-5-21-1-2-3-1001", false, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			sid, err := windows.StringToSid(c.principal)
			if err != nil {
				t.Fatal(err)
			}
			for _, directory := range []bool{false, true} {
				sd, err := admissionPrincipalSD(sid, c.administrative, directory)
				if err != nil {
					t.Fatal(err)
				}
				owner, _, err := sd.Owner()
				if err != nil || owner.String() != c.owner {
					t.Fatal("incorrect private owner", sd.String(), err)
				}
				control, _, err := sd.Control()
				if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
					t.Fatal("private ACL is inherited", err)
				}
				acl, _, err := sd.DACL()
				if err != nil || acl == nil || acl.AceCount != c.aces {
					t.Fatal("incorrect private ACL size", sd.String(), err)
				}
				allowed := map[string]bool{"S-1-5-18": true, "S-1-5-32-544": true}
				if !c.administrative {
					allowed[c.principal] = true
				}
				for i := uint32(0); i < uint32(acl.AceCount); i++ {
					var ace *windows.ACCESS_ALLOWED_ACE
					if err = windows.GetAce(acl, i, &ace); err != nil {
						t.Fatal(err)
					}
					principal := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
					if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || !allowed[principal] {
						t.Fatal("private metadata grants a foreign principal", sd.String())
					}
					if directory && ace.Header.AceFlags&(windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE) != (windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE) {
						t.Fatal("private child inheritance absent", sd.String())
					}
				}
			}
		})
	}
}
