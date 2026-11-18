package engine

import (
	"bytes"
	"encoding/binary"
	"strings"
)

const maxExecutableSections = 4096

// fileSpan validates untrusted offsets without addition overflow. Executable
// validation deliberately avoids debug/pe and debug/elf, whose constructors can
// allocate tables from attacker-controlled header counts before our limits run.
func fileSpan(data []byte, offset, size uint64) bool {
	return offset <= uint64(len(data)) && size <= uint64(len(data))-offset
}

// packedSection recognizes section names that executable packers and
// protectors write. Substring markers are distinctive; short names such as
// MEW's and yC's are compared exactly.
func packedSection(name string) bool {
	lower := strings.ToLower(name)
	for _, marker := range []string{"upx", "aspack", "mpress", "petite", "themida", "vmp", "wwp32", ".nsp", "pec2", "pespin"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	switch name {
	case "MEW", "yC", ".yP", "pebundle", "PEBundle", ".packed", "PELOCKnt", "ExeS":
		return true
	}
	return false
}

// peSection is a section's virtual address and file-backed extent, with the
// raw pointer rounded down the way the Windows loader rounds it, to 512 bytes.
type peSection struct {
	va, raw, rawEnd uint64
}

// peImage is the bounded layout of a PE image. overlay is the end of the
// headers and raw section data, limited to the file. certificate is where the
// table that the security directory names after the image starts, and its
// Authenticode signature spans [certificate, certificateEnd), empty when the
// table holds none; both are len(data) without a table.
type peImage struct {
	sections                    []peSection
	headers                     uint64
	overlay                     int
	certificate, certificateEnd int
	entry                       int  // file offset of the entry point, -1 when not file-backed
	packed                      bool // a packer's section name, entry code or header layout
}

// parsePEImage follows what the Windows loader accepts rather than what linkers
// emit: a header inside the DOS header, section data that starts inside the
// headers, unaligned raw pointers, a raw size beyond the virtual size, a
// SizeOfHeaders beyond the file and COFF symbol-table fields (ignored for
// images) are all loadable, and packers use them. Such layouts mark the image
// as packed instead of malformed. A section whose virtual extent passes
// SizeOfImage is accepted as the reference engine accepts it: only
// file-backed bytes are read here. Section data outside the file stays
// malformed. This is structural validation, not PE unpacking or an assertion
// of safety.
// notPE reports an MZ file the reference does not analyze as a PE at all: a
// 16-bit (NE, LE) program or font, or a header offset of zero or past the end
// of the file. Such a file is not reported, while a file too short for a DOS
// header, like a damaged PE header, is malformed (the reference raises its
// broken-executable alert there).
func notPE(data []byte) bool {
	if len(data) < 64 {
		return false
	}
	at := uint64(binary.LittleEndian.Uint32(data[0x3c:0x40]))
	return at == 0 || at+4 > uint64(len(data)) || !bytes.Equal(data[at:at+4], []byte{'P', 'E', 0, 0})
}

func parsePEImage(data []byte) (*peImage, error) {
	return parsePEImageView(data, false)
}

// parsePEScanImage provides the bounded physical section view accepted by the
// reference scanner. A declared raw size can exceed EOF while its start and
// header remain readable. Scanning clips that extent; signed-image hashing and
// trust continue to use parsePEImage's strict, complete-extent validation.
func parsePEScanImage(data []byte) (*peImage, error) {
	return parsePEImageView(data, true)
}

func parsePEImageView(data []byte, scanExtent bool) (*peImage, error) {
	if len(data) < 64 || !bytes.Equal(data[:2], []byte("MZ")) {
		return nil, ErrMalformed
	}
	order := binary.LittleEndian
	peOffset := uint64(order.Uint32(data[0x3c:0x40]))
	if peOffset < 4 || !fileSpan(data, peOffset, 24) || !bytes.Equal(data[peOffset:peOffset+4], []byte{'P', 'E', 0, 0}) {
		return nil, ErrMalformed
	}
	coff := data[peOffset+4 : peOffset+24]
	sections := uint64(order.Uint16(coff[2:4]))
	if sections == 0 {
		return nil, ErrMalformed
	}
	// This COFF field is already bounded to 65,535. The complete section
	// table must fit the input before we allocate or visit any entries.
	optionalSize := uint64(order.Uint16(coff[16:18]))
	optionalOffset := peOffset + 24
	if optionalSize < 2 || !fileSpan(data, optionalOffset, optionalSize) {
		return nil, ErrMalformed
	}
	optional := data[optionalOffset : optionalOffset+optionalSize]
	var directoryStart uint64
	switch order.Uint16(optional[:2]) {
	case 0x10b:
		directoryStart = 96
	case 0x20b:
		directoryStart = 112
	default:
		return nil, ErrUnsupported
	}
	if optionalSize < directoryStart {
		return nil, ErrMalformed
	}
	directories := uint64(order.Uint32(optional[directoryStart-4 : directoryStart]))
	if directories > 16 {
		return nil, ErrUnsupported
	}
	if directories > (optionalSize-directoryStart)/8 {
		return nil, ErrMalformed
	}
	tableOffset := optionalOffset + optionalSize
	if !fileSpan(data, tableOffset, sections*40) {
		return nil, ErrMalformed
	}
	img := &peImage{
		headers:        uint64(order.Uint32(optional[60:64])),
		certificate:    len(data),
		certificateEnd: len(data),
		entry:          -1,
		packed:         peOffset < 64,
	}
	fileAlign := uint64(order.Uint32(optional[36:40]))
	imageSize := uint64(order.Uint32(optional[56:60]))
	if img.headers < tableOffset+sections*40 || imageSize == 0 {
		return nil, ErrMalformed
	}
	end := min(img.headers, uint64(len(data)))
	img.sections = make([]peSection, 0, sections)
	for i := uint64(0); i < sections; i++ {
		h := data[tableOffset+i*40 : tableOffset+(i+1)*40]
		name := h[:8]
		if end := bytes.IndexByte(name, 0); end >= 0 {
			name = name[:end] // MEW writes code after the name's terminator
		}
		if packedSection(string(name)) {
			img.packed = true
		}
		virtualOffset := uint64(order.Uint32(h[12:16]))
		rawSize := uint64(order.Uint32(h[16:20]))
		rawOffset := uint64(order.Uint32(h[20:24]))
		if fileAlign >= 0x200 {
			rawOffset &^= 0x1ff
		}
		section := peSection{va: virtualOffset}
		if rawSize != 0 {
			if !fileSpan(data, rawOffset, rawSize) {
				if !scanExtent || rawOffset >= uint64(len(data)) {
					return nil, ErrMalformed
				}
				// The strict certificate reader cannot establish the signed
				// image layout after clipping. Preserve the prior incomplete
				// result for a declared security table rather than bypass its
				// blocklist/trust analysis through the unsigned scan view.
				if directories > 4 {
					security := optional[directoryStart+32 : directoryStart+40]
					if order.Uint32(security[:4]) != 0 || order.Uint32(security[4:]) != 0 {
						return nil, ErrMalformed
					}
				}
				rawSize = uint64(len(data)) - rawOffset
			}
			if rawOffset < img.headers {
				img.packed = true // section data inside the headers
			}
			section.raw, section.rawEnd = rawOffset, rawOffset+rawSize
			end = max(end, rawOffset+rawSize)
		}
		img.sections = append(img.sections, section)
	}
	img.overlay = int(end)
	// The certificate table's pointer is a file offset, not an RVA. Signing
	// tools append the table, so overlay data ends where it starts. Its DER
	// signature is not an installer payload and holds short archive magics by
	// chance. Only that signature is set apart: bytes an entry carries after
	// it (certificate padding keeps a signature valid on most systems), and a
	// table that holds no signature, are overlay data where payloads hide.
	if directories > 4 {
		dir := optional[directoryStart+32 : directoryStart+40]
		offset, size := uint64(order.Uint32(dir[:4])), uint64(order.Uint32(dir[4:8]))
		if size != 0 && offset >= end && offset < uint64(len(data)) {
			img.certificate = int(offset)
			img.certificateEnd = int(offset) + authenticodeSignatures(data[offset:min(offset+size, uint64(len(data)))])
		}
	}
	entry := uint64(order.Uint32(optional[16:20]))
	if entry < img.headers {
		if entry < uint64(len(data)) {
			img.entry = int(entry)
		}
	} else {
		for i := len(img.sections) - 1; i >= 0; i-- {
			s := img.sections[i]
			if s.rawEnd > s.raw && entry >= s.va && entry-s.va < s.rawEnd-s.raw {
				img.entry = int(s.raw + entry - s.va)
				break
			}
		}
	}
	if packerEntry(data, img) {
		img.packed = true
	}
	return img, nil
}

// authenticodeSignatures returns the length of the run of signature entries
// that starts a certificate table, or 0 when it starts with none. UEFI
// binaries carry one entry per signature, each aligned to eight bytes. The
// run ends at an entry that carries more than alignment padding after its
// signature, or at one that holds no signature.
func authenticodeSignatures(table []byte) int {
	end := 0
	for pos := 0; pos < len(table); {
		n := authenticodeSignature(table[pos:])
		if n == 0 {
			break
		}
		end = pos + n
		next := pos + int((uint64(binary.LittleEndian.Uint32(table[pos:]))+7)&^7)
		if next-end > 7 {
			break
		}
		pos = next
	}
	return end
}

// authenticodeSignature returns the length of the header and DER signature
// of the first WIN_CERTIFICATE entry of a certificate table, or 0 when the
// table does not start with one. An entry is a 32-bit length that includes
// the 8-byte header, a revision (1.0 or 2.0) and a type; Authenticode's type
// 2 holds PKCS#7 SignedData, a DER SEQUENCE whose length is definite.
func authenticodeSignature(table []byte) int {
	if len(table) < 10 {
		return 0
	}
	order := binary.LittleEndian
	length := uint64(order.Uint32(table))
	revision, kind := order.Uint16(table[4:]), order.Uint16(table[6:])
	if length > uint64(len(table)) || revision != 0x0100 && revision != 0x0200 || kind != 2 || table[8] != 0x30 {
		return 0
	}
	header, size := uint64(2), uint64(table[9])
	if size >= 0x80 {
		digits := size - 0x80
		if digits == 0 || digits > 4 || uint64(len(table)) < 10+digits {
			return 0 // indefinite (BER) or oversized lengths are not DER
		}
		size = 0
		for _, c := range table[10 : 10+digits] {
			size = size<<8 | uint64(c)
		}
		header += digits
	}
	if end := 8 + header + size; end <= length {
		return int(end)
	}
	return 0
}

// validatePE returns the overlay offset of a loadable image. A packed image
// is reported as unsupported: its code and data need an unpacker.
func validatePE(data []byte) (overlayOffset int, err error) {
	img, err := parsePEImage(data)
	if err != nil {
		return 0, err
	}
	if img.packed {
		return img.overlay, ErrUnsupported
	}
	return img.overlay, nil
}

// peEntryPattern is a packer's entry-point code; 0x100 in a pattern matches
// any byte. The patterns are the stubs of widely seen packers whose output the
// reference engine unpacks before scanning.
type peEntryPattern struct {
	offset  int
	pattern []int
}

const anyByte = 0x100

var packerEntryPatterns = []peEntryPattern{
	// UPX: pushad; mov esi, packed; lea edi, [esi-...]; push edi; or ebp, -1
	{0, []int{0x60, 0xbe, anyByte, anyByte, anyByte, anyByte, 0x8d, 0xbe, anyByte, anyByte, anyByte, anyByte, 0x57, 0x83, 0xcd, 0xff}},
	// FSG 2.0: xchg [..], esp; popad; xchg eax, esp; push ebp; movsb
	{0, []int{0x87, 0x25, anyByte, anyByte, anyByte, anyByte, 0x61, 0x94, 0x55, 0xa4, 0xb6, 0x80, 0xff, 0x13}},
	// FSG 1.3x: mov ebx/edi/esi, ...; push ebx; call +0x0a with its bit reader
	{0, []int{0xbb, anyByte, anyByte, anyByte, anyByte, 0xbf, anyByte, anyByte, anyByte, anyByte, 0xbe, anyByte, anyByte, anyByte, anyByte, 0x53, 0xe8, 0x0a, 0x00, 0x00, 0x00, 0x02, 0xd2, 0x75, 0x05, 0x8a, 0x16, 0x46, 0x12, 0xd2, 0xc3}},
	// PESpin 1.x: call $+5; mov ebx, [esp]; add ebx, ...
	{4, []int{0xe8, 0x00, 0x00, 0x00, 0x00, 0x8b, 0x1c, 0x24, 0x83, 0xc3}},
	// yC: push ebp; mov ebp, esp; push ebx/esi/edi; pushad; call $+5; pop ebp; sub ebp, ...
	{0, []int{0x55, 0x8b, 0xec, 0x53, 0x56, 0x57, 0x60, 0xe8, 0x00, 0x00, 0x00, 0x00, 0x5d, 0x81, 0xed}},
	// WWPack32 decompressor stub with its banner
	{0, []int{0x53, 0x55, 0x8b, 0xe8, 0x33, 0xdb, 0xeb, 0x60, 0x0d, 0x0a, 0x0d, 0x0a, 0x57, 0x57, 0x50, 0x61, 0x63, 0x6b, 0x33, 0x32}},
	// ASPack 2.x: pushad; call +3; jmp ...; pop ebp
	{0, []int{0x60, 0xe8, 0x03, 0x00, 0x00, 0x00, 0xe9, 0xeb, 0x04, 0x5d, 0x45, 0x55, 0xc3, 0xe8, 0x01}},
	// PEtite 2.x: mov eax, ...; push ...; push fs:[0]; mov fs:[0], esp
	{0, []int{0xb8, anyByte, anyByte, anyByte, anyByte, 0x68, anyByte, anyByte, anyByte, anyByte, 0x64, 0xff, 0x35, 0x00, 0x00, 0x00, 0x00, 0x64, 0x89, 0x25, 0x00, 0x00, 0x00, 0x00}},
	// Upack: mov esi, ...; lodsd; push eax; push dword [esi+0x34]; jmp
	{0, []int{0xbe, anyByte, anyByte, anyByte, anyByte, 0xad, 0x50, 0xff, 0x76, 0x34, 0xeb}},
	// NsPack: pushfd; pushad; call $+5; pop ebp
	{0, []int{0x9c, 0x60, 0xe8, 0x00, 0x00, 0x00, 0x00, 0x5d, 0xb8, 0x07, 0x00, 0x00, 0x00}},
}

// packerEntry reports entry-point code of a known packer stub.
func packerEntry(data []byte, img *peImage) bool {
	if img.entry < 0 {
		return false
	}
	code := data[img.entry:]
	for _, p := range packerEntryPatterns {
		if len(code) < p.offset+len(p.pattern) {
			continue
		}
		match := true
		for i, want := range p.pattern {
			if want != anyByte && code[p.offset+i] != byte(want) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// validateELF performs bounded ELF32/ELF64 validation in either byte order.
// Extended numbering and compressed/recognized packed sections are explicitly
// unsupported. Section/segment contents are already matched as raw bytes; this
// function does not claim to unpack an executable.
func validateELF(data []byte) error {
	if len(data) < 16 || !bytes.Equal(data[:4], []byte{0x7f, 'E', 'L', 'F'}) || data[6] != 1 {
		return ErrMalformed
	}
	var order binary.ByteOrder
	switch data[5] {
	case 1:
		order = binary.LittleEndian
	case 2:
		order = binary.BigEndian
	default:
		return ErrMalformed
	}
	var phoff, shoff uint64
	var phsize, phnum, shsize, shnum, shstr uint16
	class := data[4]
	switch class {
	case 1:
		if len(data) < 52 || order.Uint16(data[40:42]) != 52 {
			return ErrMalformed
		}
		phoff, shoff = uint64(order.Uint32(data[28:32])), uint64(order.Uint32(data[32:36]))
		phsize, phnum = order.Uint16(data[42:44]), order.Uint16(data[44:46])
		shsize, shnum, shstr = order.Uint16(data[46:48]), order.Uint16(data[48:50]), order.Uint16(data[50:52])
	case 2:
		if len(data) < 64 || order.Uint16(data[52:54]) != 64 {
			return ErrMalformed
		}
		phoff, shoff = order.Uint64(data[32:40]), order.Uint64(data[40:48])
		phsize, phnum = order.Uint16(data[54:56]), order.Uint16(data[56:58])
		shsize, shnum, shstr = order.Uint16(data[58:60]), order.Uint16(data[60:62]), order.Uint16(data[62:64])
	default:
		return ErrUnsupported
	}
	if order.Uint32(data[20:24]) != 1 {
		return ErrMalformed
	}
	if phnum == 0xffff || shstr == 0xffff || (shnum == 0 && shoff != 0) {
		return ErrUnsupported
	}
	if phnum > maxExecutableSections || shnum > maxExecutableSections {
		return ErrLimit
	}
	wantPH, wantSH := uint16(32), uint16(40)
	if class == 2 {
		wantPH, wantSH = 56, 64
	}
	if (phnum > 0 && (phoff == 0 || phsize != wantPH)) ||
		(shnum > 0 && (shoff == 0 || shsize != wantSH)) ||
		!fileSpan(data, phoff, uint64(phnum)*uint64(phsize)) ||
		!fileSpan(data, shoff, uint64(shnum)*uint64(shsize)) ||
		(shstr != 0 && shstr >= shnum) {
		return ErrMalformed
	}
	for i := uint64(0); i < uint64(phnum); i++ {
		p := data[phoff+i*uint64(phsize) : phoff+(i+1)*uint64(phsize)]
		kind := order.Uint32(p[:4])
		var offset, address, fileSize, memSize, alignment uint64
		if class == 1 {
			offset, address = uint64(order.Uint32(p[4:8])), uint64(order.Uint32(p[8:12]))
			fileSize, memSize = uint64(order.Uint32(p[16:20])), uint64(order.Uint32(p[20:24]))
			alignment = uint64(order.Uint32(p[28:32]))
		} else {
			offset, address = order.Uint64(p[8:16]), order.Uint64(p[16:24])
			fileSize, memSize, alignment = order.Uint64(p[32:40]), order.Uint64(p[40:48]), order.Uint64(p[48:56])
		}
		if !fileSpan(data, offset, fileSize) || memSize > ^uint64(0)-address ||
			(class == 1 && address+memSize > 1<<32) || (kind == 1 && fileSize > memSize) {
			return ErrMalformed
		}
		if alignment > 1 && (alignment&(alignment-1) != 0 || (kind == 1 && address%alignment != offset%alignment)) {
			return ErrMalformed
		}
	}
	var names []byte
	if shstr != 0 {
		h := data[shoff+uint64(shstr)*uint64(shsize) : shoff+(uint64(shstr)+1)*uint64(shsize)]
		offset, size := elfSectionSpan(h, class, order)
		if order.Uint32(h[4:8]) != 3 || !fileSpan(data, offset, size) {
			return ErrMalformed
		}
		names = data[offset : offset+size]
	}
	for i := uint64(0); i < uint64(shnum); i++ {
		h := data[shoff+i*uint64(shsize) : shoff+(i+1)*uint64(shsize)]
		kind := order.Uint32(h[4:8])
		if kind == 0 {
			continue
		}
		offset, size := elfSectionSpan(h, class, order)
		if kind != 8 && !fileSpan(data, offset, size) { // SHT_NOBITS has no file bytes
			return ErrMalformed
		}
		flags := uint64(order.Uint32(h[8:12]))
		if class == 2 {
			flags = order.Uint64(h[8:16])
		}
		if flags&0x800 != 0 { // SHF_COMPRESSED
			return ErrUnsupported
		}
		nameOffset := uint64(order.Uint32(h[:4]))
		if nameOffset != 0 && names != nil {
			if nameOffset >= uint64(len(names)) {
				return ErrMalformed
			}
			nameBytes := names[nameOffset:min(uint64(len(names)), nameOffset+257)]
			end := bytes.IndexByte(nameBytes, 0)
			if end < 0 {
				return ErrLimit
			}
			if packedSection(string(nameBytes[:end])) || bytes.HasPrefix(nameBytes[:end], []byte(".zdebug")) {
				return ErrUnsupported
			}
		}
	}
	return nil
}

func elfSectionSpan(h []byte, class byte, order binary.ByteOrder) (offset, size uint64) {
	if class == 1 {
		return uint64(order.Uint32(h[16:20])), uint64(order.Uint32(h[20:24]))
	}
	return order.Uint64(h[24:32]), order.Uint64(h[32:40])
}
