//go:build windows

package ransomware

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type canaryDir struct{ file *os.File }

func canaryNTFile(root windows.Handle, name string, access, disposition, options, attrs uint32) (*os.File, error) {
	object, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: root, ObjectName: object, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &oa, &iosb, nil, attrs, windows.FILE_SHARE_READ, disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		if status, ok := err.(windows.NTStatus); ok {
			err = status.Errno()
		}
		return nil, err
	}
	return os.NewFile(uintptr(handle), name), nil
}
func openCanaryDir(path string) (*canaryDir, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || len(path) > 4096 || strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\??\`) {
		return nil, errors.New("canary directory must be an ordinary clean absolute path")
	}
	ntPath := `\??\` + path
	if strings.HasPrefix(path, `\\`) {
		ntPath = `\??\UNC\` + strings.TrimPrefix(path, `\\`)
	}
	f, err := canaryNTFile(0, ntPath, windows.FILE_LIST_DIRECTORY|windows.FILE_READ_ATTRIBUTES, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE, 0)
	if err != nil {
		return nil, err
	}
	return &canaryDir{f}, nil
}
func (d *canaryDir) Close() error { return d.file.Close() }
func (d *canaryDir) open(name string, create bool) (*os.File, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\:`) {
		return nil, errors.New("canary requires a single ordinary filename")
	}
	access, disposition, attrs := uint32(windows.FILE_GENERIC_READ), uint32(windows.FILE_OPEN), uint32(0)
	if create {
		access = windows.FILE_GENERIC_WRITE | windows.FILE_READ_ATTRIBUTES
		disposition = windows.FILE_CREATE
		attrs = windows.FILE_ATTRIBUTE_HIDDEN
	}
	return canaryNTFile(windows.Handle(d.file.Fd()), name, access, disposition, windows.FILE_NON_DIRECTORY_FILE, attrs)
}
