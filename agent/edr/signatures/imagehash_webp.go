package signatures

import (
	"bytes"
	"encoding/binary"
)

// webpLosslessProfile admits the two still-image VP8L layouts checked
// against the pinned reference: a single simple chunk or a VP8X container
// with optional ICC, EXIF and XMP chunks.
func webpLosslessProfile(data []byte) bool {
	return simpleWebPLossless(data) || extendedWebPLossless(data)
}

// webpLossyProfile admits bounded still VP8 with or without a separate
// alpha plane. Animation and unknown layouts remain incomplete.
func webpLossyProfile(data []byte) bool {
	return simpleWebPLossy(data) || extendedWebPLossy(data) || extendedWebPAlphaLossy(data)
}

func webpRiffExact(data []byte) bool {
	return len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) &&
		bytes.Equal(data[8:12], []byte("WEBP")) &&
		uint64(len(data)) == 8+uint64(binary.LittleEndian.Uint32(data[4:8]))
}

// simpleWebPLossless requires a complete RIFF file with one VP8L chunk.
func simpleWebPLossless(data []byte) bool {
	if !webpRiffExact(data) || len(data) < 25 || !bytes.Equal(data[12:16], []byte("VP8L")) || data[20] != 0x2f {
		return false
	}
	chunkSize := uint64(binary.LittleEndian.Uint32(data[16:20]))
	return uint64(len(data)) == 20+chunkSize+(chunkSize&1) &&
		(chunkSize&1 == 0 || data[len(data)-1] == 0)
}

// webpChunkAt keeps every size calculation in uint64 before slicing, so
// malformed four-byte lengths cannot wrap an int or escape the RIFF body.
func webpChunkAt(data []byte, off int) (string, []byte, int, bool) {
	if off < 12 || off > len(data)-8 {
		return "", nil, 0, false
	}
	size := uint64(binary.LittleEndian.Uint32(data[off+4 : off+8]))
	end := uint64(off) + 8 + size + (size & 1)
	if end > uint64(len(data)) || size&1 != 0 && data[end-1] != 0 {
		return "", nil, 0, false
	}
	return string(data[off : off+4]), data[off+8 : off+8+int(size)], int(end), true
}

// extendedWebPLossless accepts a VP8X still image with an optional ICCP
// chunk before VP8L and at most one EXIF and XMP chunk after it. Requiring
// flags, dimensions, chunk order and terminal length to agree avoids
// deciding hashes on unvalidated or ambiguous containers.
func extendedWebPLossless(data []byte) bool {
	if !webpRiffExact(data) {
		return false
	}
	kind, header, pos, ok := webpChunkAt(data, 12)
	if !ok || kind != "VP8X" || len(header) != 10 {
		return false
	}
	flags := header[0]
	if flags&^0x3c != 0 || header[1] != 0 || header[2] != 0 || header[3] != 0 {
		return false
	}
	canvasWidth := 1 + int(header[4]) + (int(header[5]) << 8) + (int(header[6]) << 16)
	canvasHeight := 1 + int(header[7]) + (int(header[8]) << 8) + (int(header[9]) << 16)
	if canvasWidth <= 0 || canvasHeight <= 0 || canvasWidth > 512 || canvasHeight > 512 {
		return false
	}
	if flags&0x20 != 0 {
		kind, _, pos, ok = webpChunkAt(data, pos)
		if !ok || kind != "ICCP" {
			return false
		}
	}
	kind, body, pos, ok := webpChunkAt(data, pos)
	if !ok || kind != "VP8L" || len(body) < 5 || body[0] != 0x2f {
		return false
	}
	bits := binary.LittleEndian.Uint32(body[1:5])
	if bits>>29 != 0 || 1+int(bits&0x3fff) != canvasWidth ||
		1+int((bits>>14)&0x3fff) != canvasHeight || (bits&(1<<28) != 0) != (flags&0x10 != 0) {
		return false
	}
	seen := byte(0)
	for pos < len(data) {
		kind, _, next, ok := webpChunkAt(data, pos)
		if !ok {
			return false
		}
		switch kind {
		case "EXIF":
			if flags&0x08 == 0 || seen&0x08 != 0 {
				return false
			}
			seen |= 0x08
		case "XMP ":
			if flags&0x04 == 0 || seen&0x04 != 0 {
				return false
			}
			seen |= 0x04
		default:
			return false
		}
		pos = next
	}
	return pos == len(data) && seen == flags&0x0c
}

func vp8StillDimensions(body []byte) (int, int, bool) {
	if len(body) < 10 || body[0]&1 != 0 || !bytes.Equal(body[3:6], []byte{0x9d, 0x01, 0x2a}) {
		return 0, 0, false
	}
	w := int(binary.LittleEndian.Uint16(body[6:8]) & 0x3fff)
	h := int(binary.LittleEndian.Uint16(body[8:10]) & 0x3fff)
	// The 1–32 by 1–32 still VP8 profile is compared against independent
	// native hashes in simple, VP8X and raw-ALPH containers. The shared
	// resize paths retain every differing numeric-profile bit as uncertain.
	regular := w >= 16 && h >= 16 && w <= 512 && h <= 512
	tiny := w >= 1 && h >= 1 && w <= 32 && h <= 32
	return w, h, regular || tiny
}

func simpleWebPLossy(data []byte) bool {
	if !webpRiffExact(data) {
		return false
	}
	kind, body, end, ok := webpChunkAt(data, 12)
	_, _, valid := vp8StillDimensions(body)
	return ok && kind == "VP8 " && end == len(data) && valid
}

func extendedWebPLossy(data []byte) bool { return extendedWebPLossyMode(data, false) }

func extendedWebPAlphaLossy(data []byte) bool { return extendedWebPLossyMode(data, true) }

func extendedWebPLossyMode(data []byte, alpha bool) bool {
	if !webpRiffExact(data) {
		return false
	}
	kind, header, pos, ok := webpChunkAt(data, 12)
	if !ok || kind != "VP8X" || len(header) != 10 {
		return false
	}
	flags := header[0]
	if flags&^0x3c != 0 || (flags&0x10 != 0) != alpha || header[1] != 0 || header[2] != 0 || header[3] != 0 {
		return false
	}
	w := 1 + int(header[4]) + (int(header[5]) << 8) + (int(header[6]) << 16)
	h := 1 + int(header[7]) + (int(header[8]) << 8) + (int(header[9]) << 16)
	if w > 512 || h > 512 {
		return false
	}
	if flags&0x20 != 0 {
		kind, _, pos, ok = webpChunkAt(data, pos)
		if !ok || kind != "ICCP" {
			return false
		}
	}
	if alpha {
		kind, body, next, ok := webpChunkAt(data, pos)
		if !ok || kind != "ALPH" || len(body) < 1 || body[0]&0xe2 != 0 {
			return false
		}
		pos = next
	}
	kind, body, pos, ok := webpChunkAt(data, pos)
	if !ok || kind != "VP8 " {
		return false
	}
	width, height, valid := vp8StillDimensions(body)
	if !valid || width != w || height != h {
		return false
	}
	seen := byte(0)
	for pos < len(data) {
		kind, _, next, ok := webpChunkAt(data, pos)
		if !ok {
			return false
		}
		switch kind {
		case "EXIF":
			if flags&0x08 == 0 || seen&0x08 != 0 {
				return false
			}
			seen |= 0x08
		case "XMP ":
			if flags&0x04 == 0 || seen&0x04 != 0 {
				return false
			}
			seen |= 0x04
		default:
			return false
		}
		pos = next
	}
	return pos == len(data) && seen == flags&0x0c
}

// webpVP8Only keeps the VP8 bitstream while removing the validated ALPH
// wrapper. Its raw YCbCr planes allow the same fancy chroma conversion used
// for opaque VP8; the full alpha image must already have decoded cleanly.
func webpVP8Only(data []byte) []byte {
	for pos := 12; pos < len(data); {
		kind, body, next, ok := webpChunkAt(data, pos)
		if !ok {
			return nil
		}
		if kind == "VP8 " {
			out := make([]byte, 20+len(body)+(len(body)&1))
			copy(out, "RIFF")
			binary.LittleEndian.PutUint32(out[4:8], uint32(len(out)-8))
			copy(out[8:12], "WEBP")
			copy(out[12:16], "VP8 ")
			binary.LittleEndian.PutUint32(out[16:20], uint32(len(body)))
			copy(out[20:], body)
			return out
		}
		pos = next
	}
	return nil
}
