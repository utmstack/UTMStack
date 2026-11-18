// Package sevenzip reads 7-Zip archives for the scanner: the archive and
// header layout of the public 7z format description (7zFormat.txt in the
// 7-Zip documentation) and the coders the reference scanner decodes (Copy,
// LZMA, LZMA2, PPMd, the BCJ, BCJ2 and branch filters). Other coders and
// encrypted data are reported as such; their members are not extracted, as
// in the reference.
package sevenzip

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"unicode/utf16"
)

var (
	// ErrFormat: the archive or one of its headers cannot be read.
	ErrFormat = errors.New("7z: malformed archive")
	// ErrUnsupported: a coder this reader (and the reference) does not
	// decode.
	ErrUnsupported = errors.New("7z: unsupported coder")
	// ErrEncrypted: encrypted data or headers.
	ErrEncrypted = errors.New("7z: encrypted")
	// ErrData: a coder's data is corrupt.
	ErrData = errors.New("7z: corrupt data")
	// ErrLimit: a size beyond the caller's limit.
	ErrLimit = errors.New("7z: size limit")
	// ErrAborted: the caller's Done channel closed during decoding.
	ErrAborted = errors.New("7z: decoding stopped")
	// ErrNothing: the reference reads no member of such an archive: a cut
	// start header, a major version other than 0, a header past the end of
	// the data or a header whose CRC does not match.
	ErrNothing = errors.New("7z: archive the reference does not read")
)

// Signature starts every 7z archive.
var Signature = []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c}

const (
	idEnd = iota
	idHeader
	idArchiveProperties
	idAdditionalStreamsInfo
	idMainStreamsInfo
	idFilesInfo
	idPackInfo
	idUnpackInfo
	idSubStreamsInfo
	idSize
	idCRC
	idFolder
	idCodersUnpackSize
	idNumUnpackStream
	idEmptyStream
	idEmptyFile
	idAnti
	idName
	idCTime
	idATime
	idMTime
	idWinAttributes
	idComment
	idEncodedHeader
	idStartPos
	idDummy
)

// Coder is one coder of a folder.
type Coder struct {
	ID            []byte
	NumIn, NumOut int
	Props         []byte
}

// Folder is a coder graph whose final output holds one or more files.
type Folder struct {
	Coders      []Coder
	BindPairs   [][2]int // in index, out index
	Packed      []int    // in indices fed by pack streams, in pack order
	UnpackSizes []uint64 // per out stream
	CRC         uint32
	HasCRC      bool
	packStart   int // first pack stream of this folder
}

// File is one archive entry.
type File struct {
	Name      string
	HasStream bool
	IsDir     bool
	Size      uint64
	CRC       uint32
	HasCRC    bool
	Folder    int    // -1 without a stream
	Offset    uint64 // within the folder's output
}

// Archive is a parsed archive.
type Archive struct {
	data      []byte
	packPos   uint64
	packSizes []uint64
	Folders   []Folder
	Files     []File
	limit     uint64
	done      <-chan struct{}
}

type reader struct {
	b   []byte
	pos int
	bad bool
}

func (r *reader) byte() byte {
	if r.pos >= len(r.b) {
		r.bad = true
		return 0
	}
	v := r.b[r.pos]
	r.pos++
	return v
}

func (r *reader) bytes(n uint64) []byte {
	if n > uint64(len(r.b)-r.pos) {
		r.bad = true
		r.pos = len(r.b)
		return nil
	}
	v := r.b[r.pos : r.pos+int(n)]
	r.pos += int(n)
	return v
}

func (r *reader) number() uint64 {
	first := r.byte()
	mask := byte(0x80)
	var v uint64
	for i := 0; i < 8; i++ {
		if first&mask == 0 {
			return v | (uint64(first)&(uint64(mask)-1))<<(8*i)
		}
		v |= uint64(r.byte()) << (8 * i)
		mask >>= 1
	}
	return v
}

// count reads a number used as a count, at most max.
func (r *reader) count(max int) int {
	v := r.number()
	if v > uint64(max) {
		r.bad = true
		return 0
	}
	return int(v)
}

func (r *reader) uint32() uint32 {
	b := r.bytes(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func (r *reader) bits(n int) []bool {
	out := make([]bool, n)
	var b byte
	for i := range out {
		if i%8 == 0 {
			b = r.byte()
		}
		out[i] = b&(0x80>>(i%8)) != 0
	}
	return out
}

// defined reads an all-defined flag or a bit vector of n entries.
func (r *reader) defined(n int) []bool {
	if r.byte() != 0 {
		out := make([]bool, n)
		for i := range out {
			out[i] = true
		}
		return out
	}
	return r.bits(n)
}

// Limits bound what the reader allocates and decodes.
type Limits struct {
	MaxFiles  int
	MaxUnpack uint64 // per folder output and in total for headers
	// Done, when closed, stops decoding with ErrAborted.
	Done <-chan struct{}
}

// Open parses the archive in data (starting at its signature).
func Open(data []byte, lim Limits) (*Archive, error) {
	if !bytes.HasPrefix(data, Signature) {
		return nil, ErrFormat
	}
	if len(data) < 32 {
		return nil, ErrNothing // a cut start header
	}
	// The start header's own CRC and the minor version are not checked,
	// as in the reference.
	if data[6] != 0 {
		return nil, ErrNothing
	}
	le := binary.LittleEndian
	next, size := le.Uint64(data[12:]), le.Uint64(data[20:])
	if next > uint64(len(data))-32 || size > uint64(len(data))-32-next {
		return nil, ErrNothing
	}
	header := data[32+next : 32+next+size]
	if crc32.ChecksumIEEE(header) != le.Uint32(data[28:]) {
		return nil, ErrNothing
	}
	a := &Archive{data: data, limit: lim.MaxUnpack, done: lim.Done}
	if size == 0 {
		return a, nil
	}
	for round := 0; ; round++ {
		r := &reader{b: header}
		switch r.byte() {
		case idHeader:
			if err := a.readHeader(r, lim); err != nil {
				return nil, err
			}
			return a, nil
		case idEncodedHeader:
			if round > 4 {
				return nil, ErrFormat
			}
			e := &Archive{data: data, limit: lim.MaxUnpack, done: lim.Done}
			if err := e.readStreams(r, lim); err != nil {
				return nil, err
			}
			if len(e.Folders) == 0 {
				return nil, ErrFormat
			}
			out, err := e.Unpack(0)
			if err != nil {
				return nil, err
			}
			header = out
		default:
			return nil, ErrFormat
		}
	}
}

func (a *Archive) readHeader(r *reader, lim Limits) error {
	id := r.byte()
	if id == idArchiveProperties {
		for {
			t := r.byte()
			if t == idEnd || r.bad {
				break
			}
			r.bytes(r.number())
		}
		id = r.byte()
	}
	if id == idAdditionalStreamsInfo {
		// Not used by archives the 7-Zip tools write.
		return ErrFormat
	}
	var sub *subStreams
	if id == idMainStreamsInfo {
		s, err := a.readStreamsInfo(r, lim)
		if err != nil {
			return err
		}
		sub = s
		id = r.byte()
	}
	if id == idFilesInfo {
		if err := a.readFiles(r, lim, sub); err != nil {
			return err
		}
		id = r.byte()
	}
	if id != idEnd || r.bad {
		return ErrFormat
	}
	return nil
}

// readStreams reads the streams info of an encoded header.
func (a *Archive) readStreams(r *reader, lim Limits) error {
	_, err := a.readStreamsInfo(r, lim)
	return err
}

type subStreams struct {
	counts []int
	sizes  []uint64
	crcs   []uint32
	hasCRC []bool
}

func (a *Archive) readStreamsInfo(r *reader, lim Limits) (*subStreams, error) {
	id := r.byte()
	if id == idPackInfo {
		a.packPos = r.number()
		n := r.count(1 << 16)
		id = r.byte()
		if id == idSize {
			a.packSizes = make([]uint64, n)
			for i := range a.packSizes {
				a.packSizes[i] = r.number()
			}
			id = r.byte()
		}
		if id == idCRC {
			for _, d := range r.defined(n) {
				if d {
					r.uint32()
				}
			}
			id = r.byte()
		}
		if id != idEnd {
			return nil, ErrFormat
		}
		id = r.byte()
	}
	if id == idUnpackInfo {
		if r.byte() != idFolder {
			return nil, ErrFormat
		}
		n := r.count(1 << 16)
		if r.byte() != 0 {
			return nil, ErrFormat // folders in another stream
		}
		a.Folders = make([]Folder, n)
		pack := 0
		for i := range a.Folders {
			if err := r.folder(&a.Folders[i]); err != nil {
				return nil, err
			}
			a.Folders[i].packStart = pack
			pack += len(a.Folders[i].Packed)
		}
		if r.byte() != idCodersUnpackSize {
			return nil, ErrFormat
		}
		for i := range a.Folders {
			f := &a.Folders[i]
			outs := 0
			for _, c := range f.Coders {
				outs += c.NumOut
			}
			f.UnpackSizes = make([]uint64, outs)
			for j := range f.UnpackSizes {
				f.UnpackSizes[j] = r.number()
			}
		}
		id = r.byte()
		if id == idCRC {
			for i, d := range r.defined(n) {
				if d {
					a.Folders[i].CRC, a.Folders[i].HasCRC = r.uint32(), true
				}
			}
			id = r.byte()
		}
		if id != idEnd {
			return nil, ErrFormat
		}
		id = r.byte()
		if pack > len(a.packSizes) {
			return nil, ErrFormat
		}
	}
	sub := &subStreams{counts: make([]int, len(a.Folders))}
	for i := range sub.counts {
		sub.counts[i] = 1
	}
	if id == idSubStreamsInfo {
		id = r.byte()
		if id == idNumUnpackStream {
			total := 0
			for i := range sub.counts {
				sub.counts[i] = r.count(lim.MaxFiles + 1)
				total += sub.counts[i]
				if total > lim.MaxFiles+1 {
					return nil, ErrLimit
				}
			}
			id = r.byte()
		}
		for i, f := range a.Folders {
			if sub.counts[i] == 0 {
				continue
			}
			size, _ := f.outSize()
			var sum uint64
			for j := 1; j < sub.counts[i]; j++ {
				if id == idSize {
					v := r.number()
					sub.sizes = append(sub.sizes, v)
					sum += v
				}
			}
			if id != idSize && sub.counts[i] > 1 {
				return nil, ErrFormat
			}
			if sum > size {
				return nil, ErrFormat
			}
			sub.sizes = append(sub.sizes, size-sum)
		}
		if id == idSize {
			id = r.byte()
		}
		// Digests for the streams whose folder CRC does not already give
		// them (folders with exactly one stream and a CRC).
		unknown := 0
		for i, f := range a.Folders {
			if sub.counts[i] != 1 || !f.HasCRC {
				unknown += sub.counts[i]
			}
		}
		for id != idEnd && !r.bad {
			if id == idCRC {
				defined := r.defined(unknown)
				digests := make([]uint32, unknown)
				for i, d := range defined {
					if d {
						digests[i] = r.uint32()
					}
				}
				k := 0
				for i, f := range a.Folders {
					if sub.counts[i] == 1 && f.HasCRC {
						sub.crcs = append(sub.crcs, f.CRC)
						sub.hasCRC = append(sub.hasCRC, true)
						continue
					}
					for j := 0; j < sub.counts[i]; j++ {
						sub.crcs = append(sub.crcs, digests[k])
						sub.hasCRC = append(sub.hasCRC, defined[k])
						k++
					}
				}
			} else {
				r.bytes(r.number())
			}
			id = r.byte()
		}
		id = r.byte()
	} else if len(a.Folders) > 0 {
		for _, f := range a.Folders {
			size, _ := f.outSize()
			sub.sizes = append(sub.sizes, size)
			sub.crcs = append(sub.crcs, f.CRC)
			sub.hasCRC = append(sub.hasCRC, f.HasCRC)
		}
	}
	if id != idEnd || r.bad {
		return nil, ErrFormat
	}
	return sub, nil
}

func (r *reader) folder(f *Folder) error {
	n := r.count(64)
	if n == 0 {
		return ErrFormat
	}
	ins, outs := 0, 0
	for i := 0; i < n; i++ {
		flags := r.byte()
		c := Coder{ID: r.bytes(uint64(flags & 0xf)), NumIn: 1, NumOut: 1}
		if flags&0x10 != 0 {
			c.NumIn, c.NumOut = r.count(64), r.count(64)
		}
		if flags&0x20 != 0 {
			c.Props = r.bytes(r.number())
		}
		if flags&0x80 != 0 || r.bad {
			return ErrFormat // alternative methods are not written by 7-Zip
		}
		ins += c.NumIn
		outs += c.NumOut
		f.Coders = append(f.Coders, c)
	}
	if outs == 0 || ins < outs-1 {
		return ErrFormat
	}
	for i := 0; i < outs-1; i++ {
		in, out := r.count(ins-1), r.count(outs-1)
		f.BindPairs = append(f.BindPairs, [2]int{in, out})
	}
	packed := ins - (outs - 1)
	if packed == 1 {
		for i := 0; i < ins; i++ {
			if f.bindIn(i) < 0 {
				f.Packed = append(f.Packed, i)
				break
			}
		}
		if len(f.Packed) != 1 {
			return ErrFormat
		}
	} else {
		for i := 0; i < packed; i++ {
			f.Packed = append(f.Packed, r.count(ins-1))
		}
	}
	if r.bad {
		return ErrFormat
	}
	return nil
}

// bindIn is the bind pair index feeding in stream i, or -1.
func (f *Folder) bindIn(i int) int {
	for k, bp := range f.BindPairs {
		if bp[0] == i {
			return k
		}
	}
	return -1
}

// mainOut is the out stream no bind pair consumes.
func (f *Folder) mainOut() int {
	outs := len(f.UnpackSizes)
	for o := 0; o < outs; o++ {
		bound := false
		for _, bp := range f.BindPairs {
			if bp[1] == o {
				bound = true
			}
		}
		if !bound {
			return o
		}
	}
	return -1
}

func (f *Folder) outSize() (uint64, bool) {
	o := f.mainOut()
	if o < 0 || o >= len(f.UnpackSizes) {
		return 0, false
	}
	return f.UnpackSizes[o], true
}

func (a *Archive) readFiles(r *reader, lim Limits, sub *subStreams) error {
	n := r.count(lim.MaxFiles + 1)
	if n > lim.MaxFiles {
		return ErrLimit
	}
	a.Files = make([]File, n)
	var empty, emptyFile []bool
	for {
		t := r.byte()
		if t == idEnd || r.bad {
			break
		}
		size := r.number()
		body := r.bytes(size)
		pr := &reader{b: body}
		switch t {
		case idEmptyStream:
			empty = pr.bits(n)
		case idEmptyFile:
			k := 0
			for _, e := range empty {
				if e {
					k++
				}
			}
			emptyFile = pr.bits(k)
		case idName:
			if pr.byte() != 0 {
				return ErrFormat
			}
			rest := pr.b[pr.pos:]
			for i := range a.Files {
				var u []uint16
				for {
					if len(rest) < 2 {
						return ErrFormat
					}
					c := binary.LittleEndian.Uint16(rest)
					rest = rest[2:]
					if c == 0 {
						break
					}
					u = append(u, c)
				}
				a.Files[i].Name = string(utf16.Decode(u))
			}
		}
	}
	if r.bad {
		return ErrFormat
	}
	stream, emptyIndex := 0, 0
	folder, inFolder := 0, 0
	var offset uint64
	for i := range a.Files {
		f := &a.Files[i]
		f.Folder = -1
		if empty != nil && empty[i] {
			f.IsDir = emptyFile == nil || emptyIndex >= len(emptyFile) || !emptyFile[emptyIndex]
			emptyIndex++
			continue
		}
		f.HasStream = true
		for sub == nil || folder < len(sub.counts) && inFolder >= sub.counts[folder] {
			if sub == nil {
				return ErrFormat
			}
			folder++
			inFolder, offset = 0, 0
		}
		if folder >= len(sub.counts) || stream >= len(sub.sizes) {
			return ErrFormat
		}
		f.Folder, f.Offset, f.Size = folder, offset, sub.sizes[stream]
		if stream < len(sub.crcs) {
			f.CRC, f.HasCRC = sub.crcs[stream], sub.hasCRC[stream]
		}
		offset += f.Size
		inFolder++
		stream++
	}
	return nil
}

// packStream returns pack stream i of the archive.
func (a *Archive) packStream(i int) ([]byte, error) {
	if i >= len(a.packSizes) {
		return nil, ErrFormat
	}
	start := 32 + a.packPos
	for _, s := range a.packSizes[:i] {
		start += s
	}
	size := a.packSizes[i]
	if start > uint64(len(a.data)) || size > uint64(len(a.data))-start {
		return nil, ErrFormat
	}
	return a.data[start : start+size], nil
}

// Unpack decodes folder i's main output.
func (a *Archive) Unpack(i int) ([]byte, error) {
	if i >= len(a.Folders) {
		return nil, ErrFormat
	}
	f := &a.Folders[i]
	size, ok := f.outSize()
	if !ok {
		return nil, ErrFormat
	}
	if size > a.limit {
		return nil, ErrLimit
	}
	out, err := a.decodeOut(f, f.mainOut(), 0)
	if err != nil {
		return nil, err
	}
	if uint64(len(out)) != size {
		return nil, ErrData
	}
	return out, nil
}

func (a *Archive) decodeOut(f *Folder, out, depth int) ([]byte, error) {
	if depth > 8 {
		return nil, ErrFormat
	}
	ci, inBase, outBase := -1, 0, 0
	for k, c := range f.Coders {
		if out < outBase+c.NumOut {
			ci = k
			break
		}
		inBase += c.NumIn
		outBase += c.NumOut
	}
	if ci < 0 || f.Coders[ci].NumOut != 1 {
		return nil, ErrUnsupported
	}
	c := f.Coders[ci]
	if out >= len(f.UnpackSizes) || f.UnpackSizes[out] > a.limit {
		return nil, ErrLimit
	}
	inputs := make([][]byte, c.NumIn)
	for j := range inputs {
		in := inBase + j
		if bp := f.bindIn(in); bp >= 0 {
			b, err := a.decodeOut(f, f.BindPairs[bp][1], depth+1)
			if err != nil {
				return nil, err
			}
			inputs[j] = b
			continue
		}
		found := false
		for k, p := range f.Packed {
			if p == in {
				b, err := a.packStream(f.packStart + k)
				if err != nil {
					return nil, err
				}
				inputs[j], found = b, true
			}
		}
		if !found {
			return nil, ErrFormat
		}
	}
	return decode(c, inputs, f.UnpackSizes[out], a.done)
}
