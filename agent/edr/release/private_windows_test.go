//go:build windows

package release

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsReleaseStagingAndPublishedAccess(t *testing.T) {
	stage, err := MakePrivateTempDir(t.TempDir(), ".private-")
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPrivateDirectory(stage); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(stage, "owned.exe")
	if err := os.WriteFile(file, []byte("inert permission control"), 0600); err != nil {
		t.Fatal(err)
	}
	inspect := func(wantUsers bool) {
		t.Helper()
		sd, err := windows.GetNamedSecurityInfo(file, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		acl, _, err := sd.DACL()
		if err != nil || acl == nil {
			t.Fatal("missing ACL", err)
		}
		users := false
		for i := uint32(0); i < uint32(acl.AceCount); i++ {
			var ace *windows.ACCESS_ALLOWED_ACE
			if err := windows.GetAce(acl, i, &ace); err != nil {
				t.Fatal(err)
			}
			sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
			if ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE && sid.IsWellKnown(windows.WinBuiltinUsersSid) {
				users = true
				if ace.Mask&(windows.FILE_WRITE_DATA|windows.FILE_APPEND_DATA|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER|windows.GENERIC_ALL|windows.GENERIC_WRITE) != 0 {
					t.Fatal("ordinary users can alter published artifact")
				}
				if ace.Mask&windows.FILE_GENERIC_READ != windows.FILE_GENERIC_READ || ace.Mask&windows.FILE_GENERIC_EXECUTE != windows.FILE_GENERIC_EXECUTE {
					t.Fatal("ordinary users cannot read/execute published artifact")
				}
			}
		}
		if users != wantUsers {
			t.Fatalf("ordinary users access=%v want=%v", users, wantUsers)
		}
	}
	inspect(false)
	writer, err := os.OpenFile(file, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	hardenErr := PreparePublishedFile(file)
	closeErr := writer.Close()
	if !errors.Is(hardenErr, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("retained writer was not refused: %v", hardenErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	inspect(false)
	path, err := windows.UTF16PtrFromString(file)
	if err != nil {
		t.Fatal(err)
	}
	deleter, err := windows.CreateFile(path, windows.DELETE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	hardenErr = PreparePublishedFile(file)
	closeErr = windows.CloseHandle(deleter)
	if !errors.Is(hardenErr, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("retained delete handle was not refused: %v", hardenErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	inspect(false)
	if err := PreparePublishedFile(file); err != nil {
		t.Fatal(err)
	}
	inspect(true)
}

func TestWindowsReleaseRejectsForeignWritableParent(t *testing.T) {
	parent := t.TempDir()
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err = windows.SetNamedSecurityInfo(parent, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	if stage, err := MakePrivateTempDir(parent, ".refuse-"); err == nil {
		os.RemoveAll(stage)
		t.Fatal("foreign-writable parent accepted")
	}
}
