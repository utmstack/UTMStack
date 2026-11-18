package ransomware

import "testing"

const (
	decoyPath  = `C:\00__accounts.xlsx`
	otherPath  = `C:\Users\x\report.xlsx`
	decoyFO    = uint64(0xFFFF8A0D12345678)
	otherFO    = uint64(0xFFFF8A0D87654321)
	openOpts   = uint32(1<<24 | 0x60) // FILE_OPEN | FILE_SYNCHRONOUS_IO_NONALERT | FILE_NON_DIRECTORY_FILE
	writerPID  = 1308
	readerPID  = 2872
	createKind = kfCreate
)

func newTestDecoder() *kernelFileDecoder {
	return newKernelFileDecoder(func(p string) bool { return p == decoyPath })
}

func openRec(pid int, fo uint64, path string, opts uint32) kernelFileRecord {
	return kernelFileRecord{EventID: createKind, PID: pid, FileObject: fo, Path: path, CreateOptions: opts, HasCreateOptions: true}
}

// feed runs records through the decoder and returns what it emits.
func feed(d *kernelFileDecoder, recs ...kernelFileRecord) []FileEvent {
	var out []FileEvent
	for _, r := range recs {
		if fe, ok := d.decode(r); ok {
			out = append(out, fe)
		}
	}
	return out
}

func TestDecoder_ReadOnlyAccessEmitsNothing(t *testing.T) {
	// Test-Path / Get-Item / GetFileAttributesW / a plain read: open, read,
	// query, query security, close. None of it modifies the decoy.
	d := newTestDecoder()
	got := feed(d,
		openRec(readerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: 15, PID: readerPID, FileObject: decoyFO},                                   // Read
		kernelFileRecord{EventID: 22, PID: readerPID, FileObject: decoyFO, InfoClass: 5, HasInfoClass: true}, // QueryInformation
		kernelFileRecord{EventID: 32, PID: readerPID, FileObject: decoyFO},                                   // QuerySecurity
		kernelFileRecord{EventID: 13, PID: readerPID, FileObject: decoyFO},                                   // Cleanup
		kernelFileRecord{EventID: kfClose, PID: readerPID, FileObject: decoyFO},
	)
	if len(got) != 0 {
		t.Fatalf("read-only access must emit nothing, got %+v", got)
	}
	// An open that carries no create options (unparsable) is still just an open.
	if got := feed(d, kernelFileRecord{EventID: kfCreate, PID: readerPID, FileObject: decoyFO, Path: decoyPath}); len(got) != 0 {
		t.Fatalf("an open without options must emit nothing, got %+v", got)
	}
}

func TestDecoder_WriteThroughDecoyOpen(t *testing.T) {
	d := newTestDecoder()
	got := feed(d,
		openRec(writerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO},
	)
	if len(got) != 1 || got[0] != (FileEvent{PID: writerPID, Path: decoyPath, Op: OpWrite}) {
		t.Fatalf("write to an open decoy = %+v", got)
	}
}

func TestDecoder_PagingWriteIsNotAttributed(t *testing.T) {
	// A paging write (IRP_PAGING_IO) is the cache or memory manager writing back
	// data a process already wrote. It runs in whichever thread forced the
	// flush: the lazy writer (System) or, for a non-cached read, the reader
	// itself. It must not be blamed on that thread's process.
	d := newTestDecoder()
	got := feed(d,
		openRec(writerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: kfWrite, PID: readerPID, FileObject: decoyFO, IOFlags: 0x43, HasIOFlags: true}, // PAGING_IO|NOCACHE|SYNCHRONOUS_PAGING_IO
	)
	if len(got) != 0 {
		t.Fatalf("a paging write must not be attributed, got %+v", got)
	}
	// Writes a process issued itself, cached or non-cached, are attributed.
	for _, flags := range []uint32{0xA00, 0xA01, 0x60A00} {
		got := feed(d, kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO, IOFlags: flags, HasIOFlags: true})
		if len(got) != 1 || got[0] != (FileEvent{PID: writerPID, Path: decoyPath, Op: OpWrite}) {
			t.Errorf("write with IOFlags %#x = %+v, want an attributed write", flags, got)
		}
	}
}

func TestDecoder_WriteToOtherFileEmitsNothing(t *testing.T) {
	d := newTestDecoder()
	got := feed(d,
		openRec(writerPID, otherFO, otherPath, openOpts),
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: otherFO},
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: 0},
	)
	if len(got) != 0 {
		t.Fatalf("writes to non-decoy files must emit nothing, got %+v", got)
	}
}

func TestDecoder_SetInformationClasses(t *testing.T) {
	cases := []struct {
		class uint32
		op    FileOp
		emit  bool
	}{
		{4, 0, false},  // FileBasicInformation: attributes/timestamps only
		{14, 0, false}, // FilePositionInformation
		{16, 0, false}, // FileModeInformation
		{11, 0, false}, // FileLinkInformation: adds a name, content untouched
		{43, 0, false}, // FileIoPriorityHintInformation
		{10, OpRename, true},
		{65, OpRename, true},
		{56, OpRename, true},
		{66, OpRename, true},
		{40, OpRename, true},
		{13, OpDelete, true},
		{64, OpDelete, true},
		{19, OpSetInfo, true},
		{20, OpSetInfo, true},
		{39, OpSetInfo, true},
	}
	for _, c := range cases {
		d := newTestDecoder()
		got := feed(d,
			openRec(writerPID, decoyFO, decoyPath, openOpts),
			kernelFileRecord{EventID: kfSetInformation, PID: writerPID, FileObject: decoyFO, InfoClass: c.class, HasInfoClass: true},
		)
		if !c.emit {
			if len(got) != 0 {
				t.Errorf("info class %d must not emit, got %+v", c.class, got)
			}
			continue
		}
		if len(got) != 1 || got[0].Op != c.op || got[0].Path != decoyPath {
			t.Errorf("info class %d = %+v, want op %d", c.class, got, c.op)
		}
	}
	// A SetInformation without a parsed class is not scored.
	d := newTestDecoder()
	if got := feed(d, openRec(writerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: kfSetInformation, PID: writerPID, FileObject: decoyFO}); len(got) != 0 {
		t.Fatalf("SetInformation without a class must not emit, got %+v", got)
	}
}

func TestDecoder_DeleteAndRenameEvents(t *testing.T) {
	for _, c := range []struct {
		id uint16
		op FileOp
	}{{kfSetDelete, OpDelete}, {kfRename, OpRename}, {kfDeletePath, OpDelete}, {kfRenamePath, OpRename}} {
		d := newTestDecoder()
		got := feed(d,
			openRec(writerPID, decoyFO, decoyPath, openOpts),
			kernelFileRecord{EventID: c.id, PID: writerPID, FileObject: decoyFO},
		)
		if len(got) != 1 || got[0] != (FileEvent{PID: writerPID, Path: decoyPath, Op: c.op}) {
			t.Errorf("event %d on an open decoy = %+v, want op %d", c.id, got, c.op)
		}
	}
}

func TestDecoder_PathEventsMatchDecoyWithoutOpen(t *testing.T) {
	// DeletePath/RenamePath carry the path themselves, so a decoy is recognised
	// even when its open was not seen (e.g. the feed started after the open).
	d := newTestDecoder()
	got := feed(d,
		kernelFileRecord{EventID: kfDeletePath, PID: writerPID, FileObject: otherFO, Path: decoyPath},
		kernelFileRecord{EventID: kfRenamePath, PID: writerPID, FileObject: otherFO, Path: decoyPath},
		kernelFileRecord{EventID: kfDeletePath, PID: writerPID, FileObject: otherFO, Path: otherPath},
	)
	if len(got) != 2 || got[0].Op != OpDelete || got[1].Op != OpRename {
		t.Fatalf("path events = %+v", got)
	}
}

func TestDecoder_CreateDispositions(t *testing.T) {
	cases := []struct {
		name string
		opts uint32
		op   FileOp
		emit bool
	}{
		{"FILE_OPEN", 1 << 24, 0, false},
		{"FILE_OPEN_IF", 3 << 24, 0, false},
		{"FILE_CREATE", 2 << 24, 0, false},
		{"FILE_SUPERSEDE (never sent by CreateFile; also what a missing disposition reads as)", 0, 0, false},
		{"FILE_OVERWRITE (TRUNCATE_EXISTING)", 4<<24 | 0x60, OpWrite, true},
		{"FILE_OVERWRITE_IF (CREATE_ALWAYS)", 5<<24 | 0x60, OpWrite, true},
		{"FILE_DELETE_ON_CLOSE", 1<<24 | 0x1000, OpDelete, true},
	}
	for _, c := range cases {
		d := newTestDecoder()
		got := feed(d, openRec(writerPID, decoyFO, decoyPath, c.opts))
		if !c.emit {
			if len(got) != 0 {
				t.Errorf("%s must not emit, got %+v", c.name, got)
			}
			continue
		}
		if len(got) != 1 || got[0] != (FileEvent{PID: writerPID, Path: decoyPath, Op: c.op}) {
			t.Errorf("%s = %+v, want op %d", c.name, got, c.op)
		}
	}
	// A destructive disposition on a non-decoy file is not reported.
	d := newTestDecoder()
	if got := feed(d, openRec(writerPID, otherFO, otherPath, 5<<24)); len(got) != 0 {
		t.Fatalf("non-decoy overwrite must not emit, got %+v", got)
	}
}

func TestDecoder_CloseEndsCorrelation(t *testing.T) {
	d := newTestDecoder()
	got := feed(d,
		openRec(writerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: kfClose, PID: writerPID, FileObject: decoyFO},
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO},
	)
	if len(got) != 0 {
		t.Fatalf("a write after Close must not be attributed to the decoy, got %+v", got)
	}
}

func TestDecoder_ReusedFileObjectIsNotADecoy(t *testing.T) {
	// If the decoy's Close was lost, a later open of another file that reuses the
	// same FileObject address must replace the stale mapping.
	d := newTestDecoder()
	got := feed(d,
		openRec(readerPID, decoyFO, decoyPath, openOpts),
		openRec(writerPID, decoyFO, otherPath, openOpts),
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO},
	)
	if len(got) != 0 {
		t.Fatalf("write through a reused FileObject was attributed to the decoy: %+v", got)
	}
}

func TestDecoder_OtherProcessWritingSharedHandleIsAttributed(t *testing.T) {
	// Attribution is the process that issues the write, not the one that opened.
	d := newTestDecoder()
	got := feed(d,
		openRec(readerPID, decoyFO, decoyPath, openOpts),
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO},
	)
	if len(got) != 1 || got[0].PID != writerPID {
		t.Fatalf("write = %+v, want pid %d", got, writerPID)
	}
}

func TestDecoder_CorrelationCapacityIsStickyAndPreservesExistingOpens(t *testing.T) {
	d := newKernelFileDecoder(func(string) bool { return true })
	for i := 1; i <= maxTrackedDecoyOpens; i++ {
		d.decode(openRec(writerPID, uint64(i), decoyPath, openOpts))
	}
	if d.Err() != nil || len(d.opens) != maxTrackedDecoyOpens {
		t.Fatal("capacity prematurely degraded", d.Err(), len(d.opens))
	}
	// This many genuine concurrently open handles requires no lost Close event.
	d.decode(openRec(readerPID, maxTrackedDecoyOpens+1, decoyPath, openOpts))
	if d.Err() == nil || len(d.opens) != maxTrackedDecoyOpens {
		t.Fatal("overflow hidden or existing state evicted", d.Err(), len(d.opens))
	}
	failure := d.Err()
	event, ok := d.decode(kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: 1})
	if !ok || event.PID != writerPID || event.Path != decoyPath {
		t.Fatal("older writable correlation evicted", event)
	}
	if _, ok := d.decode(kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: maxTrackedDecoyOpens + 1}); ok {
		t.Fatal("unadmitted handle fabricated correlation")
	}
	d.decode(kernelFileRecord{EventID: kfClose, FileObject: 2})
	d.decode(openRec(readerPID, maxTrackedDecoyOpens+1, decoyPath, openOpts))
	if _, ok := d.decode(kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: maxTrackedDecoyOpens + 1}); !ok {
		t.Fatal("released capacity unusable")
	}
	if d.Err() != failure {
		t.Fatal("recovered capacity erased prior coverage loss")
	}
}

func TestDecoder_SetLinkNeverBecomesDestructiveRename(t *testing.T) {
	d := newTestDecoder()
	d.decode(openRec(readerPID, decoyFO, decoyPath, openOpts))
	for _, r := range []kernelFileRecord{
		{EventID: kfSetLink, PID: readerPID, FileObject: decoyFO},
		{EventID: 28, PID: readerPID, FileObject: decoyFO, Path: decoyPath},
		{EventID: kfSetInformation, PID: readerPID, FileObject: decoyFO, InfoClass: 11, HasInfoClass: true},
	} {
		if event, ok := d.decode(r); ok {
			t.Fatal("benign hard-link classified destructive", event)
		}
	}
	if kernelFileEventWanted(kfSetLink) {
		t.Fatal("SetLink must be filtered before property decoding")
	}
}

func TestDecoder_NilDecoySetEmitsNothing(t *testing.T) {
	d := newKernelFileDecoder(nil)
	got := feed(d,
		openRec(writerPID, decoyFO, decoyPath, 5<<24),
		kernelFileRecord{EventID: kfWrite, PID: writerPID, FileObject: decoyFO},
		kernelFileRecord{EventID: kfDeletePath, PID: writerPID, Path: decoyPath},
	)
	if len(got) != 0 {
		t.Fatalf("without a decoy set nothing is a decoy, got %+v", got)
	}
}

func TestKernelFileEventWanted(t *testing.T) {
	for _, id := range []uint16{kfCreate, kfClose, kfWrite, kfSetInformation, kfSetDelete, kfRename, kfDeletePath, kfRenamePath, kfCreateNewFile} {
		if !kernelFileEventWanted(id) {
			t.Errorf("event %d must be decoded", id)
		}
	}
	// NameCreate, NameDelete, Cleanup, Read, DirEnum, Flush, QueryInformation,
	// FSCTL, OperationEnd, DirNotify, SetLinkPath, SetSecurity, QuerySecurity,
	// SetEA, QueryEA.
	for _, id := range []uint16{10, 11, 13, 15, 20, 21, 22, 23, 24, 25, 28, 29, 31, 32, 33, 34} {
		if kernelFileEventWanted(id) {
			t.Errorf("event %d must be skipped before parsing", id)
		}
	}
}

func TestParseETWUint(t *testing.T) {
	cases := []struct {
		in   string
		want uint64
		ok   bool
	}{
		{"0xFFFF8A0D12345678", 0xFFFF8A0D12345678, true},
		{"0x1a", 0x1a, true},
		{"0X1A", 0x1a, true},
		{"83886176", 83886176, true},
		{" 12 ", 12, true},
		{"", 0, false},
		{"0x", 0, false},
		{"n/a", 0, false},
		{"-1", 0, false},
	}
	for _, c := range cases {
		got, ok := parseETWUint(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("parseETWUint(%q) = %d,%v want %d,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}
