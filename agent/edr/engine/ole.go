package engine

import (
	"github.com/utmstack/UTMStack/agent/edr/signatures"

	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
)

// VBA compression is implemented from Microsoft's specification:
// https://learn.microsoft.com/en-us/openspecs/office_file_formats/ms-ovba/4742b896-b32b-4eb0-8372-fbf01e3c65fd
// The compound-file reader (readOLE below) follows the reference's reader as
// learned from its outputs (docs/edr/native-engine/parity/cfb.md), not the
// format specification: it reads lazily and keeps whatever it can.
const (
	oleFree  = uint32(0xffffffff)
	oleEnd   = uint32(0xfffffffe)
	oleFAT   = uint32(0xfffffffd)
	oleDIFAT = uint32(0xfffffffc)
	// oleChainLimit bounds the sectors of one cached chain (directory, mini
	// stream, mini FAT, FAT extension) before it repeats; a file that needs
	// more is left incomplete.
	oleChainLimit = 1 << 22
)

var oleMagic = []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}

// oleStream is a stream entry the reference reads.
type oleStream struct {
	name        string // the reference's name for it (oleName)
	key         string // its own name in upper case, for macro project lookups
	parent      uint32 // the entry whose child chain holds it
	start, size uint32
	nameLength  int
}

type oleFile struct {
	s          *scanState
	data       []byte
	shift      uint
	big, small int64
	maxBlock   int64
	fatRefused bool
	difat      [109]uint32
	cutoff     uint32
	rootKnown  bool
	xbat       oleChain
	dir        oleChain
	root       oleChain
	miniFAT    oleChain
	seen       map[uint32]struct{}
	streams    []oleStream
	encrypted  bool
	limited    bool  // a chain went past oleChainLimit
	undecided  error // the reference's reading cannot be told
}

// ole scans every stream the reference reads, using the caller's global
// scan budgets. Packaged OLE1 objects and PowerPoint embedded storages are
// unwrapped, and the encryption markers of the binary Office formats leave
// it incomplete. When the file holds a _VBA_PROJECT stream, every storage
// with a dir stream is a macro project: its dump and its modules' source are
// matched as the reference matches them (vbaproject.go).
func (s *scanState) ole(data []byte, depth int) (string, error) {
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	if int64(len(data)) > s.maxFile {
		return "", ErrLimit
	}
	f, err := readOLE(s, data)
	if err != nil {
		return "", err
	}
	incomplete := f.undecided
	if f.encrypted {
		incomplete = ErrUnsupported
	}
	// The reference reads macro projects only in a file with a _VBA_PROJECT
	// stream; their streams are kept for the project pass.
	vbaFound := false
	projects := make(map[uint32]map[string][]byte)
	for _, st := range f.streams {
		switch st.name {
		case "_vba_project":
			vbaFound = true
		case "dir":
			projects[st.parent] = make(map[string][]byte)
		}
	}
	if !vbaFound {
		clear(projects)
	}
	for _, st := range f.streams {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if s.files >= s.maxFiles {
			incomplete = ErrLimit
			break
		}
		b, err := f.read(st)
		if err != nil {
			incomplete = err
			continue
		}
		if project := projects[st.parent]; project != nil {
			project[st.key] = b
		}
		if err := oleEncryption(st.name, b); err != nil {
			incomplete = err
		}
		sig, err := s.visit(b, depth+1)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
		switch st.name {
		case "_1_ole10native":
			sig, err = s.ole10Native(b, depth+1)
		case "powerpoint document":
			sig, err = s.pptStorages(b, depth+1)
		case "worddocument":
			sig, err = s.word6Macros(b, depth+1)
		case "workbook":
			sig, err = s.workbookPictures(b, depth+1)
		}
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	if f.limited {
		incomplete = ErrLimit
	}
	storages := make([]uint32, 0, len(projects))
	for id := range projects {
		storages = append(storages, id)
	}
	slices.Sort(storages)
	for _, id := range storages {
		sig, err := s.vbaProject(projects[id], depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// vbaProject matches one macro project: the reference's dump of it, a
// SCRIPT layer matched as a file against root 0, and each module's source,
// matched as the reference's module reader matches it. That reader finds
// modules through the _VBA_PROJECT cache; this engine takes them from the
// dir stream instead (the same modules and offsets in 766 of 767 modules of
// real files). Where the reference's decompressor garbles a module, the
// source as Office reads it is matched as well.
func (s *scanState) vbaProject(streams map[string][]byte, depth int) (string, error) {
	if depth+1 > s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	cache, ok := streams["_VBA_PROJECT"]
	if !ok {
		// A project storage without its own _VBA_PROJECT stream was not
		// observed.
		return "", unknownRepresentation(signatures.UnknownVBA)
	}
	// All decompression work, kept or discarded, and the dump are charged.
	budget := min(s.maxFile, s.remaining)
	initial := budget
	defer func() { s.spend(int(initial - max(budget, 0))) }()
	dump, err := vbaProjectDump(s.ctx, streams["DIR"], streams, &budget)
	switch {
	case errors.Is(err, errVBADumpUnknown):
		incomplete = unknownRepresentation(signatures.UnknownVBA)
	case err != nil:
		return "", err
	case dump.scanned:
		s.layers.Push(signatures.Layer{Type: signatures.TypeScript, Size: int64(len(dump.text))})
		res, err := s.db.MatchObject(s.ctx, dump.text, signatures.MatchContext{Type: signatures.TypeScript, Mode: signatures.FileMode, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth + 1), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
		s.layers.Pop()
		if res.Name != "" {
			return res.Name, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	// Module source. Without a compiled cache (seven bytes or fewer) the
	// reference's module reader finds nothing, so the source matched here is
	// a superset.
	decoded, err := decompressVBA(s.ctx, streams["DIR"], max(budget, 0))
	budget -= int64(len(decoded))
	var project *oleVBAProject
	if err == nil {
		project, err = parseVBADirectory(s.ctx, decoded, s.maxFiles-s.files)
	}
	switch {
	case s.ctx.Err() != nil:
		return "", s.ctx.Err()
	case errors.Is(err, ErrLimit):
		return "", err
	case err != nil && len(cache) > 7:
		return "", unknownRepresentation(signatures.UnknownVBA)
	case err != nil:
		return "", incomplete
	}
	names := make([]string, 0, len(project.modules))
	for name := range project.modules {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		stream, offset := streams[name], project.modules[name]
		if uint64(offset) >= uint64(len(stream)) {
			continue
		}
		text, err := vbaInflate(s.ctx, stream[offset:], &budget)
		switch {
		case errors.Is(err, errVBAInflate):
		case err != nil:
			return "", err
		default:
			if sig, err := s.vbaModule(text, depth); sig != "" {
				return sig, nil
			} else if err != nil {
				incomplete = err
			}
		}
		source, err := decompressVBA(s.ctx, stream[offset:], max(budget, 0))
		budget -= int64(len(source))
		switch {
		case s.ctx.Err() != nil:
			return "", s.ctx.Err()
		case errors.Is(err, ErrLimit):
			return "", err
		case err != nil || bytes.Equal(source, text):
		default:
			if sig, err := s.vbaModule(source, depth); sig != "" {
				return sig, nil
			} else if err != nil {
				incomplete = err
			}
		}
	}
	return "", incomplete
}

// vbaModule matches a macro module's source the way the reference's module
// reader does: as a buffer (no offsets from its end, no hashes) against the
// OLE2 root and root 0, on a layer of its own typed OLE2 for logical rules.
func (s *scanState) vbaModule(source []byte, depth int) (string, error) {
	if len(source) == 0 {
		return "", nil
	}
	s.layers.Push(signatures.Layer{Type: signatures.TypeMSOLE2, Size: int64(len(source))})
	defer s.layers.Pop()
	res, err := s.db.MatchObject(s.ctx, source, signatures.MatchContext{Type: signatures.TypeMSOLE2, Mode: signatures.BufferMode, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth + 1), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
	return res.Name, err
}

// workbookPictures scans the pictures of a workbook's drawing group, as the
// reference does (biffpictures.go).
func (s *scanState) workbookPictures(b []byte, depth int) (string, error) {
	pictures, decided := biffPictures(b)
	var incomplete error
	if !decided {
		incomplete = unknownRepresentation(signatures.UnknownWorkbookPictures)
	}
	for _, p := range pictures {
		sig, err := s.visit(p, depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// oleEncryption recognizes, by the reference's stream name (oleName), the
// encryption markers of the binary Office formats, whose protected content cannot be decoded here:
// the Word FIB fEncrypted flag (MS-DOC 2.5.2), a FilePass record in the
// workbook globals (MS-XLS 2.4.117) and the encrypted CurrentUserAtom token
// (MS-PPT 2.3.2).
func oleEncryption(name string, b []byte) error {
	switch name {
	case "worddocument":
		if len(b) >= 12 && binary.LittleEndian.Uint16(b[10:12])&0x0100 != 0 {
			return ErrUnsupported
		}
	case "workbook", "book":
		return biffEncryption(b)
	case "current user":
		if len(b) >= 16 && binary.LittleEndian.Uint16(b[2:4]) == 0x0ff6 && binary.LittleEndian.Uint32(b[12:16]) == 0xf3d1c4df {
			return ErrUnsupported
		}
	}
	return nil
}

// biffEncryption walks the workbook globals substream, from its BOF record to
// its EOF record, for a FilePass record, which follows the BOF record of an
// encrypted workbook. Record headers stay in plain text when record data is
// encrypted. A stream that is not a workbook or whose records break off
// before a FilePass record shows no encryption; the reference reads such
// streams as they are.
func biffEncryption(b []byte) error {
	const bof, eof, filePass = 0x0809, 0x000a, 0x002f
	if len(b) < 4 || binary.LittleEndian.Uint16(b) != bof {
		return nil
	}
	for pos := 0; len(b)-pos >= 4; {
		typ, size := binary.LittleEndian.Uint16(b[pos:]), int(binary.LittleEndian.Uint16(b[pos+2:]))
		switch {
		case typ == filePass:
			return ErrUnsupported
		case typ == eof || size > len(b)-pos-4:
			return nil
		}
		pos += 4 + size
	}
	return nil
}

// ole10Native unwraps the OLE1 packager layout of an "\x01Ole10Native"
// stream: total size, flags, NUL-terminated label and source path, a
// reserved word, a length-prefixed NUL-terminated temporary path, then the
// embedded file's length and data. Newer packagers append Unicode copies of
// the names after the data. Native data of any other layout is scanned from
// after the size field but remains incomplete, as do truncated data.
func (s *scanState) ole10Native(b []byte, depth int) (string, error) {
	if len(b) < 4 {
		return "", ErrMalformed
	}
	native, truncated := b[4:], true
	if total := binary.LittleEndian.Uint32(b); uint64(total) <= uint64(len(native)) {
		native, truncated = native[:total], false
	}
	payload, packaged, complete := olePackage(native)
	if !packaged {
		payload = native
	}
	if sig, err := s.visit(payload, depth+1); sig != "" || err != nil {
		return sig, err
	}
	switch {
	case truncated || packaged && !complete:
		return "", ErrMalformed
	case !packaged:
		return "", ErrUnsupported
	}
	return "", nil
}

// olePackage reports whether b has the packager header, and whether its
// declared data length fits; the payload is truncated to b otherwise.
func olePackage(b []byte) (payload []byte, packaged, complete bool) {
	const maxName = 4096
	pos := 2 // flags
	for i := 0; i < 2; i++ {
		if pos > len(b) {
			return nil, false, false
		}
		n := bytes.IndexByte(b[pos:min(len(b), pos+maxName+1)], 0)
		if n < 0 {
			return nil, false, false
		}
		pos += n + 1
	}
	pos += 4 // reserved
	if pos > len(b) || len(b)-pos < 4 {
		return nil, false, false
	}
	temp := binary.LittleEndian.Uint32(b[pos:])
	pos += 4
	if temp == 0 || temp > maxName || uint64(temp) > uint64(len(b)-pos) || bytes.IndexByte(b[pos:pos+int(temp)], 0) != int(temp)-1 {
		return nil, false, false
	}
	pos += int(temp)
	if len(b)-pos < 4 {
		return nil, false, false
	}
	size := binary.LittleEndian.Uint32(b[pos:])
	pos += 4
	if uint64(size) > uint64(len(b)-pos) {
		return b[pos:], true, false
	}
	return b[pos : pos+int(size)], true, true
}

// pptStorages walks the top-level records of a "PowerPoint Document" stream
// as the reference does (docs/2026-10-14-cfb-oracle/pptlab.py): the walk ends
// with fewer than 8 bytes left or at a record that runs past the stream, and
// a record of length 0 anywhere fails it, so that nothing of the stream is
// extracted, not even objects found before it. Every ExOleObjStg record
// (MS-PPT 2.10.34), whatever its version and instance, is read as a 32-bit
// size, which the reference does not use, and zlib data
// (pptCompressedStorage). A CryptSession10Container marks an encrypted
// document.
func (s *scanState) pptStorages(b []byte, depth int) (string, error) {
	const exOleObjStg, cryptSession = 0x1011, 0x2f14
	var incomplete error
	var storages [][]byte
	for pos, records := 0, 0; len(b)-pos >= 8; records++ {
		if records&4095 == 0 {
			if err := s.ctx.Err(); err != nil {
				return "", err
			}
		}
		typ := binary.LittleEndian.Uint16(b[pos+2:])
		size := binary.LittleEndian.Uint32(b[pos+4:])
		if size == 0 {
			return "", incomplete
		}
		if uint64(size) > uint64(len(b)-pos-8) {
			break
		}
		body := b[pos+8 : pos+8+int(size)]
		pos += 8 + int(size)
		switch typ {
		case cryptSession:
			incomplete = ErrUnsupported
		case exOleObjStg:
			storages = append(storages, body)
		}
	}
	for _, body := range storages {
		sig, err := s.pptCompressedStorage(body, depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// pptCompressedStorage analyzes whatever the zlib data after the record's
// size field decodes to, as the reference does: also when the data is
// damaged, ends early or decodes to more or fewer bytes than declared. Data
// without a zlib header yields nothing.
func (s *scanState) pptCompressedStorage(b []byte, depth int) (string, error) {
	if len(b) < 4 {
		return "", nil
	}
	r, err := zlib.NewReader(bytes.NewReader(b[4:]))
	if err != nil {
		return "", nil
	}
	limit := min(s.maxFile, s.remaining)
	if limit < 0 {
		return "", ErrLimit
	}
	out, _ := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, limit+1))
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if int64(len(out)) > limit {
		s.spend(len(out))
		return "", ErrLimit
	}
	return s.visit(out, depth+1)
}

// readOLE reads a compound file's directory as the reference does:
//   - Fewer than 517 bytes, two nonzero bytes at offset 34, a mini sector
//     shift of 0 or above the sector shift, or a sector shift above 28: the
//     reference reads nothing. Sector shifts below 9 are left undecided.
//     No other header field is checked; the declared FAT, mini FAT, extension
//     and directory counts are not used.
//   - Block n (2^shift bytes) is at (n+1)<<shift. A block the file ends
//     inside is zero-filled; one at or past the end cannot be read. A block
//     number above (file size - sector size) / mini sector size, in either
//     kind of chain, ends the chain.
//   - FAT entries come from the header's 109 FAT sector numbers, then from
//     the extension chain; a negative FAT count makes every lookup fail.
//   - Directory entry n is the (n%4)-th 128-byte entry of the (n/4)-th
//     directory sector, in 4096-byte sectors too, where entries 4-31 of
//     each sector are therefore never read.
//   - The tree is walked from entry 0: an entry, then its child chain one
//     level deeper, then its right and left links. An unreadable link, a
//     negative one, an unused entry, a type other than storage or stream
//     (or root, at entry 0 only) skip that link. Coming back to an entry
//     already walked, or an entry deeper than the recursion limit, ends the
//     whole walk. Only entry 0 as root gives the mini stream its start; any
//     entry's child chain is walked, a stream's too.
//   - A stream is read from the mini stream when its 32-bit size is below
//     the header's cutoff, from big blocks otherwise (stream reading below).
//   - When the walk finds a macro project, a Word or PowerPoint document
//     stream or an OLE1 native stream (or Excel 4 macros or images in a
//     workbook), the reference writes streams to files by name and skips
//     names longer than 64 bytes; otherwise it keeps them.
func readOLE(s *scanState, data []byte) (*oleFile, error) {
	f := &oleFile{s: s, data: data, seen: make(map[uint32]struct{})}
	if len(data) < 517 || !bytes.HasPrefix(data, oleMagic) {
		return f, nil
	}
	u16 := func(off int) uint16 { return binary.LittleEndian.Uint16(data[off:]) }
	u32 := func(off int) uint32 { return binary.LittleEndian.Uint32(data[off:]) }
	shift, mshift := uint(u16(30)), uint(u16(32))
	if u16(34) != 0 || mshift == 0 || mshift > shift || shift > 28 {
		return f, nil
	}
	if shift < 9 {
		f.undecided = unknownRepresentation(signatures.UnknownCompoundFile)
		return f, nil
	}
	f.shift = shift
	f.big, f.small = 1<<shift, 1<<mshift
	f.maxBlock = (int64(len(data)) - f.big) / f.small
	f.fatRefused = int32(u32(44)) < 0
	f.cutoff = u32(56)
	for i := range f.difat {
		f.difat[i] = u32(76 + i*4)
	}
	f.dir = oleChain{start: u32(48)}
	f.miniFAT = oleChain{start: u32(60)}
	f.xbat = oleChain{start: u32(68), xbat: true}
	if err := f.walk(); err != nil {
		return nil, err
	}
	folder, workbook := false, false
	for _, st := range f.streams {
		switch st.name {
		case "_vba_project", "worddocument", "powerpoint document", "_1_ole10native":
			folder = true
		case "workbook", "book":
			workbook = true
		}
	}
	kept := f.streams[:0]
	for _, st := range f.streams {
		if st.nameLength > 64 {
			if folder {
				continue
			}
			if workbook {
				// Kept or skipped depending on the reference's workbook
				// scan, which this engine does not reproduce.
				f.undecided = unknownRepresentation(signatures.UnknownCompoundFile)
				continue
			}
		}
		kept = append(kept, st)
	}
	f.streams = kept
	return f, nil
}

func allZeroOLE(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// block returns big block n, zero-filled when the file ends inside it.
func (f *oleFile) block(n uint32) ([]byte, bool) {
	off := (int64(n) + 1) << f.shift
	if off >= int64(len(f.data)) {
		return nil, false
	}
	if end := off + f.big; end <= int64(len(f.data)) {
		return f.data[off:end], true
	}
	b := make([]byte, f.big)
	copy(b, f.data[off:])
	return b, true
}

// next returns the FAT entry of big block n.
func (f *oleFile) next(n uint32) (uint32, bool) {
	if f.fatRefused {
		return 0, false
	}
	per := uint32(f.big / 4)
	index := n / per
	sector := uint32(0)
	if index < 109 {
		sector = f.difat[index]
	} else {
		k, entries := index-109, per-1
		x, ok := f.step(&f.xbat, k/entries)
		if !ok {
			return 0, false
		}
		b, ok := f.block(x)
		if !ok {
			return 0, false
		}
		sector = binary.LittleEndian.Uint32(b[(k%entries)*4:])
	}
	b, ok := f.block(sector)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[(n%per)*4:]), true
}

// oleChain caches the sectors of a chain the reference walks step by step
// (k steps from its start for the k-th sector): the directory, the mini
// stream, the mini FAT and the FAT extension, whose next sector is the last
// word of each sector. A chain that comes back to a sector repeats.
type oleChain struct {
	start   uint32
	xbat    bool
	sectors []uint32
	index   map[uint32]int
	loop    int
	failed  bool
}

func (f *oleFile) step(c *oleChain, k uint32) (uint32, bool) {
	if c.sectors == nil {
		c.sectors, c.index, c.loop = []uint32{c.start}, map[uint32]int{c.start: 0}, -1
	}
	for uint64(len(c.sectors)) <= uint64(k) && c.loop < 0 && !c.failed {
		if len(c.sectors) >= oleChainLimit || len(c.sectors)&4095 == 0 && f.s.ctx.Err() != nil {
			f.limited = true
			return 0, false
		}
		last := c.sectors[len(c.sectors)-1]
		var next uint32
		ok := false
		if c.xbat {
			var b []byte
			if b, ok = f.block(last); ok {
				next = binary.LittleEndian.Uint32(b[f.big-4:])
			}
		} else {
			next, ok = f.next(last)
		}
		if !ok {
			c.failed = true
			break
		}
		if at, again := c.index[next]; again {
			c.loop = at
			break
		}
		c.index[next] = len(c.sectors)
		c.sectors = append(c.sectors, next)
	}
	switch {
	case uint64(k) < uint64(len(c.sectors)):
		return c.sectors[k], true
	case c.loop >= 0:
		period := uint32(len(c.sectors) - c.loop)
		return c.sectors[uint32(c.loop)+(k-uint32(c.loop))%period], true
	}
	return 0, false
}

func (f *oleFile) smallBlock(n uint32) ([]byte, bool) {
	if !f.rootKnown {
		return nil, false
	}
	per := uint32(f.big / f.small)
	sector, ok := f.step(&f.root, n/per)
	if !ok {
		return nil, false
	}
	b, ok := f.block(sector)
	if !ok {
		return nil, false
	}
	off := int64(n%per) * f.small
	return b[off : off+f.small], true
}

func (f *oleFile) smallNext(n uint32) (uint32, bool) {
	per := uint32(f.big / 4)
	sector, ok := f.step(&f.miniFAT, n/per)
	if !ok {
		return 0, false
	}
	b, ok := f.block(sector)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[(n%per)*4:]), true
}

func (f *oleFile) entry(n uint32) ([]byte, bool) {
	sector, ok := f.step(&f.dir, n/4)
	if !ok {
		return nil, false
	}
	b, ok := f.block(sector)
	if !ok {
		return nil, false
	}
	return b[(n%4)*128 : (n%4)*128+128], true
}

// read returns a stream as the reference reads it: blocks follow the chain
// until the size is reached, and a free, end or other negative entry, a
// block number past the bound, a block already read for this stream or an
// unreadable block ends it; what was read is kept. A size above the
// remaining scan budget is a limit, as in the reference.
func (f *oleFile) read(st oleStream) ([]byte, error) {
	if int64(st.size) > f.s.remaining {
		return nil, ErrLimit
	}
	small := st.size < f.cutoff
	clear(f.seen)
	var out []byte
	for cur := st.start; uint64(len(out)) < uint64(st.size); {
		if len(f.seen)&4095 == 4095 {
			if err := f.s.ctx.Err(); err != nil {
				return nil, err
			}
		}
		if int32(cur) < 0 || int64(cur) > f.maxBlock {
			break
		}
		if _, again := f.seen[cur]; again {
			break
		}
		f.seen[cur] = struct{}{}
		var b []byte
		var ok bool
		if small {
			b, ok = f.smallBlock(cur)
		} else {
			b, ok = f.block(cur)
		}
		if !ok {
			break
		}
		out = append(out, b...)
		if small {
			cur, ok = f.smallNext(cur)
		} else {
			cur, ok = f.next(cur)
		}
		if !ok {
			break
		}
	}
	if uint64(len(out)) > uint64(st.size) {
		out = out[:st.size]
	}
	return out, nil
}

func (f *oleFile) walk() error {
	type pending struct {
		id, parent uint32
		level      int
	}
	walked := make(map[uint32]struct{})
	stack := []pending{{0, 0, 0}}
	for len(stack) > 0 {
		if len(walked)&4095 == 4095 {
			if err := f.s.ctx.Err(); err != nil {
				return err
			}
		}
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if int32(p.id) < 0 {
			continue
		}
		if _, again := walked[p.id]; again {
			return nil
		}
		e, ok := f.entry(p.id)
		if !ok {
			continue
		}
		walked[p.id] = struct{}{}
		if p.level > f.s.maxDepth {
			// The reference stops here too, and reports the limit.
			f.limited = true
			return nil
		}
		switch e[66] {
		case 1:
		case 2:
			name, key := oleName(e)
			f.streams = append(f.streams, oleStream{name: name, key: key, parent: p.parent,
				start: binary.LittleEndian.Uint32(e[116:]), size: binary.LittleEndian.Uint32(e[120:]),
				nameLength: int(binary.LittleEndian.Uint16(e[64:]))})
		case 5:
			if p.id != 0 {
				continue
			}
			f.root = oleChain{start: binary.LittleEndian.Uint32(e[116:])}
			f.rootKnown = true
		default:
			continue
		}
		switch name, _ := oleName(e); name {
		case "encryptedpackage", "encryptioninfo", "encryptedsummary", "_6_dataspaces", "dataspaces":
			f.encrypted = true
		}
		stack = append(stack, pending{binary.LittleEndian.Uint32(e[68:]), p.parent, p.level},
			pending{binary.LittleEndian.Uint32(e[72:]), p.parent, p.level},
			pending{binary.LittleEndian.Uint32(e[76:]), p.id, p.level + 1})
	}
	return nil
}

// oleName returns the reference's name for a directory entry and the
// entry's own name in upper case. The name has (length-1)/2 UTF-16 units
// from the length field, none when it is below 2 or above 128 (units past
// the name area come from the rest of the entry). The reference lowercases
// printable ASCII, writes control characters as _N_ and other units u as
// three letters, a+(u>>8&15), a+(u>>12&15), a+(u&15), between underscores.
func oleName(e []byte) (string, string) {
	length := int(binary.LittleEndian.Uint16(e[64:]))
	if length < 2 || length > 128 {
		return "", ""
	}
	units := make([]uint16, (length-1)/2)
	var name strings.Builder
	for i := range units {
		u := binary.LittleEndian.Uint16(e[i*2:])
		units[i] = u
		switch {
		case u < 0x20:
			name.WriteString("_" + strconv.Itoa(int(u)) + "_")
		case u < 0x7f:
			if u >= 'A' && u <= 'Z' {
				u += 'a' - 'A'
			}
			name.WriteByte(byte(u))
		default:
			name.Write([]byte{'_', byte('a' + u>>8&15), byte('a' + u>>12&15), byte('a' + u&15), '_'})
		}
	}
	return name.String(), strings.ToUpper(string(utf16.Decode(units)))
}

func decompressVBA(ctx context.Context, data []byte, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, ErrLimit
	}
	work := limit + int64(len(data))
	if work < limit {
		return nil, ErrLimit
	}
	return decompressVBAWork(ctx, data, limit, &work)
}

func decompressVBAWork(ctx context.Context, data []byte, limit int64, work *int64) ([]byte, error) {
	if len(data) < 4 || data[0] != 1 {
		return nil, ErrMalformed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Errors return the output decoded so far, which the caller charges.
	var out []byte
	for pos := 1; pos < len(data); {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if len(data)-pos < 2 {
			return out, ErrMalformed
		}
		h := binary.LittleEndian.Uint16(data[pos:])
		if h&0x7000 != 0x3000 {
			return out, ErrMalformed
		}
		size := int(h&0x0fff) + 3
		if size > len(data)-pos {
			return out, ErrMalformed
		}
		end, start := pos+size, len(out)
		pos += 2
		*work -= int64(size)
		if *work < 0 {
			return out, ErrLimit
		}
		if h&0x8000 == 0 {
			if size != 4098 {
				return out, ErrMalformed
			}
			if int64(len(out))+4096 > limit {
				return out, ErrLimit
			}
			*work -= 4096
			if *work < 0 {
				return out, ErrLimit
			}
			out = append(out, data[pos:end]...)
			pos = end
			continue
		}
		for pos < end {
			flags := data[pos]
			pos++
			for bit := uint(0); bit < 8 && pos < end; bit++ {
				if flags&(1<<bit) == 0 {
					if len(out)-start >= 4096 {
						return out, ErrMalformed
					}
					if int64(len(out)) >= limit {
						return out, ErrLimit
					}
					*work--
					if *work < 0 {
						return out, ErrLimit
					}
					out = append(out, data[pos])
					pos++
					continue
				}
				if end-pos < 2 || len(out) == start {
					return out, ErrMalformed
				}
				token := binary.LittleEndian.Uint16(data[pos:])
				pos += 2
				bitCount := max(4, bits.Len(uint(len(out)-start-1)))
				lengthMask := uint16(0xffff >> bitCount)
				length, offset := int(token&lengthMask)+3, int(token>>(16-bitCount))+1
				if offset > len(out)-start || len(out)-start+length > 4096 {
					return out, ErrMalformed
				}
				if int64(len(out))+int64(length) > limit {
					return out, ErrLimit
				}
				*work -= int64(length)
				if *work < 0 {
					return out, ErrLimit
				}
				for j := 0; j < length; j++ {
					out = append(out, out[len(out)-offset])
				}
			}
		}
		if len(out) == start {
			return out, ErrMalformed
		}
	}
	return out, nil
}
