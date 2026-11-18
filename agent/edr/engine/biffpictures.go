package engine

import "encoding/binary"

// Pictures in an Excel binary workbook, extracted as the reference does
// (learned from its debug output and verdicts on generated workbooks and 380
// real ones; docs/edr/native-engine/parity/macros.md section 7). The
// reference walks every record of a "Workbook" stream; a record that runs
// past the end of the stream stops it and no picture is extracted. The data
// of the first MSODRAWINGGROUP record and the records of either type right
// after it is an OfficeArtDggContainer; each OfficeArtBStoreContainer in it is
// walked by length: a File Blip Store Entry is skipped by its fixed part and
// name (an embedded picture is the next record), a picture record of a known
// type and instance is extracted without its header, other records are
// skipped. Each picture is scanned as a file.

// biffPictureHeader is the header length of the picture records the
// reference extracts, by record type and instance: one or two 16-byte
// identifiers, then a tag byte (bitmaps) or a 34-byte metafile header.
var biffPictureHeader = map[uint16]map[uint16]int{
	0xf01a: {0x3d4: 50, 0x3d5: 66},                       // EMF
	0xf01b: {0x216: 50, 0x217: 66},                       // WMF
	0xf01c: {0x542: 50, 0x543: 66},                       // PICT
	0xf01d: {0x46a: 17, 0x46b: 33, 0x6e2: 17, 0x6e3: 33}, // JPEG
	0xf01e: {0x6e0: 17, 0x6e1: 33},                       // PNG
	0xf01f: {0x7a8: 17, 0x7a9: 33},                       // DIB
	0xf029: {0x6e4: 17, 0x6e5: 33},                       // TIFF
	0xf02a: {0x6e2: 17, 0x6e3: 33},                       // JPEG (CMYK)
}

// biffPictures returns the pictures the reference extracts from a workbook
// stream: the first MSODRAWINGGROUP record and the MSODRAWINGGROUP and
// CONTINUE records right after it (Excel splits a large drawing group so).
// decided is false where this engine cannot tell what the reference
// extracts: a record longer than any observed (8,225 bytes are read, 40,837
// stop the walk), a second picture store, or record forms not observed.
func biffPictures(b []byte) (pictures [][]byte, decided bool) {
	var group []byte
	inGroup := false
	for pos := 0; len(b)-pos >= 4; {
		typ := binary.LittleEndian.Uint16(b[pos:])
		n := int(binary.LittleEndian.Uint16(b[pos+2:]))
		if n > len(b)-pos-4 {
			return nil, true
		}
		if n > 8225 {
			return nil, false
		}
		data := b[pos+4 : pos+4+n]
		pos += 4 + n
		switch {
		case typ == 0x00eb && group == nil:
			group, inGroup = append([]byte{}, data...), true
		case (typ == 0x00eb || typ == 0x003c) && inGroup:
			group = append(group, data...)
		default:
			inGroup = false
		}
	}
	if group == nil {
		return nil, true
	}
	if len(group) < 8 {
		return nil, false
	}
	ver, typ, size := artHeader(group)
	if ver != 0xf || typ != 0xf000 {
		return nil, true
	}
	end := 8 + min(int(size), len(group)-8)
	stores := 0
	for pos := 8; end-pos >= 8; {
		ver, typ, size := artHeader(group[pos:])
		body := pos + 8
		next := body + int(size)
		if next > end || next < body {
			next = end
		}
		if typ == 0xf001 && ver == 0xf {
			stores++
			if stores > 1 {
				return nil, false
			}
			pics, ok := blipStore(group[body:next])
			if !ok {
				return nil, false
			}
			pictures = append(pictures, pics...)
		}
		pos = next
	}
	return pictures, true
}

func artHeader(b []byte) (ver uint16, typ uint16, size uint32) {
	v := binary.LittleEndian.Uint16(b)
	return v & 0xf, binary.LittleEndian.Uint16(b[2:]), binary.LittleEndian.Uint32(b[4:])
}

// blipStore walks an OfficeArtBStoreContainer's data. A File Blip Store
// Entry whose picture size is 0 stops the walk; otherwise the picture record
// after its fixed part and name is read within that size, and the walk goes
// on after the entry.
func blipStore(b []byte) (pictures [][]byte, ok bool) {
	for pos := 0; len(b)-pos >= 8; {
		ver, typ, size := artHeader(b[pos:])
		body := pos + 8
		next := len(b)
		if uint64(size) <= uint64(len(b)-body) {
			next = body + int(size)
		}
		switch {
		case typ == 0xf007 && ver == 2:
			if len(b)-body < 36 {
				return nil, false
			}
			pictureSize := binary.LittleEndian.Uint32(b[body+20:])
			if pictureSize == 0 {
				return pictures, true
			}
			inner := body + 36 + int(b[body+33])
			if p, ok := blipPicture(b, inner, pictureSize); !ok {
				return nil, false
			} else if p != nil {
				pictures = append(pictures, p)
			}
		case typ == 0xf007:
			return nil, false
		default:
			if p, ok := blipPicture(b, pos, size+8); !ok {
				return nil, false
			} else if p != nil {
				pictures = append(pictures, p)
			}
		}
		pos = next
	}
	return pictures, true
}

// blipPicture is the picture of a picture record at pos, read within limit
// bytes (header included); nil when the record is no picture the reference
// extracts.
func blipPicture(b []byte, pos int, limit uint32) ([]byte, bool) {
	if len(b)-pos < 8 {
		return nil, true
	}
	ver, typ, _ := artHeader(b[pos:])
	header, picture := biffPictureHeader[typ]
	if !picture {
		return nil, true
	}
	if ver != 0 {
		return nil, false
	}
	hl, valid := header[binary.LittleEndian.Uint16(b[pos:])>>4]
	if !valid {
		return nil, true
	}
	if limit < uint32(8+hl) {
		return nil, false
	}
	end := len(b)
	if uint64(limit) <= uint64(len(b)-pos) {
		end = pos + int(limit)
	}
	if pos+8+hl >= end {
		return nil, false
	}
	return b[pos+8+hl : end], true
}
