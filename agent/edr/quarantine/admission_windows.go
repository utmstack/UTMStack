//go:build windows

package quarantine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

func admissionUserSID() (*windows.SID, error) {
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return u.User.Sid, nil
}
func admissionSD(directory bool) (*windows.SECURITY_DESCRIPTOR, error) {
	sid, err := admissionUserSID()
	if err != nil {
		return nil, err
	}
	administrative := false
	if !sid.IsWellKnown(windows.WinLocalSystemSid) {
		// Use this process token's enabled owner groups, matching admissionUserSID.
		// An impersonated thread or a filtered administrator must not confer the
		// canonical privileged owner on another process identity.
		groups, e := windows.GetCurrentProcessToken().GetTokenGroups()
		if e != nil {
			return nil, e
		}
		for _, group := range groups.AllGroups() {
			if group.Sid.IsWellKnown(windows.WinBuiltinAdministratorsSid) && group.Attributes&windows.SE_GROUP_ENABLED != 0 && group.Attributes&windows.SE_GROUP_OWNER != 0 && group.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
				administrative = true
				break
			}
		}
	}
	return admissionPrincipalSD(sid, administrative, directory)
}

func admissionPrincipalSD(sid *windows.SID, administrative, directory bool) (*windows.SECURITY_DESCRIPTOR, error) {
	if sid == nil {
		return nil, errors.New("private metadata requires the actual process principal")
	}
	flags := ""
	if directory {
		flags = "OICI"
	}
	owner := sid.String()
	grants := fmt.Sprintf("(A;%s;FA;;;SY)(A;%s;FA;;;BA)", flags, flags)
	switch {
	case sid.IsWellKnown(windows.WinLocalSystemSid):
		owner = "SY"
	case administrative:
		// BA is already trusted by existing readers. An individual administrator
		// owner/ACE cannot be corroborated later by the SYSTEM service. The actor
		// recorded in each approval remains the actual individual process SID.
		owner = "BA"
	default:
		// Preserve operation under a configured non-administrative service account.
		grants += fmt.Sprintf("(A;%s;FA;;;%s)", flags, owner)
	}
	return windows.SecurityDescriptorFromString("O:" + owner + "D:P" + grants)
}
func admissionNTFile(root windows.Handle, name string, access, disposition, options uint32, sd *windows.SECURITY_DESCRIPTOR, share uint32) (*os.File, error) {
	object, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: root, ObjectName: object, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE, SecurityDescriptor: sd}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, share, disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT, 0, 0)
	runtime.KeepAlive(sd)
	if err != nil {
		return nil, restoreWindowsError(err)
	}
	return os.NewFile(uintptr(handle), name), nil
}
func admissionFileID(f *os.File) (string, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		return "", err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return "", ErrAdmissionChanged
	}
	return fmt.Sprintf("%08x:%08x%08x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow), nil
}
func admissionTrustedSID(sid, current *windows.SID) bool {
	return sid != nil && (sid.Equals(current) || sid.IsWellKnown(windows.WinLocalSystemSid) || sid.IsWellKnown(windows.WinBuiltinAdministratorsSid))
}
func admissionPrivateDirectory(f *os.File) error {
	sd, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	current, err := admissionUserSID()
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil || !admissionTrustedSID(owner, current) {
		return errors.New("quarantine directory has an untrusted owner")
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("quarantine directory requires a protected DACL")
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount == 0 || acl.AceCount > 3 {
		return errors.New("quarantine directory DACL is not private")
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err = windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || !admissionTrustedSID((*windows.SID)(unsafe.Pointer(&ace.SidStart)), current) {
			return errors.New("quarantine directory grants a foreign principal access")
		}
	}
	return nil
}
func openAdmissionRoot(path string, harden bool) (*restoreDirectory, error) {
	parent, err := platformRestoreDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	access := uint32(windows.FILE_LIST_DIRECTORY | windows.FILE_TRAVERSE | windows.FILE_READ_ATTRIBUTES | windows.READ_CONTROL)
	if harden {
		access |= windows.WRITE_DAC
	}
	f, err := admissionNTFile(windows.Handle(parent.file.Fd()), filepath.Base(path), access, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE)
	if err != nil {
		return nil, err
	}
	d := &restoreDirectory{file: f, path: path}
	if harden {
		old, e := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
		if e != nil {
			d.Close()
			return nil, e
		}
		current, e := admissionUserSID()
		if e != nil {
			d.Close()
			return nil, e
		}
		owner, _, e := old.Owner()
		if e != nil || !admissionTrustedSID(owner, current) {
			d.Close()
			return nil, errors.New("quarantine directory owner is untrusted")
		}
		sd, e := admissionSD(true)
		if e != nil {
			d.Close()
			return nil, e
		}
		acl, _, e := sd.DACL()
		if e != nil {
			d.Close()
			return nil, e
		}
		if e = windows.SetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); e != nil {
			d.Close()
			return nil, e
		}
	}
	if err = admissionPrivateDirectory(f); err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}
func admissionOpenChild(d *restoreDirectory, name string) (*restoreDirectory, error) {
	f, err := admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_LIST_DIRECTORY|windows.FILE_TRAVERSE|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE)
	if err != nil {
		return nil, err
	}
	child := &restoreDirectory{file: f, path: filepath.Join(d.path, name)}
	if err = admissionPrivateDirectory(f); err != nil {
		child.Close()
		return nil, err
	}
	return child, nil
}
func admissionMakeDirectory(d *restoreDirectory, name string) (*restoreDirectory, error) {
	sd, err := admissionSD(true)
	if err != nil {
		return nil, err
	}
	f, err := admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_LIST_DIRECTORY|windows.FILE_TRAVERSE|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_CREATE, windows.FILE_DIRECTORY_FILE, sd, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE)
	if err != nil {
		return nil, err
	}
	child := &restoreDirectory{file: f, path: filepath.Join(d.path, name)}
	if err = admissionPrivateDirectory(f); err != nil {
		child.Close()
		return nil, err
	}
	return child, nil
}
func admissionRemoveDirectory(d *restoreDirectory, name string) error {
	f, err := admissionNTFile(windows.Handle(d.file.Fd()), name, windows.DELETE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return err
	}
	err = disposeRestoreFile(f, true)
	return errors.Join(err, f.Close())
}
func admissionCreateFile(d *restoreDirectory, name string) (*os.File, error) {
	sd, err := admissionSD(false)
	if err != nil {
		return nil, err
	}
	return admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_WRITE_THROUGH, sd, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
}
func admissionReadFile(d *restoreDirectory, name string) (*os.File, error) {
	return admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_GENERIC_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_DELETE)
}
func admissionReadPinnedSource(d *restoreDirectory, name string) (*os.File, error) {
	// The caller continuously owns the strict source pin. This second handle
	// reads identity metadata only, and must accommodate that pin's append and
	// delete access. It does not relax the original handle's writer exclusion.
	return admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
}
func admissionOpenSource(d *restoreDirectory, name string) (*os.File, error) {
	const access = windows.FILE_GENERIC_READ | windows.DELETE | windows.WRITE_DAC | windows.WRITE_OWNER | windows.FILE_WRITE_ATTRIBUTES
	// Append access permits flushing without the write-data access that refuses
	// running images. A failed open has not pinned or changed any source: retry
	// the original access for readonly attributes or a DACL denying data writes.
	// Never close a successful pin or relax its writer/delete sharing exclusion.
	if f, err := admissionNTFile(windows.Handle(d.file.Fd()), name, access|windows.FILE_APPEND_DATA, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ); err == nil {
		return f, nil
	}
	return admissionNTFile(windows.Handle(d.file.Fd()), name, access, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ)
}
func admissionSyncDirectory(*restoreDirectory) error     { return nil } // Process-crash replay; native power-loss qualification is separate.
func admissionSnapshotOwner(*os.File, os.FileInfo) error { return nil } // Original security is restored from authenticated admission metadata.
func admissionClaim(input *os.File, _ *restoreDirectory, _ string, to *restoreDirectory, name string) error {
	return renameRestoreHandle(input, to, name)
}
func admissionRollback(input *os.File, from *restoreDirectory, old string, to *restoreDirectory, name string) error {
	// Windows denies source replacement while input is open without delete sharing.
	return renameRestoreHandle(input, to, name)
}
func admissionRename(from *restoreDirectory, old string, to *restoreDirectory, name string, replace bool) error {
	f, err := admissionNTFile(windows.Handle(from.file.Fd()), old, windows.DELETE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return err
	}
	defer f.Close()
	return renameAdmissionHandle(f, to, name, replace)
}
func renameAdmissionHandle(f *os.File, parent *restoreDirectory, name string, replace bool) error {
	encoded, err := windows.UTF16FromString(name)
	if err != nil {
		return err
	}
	type renameInfo struct {
		ReplaceIfExists uint32
		RootDirectory   windows.Handle
		FileNameLength  uint32
		FileName        [1]uint16
	}
	var layout renameInfo
	buffer := make([]byte, int(unsafe.Sizeof(layout))+(len(encoded)-1)*2)
	info := (*renameInfo)(unsafe.Pointer(&buffer[0]))
	if replace {
		info.ReplaceIfExists = 1
	}
	info.RootDirectory = windows.Handle(parent.file.Fd())
	info.FileNameLength = uint32((len(encoded) - 1) * 2)
	copy(unsafe.Slice(&info.FileName[0], len(encoded)-1), encoded[:len(encoded)-1])
	var iosb windows.IO_STATUS_BLOCK
	return restoreWindowsError(windows.NtSetInformationFile(windows.Handle(f.Fd()), &iosb, &buffer[0], uint32(len(buffer)), windows.FileRenameInformation))
}
func admissionOriginalSecurity(f *os.File) (string, error) {
	sd, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return "", err
	}
	return sd.String(), nil
}
func admissionProtectCaptured(f *os.File) error {
	sd, err := admissionSD(false)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, owner, nil, acl, nil)
}

func restoreAdmissionSecurity(f *os.File, text string) (result error) {
	if text == "" {
		return nil
	}
	sd, err := windows.SecurityDescriptorFromString(text)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	group, _, err := sd.Group()
	if err != nil {
		return err
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	protection := windows.SECURITY_INFORMATION(windows.UNPROTECTED_DACL_SECURITY_INFORMATION)
	if control&windows.SE_DACL_PROTECTED != 0 {
		protection = windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	apply := func() error {
		return windows.SetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|protection, owner, group, acl, nil)
	}
	if err = apply(); err == nil {
		return nil
	}
	// Privilege changes run on a disposable OS thread. It is intentionally
	// never unlocked: Go destroys the thread when this goroutine exits, including
	// every failure path. The caller's identity and process token stay unchanged.
	done := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		err := func() error {
			if err := windows.ImpersonateSelf(windows.SecurityImpersonation); err != nil {
				return err
			}
			defer windows.RevertToSelf()
			var token windows.Token
			if err := windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, true, &token); err != nil {
				return err
			}
			defer token.Close()
			name, err := windows.UTF16PtrFromString("SeRestorePrivilege")
			if err != nil {
				return err
			}
			var luid windows.LUID
			if err = windows.LookupPrivilegeValue(nil, name, &luid); err != nil {
				return err
			}
			privileges := windows.Tokenprivileges{PrivilegeCount: 1, Privileges: [1]windows.LUIDAndAttributes{{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED}}}
			if err = windows.AdjustTokenPrivileges(token, false, &privileges, 0, nil, nil); err != nil {
				return err
			}
			return apply()
		}()
		done <- err
	}()
	return <-done
}

func admissionRestoreSourceSecurity(f *os.File, security string) error {
	return restoreAdmissionSecurity(f, security)
}
