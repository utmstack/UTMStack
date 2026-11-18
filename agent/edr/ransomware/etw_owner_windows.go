//go:build windows

package ransomware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/0xrawsec/golang-etw/etw"
	"golang.org/x/sys/windows"
)

const traceDirectorySDDL = "O:SYG:SYD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"
const traceFileSDDL = "O:SYG:SYD:P(A;;FA;;;SY)(A;;FA;;;BA)"

type traceOwnerStore struct {
	dir  *os.File
	lock *os.File
}

func traceNTFile(root windows.Handle, name string, access, disposition, options uint32, sd *windows.SECURITY_DESCRIPTOR, share uint32) (*os.File, error) {
	object, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: root, ObjectName: object, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE, SecurityDescriptor: sd}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, share, disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	runtime.KeepAlive(sd)
	if status, ok := err.(windows.NTStatus); ok {
		err = status.Errno()
	}
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handle), name), nil
}

// Trust the opened kernel object, not an environment-provided path or a
// previous pathname check. Only SYSTEM/administrators can mutate these records.
func verifyTraceOwnerFile(file *os.File, directory bool) error {
	h := windows.Handle(file.Fd())
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil || !owner.IsWellKnown(windows.WinLocalSystemSid) {
		return errors.New("ETW ownership object is not SYSTEM-owned")
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("ETW ownership DACL is not protected")
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil || acl.AceCount != 2 {
		return errors.New("ETW ownership DACL is not restricted")
	}
	var system, admin bool
	flags := byte(0)
	if directory {
		flags = windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE
	}
	for i := uint32(0); i < 2; i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != flags || ace.Mask != 0x001f01ff {
			return errors.New("ETW ownership ACE grants unexpected rights")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		switch {
		case sid.IsWellKnown(windows.WinLocalSystemSid):
			if system {
				return errors.New("duplicate ETW owner ACE")
			}
			system = true
		case sid.IsWellKnown(windows.WinBuiltinAdministratorsSid):
			if admin {
				return errors.New("duplicate ETW administrator ACE")
			}
			admin = true
		default:
			return errors.New("ETW ownership ACE has a foreign trustee")
		}
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory || !directory && info.NumberOfLinks != 1 {
		return errors.New("ETW ownership object type or link count is unsafe")
	}
	return nil
}

func openTraceOwnerStore(ctx context.Context) (*traceOwnerStore, error) {
	programData, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(programData) || filepath.Clean(programData) != programData || len(programData) > 4096 || len(filepath.VolumeName(programData)) != 2 {
		return nil, errors.New("ETW ownership requires a local ordinary ProgramData directory")
	}
	// Do not deny ordinary OS writes/renames on the shared ProgramData parent.
	// Subsequent operations use the pinned private child handle, never a path.
	parentFile, err := traceNTFile(0, `\??\`+programData, windows.FILE_LIST_DIRECTORY|windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return nil, err
	}
	defer parentFile.Close()
	sd, err := windows.SecurityDescriptorFromString(traceDirectorySDDL)
	if err != nil {
		return nil, err
	}
	f, err := traceNTFile(windows.Handle(parentFile.Fd()), "UTMStackEDRTraceOwners", windows.FILE_LIST_DIRECTORY|windows.FILE_READ_ATTRIBUTES|windows.READ_CONTROL, windows.FILE_OPEN_IF, windows.FILE_DIRECTORY_FILE, sd, windows.FILE_SHARE_READ)
	if err != nil {
		return nil, err
	}
	store := &traceOwnerStore{dir: f}
	ok := false
	defer func() {
		if !ok {
			store.Close()
		}
	}()
	if err = verifyTraceOwnerFile(f, true); err != nil {
		return nil, err
	}
	if err = store.acquire(ctx); err != nil {
		return nil, err
	}
	ok = true
	return store, nil
}

// Startup and normal record removal share one bounded cross-process lock.
func (s *traceOwnerStore) acquire(ctx context.Context) error {
	if s.lock != nil {
		return nil
	}
	fileSD, err := windows.SecurityDescriptorFromString(traceFileSDDL)
	if err != nil {
		return err
	}
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	for {
		s.lock, err = traceNTFile(windows.Handle(s.dir.Fd()), "startup.lock", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.READ_CONTROL, windows.FILE_OPEN_IF, windows.FILE_NON_DIRECTORY_FILE, fileSD, 0)
		if err == nil {
			break
		}
		if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("ETW ownership lock deadline")
		case <-time.After(25 * time.Millisecond):
		}
	}
	if err = verifyTraceOwnerFile(s.lock, false); err != nil {
		return err
	}
	return nil
}
func (s *traceOwnerStore) unlock() {
	if s.lock != nil {
		_ = s.lock.Close()
		s.lock = nil
	}
}
func (s *traceOwnerStore) Close() {
	s.unlock()
	if s.dir != nil {
		_ = s.dir.Close()
		s.dir = nil
	}
}

func (s *traceOwnerStore) openRecord(name string, create bool) (*os.File, error) {
	if name == "" || strings.ContainsAny(name, `/\:`) || name == "." || name == ".." {
		return nil, errors.New("invalid ETW ownership filename")
	}
	disposition := uint32(windows.FILE_OPEN)
	access := uint32(windows.FILE_GENERIC_READ | windows.READ_CONTROL | windows.DELETE)
	var sd *windows.SECURITY_DESCRIPTOR
	var err error
	if create {
		disposition = windows.FILE_CREATE
		access |= windows.FILE_GENERIC_WRITE
		sd, err = windows.SecurityDescriptorFromString(traceFileSDDL)
		if err != nil {
			return nil, err
		}
	}
	f, err := traceNTFile(windows.Handle(s.dir.Fd()), name, access, disposition, windows.FILE_NON_DIRECTORY_FILE, sd, windows.FILE_SHARE_READ)
	if err != nil {
		return nil, err
	}
	if err = verifyTraceOwnerFile(f, false); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
func deleteTraceRecord(file *os.File) error {
	var iosb windows.IO_STATUS_BLOCK
	flag := byte(1)
	err := windows.NtSetInformationFile(windows.Handle(file.Fd()), &iosb, &flag, 1, windows.FileDispositionInformation)
	if status, ok := err.(windows.NTStatus); ok {
		return status.Errno()
	}
	return err
}
func (s *traceOwnerStore) persist(record traceOwnerRecord) (result error) {
	if err := record.validate(); err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if len(data) > maxTraceOwnerBytes {
		return errors.New("ETW owner record size limit")
	}
	f, err := s.openRecord(record.filename(), true)
	if err != nil {
		return err
	}
	defer func() {
		if result != nil {
			_ = deleteTraceRecord(f)
		}
		f.Close()
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
func readTraceRecord(file *os.File) (traceOwnerRecord, error) {
	var record traceOwnerRecord
	data, err := io.ReadAll(io.LimitReader(file, maxTraceOwnerBytes+1))
	if err != nil {
		return record, err
	}
	if len(data) > maxTraceOwnerBytes {
		return record, errors.New("ETW ownership record size limit")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err = d.Decode(&record); err != nil {
		return record, err
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return record, errors.New("trailing ETW ownership content")
	}
	return record, record.validate()
}

func traceProcessState(record traceOwnerRecord) traceOwnerProcess {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, record.PID)
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return traceProcessGone
	}
	if err != nil {
		return traceProcessUnknown
	}
	defer windows.CloseHandle(h)
	var creation, exit, kernel, user windows.Filetime
	if windows.GetProcessTimes(h, &creation, &exit, &kernel, &user) != nil {
		return traceProcessUnknown
	}
	id := int64(uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime))
	if id <= 0 {
		return traceProcessUnknown
	}
	if id != record.Creation {
		return traceProcessReused
	}
	status, e := windows.WaitForSingleObject(h, 0)
	if e != nil {
		return traceProcessUnknown
	}
	if status == windows.WAIT_OBJECT_0 {
		return traceProcessGone
	}
	if status == uint32(windows.WAIT_TIMEOUT) {
		return traceProcessSameAlive
	}
	return traceProcessUnknown
}

func queriedTraceName(p *etw.EventTraceProperties) (string, error) {
	total, offset := p.Wnode.BufferSize, p.LoggerNameOffset
	if total > uint32(unsafe.Sizeof(*p))+2050 || offset < uint32(unsafe.Sizeof(*p)) || offset > total || total-offset < 2 || offset%2 != 0 {
		return "", errors.New("invalid native ETW session name bounds")
	}
	units := unsafe.Slice((*uint16)(unsafe.Add(unsafe.Pointer(p), uintptr(offset))), int((total-offset)/2))
	for i, v := range units {
		if v == 0 {
			return windows.UTF16ToString(units[:i]), nil
		}
	}
	return "", errors.New("unterminated native ETW session name")
}
func queryOwnedTrace(name string, handle uint64) (traceOwnerSnapshot, error) {
	// Reserve the documented maximum name space for native QUERY output.
	p := fileTraceProperties(strings.Repeat(" ", 1024))
	var nativeName *uint16
	var err error
	if handle == 0 {
		nativeName, err = windows.UTF16PtrFromString(name)
		if err != nil {
			return traceOwnerSnapshot{}, err
		}
	}
	err = etw.ControlTrace(syscall.Handle(handle), nativeName, p, etw.EVENT_TRACE_CONTROL_QUERY)
	if errors.Is(err, windows.ERROR_WMI_INSTANCE_NOT_FOUND) {
		return traceOwnerSnapshot{Known: true}, nil
	}
	if err != nil {
		return traceOwnerSnapshot{}, err
	}
	actualName, err := queriedTraceName(p)
	if err != nil {
		return traceOwnerSnapshot{}, err
	}
	return traceOwnerSnapshot{Known: true, Exists: true, Name: actualName, GUID: strings.ToUpper(p.Wnode.Guid.String()), Handle: p.Wnode.Union1}, nil
}
func stopVerifiedOrphan(record traceOwnerRecord, native traceOwnerSnapshot) error {
	// Query the freshly obtained handle again; never issue STOP by a reusable
	// session name or by a persisted handle from a prior process lifetime.
	again, err := queryOwnedTrace(record.Name, native.Handle)
	if err != nil {
		return err
	}
	if !again.Exists || again.Name != record.Name || again.GUID != record.GUID || again.Handle != native.Handle {
		return errors.New("ETW ownership changed before orphan stop")
	}
	return etw.ControlTrace(syscall.Handle(again.Handle), nil, fileTraceProperties(record.Name), etw.EVENT_TRACE_CONTROL_STOP)
}
func (s *traceOwnerStore) cleanupOrphans() error {
	names, err := s.dir.Readdirnames(maxTraceOwnerRecords + 2)
	if err != nil && err != io.EOF {
		return err
	}
	if len(names) > maxTraceOwnerRecords+1 {
		return errors.New("ETW ownership record count limit")
	}
	count := 0
	for _, name := range names {
		if name == "startup.lock" {
			continue
		}
		count++
		if count > maxTraceOwnerRecords {
			return errors.New("ETW ownership record count limit")
		}
		f, err := s.openRecord(name, false)
		if err != nil {
			return err
		}
		record, e := readTraceRecord(f)
		if e == nil && record.filename() != name {
			e = errors.New("ETW ownership record name mismatch")
		}
		if e == nil {
			native, qerr := queryOwnedTrace(record.Name, 0)
			if qerr != nil {
				e = qerr
			} else {
				var action traceOrphanAction
				action, e = decideTraceOrphan(record, traceProcessState(record), native)
				if e == nil && action == traceStopOwnedOrphan {
					e = stopVerifiedOrphan(record, native)
				}
				if e == nil && action != traceLeaveActive {
					e = deleteTraceRecord(f)
					count--
				}
			}
		}
		f.Close()
		if e != nil {
			return fmt.Errorf("ransomware ETW ownership remains unverified: %w", e)
		}
	}
	if count >= maxTraceOwnerRecords {
		return errors.New("ETW ownership record count limit")
	}
	return nil
}
