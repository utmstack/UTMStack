//go:build windows && edr_acceptance

package watcher

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type acceptanceCursors struct {
	initialID   uint64
	initialNext int64
	ready       chan struct{}
	once        sync.Once
	mu          sync.Mutex
	id          uint64
	next        int64
}

func (c *acceptanceCursors) LoadUSN(string) (uint64, int64) {
	c.once.Do(func() { close(c.ready) })
	return c.initialID, c.initialNext
}
func (c *acceptanceCursors) SaveUSN(_ string, id uint64, next int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.id, c.next = id, next
}

type acceptanceSink struct {
	path  string
	found chan string
}

func (s acceptanceSink) Enqueue(path, op string) {
	if strings.EqualFold(path, s.path) {
		select {
		case s.found <- op:
		default:
		}
	}
}

func TestWindowsAcceptanceUSNCreateAndExtend(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "inert-usn-fixture.txt")
	cursors := &acceptanceCursors{ready: make(chan struct{})}
	sink := acceptanceSink{path: path, found: make(chan string, 16)}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	w := New([]string{filepath.VolumeName(dir)}, sink, cursors, func(p string) bool { return !strings.HasPrefix(strings.ToLower(p), strings.ToLower(dir)) })
	attachWatcherReadDiagnostic(t, w)
	go w.Run(ctx)
	select {
	case <-cursors.ready:
	case <-ctx.Done():
		t.Fatal("USN query did not reach cursor initialization; journal/access unavailable")
	}
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(path, []byte("inert creation\n"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case op := <-sink.found:
		if op != "usn" {
			t.Fatalf("source=%s", op)
		}
	case <-ctx.Done():
		t.Fatal("real NTFS create was not delivered")
	}
	for len(sink.found) > 0 {
		<-sink.found
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("inert extension\n"))
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("append=%v close=%v", err, closeErr)
	}
	select {
	case <-sink.found:
	case <-ctx.Done():
		t.Fatal("real NTFS extension/close was not delivered")
	}
	cancel()
	time.Sleep(600 * time.Millisecond)
	cursors.mu.Lock()
	defer cursors.mu.Unlock()
	if cursors.id == 0 || cursors.next <= 0 {
		t.Fatalf("cursor was not saved: id=%d next=%d", cursors.id, cursors.next)
	}
	t.Logf("NTFS create/extend delivered, journal=%d cursor=%d", cursors.id, cursors.next)
}

func TestWindowsAcceptanceUSNInvalidSavedCursorResync(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	dir := t.TempDir()
	volume := filepath.VolumeName(dir)
	name, err := windows.UTF16PtrFromString(`\\.\` + volume)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	var journal usnJournalData
	var size uint32
	err = windows.DeviceIoControl(handle, fsctlQueryUSNJournal, nil, 0, (*byte)(unsafe.Pointer(&journal)), uint32(unsafe.Sizeof(journal)), &size, nil)
	windows.CloseHandle(handle)
	if err != nil {
		t.Fatal(err)
	}
	cursors := &acceptanceCursors{ready: make(chan struct{}), initialID: journal.UsnJournalID, initialNext: journal.NextUsn + 1<<30}
	path := filepath.Join(dir, "owned-after-resync.txt")
	sink := acceptanceSink{path: path, found: make(chan string, 16)}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	watcher := New([]string{volume}, sink, cursors, func(p string) bool { return !strings.HasPrefix(strings.ToLower(p), strings.ToLower(dir)) })
	attachWatcherReadDiagnostic(t, watcher)
	go watcher.Run(ctx)
	select {
	case <-cursors.ready:
	case <-ctx.Done():
		t.Fatal("journal was not queried")
	}
	time.Sleep(200 * time.Millisecond)
	if err := os.WriteFile(path, []byte("inert post-gap creation"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sink.found:
	case <-ctx.Done():
		t.Fatal("invalid saved cursor permanently blocked new NTFS events")
	}
	cancel()
	time.Sleep(600 * time.Millisecond)
	cursors.mu.Lock()
	defer cursors.mu.Unlock()
	if cursors.id != journal.UsnJournalID || cursors.next <= 0 || cursors.next >= cursors.initialNext {
		t.Fatalf("cursor did not resynchronize: %+v", cursors)
	}
	t.Logf("invalid cursor=%d resynchronized=%d; actual owned NTFS creation delivered", cursors.initialNext, cursors.next)
}

func attachWatcherReadDiagnostic(t *testing.T, watcher *Watcher) {
	t.Helper()
	records := make(chan string, 32)
	watcher.readDiagnostic = func(start int64, got uint32, next int64, err error) {
		select {
		case records <- fmt.Sprintf("start=%d returned=%d bytes=%d error=%v", start, next, got, err):
		default:
		}
	}
	t.Cleanup(func() {
		for len(records) > 0 {
			t.Log("bounded USN read:", <-records)
		}
	})
}
func TestWindowsAcceptanceOwnedNTFSResolution(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "owned-resolution.txt")
	if err := os.WriteFile(path, []byte("inert owned resolution"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		t.Fatal(err)
	}
	name, _ := windows.UTF16PtrFromString(`\\.\` + filepath.VolumeName(dir))
	volume, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(volume)
	id := uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)

	var descriptor struct {
		Size   uint32
		Type   uint32
		FileID [2]uint64
	}
	descriptor.Size = uint32(unsafe.Sizeof(descriptor))
	descriptor.FileID[0] = id
	direct, openErr := openFileByID(volume, unsafe.Pointer(&descriptor))
	t.Logf("direct owned OpenFileByID: size=%d alignment=%d error=%v", descriptor.Size, unsafe.Alignof(descriptor), openErr)
	if openErr == nil {
		windows.CloseHandle(direct)
	}
	resolved := resolvePath(volume, id)
	t.Logf("owned path=%q resolved=%q reference=%d", path, resolved, id)
	if !strings.EqualFold(resolved, path) {
		t.Fatal("owned NTFS reference did not resolve to its expected path")
	}
}

func TestWindowsAcceptanceUSNReadVariants(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	dir := t.TempDir()
	name, _ := windows.UTF16PtrFromString(`\\.\` + filepath.VolumeName(dir))
	volume, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(volume)
	query := func() usnJournalData {
		var jd usnJournalData
		var got uint32
		if err := windows.DeviceIoControl(volume, fsctlQueryUSNJournal, nil, 0, (*byte)(unsafe.Pointer(&jd)), uint32(unsafe.Sizeof(jd)), &got, nil); err != nil {
			t.Fatal(err)
		}
		return jd
	}
	before := query()
	path := filepath.Join(dir, "owned-read-variant.txt")
	if err := os.WriteFile(path, []byte("inert read variant"), 0600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	after := query()
	t.Logf("before=%+v after=%+v", before, after)
	aligned := alignedUSNBuffer()
	extra := make([]uint64, 65536/8+1)
	unaligned := unsafe.Slice((*byte)(unsafe.Pointer(&extra[0])), len(extra)*8)[1:65537]
	buffers := [][]byte{unaligned, aligned}
	for _, start := range []int64{before.FirstUsn, before.NextUsn} {
		for version := 0; version < 2; version++ {
			for maskIndex, mask := range []uint32{usnReasonMask, 0xffffffff} {
				for bufferIndex, buffer := range buffers {
					var request struct {
						readUsnData
						MinMajor uint16
						MaxMajor uint16
					}
					request.StartUsn = start
					request.ReasonMask = mask
					request.UsnJournalID = before.UsnJournalID
					request.MinMajor = 2
					request.MaxMajor = 2
					size := uint32(unsafe.Sizeof(request.readUsnData))
					if version == 1 {
						size = uint32(unsafe.Sizeof(request))
					}
					var got uint32
					err := windows.DeviceIoControl(volume, fsctlReadUSNJournal, (*byte)(unsafe.Pointer(&request)), size, &buffer[0], uint32(len(buffer)), &got, nil)
					t.Logf("variant start=%d version=%d input_size=%d mask=%d buffer=%d pointer_mod4096=%d got=%d error=%v", start, version, size, maskIndex, bufferIndex, uintptr(unsafe.Pointer(&buffer[0]))%4096, got, err)
					if bufferIndex == 0 {
						if !errors.Is(err, windows.ERROR_INVALID_USER_BUFFER) || got != 0 {
							t.Errorf("misaligned native buffer was not refused: got=%d error=%v", got, err)
						}
					} else if err != nil || got < 8 {
						t.Errorf("aligned native read failed: got=%d error=%v", got, err)
					}
				}
			}
		}
	}
}
