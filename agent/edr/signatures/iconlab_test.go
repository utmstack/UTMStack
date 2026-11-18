package signatures

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math/rand"
)

// Generated icon programs for the icon oracle tests. The generator is
// deterministic (math/rand sources are stable), so the programs the
// reference hashed are rebuilt byte for byte from their seeds.

// labIcon is one generated icon: an image of w x h pixels at a bit depth,
// with a palette for depths up to 8, an AND mask for depths below 32, or a
// PNG image for a 256-pixel entry.
type labIcon struct {
	w, bpp      int
	h           int       // optional height; zero means a square icon
	headerExtra int       // additional bytes in a V4/V5-style bitmap header
	palette     [][3]byte // depths 1, 4 and 8
	index       []int     // palette index per pixel, top-down
	raw16       []uint16  // reference 16-bit bitmap words, top-down
	rgba        [][4]byte // depths 24 and 32 (alpha used by 32), top-down
	mask        []bool    // true is transparent (depths below 32)
	png         []byte
}

// bitmap is the icon's resource data: a 40-byte bitmap header, the palette,
// the color rows bottom-up, then the AND mask rows.
func (ic labIcon) bitmap() []byte {
	if ic.png != nil {
		return ic.png
	}
	le := binary.LittleEndian
	h := ic.w
	if ic.h != 0 {
		h = ic.h
	}
	var b bytes.Buffer
	hdr := make([]byte, 40+ic.headerExtra)
	le.PutUint32(hdr[0:], uint32(len(hdr)))
	le.PutUint32(hdr[4:], uint32(ic.w))
	le.PutUint32(hdr[8:], uint32(2*h))
	le.PutUint16(hdr[12:], 1)
	le.PutUint16(hdr[14:], uint16(ic.bpp))
	b.Write(hdr)
	for _, c := range ic.palette {
		b.Write([]byte{c[2], c[1], c[0], 0})
	}
	stride := (ic.w*ic.bpp + 31) / 32 * 4
	for y := h - 1; y >= 0; y-- {
		row := make([]byte, stride)
		for x := 0; x < ic.w; x++ {
			i := y*ic.w + x
			switch ic.bpp {
			case 1, 4, 8:
				bit := x * ic.bpp
				row[bit/8] |= byte(ic.index[i] << (8 - ic.bpp - bit%8))
			case 16:
				le.PutUint16(row[2*x:], ic.raw16[i])
			case 24:
				p := ic.rgba[i]
				copy(row[3*x:], []byte{p[2], p[1], p[0]})
			case 32:
				p := ic.rgba[i]
				copy(row[4*x:], []byte{p[2], p[1], p[0], p[3]})
			}
		}
		b.Write(row)
	}
	maskStride := (ic.w + 31) / 32 * 4
	for y := h - 1; y >= 0; y-- {
		row := make([]byte, maskStride)
		for x := 0; x < ic.w; x++ {
			if ic.mask != nil && ic.mask[y*ic.w+x] {
				row[x/8] |= 0x80 >> (x % 8)
			}
		}
		b.Write(row)
	}
	return b.Bytes()
}

// labProgram is a 32-bit program whose resources hold the icons (type 3,
// ids 1..n) and one group icon (type 14) listing them in order.
func labProgram(icons []labIcon) []byte {
	le := binary.LittleEndian
	const rva = 0x1000
	n := len(icons)
	datas := make([][]byte, n+1)
	group := make([]byte, 6, 6+14*n)
	le.PutUint16(group[2:], 1)
	le.PutUint16(group[4:], uint16(n))
	for i, ic := range icons {
		datas[i] = ic.bitmap()
		e := make([]byte, 14)
		h := ic.w
		if ic.h != 0 {
			h = ic.h
		}
		e[0], e[1] = byte(ic.w), byte(h)
		if ic.bpp <= 8 {
			e[2] = byte(len(ic.palette))
		}
		le.PutUint16(e[4:], 1)
		le.PutUint16(e[6:], uint16(ic.bpp))
		le.PutUint32(e[8:], uint32(len(datas[i])))
		le.PutUint16(e[12:], uint16(i+1))
		group = append(group, e...)
	}
	datas[n] = group
	dirSize := func(entries int) int { return 16 + 8*entries }
	align := func(v, a int) int { return (v + a - 1) / a * a }
	off := dirSize(2)
	icDir := off
	off += dirSize(n)
	grDir := off
	off += dirSize(1)
	langs := make([]int, n+1)
	for i := range langs {
		langs[i] = off
		off += dirSize(1)
	}
	dataEntries := make([]int, n+1)
	for i := range dataEntries {
		dataEntries[i] = off
		off += 16
	}
	dataAt := make([]int, n+1)
	for i := range dataAt {
		off = align(off, 4)
		dataAt[i] = off
		off += len(datas[i])
	}
	sec := make([]byte, off)
	putEntry := func(dir, i int, id, target uint32) {
		le.PutUint16(sec[dir+14:], uint16(max(i+1, int(le.Uint16(sec[dir+14:])))))
		le.PutUint32(sec[dir+16+8*i:], id)
		le.PutUint32(sec[dir+20+8*i:], target)
	}
	putEntry(0, 0, 3, 0x80000000|uint32(icDir))
	putEntry(0, 1, 14, 0x80000000|uint32(grDir))
	for i := 0; i < n; i++ {
		putEntry(icDir, i, uint32(i+1), 0x80000000|uint32(langs[i]))
	}
	putEntry(grDir, 0, 1, 0x80000000|uint32(langs[n]))
	for i := 0; i <= n; i++ {
		putEntry(langs[i], 0, 0x409, uint32(dataEntries[i]))
		le.PutUint32(sec[dataEntries[i]:], uint32(rva+dataAt[i]))
		le.PutUint32(sec[dataEntries[i]+4:], uint32(len(datas[i])))
		copy(sec[dataAt[i]:], datas[i])
	}
	raw := align(len(sec), 0x200)
	img := make([]byte, 0x200+raw)
	copy(img, "MZ")
	le.PutUint32(img[0x3c:], 0x40)
	copy(img[0x40:], "PE\x00\x00")
	c := 0x44
	le.PutUint16(img[c:], 0x14c)
	le.PutUint16(img[c+2:], 1)
	le.PutUint16(img[c+16:], 0xe0)
	le.PutUint16(img[c+18:], 0x0102)
	o := c + 20
	le.PutUint16(img[o:], 0x10b)
	img[o+2] = 6
	le.PutUint32(img[o+4:], uint32(raw))
	le.PutUint32(img[o+16:], rva)
	le.PutUint32(img[o+20:], rva)
	le.PutUint32(img[o+28:], 0x400000)
	le.PutUint32(img[o+32:], 0x1000)
	le.PutUint32(img[o+36:], 0x200)
	le.PutUint16(img[o+40:], 4)
	le.PutUint16(img[o+48:], 4)
	le.PutUint32(img[o+56:], uint32(rva+align(len(sec), 0x1000)))
	le.PutUint32(img[o+60:], 0x200)
	le.PutUint16(img[o+68:], 2)
	le.PutUint32(img[o+72:], 0x100000)
	le.PutUint32(img[o+76:], 0x1000)
	le.PutUint32(img[o+80:], 0x100000)
	le.PutUint32(img[o+84:], 0x1000)
	le.PutUint32(img[o+92:], 16)
	le.PutUint32(img[o+112:], rva)
	le.PutUint32(img[o+116:], uint32(len(sec)))
	s := o + 0xe0
	copy(img[s:], ".rsrc")
	le.PutUint32(img[s+8:], uint32(len(sec)))
	le.PutUint32(img[s+12:], rva)
	le.PutUint32(img[s+16:], uint32(raw))
	le.PutUint32(img[s+20:], 0x200)
	le.PutUint32(img[s+36:], 0x60000060)
	copy(img[0x200:], sec)
	return img
}

// labPaint fills an image with a background and a few rectangles. gray
// keeps every color gray; sparse colors only a few pixels.
func labPaint(r *rand.Rand, w int, gray, sparse bool) [][3]byte {
	pick := func() [3]byte {
		if gray {
			v := byte(r.Intn(256))
			return [3]byte{v, v, v}
		}
		return [3]byte{byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256))}
	}
	px := make([][3]byte, w*w)
	bg := pick()
	if sparse {
		v := byte(r.Intn(256))
		bg = [3]byte{v, v, v}
	}
	for i := range px {
		px[i] = bg
	}
	if sparse {
		// colored pixels near the 5% share the spread depends on
		n := w * w * (3 + r.Intn(6)) / 100
		for k := 0; k < n; k++ {
			px[r.Intn(w*w)] = [3]byte{byte(128 + r.Intn(128)), byte(r.Intn(64)), byte(r.Intn(256))}
		}
		return px
	}
	for k := 0; k < 1+r.Intn(5); k++ {
		x0, y0 := r.Intn(w), r.Intn(w)
		x1, y1 := min(w, x0+1+r.Intn(w/2+1)), min(w, y0+1+r.Intn(w/2+1))
		c := pick()
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				px[y*w+x] = c
			}
		}
	}
	return px
}

// labMakeIcon builds an icon of the depth from painted pixels.
func labMakeIcon(r *rand.Rand, w, bpp int, gray, sparse bool) labIcon {
	px := labPaint(r, w, gray, sparse)
	ic := labIcon{w: w, bpp: bpp}
	transparent := r.Intn(3) == 0
	switch bpp {
	case 1, 4, 8:
		size := 1 << bpp
		index := map[[3]byte]int{}
		ic.index = make([]int, w*w)
		for i, c := range px {
			j, ok := index[c]
			if !ok {
				if len(ic.palette) < size {
					j = len(ic.palette)
					ic.palette = append(ic.palette, c)
				} else {
					j = r.Intn(size)
				}
				index[c] = j
			}
			ic.index[i] = j
		}
		for len(ic.palette) < size {
			ic.palette = append(ic.palette, [3]byte{byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256))})
		}
	default:
		ic.rgba = make([][4]byte, w*w)
		for i, c := range px {
			a := byte(255)
			if bpp == 32 && transparent {
				a = []byte{0, 255, 128, 64, 200}[r.Intn(5)]
			}
			ic.rgba[i] = [4]byte{c[0], c[1], c[2], a}
		}
	}
	if bpp != 32 && transparent {
		ic.mask = make([]bool, w*w)
		for i := range ic.mask {
			ic.mask[i] = r.Intn(4) == 0
		}
	}
	return ic
}

// labPNG is a 256-pixel PNG icon.
func labPNG() labIcon {
	im := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			im.Set(x, y, color.RGBA{uint8(x), uint8(y), 99, 255})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, im)
	return labIcon{w: 0, bpp: 32, png: b.Bytes()}
}

// labPrograms are the generated programs of the icon oracle: every size
// from 8 to 130 pixels, every bit depth at native, halved and resampled
// sizes, gray and nearly gray icons, and a PNG entry.
func labPrograms() [][]byte {
	var out [][]byte
	for _, icons := range labIconSets() {
		out = append(out, labProgram(icons))
	}
	return out
}

// labIconSets are the icons of each generated program.
func labIconSets() [][]labIcon {
	r := rand.New(rand.NewSource(20261003))
	var all [][]labIcon
	var sizes []labIcon
	for w := 8; w <= 130; w++ {
		sizes = append(sizes, labMakeIcon(r, w, []int{32, 32, 8, 4, 24}[w%5], w%7 == 0, false))
	}
	for len(sizes) > 0 {
		k := min(41, len(sizes))
		all = append(all, sizes[:k])
		sizes = sizes[k:]
	}
	var depths []labIcon
	for _, bpp := range []int{1, 4, 8, 24, 32} {
		for _, w := range []int{16, 24, 32, 48, 20, 40} {
			depths = append(depths, labMakeIcon(r, w, bpp, false, false))
		}
	}
	all = append(all, depths)
	var grays []labIcon
	for i := 0; i < 40; i++ {
		w := []int{16, 24, 32, 64}[i%4]
		bpp := []int{32, 8, 4, 24}[i/4%4]
		grays = append(grays, labMakeIcon(r, w, bpp, i%2 == 0, i%2 == 1))
	}
	all = append(all, grays)
	mixed := []labIcon{labMakeIcon(r, 32, 32, false, false), labPNG(), labMakeIcon(r, 16, 8, false, false)}
	return append(all, mixed)
}
