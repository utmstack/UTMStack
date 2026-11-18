//go:build windows

package watcher

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"
	"time"
	"unsafe"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

const (
	fsctlQueryUSNJournal = 0x000900f4
	fsctlReadUSNJournal  = 0x000900bb
	usnReasonMask        = ReasonFileCreate | ReasonDataExtend | ReasonDataOverwrite | ReasonRenameNewName | ReasonClose
	maxPathChars         = 32768
)

type Watcher struct {
	volumes        []string
	sink           Sink
	cursors        CursorStore
	excluded       func(string) bool
	readDiagnostic func(int64, uint32, int64, error) // bounded native acceptance observer; nil in production
}

func New(volumes []string, sink Sink, cursors CursorStore, excluded func(string) bool) *Watcher {
	if excluded == nil {
		excluded = func(string) bool { return false }
	}
	return &Watcher{volumes: volumes, sink: sink, cursors: cursors, excluded: excluded}
}

func (w *Watcher) Run(ctx context.Context) {
	for _, vol := range w.volumes {
		go w.watchVolume(ctx, vol)
	}
	<-ctx.Done()
}

// usnJournalData mirrors USN_JOURNAL_DATA_V0.
type usnJournalData struct {
	UsnJournalID    uint64
	FirstUsn        int64
	NextUsn         int64
	LowestValidUsn  int64
	MaxUsn          int64
	MaximumSize     uint64
	AllocationDelta uint64
}

// readUsnData mirrors READ_USN_JOURNAL_DATA_V0.
type readUsnData struct {
	StartUsn          int64
	ReasonMask        uint32
	ReturnOnlyOnClose uint32
	Timeout           uint64
	BytesToWaitFor    uint64
	UsnJournalID      uint64
}

func (w *Watcher) watchVolume(ctx context.Context, vol string) {
	// vol like "C:" -> device path \\.\C:
	path, err := windows.UTF16PtrFromString(`\\.\` + vol)
	if err != nil {
		logger.Error("UTMStack EDR watcher: bad volume %s: %v", vol, err)
		return
	}
	// FILE_FLAG_BACKUP_SEMANTICS so this same volume handle can also serve as the
	// hint handle for OpenFileById during path resolution (reused per record).
	h, err := windows.CreateFile(path,
		windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		logger.Error("UTMStack EDR watcher: open %s: %v", vol, err)
		return
	}
	defer windows.CloseHandle(h)

	var jd usnJournalData
	var bytesRet uint32
	if err := windows.DeviceIoControl(h, fsctlQueryUSNJournal, nil, 0,
		(*byte)(unsafe.Pointer(&jd)), uint32(unsafe.Sizeof(jd)), &bytesRet, nil); err != nil {
		logger.Error("UTMStack EDR watcher: query journal %s: %v", vol, err)
		return
	}

	journalID, next := w.cursors.LoadUSN(vol)
	if journalID != jd.UsnJournalID || outsideJournalWindow(next, jd) {
		if journalID != 0 {
			logger.Error("UTMStack EDR watcher: USN coverage gap on %s (saved journal=%d cursor=%d, available=%d..%d); resynchronizing at tail", vol, journalID, next, jd.FirstUsn, jd.NextUsn)
		}
		journalID, next = jd.UsnJournalID, jd.NextUsn
		w.cursors.SaveUSN(vol, journalID, next)
	}

	buf := alignedUSNBuffer()
	// parentCache maps a directory's NTFS reference to its resolved path so an
	// excluded directory is resolved at most once (see the exclusion check below).
	parentCache := make(map[uint64]string, 256)
	// The cursor is persisted to edr.db, whose write generates its own USN churn.
	// Saving on every read batch created a self-sustaining feedback loop (each save
	// produced records the watcher then read, triggering another save) that pegged
	// CPU and buried real file events. Persist at most every saveInterval instead —
	// on restart at worst a couple of seconds of journal is re-scanned (idempotent).
	const saveInterval = 2 * time.Second
	var lastSave time.Time
	var lastReadError time.Time
	for {
		select {
		case <-ctx.Done():
			w.cursors.SaveUSN(vol, journalID, next) // persist final position on shutdown
			return
		default:
		}
		req := readUsnData{
			StartUsn:       next,
			ReasonMask:     usnReasonMask,
			Timeout:        0,
			BytesToWaitFor: 0,
			UsnJournalID:   journalID,
		}
		var got uint32
		err := windows.DeviceIoControl(h, fsctlReadUSNJournal,
			(*byte)(unsafe.Pointer(&req)), uint32(unsafe.Sizeof(req)),
			&buf[0], uint32(len(buf)), &got, nil)
		if w.readDiagnostic != nil {
			returned := next
			if got >= 8 {
				returned = int64(binary.LittleEndian.Uint64(buf[:8]))
			}
			w.readDiagnostic(next, got, returned, err)
		}
		if err != nil {
			if errors.Is(err, windows.ERROR_JOURNAL_ENTRY_DELETED) || errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
				var current usnJournalData
				var size uint32
				if queryErr := windows.DeviceIoControl(h, fsctlQueryUSNJournal, nil, 0, (*byte)(unsafe.Pointer(&current)), uint32(unsafe.Sizeof(current)), &size, nil); queryErr == nil && (journalID != current.UsnJournalID || outsideJournalWindow(next, current)) {
					logger.Error("UTMStack EDR watcher: USN coverage gap on %s after journal read: %v; resynchronizing at tail", vol, err)
					journalID, next = current.UsnJournalID, current.NextUsn
					clear(parentCache)
					w.cursors.SaveUSN(vol, journalID, next)
					lastSave = time.Now()
				}
			}
			// A buffer/alignment error is not evidence of an idle journal. Keep
			// failures visible without generating a self-sustaining log stream.
			if time.Since(lastReadError) >= 30*time.Second {
				logger.Error("UTMStack EDR watcher: journal read %s at %d failed: %v", vol, next, err)
				lastReadError = time.Now()
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if got < 8 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		previous := next
		next = int64(binary.LittleEndian.Uint64(buf[:8]))
		// Even a header-only result carries the next read position. A reason-
		// filtered prefix must not trap the reader before later matching records.
		if got == 8 {
			if next != previous && time.Since(lastSave) >= saveInterval {
				w.cursors.SaveUSN(vol, journalID, next)
				lastSave = time.Now()
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		changes, _ := ParseUSNBuffer(buf[8:got])
		for _, ch := range changes {
			if ch.Reason&(ReasonFileCreate|ReasonDataExtend|ReasonDataOverwrite|ReasonClose) == 0 {
				continue
			}
			// Skip records whose parent directory is excluded BEFORE resolving the
			// file itself. A volume's USN stream is dominated by high-write paths —
			// the EDR's own edr.db/log/spool churn and OS background files — almost
			// all under excluded directories. Resolving every one (an OpenFileById
			// round-trip) would peg a CPU core and starve the scanner. Parent paths
			// are cached, so an excluded directory costs one resolve.
			pdir := resolveParentCached(h, ch.ParentRefID, parentCache)
			if pdir != "" && w.excluded(pdir) {
				continue
			}
			full := resolvePath(h, ch.FileRefID)
			if full == "" {
				continue
			}
			w.sink.Enqueue(full, "usn")
		}
		if time.Since(lastSave) >= saveInterval {
			w.cursors.SaveUSN(vol, journalID, next)
			lastSave = time.Now()
		}
	}
}

// resolveParentCached resolves a directory reference to its path, memoizing
// successful lookups so the recurring parents of high-churn files (e.g. the
// EDR's own install directory) are resolved only once.
func resolveParentCached(volHandle windows.Handle, parentRef uint64, cache map[uint64]string) string {
	if p, ok := cache[parentRef]; ok {
		return p
	}
	p := resolvePath(volHandle, parentRef)
	if p != "" && len(cache) < 8192 {
		cache[parentRef] = p
	}
	return p
}

// resolvePath opens a file by its NTFS reference number (using an already-open
// handle on the same volume as the OpenFileById hint) and returns its full path
// via GetFinalPathNameByHandle. Returns "" if the file is gone (a transient
// reference — e.g. a temp/journal file already deleted — fails benignly with
// ERROR_INVALID_PARAMETER). Reusing volHandle avoids reopening \\.\<vol> on every
// USN record, which otherwise makes the watcher fall behind the journal under
// self-generated DB churn and miss real file events.
func resolvePath(volHandle windows.Handle, fileRef uint64) string {
	// FILE_ID_DESCRIPTOR for OpenFileById (Type=0 => FileId). The union member
	// is a 16-byte FILE_ID_128, so the struct is 24 bytes (Size=24); passing a
	// bare uint64 (16-byte struct) makes OpenFileById fail with ERROR_INVALID_PARAMETER.
	// The 64-bit NTFS file reference goes in the low 8 bytes of the 16-byte field.
	var idDesc struct {
		Size   uint32
		Type   uint32
		FileID [2]uint64 // preserve the native union's eight-byte alignment
	}
	idDesc.Size = uint32(unsafe.Sizeof(idDesc))
	idDesc.Type = 0
	idDesc.FileID[0] = fileRef

	fh, err := openFileByID(volHandle, unsafe.Pointer(&idDesc))
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(fh)

	bufN := make([]uint16, maxPathChars)
	n, err := windows.GetFinalPathNameByHandle(fh, &bufN[0], uint32(len(bufN)), 0)
	if err != nil || n == 0 {
		return ""
	}
	return normalizeFinalPath(windows.UTF16ToString(bufN[:n]))
}

// normalizeFinalPath drops the "\\?\" extended-length prefix that
// GetFinalPathNameByHandle returns, so surfaced object paths read as plain
// "C:\dir\file" (and "\\?\UNC\server\share" as "\\server\share").
func normalizeFinalPath(p string) string {
	if strings.HasPrefix(p, `\\?\UNC\`) {
		return `\\` + p[len(`\\?\UNC\`):]
	}
	if strings.HasPrefix(p, `\\?\`) {
		return p[len(`\\?\`):]
	}
	return p
}

var (
	modkernel32    = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileID = modkernel32.NewProc("OpenFileById")
)

func openFileByID(volHandle windows.Handle, idDesc unsafe.Pointer) (windows.Handle, error) {
	r, _, e := procOpenFileID.Call(
		uintptr(volHandle),
		uintptr(idDesc),
		uintptr(windows.GENERIC_READ),
		uintptr(windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE),
		0,
		uintptr(windows.FILE_FLAG_BACKUP_SEMANTICS),
	)
	h := windows.Handle(r)
	if h == windows.InvalidHandle {
		return 0, e
	}
	return h, nil
}

// A valid journal ID alone does not make a retained cursor readable after wrap.
func outsideJournalWindow(next int64, journal usnJournalData) bool {
	return next < journal.FirstUsn || next < journal.LowestValidUsn || next > journal.NextUsn
}

// FSCTL_READ_USN_JOURNAL requires native buffer alignment. A byte slice can
// reside at an arbitrary byte offset on a Go stack; word storage preserves the
// alignment contract whether the compiler chooses stack or heap allocation.
func alignedUSNBuffer() []byte {
	storage := make([]uint64, 64*1024/8)
	return unsafe.Slice((*byte)(unsafe.Pointer(&storage[0])), len(storage)*8)
}
