package ransomware

import (
	"errors"
	"strconv"
	"strings"
)

// Microsoft-Windows-Kernel-File event IDs the ransomware feed decodes. IDs,
// names and fields come from the provider's instrumentation manifest (Microsoft
// generated schema: github.com/microsoft/Tx/Generated/Microsoft_Windows_Kernel_File.cs;
// SetLink29 adds a hard-link and is deliberately excluded). Dump the manifest on
// the host with `wevtutil gp Microsoft-Windows-Kernel-File /ge /gm`; published
// copies: github.com/repnz/etw-providers-docs and
// github.com/jdu2600/Windows10EtwEvents). Field meanings follow the kernel FileIo
// classes Microsoft documents at learn.microsoft.com/windows/win32/etw/
// fileio-create, fileio-readwrite and fileio-info.
//
// Create (12) is emitted for EVERY IRP_MJ_CREATE, i.e. every open: read-only
// opens, attribute queries (GetFileAttributesW), Test-Path/Get-Item. It never
// means "modified" on its own. Write/SetInformation/SetDelete/Rename carry no
// file name, only the FileObject, which FileIo_Create documents as the
// "identifier that can be used for correlating operations to the same opened
// file object instance between file create and close events".
const (
	kfCreate         uint16 = 12 // Create: FileObject, CreateOptions, FileName
	kfClose          uint16 = 14 // Close: FileObject released (ends correlation)
	kfWrite          uint16 = 16 // Write: FileObject, IOFlags (no name)
	kfSetInformation uint16 = 17 // SetInformation: FileObject, InfoClass (no name)
	kfSetDelete      uint16 = 18 // SetDelete: delete disposition set on FileObject
	kfRename         uint16 = 19 // Rename: FileObject (no name)
	kfDeletePath     uint16 = 26 // DeletePath: FileObject, FilePath
	kfRenamePath     uint16 = 27 // RenamePath: FileObject, FilePath
	kfSetLink        uint16 = 29 // SetLink adds a hard-link name; no destructive modification
	kfCreateNewFile  uint16 = 30 // CreateNewFile: same fields as Create; a file was created
)

// kernelFileEventWanted reports whether an event ID is decoded at all. Reads (15),
// queries (22, 32, 34), directory enumeration (20), Cleanup (13) and the rest are
// skipped before their properties are parsed.
func kernelFileEventWanted(id uint16) bool {
	switch id {
	case kfCreate, kfClose, kfWrite, kfSetInformation, kfSetDelete, kfRename,
		kfDeletePath, kfRenamePath, kfCreateNewFile:
		return true
	}
	return false
}

// Create disposition and option values (wdm.h / NtCreateFile). Kernel-File's
// CreateOptions is the IRP_MJ_CREATE Parameters.Create.Options value: "The high
// 8 bits correspond to the value of the CreateDisposition parameter of
// ZwCreateFile, and the low 24 bits correspond to the value of the CreateOptions
// parameter" (learn.microsoft.com/windows-hardware/drivers/kernel/irp-mj-create).
const (
	fileOverwrite     = 4          // FILE_OVERWRITE: CreateFile TRUNCATE_EXISTING
	fileOverwriteIf   = 5          // FILE_OVERWRITE_IF: CreateFile CREATE_ALWAYS
	fileDeleteOnClose = 0x00001000 // FILE_DELETE_ON_CLOSE
)

// irpPagingIO is IRP_PAGING_IO (wdm.h). A Write event's IOFlags are the IRP
// flags ("IO request packet flags specified for this operation",
// learn.microsoft.com/windows/win32/etw/fileio-readwrite).
const irpPagingIO = 0x00000002

// createOp maps an open's CreateOptions to the modification the open itself
// performs: truncating the file or deleting it on close. Every other open
// (FILE_OPEN, FILE_OPEN_IF, FILE_CREATE) modifies nothing by itself.
// FILE_SUPERSEDE (0) is deliberately not scored: CreateFile never requests it,
// and 0 is also what a disposition that is not encoded would read as, so
// scoring it could turn every open back into a "write".
func createOp(opts uint32) (FileOp, bool) {
	if opts&fileDeleteOnClose != 0 {
		return OpDelete, true
	}
	switch opts >> 24 {
	case fileOverwrite, fileOverwriteIf:
		return OpWrite, true
	}
	return 0, false
}

// setInfoOp maps a FILE_INFORMATION_CLASS (values from
// learn.microsoft.com/windows-hardware/drivers/ddi/wdm/ne-wdm-_file_information_class)
// to a modification of the file's content, name or existence. Classes that only
// touch metadata or handle state are not scored: FileBasicInformation (4)
// changes attributes/timestamps, which scanners and backup tools set when they
// restore the last-access time or reset the archive bit; FilePositionInformation
// (14) and FileModeInformation (16) are handle state; FileLinkInformation (11)
// adds a name without changing the file.
func setInfoOp(class uint32) (FileOp, bool) {
	switch class {
	case 10, 40, 56, 65, 66: // Rename, ShortName, RenameBypassAccessCheck, RenameEx, RenameExBypassAccessCheck
		return OpRename, true
	case 13, 64: // Disposition, DispositionEx
		return OpDelete, true
	case 19, 20, 39: // Allocation, EndOfFile, ValidDataLength: truncate/extend the content
		return OpSetInfo, true
	}
	return 0, false
}

// kernelFileRecord is the subset of one Kernel-File event the decoder needs,
// extracted by the Windows feed. Path is already a drive-letter path.
type kernelFileRecord struct {
	EventID          uint16
	PID              int
	FileObject       uint64 // 0 = absent or unparsable
	Path             string // FileName (Create/CreateNewFile) or FilePath (DeletePath/RenamePath)
	CreateOptions    uint32
	HasCreateOptions bool
	InfoClass        uint32
	HasInfoClass     bool
	IOFlags          uint32 // Write: the IRP flags
	HasIOFlags       bool
}

// maxTrackedDecoyOpens bounds the FileObject→decoy map. Only opens of decoys
// are tracked and Close removes them. Real concurrent opens can reach the
// limit even with no event loss; rejected correlation is explicitly degraded.
const maxTrackedDecoyOpens = 4096

// kernelFileDecoder turns the Kernel-File event stream into FileEvents for
// modifications of decoy files only. It remembers which FileObjects are open on
// a decoy (from Create) so that name-less Write/SetInformation/SetDelete/Rename
// events can be attributed; the acting PID is the one that issued the event.
// Not safe for concurrent use: the ETW consumer delivers events on one thread.
type kernelFileDecoder struct {
	isDecoy func(path string) bool
	opens   map[uint64]string // FileObject → decoy path
	failure error             // sticky missing-correlation coverage, not an ETW event-loss claim
}

func newKernelFileDecoder(isDecoy func(string) bool) *kernelFileDecoder {
	return &kernelFileDecoder{isDecoy: isDecoy, opens: map[uint64]string{}}
}

func (d *kernelFileDecoder) Err() error { return d.failure }

func (d *kernelFileDecoder) decoy(path string) bool {
	return path != "" && d.isDecoy != nil && d.isDecoy(path)
}

func (d *kernelFileDecoder) decode(r kernelFileRecord) (FileEvent, bool) {
	switch r.EventID {
	case kfCreate, kfCreateNewFile:
		// A FileObject address is reused once freed. Any new open replaces what
		// we knew about the address, even if the old Close was lost.
		if r.FileObject != 0 {
			delete(d.opens, r.FileObject)
		}
		if !d.decoy(r.Path) {
			return FileEvent{}, false
		}
		if r.FileObject != 0 {
			if len(d.opens) >= maxTrackedDecoyOpens {
				if d.failure == nil {
					d.failure = errors.New("ransomware Kernel-File decoy handle correlation capacity exhausted")
				}
			} else {
				d.opens[r.FileObject] = r.Path
			}
		}
		// CreateNewFile is only tracked: a file newly created at a decoy's path
		// means the decoy was already deleted or renamed, which scored then.
		if r.EventID == kfCreate && r.HasCreateOptions {
			if op, ok := createOp(r.CreateOptions); ok {
				return FileEvent{PID: r.PID, Path: r.Path, Op: op}, true
			}
		}
		return FileEvent{}, false
	case kfClose:
		delete(d.opens, r.FileObject)
		return FileEvent{}, false
	case kfWrite:
		// A paging write is the cache or memory manager writing back data a
		// process already wrote; that process's own write was a separate,
		// non-paging Write event. The paging write runs in whichever thread
		// forced the flush: the lazy writer (System) or, before a non-cached
		// read, the reader itself. Blaming its PID would kill that reader.
		if r.HasIOFlags && r.IOFlags&irpPagingIO != 0 {
			return FileEvent{}, false
		}
		return d.onOpenDecoy(r, OpWrite)
	case kfSetInformation:
		if !r.HasInfoClass {
			return FileEvent{}, false
		}
		op, ok := setInfoOp(r.InfoClass)
		if !ok {
			return FileEvent{}, false
		}
		return d.onOpenDecoy(r, op)
	case kfSetDelete:
		return d.onOpenDecoy(r, OpDelete)
	case kfRename:
		return d.onOpenDecoy(r, OpRename)
	case kfDeletePath, kfRenamePath:
		op := OpDelete
		if r.EventID == kfRenamePath {
			op = OpRename
		}
		if d.decoy(r.Path) {
			return FileEvent{PID: r.PID, Path: r.Path, Op: op}, true
		}
		return d.onOpenDecoy(r, op)
	}
	return FileEvent{}, false
}

// onOpenDecoy emits op when the event's FileObject is an open decoy.
func (d *kernelFileDecoder) onOpenDecoy(r kernelFileRecord, op FileOp) (FileEvent, bool) {
	if r.FileObject == 0 {
		return FileEvent{}, false
	}
	path, ok := d.opens[r.FileObject]
	if !ok {
		return FileEvent{}, false
	}
	return FileEvent{PID: r.PID, Path: path, Op: op}, true
}

// parseETWUint parses an ETW property rendered by TdhFormatProperty: pointers
// come as "0x…" hex, integers as decimal.
func parseETWUint(s string) (uint64, bool) {
	s = strings.TrimSpace(s)
	base := 10
	if len(s) > 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s, base = s[2:], 16
	}
	v, err := strconv.ParseUint(s, base, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
