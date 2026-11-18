//go:build windows

package quarantine

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func admissionSyncSource(f *os.File) error {
	err := f.Sync()
	if err == nil || !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return err
	}
	// Readonly and write-denied sources retain the original strict read pin.
	// Only a missing access right selects the administrative volume barrier;
	// never mask a failed file flush caused by an I/O or device error.
	return admissionSyncSourceVolume(f)
}

func admissionSyncSourceVolume(f *os.File) error {
	defer runtime.KeepAlive(f)
	const volumeNameGUID = 1
	buf := make([]uint16, 32768)
	n, err := windows.GetFinalPathNameByHandle(windows.Handle(f.Fd()), &buf[0], uint32(len(buf)), volumeNameGUID)
	if err != nil {
		return err
	}
	if n == 0 || n >= uint32(len(buf)) {
		return errors.New("source volume GUID exceeds the bounded name buffer")
	}
	path := windows.UTF16ToString(buf[:n])
	const prefix = `\\?\Volume{`
	end := strings.Index(path, "}\\")
	if !strings.HasPrefix(path, prefix) || end != len(prefix)+36 {
		return errors.New("source synchronization requires a local volume GUID")
	}
	volume := path[:end+1]
	if _, err = windows.GUIDFromString(volume[len(prefix)-1:]); err != nil {
		return fmt.Errorf("source volume GUID is invalid: %w", err)
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		return err
	}
	root, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return err
	}
	var serial uint32
	if err = windows.GetVolumeInformation(root, nil, 0, &serial, nil, nil, nil, 0); err != nil {
		return err
	}
	if serial != info.VolumeSerialNumber {
		return errors.New("source and flush volume identities differ")
	}
	name, err := windows.UTF16PtrFromString(volume)
	if err != nil {
		return err
	}
	// No trailing slash: open the derived volume, not its root directory.
	// This performs only open/flush/close, never raw writes or volume IOCTLs.
	v, err := windows.CreateFile(name, windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return err
	}
	return errors.Join(windows.FlushFileBuffers(v), windows.CloseHandle(v))
}
