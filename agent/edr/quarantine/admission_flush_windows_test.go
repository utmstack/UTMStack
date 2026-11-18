//go:build windows && edr_acceptance

package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type flushProbeCall struct {
	Error        string `json:"error,omitempty"`
	Microseconds int64  `json:"microseconds"`
}
type flushProbeRow struct {
	Fixture     string          `json:"fixture"`
	Access      string          `json:"access"`
	Mask        uint32          `json:"mask,omitempty"`
	Share       uint32          `json:"share"`
	Open        flushProbeCall  `json:"open"`
	FileID      string          `json:"file_id,omitempty"`
	Sync        *flushProbeCall `json:"flush_file_buffers,omitempty"`
	NtNormal    *flushProbeCall `json:"nt_flush_normal,omitempty"`
	Reopen      *flushProbeCall `json:"reopen_write,omitempty"`
	SourceFlush *flushProbeCall `json:"production_source_flush,omitempty"`
	Volume      *flushProbeCall `json:"matched_volume_flush,omitempty"`
	VolumeGUID  string          `json:"volume_guid,omitempty"`
	FileSystem  string          `json:"filesystem,omitempty"`
	BeforeSHA   string          `json:"before_sha256"`
	AfterSHA    string          `json:"after_sha256"`
}

var probeNtFlushEx = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtFlushBuffersFileEx")
var probeReopen = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReOpenFile")

func probeTimed(fn func() error) flushProbeCall {
	start := time.Now()
	err := fn()
	r := flushProbeCall{Microseconds: time.Since(start).Microseconds()}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}
func probeDigest(t *testing.T, path string) string {
	t.Helper()
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		t.Fatal(e)
	}
	return hex.EncodeToString(h.Sum(nil))
}
func probeNtFlush(f *os.File, flags uint32) error {
	if e := probeNtFlushEx.Find(); e != nil {
		return e
	}
	var status windows.IO_STATUS_BLOCK
	r, _, _ := probeNtFlushEx.Call(f.Fd(), uintptr(flags), 0, 0, uintptr(unsafe.Pointer(&status)))
	runtime.KeepAlive(f)
	if uint32(r) != 0 {
		return fmt.Errorf("NTSTATUS 0x%08x", uint32(r))
	}
	return nil
}
func probeReopenWrite(f *os.File) error {
	r, _, e := probeReopen.Call(f.Fd(), uintptr(windows.GENERIC_WRITE), uintptr(windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE), 0)
	runtime.KeepAlive(f)
	if r == ^uintptr(0) {
		return e
	}
	return windows.CloseHandle(windows.Handle(r))
}

// The volume name comes from the pinned file handle, never a caller drive letter.
// Opening/flush/close is the entire volume operation; no raw writes or IOCTLs.
func probeMatchedVolume(f *os.File) (guid, filesystem string, err error) {
	buf := make([]uint16, 32768)
	n, e := windows.GetFinalPathNameByHandle(windows.Handle(f.Fd()), &buf[0], uint32(len(buf)), 1)
	if e != nil || n == 0 || n >= uint32(len(buf)) {
		return "", "", errors.Join(e, errors.New("cannot obtain bounded volume-GUID file name"))
	}
	path := windows.UTF16ToString(buf[:n])
	const prefix = `\\?\Volume{`
	if !strings.HasPrefix(path, prefix) {
		return "", "", errors.New("pinned file has no local volume GUID")
	}
	end := strings.Index(path, "}\\")
	if end < 0 {
		return "", "", errors.New("invalid volume GUID name")
	}
	guid = path[:end+1]
	if len(guid) != len(prefix)+36+1 {
		return guid, "", errors.New("unexpected volume GUID length")
	}
	var info windows.ByHandleFileInformation
	if e = windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); e != nil {
		return guid, "", e
	}
	root, e := windows.UTF16PtrFromString(guid + `\`)
	if e != nil {
		return guid, "", e
	}
	var serial, max, flags uint32
	var fs [64]uint16
	if e = windows.GetVolumeInformation(root, nil, 0, &serial, &max, &flags, &fs[0], uint32(len(fs))); e != nil {
		return guid, "", e
	}
	filesystem = windows.UTF16ToString(fs[:])
	if serial != info.VolumeSerialNumber {
		return guid, filesystem, errors.New("pinned file/volume serial mismatch")
	}
	name, e := windows.UTF16PtrFromString(guid)
	if e != nil {
		return guid, filesystem, e
	}
	v, e := windows.CreateFile(name, windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if e != nil {
		return guid, filesystem, e
	}
	defer windows.CloseHandle(v)
	return guid, filesystem, windows.FlushFileBuffers(v)
}
func TestAdmissionFlushAccessProbe(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || os.Getenv("OPENEDR_FLUSH_PROBE") != "1" {
		t.Skip("owned disposable Windows access/flush research only")
	}
	payload := []byte(strings.Repeat("owned flush control ", 256))
	allShare := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	baseMask := uint32(windows.FILE_GENERIC_READ | windows.DELETE | windows.WRITE_DAC | windows.WRITE_OWNER | windows.FILE_WRITE_ATTRIBUTES)
	profiles := []struct {
		name        string
		mask, share uint32
	}{
		{"v104-read-delete-metadata", baseMask, windows.FILE_SHARE_READ},
		{"minimal-append", baseMask | windows.FILE_APPEND_DATA, windows.FILE_SHARE_READ},
		{"minimal-write-data", baseMask | windows.FILE_WRITE_DATA, windows.FILE_SHARE_READ},
		{"generic-write", baseMask | windows.FILE_GENERIC_WRITE, windows.FILE_SHARE_READ},
		{"production-source", 0, windows.FILE_SHARE_READ},
		{"unsafe-write-sharing-control", baseMask, windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE},
	}
	var rows []flushProbeRow
	for _, fixture := range []string{"normal", "readonly-attribute", "deny-data-write-acl", "read-mapping", "write-mapping", "retained-writer", "retained-reader-no-write-sharing", "retained-reader-sharing-all", "running-owned-image", "other-volume-normal"} {
		t.Run(fixture, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "owned.txt")
			if fixture == "other-volume-normal" {
				other := os.Getenv("OPENEDR_FLUSH_OTHER_VOLUME")
				if !filepath.IsAbs(other) {
					t.Fatal("owned second NTFS volume required")
				}
				dir, e := os.MkdirTemp(other, "flush-source-")
				if e != nil {
					t.Fatal(e)
				}
				defer os.RemoveAll(dir)
				path = filepath.Join(dir, "owned.txt")
			}
			if e := os.WriteFile(path, payload, 0600); e != nil {
				t.Fatal(e)
			}
			if fixture == "running-owned-image" {
				var e error
				path, e = os.Executable()
				if e != nil {
					t.Fatal(e)
				}
			}
			before := probeDigest(t, path)
			switch fixture {
			case "retained-reader-no-write-sharing", "retained-reader-sharing-all":
				share := allShare
				if fixture == "retained-reader-no-write-sharing" {
					share &^= windows.FILE_SHARE_WRITE
				}
				p, _ := windows.UTF16PtrFromString(path)
				h, err := windows.CreateFile(p, windows.GENERIC_READ, share, nil, windows.OPEN_EXISTING, 0, 0)
				if err != nil {
					t.Fatal(err)
				}
				defer windows.CloseHandle(h)
			case "readonly-attribute":
				p, _ := windows.UTF16PtrFromString(path)
				if e := windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_READONLY); e != nil {
					t.Fatal(e)
				}
				defer windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_NORMAL)
			case "deny-data-write-acl":
				sid, e := admissionUserSID()
				if e != nil {
					t.Fatal(e)
				}
				sd, e := windows.SecurityDescriptorFromString(fmt.Sprintf("O:%sD:P(D;;0x00000006;;;%s)(A;;FA;;;SY)(A;;FA;;;BA)(A;;FA;;;%s)", sid.String(), sid.String(), sid.String()))
				if e != nil {
					t.Fatal(e)
				}
				acl, _, e := sd.DACL()
				if e != nil {
					t.Fatal(e)
				}
				p, _ := windows.UTF16PtrFromString(path)
				h, e := windows.CreateFile(p, windows.WRITE_DAC|windows.READ_CONTROL, allShare, nil, windows.OPEN_EXISTING, 0, 0)
				if e != nil {
					t.Fatal(e)
				}
				e = windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
				windows.CloseHandle(h)
				if e != nil {
					t.Fatal(e)
				}
			case "read-mapping", "write-mapping", "retained-writer":
				access := uint32(windows.GENERIC_READ)
				if fixture != "read-mapping" {
					access |= windows.GENERIC_WRITE
				}
				p, _ := windows.UTF16PtrFromString(path)
				h, e := windows.CreateFile(p, access, allShare, nil, windows.OPEN_EXISTING, 0, 0)
				if e != nil {
					t.Fatal(e)
				}
				if fixture == "retained-writer" {
					defer windows.CloseHandle(h)
				} else {
					protect, mapAccess := uint32(windows.PAGE_READONLY), uint32(windows.FILE_MAP_READ)
					if fixture == "write-mapping" {
						protect = windows.PAGE_READWRITE
						mapAccess = windows.FILE_MAP_WRITE
					}
					m, e := windows.CreateFileMapping(h, nil, protect, 0, 0, nil)
					windows.CloseHandle(h)
					if e != nil {
						t.Fatal(e)
					}
					defer windows.CloseHandle(m)
					view, e := windows.MapViewOfFile(m, mapAccess, 0, 0, 0)
					if e != nil {
						t.Fatal(e)
					}
					defer windows.UnmapViewOfFile(view)
				}
			}
			parent, e := platformRestoreDirectory(filepath.Dir(path))
			if e != nil {
				t.Fatal(e)
			}
			defer parent.Close()
			for _, profile := range profiles {
				row := flushProbeRow{Fixture: fixture, Access: profile.name, Mask: profile.mask, Share: profile.share, BeforeSHA: before}
				var f *os.File
				row.Open = probeTimed(func() error {
					var e error
					if profile.name == "production-source" {
						f, e = admissionOpenSource(parent, filepath.Base(path))
					} else {
						f, e = admissionNTFile(windows.Handle(parent.file.Fd()), filepath.Base(path), profile.mask, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE, nil, profile.share)
					}
					return e
				})
				if f != nil {
					row.FileID, e = admissionFileID(f)
					if e != nil {
						t.Fatal(e)
					}
					sync := probeTimed(f.Sync)
					row.Sync = &sync
					normal := probeTimed(func() error { return probeNtFlush(f, 0) })
					row.NtNormal = &normal

					reopen := probeTimed(func() error { return probeReopenWrite(f) })
					row.Reopen = &reopen
					if profile.name == "v104-read-delete-metadata" {
						v := probeTimed(func() error { var e error; row.VolumeGUID, row.FileSystem, e = probeMatchedVolume(f); return e })
						row.Volume = &v
					}
					if profile.name == "production-source" {
						barrier := probeTimed(func() error { return admissionSyncSource(f) })
						row.SourceFlush = &barrier
					}
					if e := f.Close(); e != nil {
						t.Fatal(e)
					}
				}
				if profile.name == "production-source" {
					refused := fixture == "retained-writer" || fixture == "write-mapping"
					if refused && row.Open.Error == "" {
						t.Fatal("production source open admitted a retained writer/mapping", row)
					}
					if !refused && (row.Open.Error != "" || row.SourceFlush == nil || row.SourceFlush.Error != "") {
						t.Fatal("production source barrier lost eligible source", row)
					}
				}
				row.AfterSHA = probeDigest(t, path)
				if row.AfterSHA != row.BeforeSHA {
					t.Fatal("flush access probe changed source bytes", row)
				}
				rows = append(rows, row)
				body, _ := json.Marshal(row)
				t.Logf("FLUSH_PROBE_ROW %s", body)
			}
		})
	}
	report := struct {
		Schema int             `json:"schema"`
		Go     string          `json:"go"`
		Arch   string          `json:"arch"`
		Rows   []flushProbeRow `json:"rows"`
	}{1, runtime.Version(), runtime.GOARCH, rows}
	if len(rows) != 60 {
		t.Fatalf("incomplete probe matrix: %d", len(rows))
	}
	output := os.Getenv("OPENEDR_FLUSH_RESULT")
	if !filepath.IsAbs(output) {
		t.Fatal("owned result path required")
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	e = json.NewEncoder(f).Encode(report)
	e = errors.Join(e, f.Sync(), f.Close())
	if e != nil {
		t.Fatal(e)
	}
}
