//go:build windows

// Read-only deletion fallback follows Go 1.25.7 Deleteat's handle-only policy.
// Copyright 2024 The Go Authors. BSD license: LICENSE.GO-BSD.

package quarantine

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

func purgeOpenFile(d *restoreDirectory, name string) (*purgeFile, error) {
	f, err := admissionNTFile(windows.Handle(d.file.Fd()), name, windows.FILE_READ_ATTRIBUTES|windows.DELETE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE)
	if err != nil {
		return nil, err
	}
	return purgeFileFromDescriptor(f)
}

var purgeReopenFile = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReOpenFile")

func purgeRemoveFile(f *purgeFile, _ *restoreDirectory, _ string) error {
	// Preserve mapped-image checks and wait for actual namespace absence. This
	// deliberately does not request POSIX deletion of objects with live handles.
	flags := uint32(windows.FILE_DISPOSITION_DELETE | windows.FILE_DISPOSITION_FORCE_IMAGE_SECTION_CHECK | windows.FILE_DISPOSITION_IGNORE_READONLY_ATTRIBUTE)
	var iosb windows.IO_STATUS_BLOCK
	err := restoreWindowsError(windows.NtSetInformationFile(windows.Handle(f.file.Fd()), &iosb, (*byte)(unsafe.Pointer(&flags)), uint32(unsafe.Sizeof(flags)), windows.FileDispositionInformationEx))
	if err == nil {
		return nil
	}
	if !errors.Is(err, windows.ERROR_INVALID_PARAMETER) && !errors.Is(err, windows.ERROR_NOT_SUPPORTED) && !errors.Is(err, windows.ERROR_INVALID_FUNCTION) {
		return err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(windows.Handle(f.file.Fd()), &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_READONLY != 0 {
		// ReOpenFile refers to this exact object, never a raced pathname. Ask
		// for write-attributes access only when the legacy fallback needs it.
		h, _, callErr := purgeReopenFile.Call(f.file.Fd(), windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS)
		if h == uintptr(windows.InvalidHandle) || h == 0 {
			if callErr == nil || errors.Is(callErr, windows.ERROR_SUCCESS) {
				callErr = windows.ERROR_INVALID_HANDLE
			}
			return callErr
		}
		basic := struct {
			CreationTime, LastAccessTime, LastWriteTime, ChangeTime int64
			Attributes                                              uint32
		}{Attributes: info.FileAttributes &^ windows.FILE_ATTRIBUTE_READONLY}
		if basic.Attributes == 0 {
			basic.Attributes = windows.FILE_ATTRIBUTE_NORMAL
		}
		err = windows.SetFileInformationByHandle(windows.Handle(h), windows.FileBasicInfo, (*byte)(unsafe.Pointer(&basic)), uint32(unsafe.Sizeof(basic)))
		closeErr := windows.CloseHandle(windows.Handle(h))
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
	}
	return disposeRestoreFile(f.file, true)
}
