package engine

import (
	"bytes"
	"encoding/binary"
	"math/bits"
)

// FileDataStoreObject is specified by MS-ONESTORE 2.6.13. These GUIDs are
// stored in the wire's mixed-endian GUID order. The object has a 36-byte
// header, cbLength payload bytes, padding to an eight-byte total length, and
// a 16-byte footer. This bounded walk handles desktop attachment records;
// web exports take the separate OneStore path below.
var oneNoteDataHeader = [...]byte{
	0xe7, 0x16, 0xe3, 0xbd, 0x65, 0x26, 0x11, 0x45,
	0xa4, 0xc4, 0x8d, 0x4d, 0x0b, 0x7a, 0x9e, 0xac,
}
var oneNoteDataFooter = [...]byte{
	0x22, 0xa7, 0xfb, 0x71, 0x79, 0x0f, 0x0b, 0x4a,
	0xbb, 0x13, 0x89, 0x92, 0x56, 0x42, 0x6b, 0x24,
}

// Desktop sections leave the second file GUID empty. OneStore web exports
// repeat the first GUID there; decide this before trying raw desktop records,
// even if later FSSHTTPB fields are malformed.
func oneNoteWebEnvelope(data []byte) bool {
	return len(data) >= 48 && bytes.Equal(data[16:32], data[32:48])
}

func (s *scanState) oneNoteAttachments(data []byte, depth int) (string, error) {
	var incomplete error
	for from := 0; from+len(oneNoteDataHeader) <= len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		rel := bytes.Index(data[from:], oneNoteDataHeader[:])
		if rel < 0 {
			break
		}
		start := from + rel
		from = start + 1
		if len(data)-start < 52 {
			continue
		}
		// Both reserved fields are required to be zero by the format. A
		// matching GUID in arbitrary file bytes cannot by itself create a
		// child object.
		if !allZero(data[start+24 : start+36]) {
			continue
		}
		size := binary.LittleEndian.Uint64(data[start+16 : start+24])
		// Write the bound as subtraction, never as a potentially wrapping
		// addition of attacker-controlled cbLength.
		available := uint64(len(data) - start - 52)
		if size > available {
			continue
		}
		padding := (8 - (size+4)%8) % 8
		if padding > available-size {
			continue
		}
		payloadStart := start + 36
		payloadEnd := payloadStart + int(size)
		footerStart := payloadEnd + int(padding)
		if !bytes.Equal(data[footerStart:footerStart+16], oneNoteDataFooter[:]) {
			continue
		}
		if sig, err := s.visit(data[payloadStart:payloadEnd], depth+1); sig != "" {
			return sig, nil
		} else if err != nil && incomplete == nil {
			incomplete = err
		}
		// A validated record has a defined extent. Nested records inside
		// its payload are handled by the child scan under the same limits.
		from = footerStart + 16
	}
	return "", incomplete
}

func allZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// oneNoteFSSHTTPB locates the OneStore web-export package described by
// MS-ONESTORE 2.8.1. The same file-type GUID also starts desktop sections, so
// the repeated file GUID, padding and Packaging Start object distinguish it.
// The stream-object walk below finds its actual terminator; a matching byte
// sequence inside an object's declared body is not a package terminator.
func oneNoteFSSHTTPB(data []byte) (int, bool) {
	if len(data) < 72 || !oneNoteWebEnvelope(data) || !allZero(data[64:68]) {
		return 0, false
	}
	h := binary.LittleEndian.Uint32(data[68:72])
	if h&3 != 2 || (h>>3)&0x3fff != 0x7a {
		return 0, false
	}
	p, ok := fssExGuidEnd(data, 72) // storage index
	if !ok || len(data)-p < 19 {    // cell-schema GUID, package start, reserved
		return 0, false
	}
	p += 16
	start := binary.LittleEndian.Uint16(data[p:])
	if start&0x1ff != 0xac || start>>9 != 1 || data[p+2] != 0 {
		return 0, false
	}
	p += 3
	return p, true
}

// fssCompact decodes MS-FSSHTTPB CompactU64 without advancing outside data.
// The low set bit selects the encoded width; 0x80 prefixes a full uint64.
func fssCompact(data []byte, at int) (uint64, int, bool) {
	if at < 0 || at >= len(data) {
		return 0, at, false
	}
	first := data[at]
	if first == 0 {
		return 0, at + 1, true
	}
	shift := bits.TrailingZeros8(first)
	if shift == 7 {
		if len(data)-at < 9 {
			return 0, at, false
		}
		return binary.LittleEndian.Uint64(data[at+1 : at+9]), at + 9, true
	}
	width := shift + 1
	if len(data)-at < width {
		return 0, at, false
	}
	var value uint64
	for i := 0; i < width; i++ {
		value |= uint64(data[at+i]) << (8 * i)
	}
	return value >> width, at + width, true
}

func fssExGuidEnd(data []byte, at int) (int, bool) {
	if at < 0 || at >= len(data) {
		return at, false
	}
	first := data[at]
	width := 0
	switch {
	case first == 0:
		width = 1
	case first&7 == 4:
		width = 17
	case first&63 == 32:
		width = 18
	case first&127 == 64:
		width = 19
	case first == 128:
		width = 21
	}
	return at + width, width != 0 && len(data)-at >= width
}

func fssSerialEnd(data []byte, at int) (int, bool) {
	if at < 0 || at >= len(data) {
		return at, false
	}
	width := 25
	if data[at] == 0 {
		width = 1
	}
	return at + width, len(data)-at >= width
}

type fssToken struct {
	typ      uint32
	length   uint64
	next     int
	compound bool
	end      bool
}

// fssReadToken reads the four MS-FSSHTTPB stream-object header forms. A start
// header's length covers bytes before the next header, not its descendants.
func fssReadToken(data []byte, at int) (fssToken, bool) {
	if at < 0 || at >= len(data) {
		return fssToken{}, false
	}
	switch data[at] & 3 {
	case 1: // 8-bit end
		return fssToken{typ: uint32(data[at] >> 2), next: at + 1, end: true}, true
	case 3: // 16-bit end
		if len(data)-at < 2 {
			return fssToken{}, false
		}
		return fssToken{typ: uint32(binary.LittleEndian.Uint16(data[at:]) >> 2), next: at + 2, end: true}, true
	case 0: // 16-bit start
		if len(data)-at < 2 {
			return fssToken{}, false
		}
		h := binary.LittleEndian.Uint16(data[at:])
		return fssToken{typ: uint32((h >> 3) & 63), length: uint64(h >> 9), next: at + 2, compound: h&4 != 0}, true
	case 2: // 32-bit start, optionally extended length
		if len(data)-at < 4 {
			return fssToken{}, false
		}
		h := binary.LittleEndian.Uint32(data[at:])
		length := uint64(h >> 17)
		next := at + 4
		if length == 0x7fff {
			var ok bool
			length, next, ok = fssCompact(data, next)
			if !ok {
				return fssToken{}, false
			}
		}
		return fssToken{typ: (h >> 3) & 0x3fff, length: length, next: next, compound: h&4 != 0}, true
	}
	return fssToken{}, false
}

type fssSpan struct{ start, end int }

// The OneNote parser used by ClamAV 1.5.4 rejects stream-object IDs outside
// this set instead of skipping their declared length and resuming at a later
// attachment. Accepting unknown IDs could turn a malformed file into a
// false-positive child scan.
func fssKnownObjectType(typ uint32) bool {
	switch typ {
	case 0x01, 0x02, 0x03, 0x05, 0x07, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e,
		0x11, 0x15, 0x16, 0x18, 0x19, 0x1a, 0x1c, 0x1d, 0x1e,
		0x6a, 0x78, 0x79, 0x7a:
		return true
	}
	return false
}

// OneStore web exports can store attachment bytes in Object Data BLOB data
// elements. Walk every stream object by its declared length and compound boundary so a
// fake Data Element inside an Object Group's raw bytes cannot be treated as a
// sibling. Validate the entire package before visiting children; a malformed
// tail cannot make a partial parse produce an alert the reference would miss.
// Even a graph-reachable BLOB is only a candidate until OneNote page and
// embedded-file property references prove that it is an attachment.
func (s *scanState) oneNoteBlobs(data []byte, depth int) (string, error) {
	pos, ok := oneNoteFSSHTTPB(data)
	if !ok {
		return "", nil
	}
	packagePos := pos
	stack := []uint32{0x15} // Data Element Package start was validated above
	var elementType uint64
	for objects := 0; objects < 1_000_000; objects++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		token, ok := fssReadToken(data, pos)
		if !ok || !fssKnownObjectType(token.typ) {
			return "", ErrMalformed
		}
		if token.end {
			if len(stack) == 0 || stack[len(stack)-1] != token.typ {
				return "", ErrMalformed
			}
			stack = stack[:len(stack)-1]
			pos = token.next
			if token.typ == 1 {
				elementType = 0
			}
			if len(stack) == 0 {
				// The enclosing OneNote Packaging object uses a 16-bit end.
				last, valid := fssReadToken(data, pos)
				if !valid || !last.end || last.typ != 0x7a || last.next != pos+2 {
					return "", ErrMalformed
				}
				// The length-delimited frame is only the first validation
				// layer. Parse typed elements and the storage graph before
				// accepting any BLOB as a possible child.
				tree, err := fssPackageTree(s.ctx, data, packagePos)
				if err != nil {
					return "", err
				}
				graph, err := parseFSSGraph(tree)
				if err != nil {
					return "", err
				}
				if _, err := graph.reachableBlobs(); err != nil {
					return "", err
				}
				attachments, err := oneNoteWebCandidates(s.ctx, graph)
				if err != nil {
					return "", err
				}
				var incomplete error
				for _, id := range attachments {
					payload, ok := graph.blobs[id]
					if !ok {
						return "", ErrMalformed
					}
					if sig, err := s.visit(payload, depth+1); sig != "" {
						return sig, nil
					} else if err != nil && incomplete == nil {
						incomplete = err
					}
				}
				return "", incomplete
			}
			continue
		}
		bodyEnd, ok := fssLeafEnd(data, token, stack[len(stack)-1])
		if !ok {
			return "", ErrMalformed
		}
		if token.typ == 1 && len(stack) == 1 && stack[0] == 0x15 {
			if !token.compound {
				return "", ErrMalformed
			}
			p, good := fssExGuidEnd(data, token.next)
			if !good {
				return "", ErrMalformed
			}
			p, good = fssSerialEnd(data, p)
			if !good {
				return "", ErrMalformed
			}
			elementType, p, good = fssCompact(data, p)
			if !good || p != bodyEnd {
				return "", ErrMalformed
			}
		} else if token.typ == 2 && len(stack) == 2 && stack[1] == 1 && elementType == 0x0a {
			if token.compound {
				return "", ErrMalformed
			}
			size, payload, good := fssCompact(data, token.next)
			if !good || payload > bodyEnd || size != uint64(bodyEnd-payload) {
				return "", ErrMalformed
			}
			// An Object Data BLOB is followed by the Data Element end.
			if bodyEnd >= len(data) || data[bodyEnd] != 0x05 {
				return "", ErrMalformed
			}
		}
		pos = bodyEnd
		if token.compound {
			if len(stack) >= 64 {
				return "", ErrLimit
			}
			stack = append(stack, token.typ)
		}
	}
	return "", ErrLimit
}
