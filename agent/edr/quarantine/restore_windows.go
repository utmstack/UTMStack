//go:build windows

package quarantine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func restoreWindowsError(err error) error {
	if status, ok := err.(windows.NTStatus); ok {
		return status.Errno()
	}
	return err
}

func restoreNTFile(root windows.Handle, name string, access, disposition, options, attrs, share uint32) (*os.File, error) {
	object, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: root, ObjectName: object, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &oa, &iosb, nil, attrs, share, disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT, 0, 0)
	if err != nil {
		return nil, restoreWindowsError(err)
	}
	return os.NewFile(uintptr(handle), name), nil
}

func platformRestoreDirectory(path string) (*restoreDirectory, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\??\`) {
		return nil, errors.New("restore directory requires an ordinary absolute path")
	}
	volume := filepath.VolumeName(path)
	ntVolume := `\??\` + volume + `\`
	if strings.HasPrefix(volume, `\\`) {
		ntVolume = `\??\UNC\` + strings.TrimPrefix(volume, `\\`) + `\`
	}
	const share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE
	f, err := restoreNTFile(0, ntVolume, windows.FILE_TRAVERSE|windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, 0, share)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(strings.TrimPrefix(path, volume), `\`), `\`) {
		if part == "" {
			continue
		}
		next, e := restoreNTFile(windows.Handle(f.Fd()), part, windows.FILE_TRAVERSE|windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, 0, share)
		_ = f.Close()
		if e != nil {
			return nil, e
		}
		f = next
	}
	return &restoreDirectory{file: f, path: path}, nil
}

func (d *restoreDirectory) stillNamed() error {
	now, err := platformRestoreDirectory(d.path)
	if err != nil {
		return err
	}
	defer now.Close()
	before, err := d.file.Stat()
	if err != nil {
		return err
	}
	after, err := now.file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, after) {
		return errors.New("destination directory changed during recovery")
	}
	return nil
}

func (d *restoreDirectory) lockItem(id string) (func(), error) {
	sum := sha256.Sum256([]byte(id))
	f, err := restoreNTFile(windows.Handle(d.file.Fd()), fmt.Sprintf(".item-lock-%02x", sum[0]), windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, windows.FILE_OPEN_IF, windows.FILE_NON_DIRECTORY_FILE, windows.FILE_ATTRIBUTE_HIDDEN, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, fmt.Errorf("%w: invalid quarantine lock", ErrRestoreIntegrity)
	}
	overlap := new(windows.Overlapped)
	err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlap)
	if err != nil {
		f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrQuarantineBusy
		}
		return nil, err
	}
	return func() { _ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, overlap); _ = f.Close() }, nil
}

func disposeRestoreFile(f *os.File, remove bool) error {
	var flag byte
	if remove {
		flag = 1
	}
	var iosb windows.IO_STATUS_BLOCK
	return restoreWindowsError(windows.NtSetInformationFile(windows.Handle(f.Fd()), &iosb, &flag, 1, windows.FileDispositionInformation))
}

func (d *restoreDirectory) removeFile(name string) error {
	f, err := restoreNTFile(windows.Handle(d.file.Fd()), name, windows.DELETE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return &os.PathError{Op: "remove quarantine entry", Path: filepath.Join(d.path, name), Err: err}
	}
	err = disposeRestoreFile(f, true)
	return errors.Join(err, f.Close())
}

// FILE_RENAME_INFORMATION uses a byte BOOLEAN followed by native alignment.
// The zero ReplaceIfExists field is essential: existing targets must survive.
func renameRestoreHandle(f *os.File, parent *restoreDirectory, name string) error {
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
	info.RootDirectory = windows.Handle(parent.file.Fd())
	info.FileNameLength = uint32((len(encoded) - 1) * 2)
	copy(unsafe.Slice(&info.FileName[0], len(encoded)-1), encoded[:len(encoded)-1])
	var iosb windows.IO_STATUS_BLOCK
	return restoreWindowsError(windows.NtSetInformationFile(windows.Handle(f.Fd()), &iosb, &buffer[0], uint32(len(buffer)), windows.FileRenameInformation))
}

func restoreOpenInput(root *restoreDirectory, name string) (*os.File, error) {
	return restoreNTFile(windows.Handle(root.file.Fd()), name, windows.FILE_GENERIC_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, 0, windows.FILE_SHARE_READ)
}
func restoreCreateOutput(stage *restoreDirectory) (*os.File, error) {
	sd, err := admissionSD(false)
	if err != nil {
		return nil, err
	}
	return admissionNTFile(windows.Handle(stage.file.Fd()), "snapshot", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_WRITE_THROUGH, sd, windows.FILE_SHARE_READ|windows.FILE_SHARE_DELETE)
}
func restoreFinishOutput(out *os.File, _ os.FileInfo, security string) error {
	return restoreAdmissionSecurity(out, security)
}
func restorePublishOutput(out *os.File, _, parent *restoreDirectory, leaf string) error {
	return renameRestoreHandle(out, parent, leaf)
}
