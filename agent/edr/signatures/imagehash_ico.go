package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
)

var errICO = errors.New("invalid or unsupported ICO image")

// An ICO directory length does not bound the decoder's physical reads. PNG
// requires length >= 8; DIB uses it to decide whether an AND mask is present.
// Keep that value separate from a zero-copy view of the remaining file.
type icoFrame struct {
	data          []byte
	declared      uint64
	width, height int
}

func (f icoFrame) matches(w, h int) bool {
	return w > 0 && h > 0 && min(w, 256) == f.width && min(h, 256) == f.height
}

// Selection follows image-rs 0.25.9: start with the last entry, then visit
// earlier entries in forward order using strict (depth, area) comparison.
// Validate every entry's bounded hints, without reading unselected bodies.
func selectedICOImage(data []byte) (icoFrame, bool) {
	if len(data) < 22 || !bytes.Equal(data[:4], []byte{0, 0, 1, 0}) {
		return icoFrame{}, false
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count == 0 || count > (len(data)-6)/16 {
		return icoFrame{}, false
	}
	dimension := func(v byte) int {
		if v == 0 {
			return 256
		}
		return int(v)
	}
	area := func(entry []byte) int { return dimension(entry[0]) * dimension(entry[1]) }
	best := data[6+16*(count-1) : 6+16*count]
	bestDepth, bestArea := binary.LittleEndian.Uint16(best[6:8]), area(best)
	for i := 0; i < count; i++ {
		entry := data[6+16*i : 6+16*(i+1)]
		planes, depth := binary.LittleEndian.Uint16(entry[4:6]), binary.LittleEndian.Uint16(entry[6:8])
		if planes > 256 || depth > 256 {
			return icoFrame{}, false
		}
		entryArea := area(entry)
		if depth > bestDepth || depth == bestDepth && entryArea > bestArea {
			best, bestDepth, bestArea = entry, depth, entryArea
		}
	}
	start := uint64(binary.LittleEndian.Uint32(best[12:16]))
	if start > uint64(len(data)) || uint64(len(data))-start < 8 {
		return icoFrame{}, false
	}
	return icoFrame{data: data[int(start):], declared: uint64(binary.LittleEndian.Uint32(best[8:12])),
		width: dimension(best[0]), height: dimension(best[1])}, true
}

type icoImage struct {
	frame  icoFrame
	config image.Config
	kind   byte // p: PNG, r: RLE, b: bitfields, u: uncompressed
	depth  int
	end    int // fixed XOR end, or the RLE decoder's actual cursor
	plain  bmpPlainInfo
	fields bmpBitfieldInfo
	rle    bmpRLEInfo
}

func readICOImage(f icoFrame) (*icoImage, error) {
	if bytes.HasPrefix(f.data, []byte("\x89PNG\r\n\x1a\n")) {
		if f.declared < 8 || !icoPNGRGBA8(f.data) {
			return nil, errICO
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(f.data))
		if err != nil || !f.matches(cfg.Width, cfg.Height) {
			return nil, errICO
		}
		return &icoImage{frame: f, config: cfg, kind: 'p', depth: 32}, nil
	}
	return readICODIB(f)
}

// The reference accepts PNGs whose expanded output is RGBA8. That includes
// indexed/RGB transparency and Adam7, but excludes RGB without alpha, gray
// alpha and 16-bit output. Full PNG validation remains with image/png.
func icoPNGRGBA8(data []byte) bool {
	if len(data) < 33 || binary.BigEndian.Uint32(data[8:12]) != 13 || !bytes.Equal(data[12:16], []byte("IHDR")) {
		return false
	}
	depth, typ := data[24], data[25]
	if typ == 6 {
		return depth == 8
	}
	if !(typ == 2 && depth == 8 || typ == 3 && (depth == 1 || depth == 2 || depth == 4 || depth == 8)) {
		return false
	}
	for pos := 33; len(data)-pos >= 12; {
		n := uint64(binary.BigEndian.Uint32(data[pos : pos+4]))
		if n > uint64(len(data)-pos-12) {
			return false
		}
		kind := string(data[pos+4 : pos+8])
		if kind == "IDAT" {
			return false
		}
		if kind == "tRNS" {
			return typ == 2 && n == 6 || typ == 3 && n > 0
		}
		pos += 12 + int(n)
	}
	return false
}

func readICODIB(f icoFrame) (*icoImage, error) {
	b := f.data
	if len(b) < 12 {
		return nil, errICO
	}
	le := binary.LittleEndian
	header := int(le.Uint32(b[:4]))
	var width, height, depth, colors, paletteBytes int
	var compression uint32
	topDown := false
	if header == 12 {
		width, height, depth = int(le.Uint16(b[4:6])), int(le.Uint16(b[6:8]))/2, int(le.Uint16(b[10:12]))
		if le.Uint16(b[8:10]) != 1 || depth != 1 && depth != 4 && depth != 8 && depth != 24 {
			return nil, errICO
		}
		paletteBytes = 3
	} else {
		switch header {
		case 40, 52, 56, 108, 124:
		default:
			return nil, errICO
		}
		if header > len(b) || le.Uint16(b[12:14]) != 1 {
			return nil, errICO
		}
		width = int(int32(le.Uint32(b[4:8])))
		h := int32(le.Uint32(b[8:12]))
		if h == -1<<31 {
			return nil, errICO
		}
		topDown = h < 0
		if topDown {
			h = -h
		}
		height = int(h) / 2
		depth, compression, colors = int(le.Uint16(b[14:16])), le.Uint32(b[16:20]), int(le.Uint32(b[32:36]))
		paletteBytes = 4
	}
	if !f.matches(width, height) || width > imageHashMaxDimension || height > imageHashMaxDimension ||
		int64(width)*int64(height) > imageHashMaxPixels || topDown && compression != 0 && compression != 3 {
		return nil, errICO
	}
	kind := byte('u')
	switch compression {
	case 0:
		switch depth {
		case 1, 2, 4, 8, 24, 32:
		case 16:
			kind = 'b'
		default:
			return nil, errICO
		}
	case 1, 2:
		if compression == 1 && depth != 8 || compression == 2 && depth != 4 {
			return nil, errICO
		}
		kind = 'r'
	case 3:
		if depth != 16 && depth != 32 {
			return nil, errICO
		}
		kind = 'b'
	default:
		return nil, errICO
	}
	out := &icoImage{frame: f, config: image.Config{ColorModel: color.NRGBAModel, Width: width, Height: height}, kind: kind, depth: depth}
	pixelStart := header
	if kind == 'b' {
		masks := [4]uint32{0x7c00, 0x3e0, 0x1f, 0}
		if compression == 3 {
			maskEnd := 52
			if header >= 56 {
				maskEnd = 56
			}
			if maskEnd > len(b) {
				return nil, errICO
			}
			for i := 0; i < 3; i++ {
				masks[i] = le.Uint32(b[40+4*i:])
			}
			if header >= 56 {
				masks[3] = le.Uint32(b[52:56])
			}
			if header == 40 || header == 108 || header == 124 {
				pixelStart += 12 // pinned image-rs V4/V5 cursor behavior
			}
		}
		out.fields = bmpBitfieldInfo{width: width, height: height, bpp: depth, topDown: topDown}
		for i, mask := range masks {
			field, err := bmpChannelFromMask(mask, uint(depth))
			if err != nil || i < 3 && field.width == 0 {
				return nil, errICO
			}
			out.fields.fields[i] = field
		}
	}
	out.plain = bmpPlainInfo{width: width, height: height, bpp: depth, topDown: topDown}
	if depth <= 8 {
		if colors == 0 {
			colors = 1 << depth
		}
		if colors > 1<<depth || pixelStart > len(b) || colors*paletteBytes > len(b)-pixelStart {
			return nil, errICO
		}
		for i := 0; i < colors; i++ {
			p := b[pixelStart+i*paletteBytes:]
			out.plain.palette[i] = [3]byte{p[2], p[1], p[0]}
		}
		pixelStart += colors * paletteBytes
	}
	if pixelStart > len(b) {
		return nil, errICO
	}
	out.plain.offset, out.fields.offset = pixelStart, pixelStart
	if kind == 'r' {
		palette := make(color.Palette, 256)
		for i, rgb := range out.plain.palette {
			palette[i] = color.NRGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}
		}
		out.rle = bmpRLEInfo{width: width, height: height, bpp: depth, offset: pixelStart, palette: palette}
	} else {
		stride := (width*depth + 31) / 32 * 4
		if stride > (len(b)-pixelStart)/height {
			return nil, errICO
		}
		out.end = pixelStart + stride*height
		if !out.maskValid(out.end) {
			return nil, errICO
		}
	}
	return out, nil
}

func (i *icoImage) maskValid(end int) bool {
	mask := uint64((i.config.Width+31)/32*4) * uint64(i.config.Height)
	if i.frame.declared >= uint64(end)+mask {
		return uint64(len(i.frame.data)) >= uint64(end)+mask
	}
	return i.frame.declared == uint64(end)
}

func (i *icoImage) decode(work *matchWork) (image.Image, error) {
	switch i.kind {
	case 'p':
		return png.Decode(bytes.NewReader(i.frame.data))
	case 'u':
		return decodeBMPPlainInfo(i.frame.data, i.plain)
	case 'b':
		return decodeBMPBitfieldsInfo(i.frame.data, i.fields)
	case 'r':
		img, end, err := decodeBMPRLEInfo(i.frame.data, i.rle, work)
		if err != nil {
			return nil, err
		}
		if !i.maskValid(end) {
			return nil, errICO
		}
		return img, nil
	}
	return nil, errICO
}
