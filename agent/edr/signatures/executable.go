package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// section is one PE section as the reference's header reader records it:
// rva, raw and size (the raw size) aligned and cut at the end of the file,
// vsz aligned (rebuilt from the raw size when zero), chr, and the four
// values as the header states them (ursz cut at the end of the file too).
type section struct {
	rva, raw, size         uint32
	vsz, chr               uint32
	urva, uvsz, uraw, ursz uint32
}

// peMeta is non-nil for every MZ object, the objects PE-target rules apply to.
// valid reports whether the header could be modeled; if not, entry and sections
// are unavailable and rules that need them cannot be decided.
type peMeta struct {
	valid bool
	// failed is an early header rejection also made by the reference. Other
	// unmodeled headers remain unknown, since the reference may accept them.
	failed   bool
	entry    int
	sections []section
	// headers is the aligned header size and size the object's length;
	// rawAddr maps addresses with them.
	headers uint32
	size    uint64
	// resDir is the resource directory address, read when the header has
	// at least three data directories (hasResDir). hasRes is the reference's
	// resource address test for icons: not a DLL, a nonempty resource
	// directory that maps into the file.
	resDir, resSize   uint32
	hasResDir, hasRes bool
}

// NumberOfSections is a 16-bit COFF field. The section-table bounds check
// below keeps allocation proportional to bytes present in the input.
const maxPESections = 65535

// peAlign and peAlignUp are the reference PEALIGN/PESALIGN, including 32-bit
// wraparound; a zero alignment leaves the value unchanged.
func peAlign(v, a uint32) uint32 {
	if a == 0 {
		return v
	}
	return v / a * a
}
func peAlignUp(v, a uint32) uint32 {
	if a == 0 {
		return v
	}
	n := v / a
	if v%a != 0 {
		n++
	}
	return n * a
}

// parsePE mirrors the reference header model used for EP, Sx, SL and SE offsets,
// EntryPoint/NumberOfSections descriptors and section hashes: aligned headers
// and sections, the 0x200 file-alignment fallback, virtual-only and truncated
// sections, and entry-point mapping. It reads only fixed header fields and the
// bounded section table, and allocates at most maxPESections entries.
func parsePE(data []byte) *peMeta {
	if len(data) < 2 || string(data[:2]) != "MZ" {
		return nil
	}
	meta := &peMeta{entry: -1}
	le := binary.LittleEndian
	size := uint64(len(data))
	if size < 64 {
		return nil // the reference cannot read a DOS header
	}
	h := uint64(le.Uint32(data[0x3c:0x40]))
	if h == 0 || h+24 > size || string(data[h:h+4]) != "PE\x00\x00" {
		return nil // no PE header (a 16-bit program or font, for example)
	}
	count := uint64(le.Uint16(data[h+6 : h+8]))
	optional := uint64(le.Uint16(data[h+20 : h+22]))
	o := h + 24
	fixed := uint64(96) // optional-header fields before the data directories
	if o+2 <= size && le.Uint16(data[o:o+2]) == 0x20b {
		fixed = 112
	}
	if count == 0 {
		meta.failed = true // the reference rejects a zero-section PE
		return meta
	}
	if optional < fixed || o+fixed > size {
		meta.failed = true // incomplete fixed optional header
		return meta
	}
	directories := min(uint64(le.Uint32(data[o+fixed-4:o+fixed])), 16) * 8
	table := o + optional
	if optional < fixed+directories || o+fixed+directories > size || table+count*40 > size {
		meta.failed = true // incomplete data directories or section table
		return meta
	}
	characteristics := le.Uint16(data[h+22 : h+24])
	entry := le.Uint32(data[o+16 : o+20])
	sectionAlign := le.Uint32(data[o+32 : o+36])
	fileAlign := le.Uint32(data[o+36 : o+40])
	headers := peAlignUp(le.Uint32(data[o+60:o+64]), sectionAlign)
	header := func(i uint64) []byte { return data[table+i*40 : table+i*40+40] }
	// A raw pointer misaligned for FileAlignment but aligned to 0x200 switches
	// the whole image to 0x200 alignment.
	for i := uint64(0); fileAlign != 0x200 && i < count; i++ {
		rawSize, pointer := le.Uint32(header(i)[16:20]), le.Uint32(header(i)[20:24])
		if fileAlign != 0 && rawSize != 0 && pointer%fileAlign != 0 && pointer%0x200 == 0 {
			fileAlign = 0x200
		}
	}
	meta.sections = make([]section, 0, count)
	for i := uint64(0); i < count; i++ {
		s := header(i)
		virtualSize, virtualAddress := le.Uint32(s[8:12]), le.Uint32(s[12:16])
		rawSize, pointer := le.Uint32(s[16:20]), le.Uint32(s[20:24])
		sec := section{rva: peAlign(virtualAddress, sectionAlign), raw: peAlign(pointer, fileAlign), size: peAlignUp(rawSize, fileAlign),
			vsz: peAlignUp(virtualSize, sectionAlign), chr: le.Uint32(s[36:40]), urva: virtualAddress, uvsz: virtualSize, uraw: pointer}
		if sec.size != 0 {
			if uint64(sec.raw) >= size || uint64(pointer) >= size {
				sec.size, rawSize = 0, 0
			} else {
				sec.size = uint32(min(uint64(sec.size), size-uint64(sec.raw)))
				rawSize = uint32(min(uint64(rawSize), size-uint64(pointer)))
			}
		}
		if virtualAddress>>31 != 0 || virtualSize>>31 != 0 || sec.size != 0 && pointer>>31 != 0 || rawSize>>31 != 0 {
			return &peMeta{entry: -1, failed: true} // reference rejects signed section fields
		}
		sec.ursz = rawSize
		if sec.vsz == 0 && sec.size != 0 {
			sec.vsz = peAlignUp(rawSize, sectionAlign)
		}
		meta.sections = append(meta.sections, sec)
	}
	meta.headers, meta.size = headers, size
	if e, ok := meta.rawAddr(entry); ok {
		meta.entry = e
	} else {
		// cli_peheader returns CL_EFORMAT when cli_rawaddr cannot map the
		// entry point. No PE-relative offset or VERSION_INFO key is scanned.
		return &peMeta{entry: -1, failed: true}
	}
	if directories >= 3*8 {
		d := data[o+fixed+16:]
		meta.resDir, meta.hasResDir = le.Uint32(d), true
		meta.resSize = le.Uint32(d[4:])
		if at, ok := meta.rawAddr(meta.resDir); ok && at != 0 && characteristics&0x2000 == 0 && meta.resSize != 0 {
			meta.hasRes = true
		}
	}
	meta.valid = true
	return meta
}

// bytecodePE is what a logical program sees of this header.
func (m *peMeta) bytecodePE() *bytecode.PEInfo {
	pe := &bytecode.PEInfo{EntryPoint: uint32(m.entry), HeaderSize: m.headers, Sections: make([]bytecode.Section, len(m.sections))}
	for i, s := range m.sections {
		pe.Sections[i] = bytecode.Section{RVA: s.rva, VSize: s.vsz, Raw: s.raw, RSize: s.size, Characteristics: s.chr,
			URVA: s.urva, UVSize: s.uvsz, URaw: s.uraw, URSize: s.ursz}
	}
	return pe
}

// rawAddr maps an address to a file offset as the reference does: an address
// below the aligned header size is its own offset, otherwise the last
// file-backed section containing it maps it.
func (m *peMeta) rawAddr(rva uint32) (int, bool) {
	if rva < m.headers {
		return int(rva), uint64(rva) < m.size
	}
	for i := len(m.sections) - 1; i >= 0; i-- {
		s := m.sections[i]
		if s.size != 0 && s.rva <= rva && s.size > rva-s.rva {
			return int(rva - s.rva + s.raw), true
		}
	}
	return 0, false
}

// resources visits the resource data entries the way the reference's
// resource lookup does: below the first type entry with the given id (an id,
// not a name, leading to a directory), every name and every language that
// leads to a data entry. visit gets the address of each data entry and
// returns true to stop. ok is false when a directory or entry could not be
// read.
func (m *peMeta) resources(data []byte, typeID uint32, visit func(entry uint32) bool) (ok bool) {
	return m.resourcesNamed(data, typeID, func(_ uint32, entry uint32) bool { return visit(entry) })
}

// resourcesNamed also supplies the resource name/ID to callers that need
// the reference's callback boundary between consecutive name groups.
func (m *peMeta) resourcesNamed(data []byte, typeID uint32, visit func(name, entry uint32) bool) (ok bool) {
	read, _ := m.walkResources(data, typeID, ^uint32(0), nil, func(name, entry uint32) (bool, error) { return visit(name, entry), nil })
	return read
}

// walkResources preserves the reference's callback order and uint16 counters.
// Directory headers and entries have separate physical read guards: a damaged
// tail cannot erase callbacks already delivered. read=false is a modeled
// terminal read failure, not permission to inspect later sibling directories.
// Optional work belongs to the enclosing scan, including cancellation.
func (m *peMeta) walkResources(data []byte, typeID, nameID uint32, work *matchWork, visit func(name, entry uint32) (bool, error)) (read bool, err error) {
	if work != nil {
		if err = work.ctx.Err(); err != nil {
			return false, err
		}
	}
	le := binary.LittleEndian
	directory := func(offset uint32) ([]byte, bool) {
		at, ok := m.rawAddr(m.resDir + offset)
		if !ok || at < 0 || at > len(data)-16 {
			return nil, false
		}
		return data[at:], true
	}
	entry := func(dir []byte, index int) (id, offset uint32, ok bool, err error) {
		if work != nil {
			if err = work.step(1); err != nil {
				return
			}
		}
		at := 16 + 8*index
		if at > len(dir)-8 {
			return
		}
		return le.Uint32(dir[at:]), le.Uint32(dir[at+4:]), true, nil
	}
	root, ok := directory(0)
	if !ok {
		return false, nil
	}
	typeStart, typeCount := 0, int(le.Uint16(root[12:]))
	if typeID>>31 == 0 {
		typeStart, typeCount = typeCount, int(le.Uint16(root[14:]))
	}
	for i := typeStart; i < typeStart+typeCount; i++ {
		typ, off, ok, err := entry(root, i)
		if !ok || err != nil {
			return false, err
		}
		if typ != typeID || off>>31 == 0 {
			continue
		}
		names, ok := directory(off & 0x7fffffff)
		if !ok {
			return false, nil
		}
		nameStart, nameCount := 0, int(le.Uint16(names[12:]))
		if nameID == ^uint32(0) {
			nameCount = int(uint16(uint32(nameCount) + uint32(le.Uint16(names[14:]))))
		} else if nameID>>31 == 0 {
			nameStart, nameCount = nameCount, int(le.Uint16(names[14:]))
		}
		for j := nameStart; j < nameStart+nameCount; j++ {
			name, off, ok, err := entry(names, j)
			if !ok || err != nil {
				return false, err
			}
			if nameID != ^uint32(0) && name != nameID || off>>31 == 0 {
				continue
			}
			languages, ok := directory(off & 0x7fffffff)
			if !ok {
				return false, nil
			}
			languageCount := uint16(uint32(le.Uint16(languages[12:])) + uint32(le.Uint16(languages[14:])))
			for k := 0; k < int(languageCount); k++ {
				_, off, ok, err := entry(languages, k)
				if !ok || err != nil {
					return false, err
				}
				if off>>31 != 0 {
					continue
				}
				stop, err := visit(name, m.resDir+off)
				if stop || err != nil {
					return true, err
				}
			}
		}
		return true, nil // only the first matching directory of this type
	}
	return true, nil
}

// hasGroupIcon reports whether the reference's icon scan would find a group
// icon (resource type 14) to compare.
func (m *peMeta) hasGroupIcon(data []byte) bool {
	found := false
	m.resources(data, 14, func(uint32) bool { found = true; return true })
	return found
}

// maxVersionResources is how many version resources the reference reads.
const maxVersionResources = 16

// versionOffsets lists the file offsets at which VI offsets match: the keys
// of the version strings of the first 16 version resources (type 16) the
// resource lookup finds. A resource that does not map into the object is
// skipped, as the reference skips it.
func (m *peMeta) versionOffsets(data []byte) []int {
	// The reference does not extract VERSION_INFO when the resource data
	// directory advertises zero bytes, even if its RVA maps to a tree.
	if !m.hasResDir || m.resSize == 0 {
		return nil
	}
	var entries []uint32
	m.resources(data, 16, func(entry uint32) bool {
		entries = append(entries, entry)
		return len(entries) == maxVersionResources
	})
	le := binary.LittleEndian
	var offsets []int
	for _, entry := range entries {
		at, found := m.rawAddr(entry)
		if !found || at+16 > len(data) {
			continue
		}
		start, found := m.rawAddr(le.Uint32(data[at:]))
		size := uint64(le.Uint32(data[at+4:]))
		if !found || size == 0 || uint64(start)+size > uint64(len(data)) {
			continue
		}
		for _, k := range versionKeys(data[start : start+int(size)]) {
			offsets = append(offsets, start+k)
		}
	}
	slices.Sort(offsets)
	return slices.Compact(offsets)
}

var (
	versionInfoKey    = utf16Key("VS_VERSION_INFO")
	stringFileInfoKey = utf16Key("StringFileInfo")
	varFileInfoKey    = utf16Key("VarFileInfo")
)

func utf16Key(s string) []byte {
	b := make([]byte, 0, 2*len(s)+2)
	for i := 0; i < len(s); i++ {
		b = append(b, s[i], 0)
	}
	return append(b, 0, 0)
}

// versionKeys lists the offsets, within one version resource, of the keys of
// its version strings, as the reference records them: a VS_VERSION_INFO
// block with its 52-byte fixed value, then its first StringFileInfo block
// (one VarFileInfo block may come first), then every string of every string
// table that has value bytes after its 32-bit aligned key. Lengths are the
// 16-bit fields of each block; string lengths are rounded up to 4 bytes. A
// block longer than its parent or too short to be one stops the reading
// there, keeping the keys already found.
func versionKeys(res []byte) (keys []int) {
	le := binary.LittleEndian
	length := func(at int) int { return int(le.Uint16(res[at:])) }
	keyIs := func(at int, key []byte) bool {
		return at+6+len(key) <= len(res) && bytes.Equal(res[at+6:at+6+len(key)], key)
	}
	const head = 6 + 0x20 + 2 + 0x34 // header, key, padding, fixed value
	if len(res) < head || length(0) > len(res) || length(0) <= head || length(2) != 0x34 ||
		!keyIs(0, versionInfoKey) || le.Uint32(res[0x28:]) != 0xfeef04bd {
		return nil
	}
	pos, rest := head, length(0)-head
	skippedVar := false
	for rest > 6 {
		size := length(pos)
		if size > rest {
			return keys
		}
		if !skippedVar && size > 6+len(varFileInfoKey) && keyIs(pos, varFileInfoKey) {
			pos, rest, skippedVar = pos+size, rest-size, true
			continue
		}
		if size <= 6+len(stringFileInfoKey) || !keyIs(pos, stringFileInfoKey) {
			return keys
		}
		table, tables := pos+6+len(stringFileInfoKey), size-6-len(stringFileInfoKey)
		for tables > 6 {
			tableSize := length(table)
			if tableSize > tables || tableSize <= 24 {
				return keys
			}
			str, strs := table+24, tableSize-24 // header, 8-character key, padding
			for strs > 6 {
				strSize := (length(str) + 3) &^ 3
				if strSize > strs || strSize <= 6+2+8 {
					return keys
				}
				key := 6
				for ; key+1 < strSize; key += 2 {
					if res[str+key] == 0 && res[str+key+1] == 0 {
						key += 2
						break
					}
				}
				if key = (key + 3) &^ 3; key < strSize && strSize-key > 2 {
					keys = append(keys, str+6)
				}
				str, strs = str+strSize, strs-strSize
			}
			table, tables = table+tableSize, tables-tableSize
		}
		return keys
	}
	return keys
}

type offsetKind uint8

const (
	offsetUnset offsetKind = iota
	offsetAbsolute
	offsetAny
	offsetEOF
	offsetEntry
	offsetEntire
	offsetLast
	offsetSection
	offsetVersionInfo
)

// Signature offset fields are parsed as signed 32-bit integers. Keeping the
// immutable representation at that width avoids retaining an oversized
// string-and-int record in every compiled body pattern.
type offset struct {
	value, shift, section int32
	kind                  offsetKind
}

func (o offset) executable() bool {
	return o.kind == offsetEntry || o.kind == offsetSection || o.kind == offsetLast || o.kind == offsetEntire || o.kind == offsetVersionInfo
}

// relative reports offsets that need the object's size or header: every kind
// but absolute and any position.
func (o offset) relative() bool { return o.kind != offsetAbsolute && o.kind != offsetAny }
func parseOffset(s string) (offset, error) {
	o := offset{kind: offsetAbsolute}
	parts := strings.Split(s, ",")
	if len(parts) > 2 {
		return o, errors.New("invalid floating offset")
	}
	s = parts[0]
	parse := func(s string) (int32, error) {
		n, e := strconv.ParseInt(s, 10, 32)
		if e != nil || n < 0 {
			return 0, errors.New("invalid nonnegative signature offset")
		}
		return int32(n), nil
	}
	var err error
	if len(parts) == 2 {
		o.shift, err = parse(parts[1])
		if err != nil {
			return o, err
		}
	}
	switch {
	case s == "*":
		if len(parts) > 1 {
			return o, errors.New("floating wildcard offsets are unsupported")
		}
		o.kind = offsetAny
	case s == "VI":
		// A match must start at a version-information string of a PE file;
		// this engine does not collect them (see unknown.go).
		if len(parts) > 1 {
			return o, errors.New("floating version-information offsets are unsupported")
		}
		o.kind = offsetVersionInfo
	case strings.HasPrefix(s, "EOF-"):
		o.kind = offsetEOF
		o.value, err = parse(s[4:])
	case s == "EP" || strings.HasPrefix(s, "EP+") || strings.HasPrefix(s, "EP-"):
		o.kind = offsetEntry
		if len(s) > 2 {
			o.value, err = parse(s[3:])
			if s[2] == '-' {
				o.value = -o.value
			}
		}
	case strings.HasPrefix(s, "SE"):
		o.kind = offsetEntire
		v, e := parse(s[2:])
		err = e
		o.section = v
		if len(parts) > 1 {
			return o, errors.New("floating entire-section offsets are unsupported")
		}
	case strings.HasPrefix(s, "SL+"):
		o.kind = offsetLast
		o.value, err = parse(s[3:])
	case strings.HasPrefix(s, "S"):
		a := strings.Split(s[1:], "+")
		if len(a) != 2 {
			return o, errors.New("unsupported executable section offset")
		}
		var v int32
		v, err = parse(a[0])
		o.section = v
		if err == nil {
			o.value, err = parse(a[1])
		}
		o.kind = offsetSection
	default:
		o.value, err = parse(s)
	}
	if err != nil {
		return o, fmt.Errorf("invalid signature offset %q: %w", s, err)
	}
	return o, nil
}
func (o offset) bounds(length int, p *peMeta) (int, int, bool) {
	lo := int64(o.value)
	hi := int64(length)
	switch o.kind {
	case offsetAny:
		return 0, length, true
	case offsetEOF:
		lo = int64(length) - int64(o.value)
	case offsetEntry:
		if p == nil || !p.valid {
			return 0, 0, false
		}
		lo = int64(p.entry) + int64(o.value)
	case offsetSection, offsetEntire, offsetLast:
		if p == nil || !p.valid {
			return 0, 0, false
		}
		index := int(o.section)
		if o.kind == offsetLast {
			index = len(p.sections) - 1
		}
		if index < 0 || index >= len(p.sections) {
			return 0, 0, false
		}
		// Sx+n and SL+n start at the aligned pointer even for a section without
		// file data; an offset past the end of the object simply cannot match.
		s := p.sections[index]
		lo = int64(s.raw) + int64(o.value)
		if o.kind == offsetEntire {
			if s.size == 0 {
				return 0, 0, false
			}
			return int(s.raw), int(int64(s.raw) + int64(s.size) - 1), true
		}
	}
	hi = lo + int64(o.shift)
	if lo < 0 || lo > int64(length) {
		return 0, 0, false
	}
	hi = min(hi, int64(length))
	return int(lo), int(hi), true
}
