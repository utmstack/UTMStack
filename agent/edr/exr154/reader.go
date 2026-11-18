// Package exr154 decodes bounded OpenEXR images for the ClamAV 1.5.4 image
// profile. The runtime is Go only. See NOTICE.md for source attribution.
package exr154

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"math"
	"math/bits"
	"sort"
)

var ErrFormat = errors.New("invalid or unsupported EXR image")
var ErrLimit = errors.New("EXR resource limit")
var le = binary.LittleEndian

type Limits struct {
	MaxDimension, MaxPixels int
	Step                    func(int64) error
}
type box struct{ x, y, w, h int }
type channel struct {
	name       string
	kind, size int
	linear     bool
}
type header struct {
	data, display                box
	channels                     []channel
	compression, blockLines, bpp int
	tileW, tileH, levels         int
	roundUp, deep                bool
	chunks, topChunks            int
	table                        int
}
type Decoder struct {
	data      []byte
	headers   []header
	selected  int
	multipart bool
	limits    Limits
}
type cursor struct {
	data []byte
	pos  int
	err  error
}

func (c *cursor) take(n int) []byte {
	if c.err != nil {
		return nil
	}
	if n < 0 || n > len(c.data)-c.pos {
		c.err = ErrFormat
		return nil
	}
	b := c.data[c.pos : c.pos+n]
	c.pos += n
	return b
}
func (c *cursor) uint() uint32 {
	b := c.take(4)
	if b == nil {
		return 0
	}
	return le.Uint32(b)
}
func (c *cursor) integer() int { return int(int32(c.uint())) }
func (c *cursor) text(max int) string {
	if c.err != nil {
		return ""
	}
	end := bytes.IndexByte(c.data[c.pos:c.pos+min(max+1, len(c.data)-c.pos)], 0)
	if end < 0 || end > max {
		c.err = ErrFormat
		return ""
	}
	b := c.take(end + 1)
	return string(b[:end])
}
func (l Limits) charge(n int64) error {
	if l.Step != nil {
		return l.Step(n)
	}
	return nil
}
func (l Limits) bounded(b box) bool {
	return b.w > 0 && b.h > 0 && b.w <= l.MaxDimension && b.h <= l.MaxDimension && int64(b.w)*int64(b.h) <= int64(l.MaxPixels)
}
func attrBox(b []byte) (box, bool) {
	if len(b) < 16 {
		return box{}, false
	}
	x, y := int(int32(le.Uint32(b))), int(int32(le.Uint32(b[4:])))
	x1, y1 := int(int32(le.Uint32(b[8:]))), int(int32(le.Uint32(b[12:])))
	x, x1 = min(x, x1), max(x, x1)
	y, y1 = min(y, y1), max(y, y1)
	const bound = math.MaxInt32 / 2
	if x <= -bound || y <= -bound || x1 >= bound || y1 >= bound {
		return box{}, false
	}
	return box{x, y, x1 - x + 1, y1 - y + 1}, true
}
func channels(b []byte) ([]channel, error) {
	c := cursor{data: b}
	var out []channel
	for c.err == nil {
		name := c.text(255)
		if c.err != nil {
			return nil, c.err
		}
		if name == "" {
			break
		}
		kind := c.uint()
		v := c.take(4)
		sx, sy := c.integer(), c.integer()
		if c.err != nil || kind > 2 || sx != 1 || sy != 1 || v[0] > 1 {
			return nil, ErrFormat
		}
		if len(out) >= 1024 {
			return nil, ErrLimit
		}
		if len(out) > 0 && out[len(out)-1].name > name {
			return nil, ErrFormat
		}
		size := 4
		if kind == 1 {
			size = 2
		}
		out = append(out, channel{name, int(kind), size, v[0] != 0})
	}
	if len(out) == 0 {
		return nil, ErrFormat
	}
	return out, nil
}
func readHeader(c *cursor, flags uint32, l Limits) (header, error) {
	h := header{compression: -1}
	haveData, haveDisplay := false, false
	hasTile := false
	blockType := ""
	declaredChunks := -1
	deepVersion := -1
	nameLimit := 32
	if flags&0x400 != 0 {
		nameLimit = 256
	}
	for c.err == nil {
		name := c.text(nameLimit)
		if c.err != nil {
			return h, c.err
		}
		if name == "" {
			break
		}
		if len(name) > 255 {
			return h, ErrFormat
		}
		kind := c.text(nameLimit)
		n := c.integer()
		v := c.take(n)
		if c.err != nil {
			return h, c.err
		}
		if kind == "" {
			return h, ErrFormat
		}
		if err := l.charge(int64(n)/64 + 1); err != nil {
			return h, err
		}
		// Known attribute values have minimum sizes; extra bytes are advisory
		// in the reference's non-pedantic image read.
		if size, ok := attributeMinimum[kind]; ok && len(v) < size {
			return h, ErrFormat
		}
		if err := validateAttribute(kind, v); err != nil {
			return h, err
		}
		if kind == "chlist" && name != "channels" {
			if _, err := channels(v); err != nil {
				return h, err
			}
		}
		switch {
		case name == "channels" && kind == "chlist":
			var err error
			h.channels, err = channels(v)
			if err != nil {
				return h, err
			}
		case name == "compression" && kind == "compression":
			if v[0] > 11 {
				return h, ErrFormat
			}
			h.compression = int(v[0])
		case name == "dataWindow" && kind == "box2i":
			h.data, haveData = attrBox(v)
		case name == "displayWindow" && kind == "box2i":
			h.display, haveDisplay = attrBox(v)
		case name == "type" && kind == "string":
			if len(v) > len("scanlineimage") {
				return h, ErrFormat
			}
			blockType = string(v)
			switch blockType {
			case "scanlineimage", "tiledimage", "deepscanline", "deeptile":
			default:
				return h, ErrFormat
			}
		case name == "tiles" && kind == "tiledesc":
			h.tileW, h.tileH = int(le.Uint32(v)), int(le.Uint32(v[4:]))
			h.levels = int(v[8] & 15)
			h.roundUp = v[8]&16 != 0
			if h.tileW <= 0 || h.tileH <= 0 || h.levels > 2 || v[8]&0xe0 != 0 {
				return h, ErrFormat
			}
			hasTile = true
		case name == "chunkCount" && kind == "int":
			declaredChunks = int(int32(le.Uint32(v)))
			if declaredChunks < 0 {
				return h, ErrFormat
			}
		case name == "version" && kind == "int":
			deepVersion = int(int32(le.Uint32(v)))
		case name == "maxSamplesPerPixel" && kind == "int":
			if int32(le.Uint32(v)) < 0 {
				return h, ErrFormat
			}
		case name == "lineOrder" && kind == "lineOrder":
			if v[0] > 2 {
				return h, ErrFormat
			}
		}
	}
	if c.err != nil || h.channels == nil || h.compression < 0 || !haveData || !haveDisplay {
		return h, ErrFormat
	}
	for _, b := range []box{h.data, h.display} {
		if b.x+b.w >= math.MaxInt32/2 || b.y+b.h >= math.MaxInt32/2 {
			return h, ErrFormat
		}
	}
	h.deep = blockType == "deepscanline" || blockType == "deeptile"
	if h.deep && (deepVersion != 1 || h.compression > 2) {
		return h, ErrFormat
	}
	switch blockType {
	case "", "scanlineimage", "tiledimage", "deepscanline", "deeptile":
	default:
		return h, ErrFormat
	}
	tiled := blockType == "tiledimage" || blockType == "deeptile" || blockType == "" && flags&0x200 != 0
	if tiled && !hasTile {
		return h, ErrFormat
	}
	if !tiled {
		h.tileW, h.tileH, h.levels = 0, 0, 0
	}
	for _, ch := range h.channels {
		h.bpp += ch.size
	}
	h.blockLines = 1
	switch h.compression {
	case 3, 5:
		h.blockLines = 16
	case 4, 6, 7, 8, 11:
		h.blockLines = 32
	case 9, 10:
		h.blockLines = 256
	}
	if !l.bounded(h.data) || !l.bounded(h.display) {
		return h, ErrLimit
	}
	if h.tileW == 0 {
		h.chunks = (h.data.h + h.blockLines - 1) / h.blockLines
		h.topChunks = h.chunks
	} else {
		count := func(w, hgt int) int { return ((w + h.tileW - 1) / h.tileW) * ((hgt + h.tileH - 1) / h.tileH) }
		h.topChunks = count(h.data.w, h.data.h)
		if h.levels == 0 {
			h.chunks = h.topChunks
		} else {
			nx, ny := levelCount(h.data.w, h.roundUp), levelCount(h.data.h, h.roundUp)
			if h.levels == 1 {
				for i := 0; i < max(nx, ny); i++ {
					h.chunks += count(levelSize(h.data.w, i, h.roundUp), levelSize(h.data.h, i, h.roundUp))
				}
			} else {
				for y := 0; y < ny; y++ {
					for x := 0; x < nx; x++ {
						h.chunks += count(levelSize(h.data.w, x, h.roundUp), levelSize(h.data.h, y, h.roundUp))
					}
				}
			}
		}
	}
	if h.chunks <= 0 || h.chunks > 1<<20 {
		return h, ErrLimit
	}
	if declaredChunks >= 0 && declaredChunks != h.chunks {
		return h, ErrFormat
	}
	return h, nil
}
func levelCount(n int, up bool) int {
	if up {
		return bits.Len(uint(n-1)) + 1
	}
	return bits.Len(uint(n))
}
func levelSize(n, level int, up bool) int {
	if up {
		return max(1, (n+(1<<level)-1)>>level)
	}
	return max(1, n>>level)
}

func Read(data []byte, l Limits) (*Decoder, error) {
	if l.MaxDimension <= 0 {
		l.MaxDimension = 1 << 14
	}
	if l.MaxPixels <= 0 {
		l.MaxPixels = 40 << 20
	}
	if len(data) < 8 || !bytes.Equal(data[:4], []byte{0x76, 0x2f, 0x31, 1}) {
		return nil, ErrFormat
	}
	flags := le.Uint32(data[4:8])
	if flags&15 != 2 || flags>>13 != 0 || flags&0x200 != 0 && flags&0x1800 != 0 {
		return nil, ErrFormat
	}
	d := &Decoder{data: data, selected: -1, multipart: flags&0x1000 != 0, limits: l}
	c := cursor{data: data, pos: 8}
	for {
		if err := l.charge(1); err != nil {
			return nil, err
		}
		h, err := readHeader(&c, flags, l)
		if err != nil {
			return nil, err
		}
		// The image adapter selects RGB layers, but its pinned dependency
		// rejects any deep-data header before layer selection, even if an
		// ordinary RGB layer follows it.
		if h.deep {
			return nil, ErrFormat
		}
		if len(d.headers) >= 256 {
			return nil, ErrLimit
		}
		d.headers = append(d.headers, h)
		if !d.multipart {
			break
		}
		if c.pos >= len(data) {
			return nil, ErrFormat
		}
		if data[c.pos] == 0 {
			c.pos++
			break
		}
	}
	for i := range d.headers {
		h := &d.headers[i]
		h.table = c.pos
		if h.chunks > (len(data)-c.pos)/8 {
			return nil, ErrFormat
		}
		c.pos += h.chunks * 8
		if !h.deep && d.selected < 0 {
			var have [3]bool
			for _, ch := range h.channels {
				switch ch.name {
				case "R":
					have[0] = true
				case "G":
					have[1] = true
				case "B":
					have[2] = true
				}
			}
			if have == [3]bool{true, true, true} {
				d.selected = i
			}
		}
	}
	if d.selected < 0 {
		return nil, ErrFormat
	}
	return d, nil
}
func (d *Decoder) Config() image.Config {
	h := d.headers[d.selected]
	return image.Config{ColorModel: color.NRGBAModel, Width: h.display.w, Height: h.display.h}
}

func (d *Decoder) Decode() (image.Image, error) {
	h := d.headers[d.selected]
	if err := d.limits.charge(int64(h.display.w)*int64(h.display.h)/8 + 1); err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, h.display.w, h.display.h))
	for i := 3; i < len(out.Pix); i += 4 {
		out.Pix[i] = 255
	}
	if err := d.limits.charge(int64(h.topChunks)*int64(bits.Len(uint(h.topChunks))+1)/8 + 1); err != nil {
		return nil, err
	}
	offsets := make([]uint64, h.topChunks)
	for i := range offsets {
		offsets[i] = le.Uint64(d.data[h.table+8*i:])
	}
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	for _, offset := range offsets {
		if err := d.limits.charge(1); err != nil {
			return nil, err
		}
		if offset > uint64(len(d.data)) {
			return nil, ErrFormat
		}
		c := cursor{data: d.data, pos: int(offset)}
		if d.multipart && c.integer() != d.selected {
			return nil, ErrFormat
		}
		x, y, w, ht := 0, 0, h.data.w, h.blockLines
		if h.tileW > 0 {
			tx, ty, lx, ly := c.integer(), c.integer(), c.integer(), c.integer()
			if tx < 0 || ty < 0 || lx < 0 || ly < 0 || lx > 31 || ly > 31 {
				return nil, ErrFormat
			}
			lw, lh := levelSize(h.data.w, lx, h.roundUp), levelSize(h.data.h, ly, h.roundUp)
			if tx >= (lw+h.tileW-1)/h.tileW || ty >= (lh+h.tileH-1)/h.tileH {
				return nil, ErrFormat
			}
			x, y = tx*h.tileW, ty*h.tileH
			w, ht = min(h.tileW, lw-x), min(h.tileH, lh-y)
		} else {
			yy := c.integer()
			// Block indices use signed division toward zero, including a
			// negative coordinate less than one block above the data window.
			y = ((yy - h.data.y) / h.blockLines) * h.blockLines
			if y < 0 || y >= h.data.h {
				return nil, ErrFormat
			}
			ht = min(ht, h.data.h-y)
		}
		n := c.integer()
		maxBlock := uint64(h.data.w) * uint64(h.blockLines) * uint64(h.bpp)
		if h.tileW > 0 {
			// The pinned 64-bit reader's header arithmetic wraps here;
			// physical reads and actual allocations stay bounded below.
			maxBlock = uint64(h.tileW) * uint64(h.tileH) * uint64(h.bpp)
		}
		if n < 0 || uint64(n) > maxBlock {
			return nil, ErrFormat
		}
		compressed := c.take(n)
		if c.err != nil {
			return nil, c.err
		}
		expected := int64(w) * int64(ht) * int64(h.bpp)
		if expected <= 0 || expected > 64<<20 {
			return nil, ErrLimit
		}
		if err := d.limits.charge(expected/8 + int64(len(compressed))/16 + 1); err != nil {
			return nil, err
		}
		raw, err := unpack(compressed, h, w, ht, int(expected), d.limits)
		if err != nil {
			return nil, err
		}
		if len(raw) != int(expected) {
			return nil, ErrFormat
		}
		pos := 0
		for row := 0; row < ht; row++ {
			if err := d.limits.charge(int64(w)/8 + 1); err != nil {
				return nil, err
			}
			oy := h.data.y + y + row - h.display.y
			var seen [3]bool
			for _, ch := range h.channels {
				rgb := -1
				switch ch.name {
				case "R":
					rgb = 0
				case "G":
					rgb = 1
				case "B":
					rgb = 2
				}
				// Specific-channel reads select the first physical occurrence,
				// even though header lookup may binary-search duplicate names.
				if rgb >= 0 {
					if seen[rgb] {
						rgb = -1
					} else {
						seen[rgb] = true
					}
				}
				line := raw[pos : pos+w*ch.size]
				pos += len(line)
				if rgb < 0 || oy < 0 || oy >= h.display.h {
					continue
				}
				for col := 0; col < w; col++ {
					ox := h.data.x + x + col - h.display.x
					if ox < 0 || ox >= h.display.w {
						continue
					}
					out.Pix[oy*out.Stride+ox*4+rgb] = sample8(line[col*ch.size:], ch.kind)
				}
			}
		}
	}
	return out, nil
}
func sample8(b []byte, kind int) byte {
	var f float32
	switch kind {
	case 0:
		f = float32(le.Uint32(b))
	case 1:
		f = halfFloat(le.Uint16(b))
	case 2:
		f = math.Float32frombits(le.Uint32(b))
	}
	if !(f < 1) {
		return 255
	}
	if f <= 0 {
		return 0
	}
	return byte(math.Round(float64(float32(f * 255))))
}
func halfFloat(h uint16) float32 {
	sign := uint32(h&0x8000) << 16
	exponent := int(h >> 10 & 31)
	fraction := uint32(h & 1023)
	if exponent == 31 {
		return math.Float32frombits(sign | 0x7f800000 | fraction<<13)
	}
	if exponent == 0 {
		if fraction == 0 {
			return math.Float32frombits(sign)
		}
		shift := bits.LeadingZeros32(fraction) - 21
		fraction = (fraction << shift) & 1023
		return math.Float32frombits(sign | uint32(113-shift)<<23 | fraction<<13)
	}
	return math.Float32frombits(sign | uint32(exponent+112)<<23 | fraction<<13)
}

var attributeMinimum = map[string]int{"box2i": 16, "box2f": 16, "int": 4, "float": 4, "double": 8, "v2i": 8, "v2f": 8, "v3i": 12, "v3f": 12, "rational": 8, "chromaticities": 32, "compression": 1, "lineOrder": 1, "envmap": 1, "tiledesc": 9, "m33f": 36, "m44f": 64, "timecode": 8, "keycode": 28, "preview": 8, "bytes": 4}

func validateAttribute(kind string, v []byte) error {
	switch kind {
	case "box2i":
		if _, ok := attrBox(v); !ok {
			return ErrFormat
		}
	case "compression":
		if v[0] > 11 {
			return ErrFormat
		}
	case "lineOrder":
		if v[0] > 2 {
			return ErrFormat
		}
	case "envmap":
		if v[0] > 1 {
			return ErrFormat
		}
	case "tiledesc":
		w, h := le.Uint32(v), le.Uint32(v[4:])
		if w == 0 || h == 0 || w >= math.MaxInt32/2 || h >= math.MaxInt32/2 || v[8]&15 > 2 || v[8]>>4 > 1 {
			return ErrFormat
		}
	case "preview":
		w, h := uint64(le.Uint32(v)), uint64(le.Uint32(v[4:]))
		if w != 0 && h > uint64(len(v)-8)/4/w {
			return ErrFormat
		}
	case "stringvector":
		c := cursor{data: v}
		for c.pos < len(v) && c.err == nil {
			c.take(c.integer())
		}
		if c.err != nil {
			return c.err
		}
	case "bytes":
		if uint64(le.Uint32(v)) > uint64(len(v)-4) {
			return ErrFormat
		}
	}
	return nil
}
