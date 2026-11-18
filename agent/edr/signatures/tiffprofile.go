package signatures

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"math"
	"sort"

	"golang.org/x/image/tiff"
)

var errTIFFProfile = errors.New("unsupported or invalid reference TIFF profile")

// tiffProfile keeps the original immutable object, and a small alternate IFD
// only when CMYK samples must be exposed as four uninterpreted channels to
// the Go TIFF codec. Offset arrays and compressed blocks keep their original
// physical addresses. Neither body rules nor hashes see the alternate view.
type tiffProfile struct {
	data, directory []byte
	header          [8]byte
	cmyk            bool
	floating        *tiffFloat
	jpeg            *tiffJPEG
	work            *matchWork
}

func readTIFFProfile(data []byte, work *matchWork) (*tiffProfile, error) {
	if len(data) < 8 || work.ctx.Err() != nil {
		return nil, errTIFFProfile
	}
	var order binary.ByteOrder
	switch string(data[:4]) {
	case "II*\x00":
		order = binary.LittleEndian
	case "MM\x00*":
		order = binary.BigEndian
	default:
		return nil, errTIFFProfile
	}
	at := uint64(order.Uint32(data[4:8]))
	if at > uint64(len(data)-2) {
		return nil, errTIFFProfile
	}
	count := int(order.Uint16(data[at:]))
	if count > 4096 || uint64(count)*12+2 > uint64(len(data))-at || work.step(int64(count)+1) != nil {
		return nil, errTIFFProfile
	}
	entries := make(map[uint16][]byte, count)
	for i := 0; i < count; i++ {
		e := data[at+2+uint64(i)*12 : at+14+uint64(i)*12]
		tag := order.Uint16(e)
		if entries[tag] != nil {
			return nil, errTIFFProfile
		}
		entries[tag] = e
	}
	values := func(tag uint16, fallback uint32) ([]uint32, error) {
		e := entries[tag]
		if e == nil {
			return []uint32{fallback}, nil
		}
		n := uint64(order.Uint32(e[4:]))
		if n == 0 || n > 16 {
			return nil, errTIFFProfile
		}
		var size uint64
		switch order.Uint16(e[2:]) {
		case 1:
			size = 1
		case 3:
			size = 2
		case 4:
			size = 4
		default:
			return nil, errTIFFProfile
		}
		raw := e[8:12]
		if n*size > 4 {
			off := uint64(order.Uint32(raw))
			if off > uint64(len(data)) || n*size > uint64(len(data))-off {
				return nil, errTIFFProfile
			}
			raw = data[off : off+n*size]
		}
		v := make([]uint32, n)
		for i := range v {
			switch size {
			case 1:
				v[i] = uint32(raw[i])
			case 2:
				v[i] = uint32(order.Uint16(raw[2*i:]))
			case 4:
				v[i] = order.Uint32(raw[4*i:])
			}
		}
		return v, nil
	}
	photo, err := values(262, math.MaxUint32)
	if err != nil || len(photo) != 1 {
		return nil, errTIFFProfile
	}
	bits, err := values(258, 1)
	if err != nil {
		return nil, errTIFFProfile
	}
	for _, b := range bits {
		if b != bits[0] {
			return nil, errTIFFProfile
		}
	}
	samples, err := values(277, 1)
	if err != nil || len(samples) != 1 {
		return nil, errTIFFProfile
	}
	planar, err := values(284, 1)
	if err != nil || len(planar) != 1 || planar[0] != 1 {
		return nil, errTIFFProfile
	}
	floatingFormat := false
	if entries[339] != nil {
		formats, err := values(339, 1)
		if err != nil {
			return nil, errTIFFProfile
		}
		for _, f := range formats {
			if !(f == 1 && bits[0] <= 16 || f == 3 && bits[0] == 32) {
				return nil, errTIFFProfile
			}
		}
		floatingFormat = formats[0] == 3
	}
	valid := false
	switch photo[0] {
	case 0, 1:
		valid = samples[0] == 1 && (bits[0] == 1 || bits[0] == 8 || bits[0] == 16) ||
			samples[0] == 2 && (bits[0] == 8 || bits[0] == 16)
	case 2:
		valid = (samples[0] == 3 || samples[0] == 4) && (bits[0] == 8 || bits[0] == 16 || bits[0] == 32)
	case 5:
		valid = samples[0] == 4 && (bits[0] == 8 || bits[0] == 16) && len(bits) == 4
	}
	if !valid {
		return nil, errTIFFProfile
	}
	if err := validateTIFFPhysicalData(data, entries, order, work); err != nil {
		return nil, err
	}
	p := &tiffProfile{data: data, cmyk: photo[0] == 5, work: work}
	compression, err := values(259, 1)
	if err != nil || len(compression) != 1 {
		return nil, errTIFFProfile
	}
	if compression[0] == 7 {
		p.jpeg, err = readTIFFJPEG(entries, order, p, values, floatingFormat)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	if photo[0] == 2 && bits[0] == 32 {
		if len(bits) != 1 && len(bits) != int(samples[0]) {
			return nil, errTIFFProfile
		}
		p.floating, err = readTIFFFloat(entries, order, p, values, floatingFormat)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	if !p.cmyk {
		return p, nil
	}
	// A private IFD describes four raw channels as unassociated RGBA. This
	// reuses the existing bounded Go raw/LZW/Deflate/PackBits/tile codec; its
	// output channels are CMYK samples, not actual RGBA colors.
	if len(data) > math.MaxUint32-2-12*(count+1)-4 {
		return nil, errTIFFProfile
	}
	shortEntry := func(tag, value uint16) []byte {
		e := make([]byte, 12)
		order.PutUint16(e, tag)
		order.PutUint16(e[2:], 3)
		order.PutUint32(e[4:], 1)
		order.PutUint16(e[8:], value)
		return e
	}
	entries[262] = shortEntry(262, 2)
	entries[338] = shortEntry(338, 2)
	tags := make([]int, 0, len(entries))
	for tag := range entries {
		tags = append(tags, int(tag))
	}
	sort.Ints(tags)
	p.directory = make([]byte, 2+12*len(tags)+4)
	order.PutUint16(p.directory, uint16(len(tags)))
	for i, tag := range tags {
		copy(p.directory[2+12*i:], entries[uint16(tag)])
	}
	copy(p.header[:], data[:8])
	order.PutUint32(p.header[4:], uint32(len(data)))
	return p, nil
}

func (p *tiffProfile) reader() *tiffView {
	return &tiffView{p: p}
}

func (p *tiffProfile) config() (image.Config, error) {
	if p.jpeg != nil {
		return p.jpeg.cfg, nil
	}
	if p.floating != nil {
		return p.floating.cfg, nil
	}
	cfg, err := tiff.DecodeConfig(p.reader())
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > imageHashMaxDimension ||
		cfg.Height > imageHashMaxDimension || int64(cfg.Width)*int64(cfg.Height) > imageHashMaxPixels {
		return image.Config{}, errTIFFProfile
	}
	return cfg, nil
}

func (p *tiffProfile) decode() (image.Image, error) {
	cfg, err := p.config()
	if err != nil || p.work.step(int64(cfg.Width)*int64(cfg.Height)/8+1) != nil {
		return nil, errTIFFProfile
	}
	if p.floating != nil {
		return p.floating.decode(p)
	}
	if p.jpeg != nil {
		return p.jpeg.decode(p)
	}
	img, err := tiff.Decode(p.reader())
	if err != nil || p.work.ctx.Err() != nil {
		return nil, errTIFFProfile
	}
	if !p.cmyk {
		return img, nil
	}
	// image-rs converts CMYK through separate f32 operations, truncates to
	// the original sample depth, then rounds RGB16 to RGB8. Integer division
	// or Go's general CMYK color model does not preserve these boundaries.
	b := img.Bounds()
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if p.work.ctx.Err() != nil || p.work.step(int64(b.Dx())+1) != nil {
			return nil, errTIFFProfile
		}
		for x := b.Min.X; x < b.Max.X; x++ {
			var c, m, yellow, k float32
			var maxValue float32
			switch raw := img.(type) {
			case *image.NRGBA:
				i := raw.PixOffset(x, y)
				c, m, yellow, k = float32(raw.Pix[i]), float32(raw.Pix[i+1]), float32(raw.Pix[i+2]), float32(raw.Pix[i+3])
				maxValue = 255
			case *image.NRGBA64:
				i := raw.PixOffset(x, y)
				c, m, yellow, k = float32(binary.BigEndian.Uint16(raw.Pix[i:])), float32(binary.BigEndian.Uint16(raw.Pix[i+2:])), float32(binary.BigEndian.Uint16(raw.Pix[i+4:])), float32(binary.BigEndian.Uint16(raw.Pix[i+6:]))
				maxValue = 65535
			default:
				return nil, errTIFFProfile
			}
			factor := float32(1) - k/maxValue
			channel := func(v float32) uint8 {
				n := uint16((maxValue - v) * factor)
				if maxValue == 255 {
					return uint8(n)
				}
				return uint8((uint32(n) + 128) / 257)
			}
			out.SetNRGBA(x, y, color.NRGBA{R: channel(c), G: channel(m), B: channel(yellow), A: 255})
		}
	}
	return out, nil
}

// All original metadata arrays and encoded blocks must be physically present
// in the original object. The appended private IFD may only supply metadata;
// it must never satisfy a truncated source array or an EOF pixel read.
func validateTIFFPhysicalData(data []byte, entries map[uint16][]byte, order binary.ByteOrder, work *matchWork) error {
	for _, e := range entries {
		var size uint64
		switch order.Uint16(e[2:]) {
		case 1, 2, 6, 7:
			size = 1
		case 3, 8:
			size = 2
		case 4, 9, 11, 13:
			size = 4
		case 5, 10, 12, 16, 17, 18:
			size = 8
		default:
			continue
		}
		n := uint64(order.Uint32(e[4:])) * size
		if n > 4 {
			off := uint64(order.Uint32(e[8:]))
			if off > uint64(len(data)) || n > uint64(len(data))-off {
				return errTIFFProfile
			}
		}
	}
	array := func(e []byte) (count int, value func(int) uint64, err error) {
		if e == nil {
			return 0, nil, errTIFFProfile
		}
		n := uint64(order.Uint32(e[4:]))
		if n == 0 || n > 1<<20 {
			return 0, nil, errTIFFProfile
		}
		var size uint64
		switch order.Uint16(e[2:]) {
		case 1:
			size = 1
		case 3:
			size = 2
		case 4:
			size = 4
		default:
			return 0, nil, errTIFFProfile
		}
		raw := e[8:12]
		if n*size > 4 {
			off := uint64(order.Uint32(raw))
			// The preceding metadata pass checked this physical slice.
			raw = data[off : off+n*size]
		}
		return int(n), func(i int) uint64 {
			switch size {
			case 1:
				return uint64(raw[i])
			case 2:
				return uint64(order.Uint16(raw[2*i:]))
			default:
				return uint64(order.Uint32(raw[4*i:]))
			}
		}, nil
	}
	for _, pair := range [][2]uint16{{273, 279}, {324, 325}} {
		if entries[pair[0]] == nil && entries[pair[1]] == nil {
			continue
		}
		n, offset, err := array(entries[pair[0]])
		if err != nil {
			return err
		}
		m, length, err := array(entries[pair[1]])
		if err != nil || n != m || work.step(int64(n)+1) != nil {
			return errTIFFProfile
		}
		for i := 0; i < n; i++ {
			off, size := offset(i), length(i)
			if off > uint64(len(data)) || size > uint64(len(data))-off {
				return errTIFFProfile
			}
		}
	}
	return nil
}

type tiffView struct {
	p   *tiffProfile
	pos int64
}

func (v *tiffView) Read(b []byte) (int, error) {
	n, err := v.ReadAt(b, v.pos)
	v.pos += int64(n)
	return n, err
}

func (v *tiffView) ReadAt(b []byte, off int64) (int, error) {
	if err := v.p.work.ctx.Err(); err != nil {
		return 0, err
	}
	if off < 0 {
		return 0, io.ErrUnexpectedEOF
	}
	total := int64(len(v.p.data)) + int64(len(v.p.directory))
	if off > total {
		return 0, io.EOF
	}
	n := int(min(int64(len(b)), total-off))
	if err := v.p.work.step(int64(n)/64 + 1); err != nil {
		return 0, err
	}
	for i := 0; i < n; {
		at := off + int64(i)
		var src []byte
		switch {
		case v.p.cmyk && at < 8:
			src = v.p.header[at:]
		case at < int64(len(v.p.data)):
			src = v.p.data[at:]
		default:
			src = v.p.directory[at-int64(len(v.p.data)):]
		}
		i += copy(b[i:n], src)
	}
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

// The codec uses the view directly, without materializing the virtual file.
var _ io.ReaderAt = (*tiffView)(nil)
var _ io.Reader = (*tiffView)(nil)
