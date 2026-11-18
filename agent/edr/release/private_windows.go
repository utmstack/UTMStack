//go:build windows

package release

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func releaseUserSID() (*windows.SID, error) {
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return u.User.Sid, nil
}
func releaseTrustedSID(sid, current *windows.SID) bool {
	return sid != nil && (sid.Equals(current) || sid.IsWellKnown(windows.WinLocalSystemSid) || sid.IsWellKnown(windows.WinBuiltinAdministratorsSid))
}
func releaseOpen(root windows.Handle, name string, access, disposition, options uint32, sd *windows.SECURITY_DESCRIPTOR) (*os.File, error) {
	object, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: root, ObjectName: object, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE, SecurityDescriptor: sd}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	share := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE)
	if options&windows.FILE_NON_DIRECTORY_FILE != 0 {
		// DACL hardening cannot revoke a retained writer's existing handle.
		// Refuse such handles before granting published read/execute access.
		share = windows.FILE_SHARE_READ
	}
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, share, disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT, 0, 0)
	runtime.KeepAlive(sd)
	if err != nil {
		if status, ok := err.(windows.NTStatus); ok {
			err = status.Errno()
		}
		return nil, err
	}
	f := os.NewFile(uintptr(handle), name)
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		f.Close()
		return nil, errors.Join(errors.New("release path is not an ordinary file or directory"), err)
	}
	return f, nil
}
func releaseNTPath(path string) (string, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || len(path) > 4096 || strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\??\`) {
		return "", errors.New("release path must be an ordinary clean absolute path")
	}
	if strings.HasPrefix(path, `\\`) {
		return `\??\UNC\` + strings.TrimPrefix(path, `\\`), nil
	}
	return `\??\` + path, nil
}
func releaseOpenDirectory(path string, access uint32) (*os.File, error) {
	nt, err := releaseNTPath(path)
	if err != nil {
		return nil, err
	}
	return releaseOpen(0, nt, access|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil)
}

// Read access in a parent is acceptable; foreign users must not replace children.
func releaseDirectoryPolicy(f *os.File, private bool) error {
	sd, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	current, err := releaseUserSID()
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil || !releaseTrustedSID(owner, current) {
		return errors.New("release directory owner is untrusted")
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	if private && control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("release directory DACL is not protected")
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount == 0 {
		return errors.New("release directory has no restrictive DACL")
	}
	const writes = windows.GENERIC_ALL | windows.GENERIC_WRITE | windows.DELETE | windows.WRITE_DAC | windows.WRITE_OWNER | windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA | windows.FILE_WRITE_EA | windows.FILE_WRITE_ATTRIBUTES | 0x40
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace.Header.AceFlags&windows.INHERIT_ONLY_ACE != 0 {
			continue
		}
		if ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE {
			continue
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return errors.New("release directory has unsupported access entry")
		}
		if !releaseTrustedSID((*windows.SID)(unsafe.Pointer(&ace.SidStart)), current) && (private || ace.Mask&writes != 0) {
			return errors.New("release directory grants foreign access")
		}
	}
	return nil
}
func releaseSD(published bool) (*windows.SECURITY_DESCRIPTOR, error) {
	sid, err := releaseUserSID()
	if err != nil {
		return nil, err
	}
	flags := "OICI"
	extra := ""
	if published {
		flags = ""
		extra = "(A;;GRGX;;;BU)"
	}
	return windows.SecurityDescriptorFromString(fmt.Sprintf("O:%sD:P(A;%s;FA;;;SY)(A;%s;FA;;;BA)(A;%s;FA;;;%s)%s", sid.String(), flags, flags, flags, sid.String(), extra))
}

// MakePrivateTempDir atomically creates a protected inheritable DACL under a
// pinned parent that cannot be replaced during creation.
func MakePrivateTempDir(parent, prefix string) (string, error) {
	if parent == "" {
		parent = os.TempDir()
	}
	parent, err := filepath.Abs(parent)
	if err != nil {
		return "", err
	}
	if strings.ContainsAny(prefix, `/\:`) {
		return "", errors.New("invalid release directory prefix")
	}
	p, err := releaseOpenDirectory(parent, windows.FILE_APPEND_DATA|windows.FILE_TRAVERSE)
	if err != nil {
		return "", err
	}
	defer p.Close()
	if err = releaseDirectoryPolicy(p, false); err != nil {
		return "", err
	}
	sd, err := releaseSD(false)
	if err != nil {
		return "", err
	}
	for range 16 {
		var random [16]byte
		if _, err = rand.Read(random[:]); err != nil {
			return "", err
		}
		name := prefix + hex.EncodeToString(random[:])
		child, err := releaseOpen(windows.Handle(p.Fd()), name, windows.FILE_LIST_DIRECTORY|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_CREATE, windows.FILE_DIRECTORY_FILE, sd)
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) || errors.Is(err, windows.ERROR_FILE_EXISTS) {
			continue
		}
		if err != nil {
			return "", err
		}
		err = releaseDirectoryPolicy(child, true)
		closeErr := child.Close()
		path := filepath.Join(parent, name)
		if err != nil || closeErr != nil {
			os.Remove(path)
			return "", errors.Join(err, closeErr)
		}
		return path, nil
	}
	return "", errors.New("release directory collision limit")
}

// CheckPrivateDirectory checks native owner and protected access entries.
func CheckPrivateDirectory(path string) error {
	f, err := releaseOpenDirectory(path, windows.FILE_LIST_DIRECTORY)
	if err != nil {
		return err
	}
	defer f.Close()
	return releaseDirectoryPolicy(f, true)
}

// PreparePublishedFile grants ordinary users read/execute only after promotion.
// Staging and rollback files retain their private inherited access policy.
func PreparePublishedFile(path string) error {
	parent, err := releaseOpenDirectory(filepath.Dir(path), windows.FILE_TRAVERSE)
	if err != nil {
		return err
	}
	defer parent.Close()
	if err = releaseDirectoryPolicy(parent, false); err != nil {
		return err
	}
	// Stream sharing checks require data/execute/delete access. Metadata-only
	// access cannot fence a writer or deleter that already holds the file.
	f, err := releaseOpen(windows.Handle(parent.Fd()), filepath.Base(path), windows.FILE_GENERIC_READ|windows.WRITE_DAC, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil)
	if err != nil {
		return err
	}
	defer f.Close()
	old, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	current, err := releaseUserSID()
	if err != nil {
		return err
	}
	owner, _, err := old.Owner()
	if err != nil || !releaseTrustedSID(owner, current) {
		return errors.New("published release owner is untrusted")
	}
	sd, err := releaseSD(true)
	if err != nil {
		return err
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
