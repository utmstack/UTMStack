package engine

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
)

// Symbian installation packages before OS 9 (EPOC releases 3-6): a header of
// UIDs, counts and pointers, then sequential file records. File records
// point at each language's data, compressed with zlib unless the header's
// "no compression" option is set; release 6 records add the uncompressed
// length and a MIME type. Option and condition records carry no files.
// Symbian OS 9 packages (SISX) use a different container and stay
// unsupported. The layout follows Alexander Thoukydides' published notes on
// the SIS file format.
const (
	sisRelease6     = 0x10003a12
	sisX            = 0x10201a7a
	sisNoCompress   = 0x0008
	sisMaxLanguages = 256
)

func (s *scanState) sis(data []byte, depth int) (string, error) {
	if binary.LittleEndian.Uint32(data) == sisX {
		return "", ErrUnsupported
	}
	if len(data) < 68 {
		return "", ErrMalformed
	}
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	u16 := func(off int) int { return int(binary.LittleEndian.Uint16(data[off:])) }
	u32 := func(off int) uint64 { return uint64(binary.LittleEndian.Uint32(data[off:])) }
	release6 := u32(4) == sisRelease6
	languages, files, options := u16(18), u16(20), u16(36)
	if languages == 0 || languages > sisMaxLanguages {
		return "", ErrMalformed
	}
	if files > s.maxFiles-s.files {
		return "", ErrLimit
	}
	pos := u32(52)
	var incomplete error
	// field reads a 32-bit field of the current record.
	field := func(n int) (uint64, bool) {
		if pos > uint64(len(data)) || uint64(len(data))-pos < uint64(4*(n+1)) {
			return 0, false
		}
		return u32(int(pos) + 4*n), true
	}
	for i := 0; i < files; i++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		kind, ok := field(0)
		if !ok {
			return "", ErrMalformed
		}
		switch kind {
		case 0, 1: // simple and multi-language files
			entries := 1
			if kind == 1 {
				entries = languages
			}
			fileType, _ := field(1)
			// type, file type, details, source and destination names: 7 fields.
			lengths, pointers := 7, 7+entries
			size := 7 + 2*entries
			if release6 {
				size += entries + 2 // original lengths, MIME type length and pointer
			}
			if _, ok := field(size - 1); !ok {
				return "", ErrMalformed
			}
			for e := 0; e < entries; e++ {
				length, _ := field(lengths + e)
				pointer, _ := field(pointers + e)
				original := int64(-1)
				if release6 {
					v, _ := field(pointers + entries + e)
					original = int64(v)
				}
				if fileType == 4 || length == 0 {
					continue // created on installation; no data in the package
				}
				if pointer > uint64(len(data)) || length > uint64(len(data))-pointer {
					return "", ErrMalformed
				}
				sig, err := s.sisFile(data[pointer:pointer+length], options&sisNoCompress == 0, original, depth)
				if sig != "" {
					return sig, nil
				}
				if err != nil {
					incomplete = err
				}
			}
			pos += uint64(4 * size)
		case 2: // options: names for each language, then a 128-bit selection
			n, ok := field(1)
			if !ok || n > uint64(s.maxFiles) {
				return "", ErrMalformed
			}
			pos += 8 + n*uint64(languages)*8 + 16
		case 3, 4: // if and else-if: an inline condition expression
			n, ok := field(1)
			if !ok {
				return "", ErrMalformed
			}
			pos += 8 + n
		case 5, 6: // else and end-if
			pos += 4
		default:
			return "", ErrMalformed
		}
	}
	return "", incomplete
}

func (s *scanState) sisFile(stored []byte, compressed bool, original int64, depth int) (string, error) {
	if !compressed {
		if original >= 0 && original != int64(len(stored)) {
			return "", ErrMalformed
		}
		return s.visit(stored, depth+1)
	}
	if original > min(s.maxFile, s.remaining) {
		return "", ErrLimit
	}
	r, err := zlib.NewReader(bytes.NewReader(stored))
	if err != nil {
		return "", ErrMalformed
	}
	defer r.Close()
	out, err := s.readChild(r, depth)
	if err != nil {
		return "", err
	}
	if original >= 0 && int64(len(out)) != original {
		s.spend(len(out))
		return "", ErrMalformed
	}
	return s.visit(out, depth+1)
}
