package signatures

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// HDR policy follows image-rs 0.25.9. See IMAGE_DECODERS_NOTICE.md.
var errHDR = errors.New("invalid or unsupported Radiance HDR image")

type hdrInfo struct{ width, height, offset int }

func hdrFloat(b []byte) bool {
	s := strings.TrimSpace(string(b))
	unsigned := strings.TrimPrefix(strings.TrimPrefix(s, "+"), "-")
	if strings.EqualFold(unsigned, "nan") && (len(s) == 3 || len(s) == 4) {
		return true // Rust accepts signed NaNs; strconv.ParseFloat does not.
	}
	if strings.Contains(s, "_") || strings.HasPrefix(strings.TrimLeft(s, "+-"), "0x") || strings.HasPrefix(strings.TrimLeft(s, "+-"), "0X") {
		return false
	}
	_, err := strconv.ParseFloat(s, 32)
	return err == nil || errors.Is(err, strconv.ErrRange)
}

func readHDRHeader(data []byte) (hdrInfo, error) {
	if !bytes.HasPrefix(data, []byte("#?RADIANCE")) {
		return hdrInfo{}, errHDR
	}
	pos := len("#?RADIANCE")
	line := func() ([]byte, bool) {
		if pos >= len(data) {
			return nil, false
		}
		n := bytes.IndexByte(data[pos:], '\n')
		if n < 0 {
			out := data[pos:]
			pos = len(data)
			return out, true
		}
		out := data[pos : pos+n]
		pos += n + 1
		return out, true
	}
	line() // the remainder of the signature line is advisory
	for {
		b, ok := line()
		if !ok {
			return hdrInfo{}, errHDR
		}
		if len(b) == 0 {
			break
		}
		if b[0] == '#' {
			continue
		}
		n := bytes.IndexByte(b, '=')
		if n <= 0 || n == len(b)-1 {
			continue
		}
		key, val := string(bytes.TrimSpace(b[:n])), b[n+1:]
		switch key {
		case "FORMAT":
			if string(bytes.TrimSpace(val)) != "32-bit_rle_rgbe" {
				return hdrInfo{}, errHDR
			}
		case "EXPOSURE", "PIXASPECT":
			if !hdrFloat(val) {
				return hdrInfo{}, errHDR
			}
		case "COLORCORR":
			fields := bytes.Fields(val)
			if len(fields) != 3 {
				return hdrInfo{}, errHDR
			}
			for _, f := range fields {
				if !hdrFloat(f) {
					return hdrInfo{}, errHDR
				}
			}
		}
	}
	b, ok := line()
	if !ok {
		return hdrInfo{}, errHDR
	}
	fields := bytes.Fields(b)
	if len(fields) != 4 || string(fields[0]) != "-Y" || string(fields[2]) != "+X" {
		return hdrInfo{}, errHDR
	}
	parse := func(b []byte) (uint64, error) { return strconv.ParseUint(strings.TrimPrefix(string(b), "+"), 10, 32) }
	h, e1 := parse(fields[1])
	w, e2 := parse(fields[3])
	if e1 != nil || e2 != nil || w == 0 || h == 0 {
		return hdrInfo{}, errHDR
	}
	return hdrInfo{int(w), int(h), pos}, nil
}

type hdrCursor struct {
	data         []byte
	pos, charged int
	work         *matchWork
	err          error
}

func (r *hdrCursor) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if r.work != nil && r.pos-r.charged >= 1024 {
		if err := r.work.step(int64(r.pos-r.charged)/4 + 1); err != nil {
			r.err = err
			return nil
		}
		r.charged = r.pos
	}
	if n < 0 || n > len(r.data)-r.pos {
		r.err = errHDR
		return nil
	}
	out := r.data[r.pos : r.pos+n]
	r.pos += n
	return out
}

func (r *hdrCursor) scanline(row [][4]byte) bool {
	first := r.take(4)
	if first == nil {
		return false
	}
	if first[0] == 2 && first[1] == 2 && first[2] < 128 {
		// The reference recognizes this marker without checking its encoded
		// width; the bounded component streams determine actual row length.
		for channel := 0; channel < 4; channel++ {
			for pos := 0; pos < len(row); {
				b := r.take(1)
				if b == nil {
					return false
				}
				n := int(b[0])
				if n <= 128 {
					if n > len(row)-pos {
						r.err = errHDR
						return false
					}
					b = r.take(n)
					if b == nil {
						return false
					}
					for i, v := range b {
						row[pos+i][channel] = v
					}
					pos += n
				} else {
					n -= 128
					if n > len(row)-pos {
						r.err = errHDR
						return false
					}
					b = r.take(1)
					if b == nil {
						return false
					}
					for i := 0; i < n; i++ {
						row[pos+i][channel] = b[0]
					}
					pos += n
				}
			}
		}
		return true
	}
	if first[0] == 1 && first[1] == 1 && first[2] == 1 {
		r.err = errHDR
		return false
	}
	previous := [4]byte{first[0], first[1], first[2], first[3]}
	row[0] = previous
	multiplier := uint64(1)
	for pos := 1; pos < len(row); {
		b := r.take(4)
		if b == nil {
			return false
		}
		if b[0] == 1 && b[1] == 1 && b[2] == 1 {
			n := uint64(b[3]) * multiplier
			multiplier *= 256 // reference's 64-bit release arithmetic
			if n > uint64(len(row)-pos) {
				r.err = errHDR
				return false
			}
			for i := 0; i < int(n); i++ {
				row[pos+i] = previous
			}
			pos += int(n)
		} else {
			previous = [4]byte{b[0], b[1], b[2], b[3]}
			row[pos] = previous
			pos++
			multiplier = 1
		}
	}
	return true
}

func decodeHDR(data []byte, info hdrInfo, work *matchWork) (image.Image, error) {
	if info.width <= 0 || info.height <= 0 || info.width > imageHashMaxDimension || info.height > imageHashMaxDimension || int64(info.width)*int64(info.height) > imageHashMaxPixels || info.offset < 0 || info.offset > len(data) {
		return nil, errHDR
	}
	out := image.NewNRGBA(image.Rect(0, 0, info.width, info.height))
	row := make([][4]byte, info.width)
	r := hdrCursor{data: data, pos: info.offset, charged: info.offset, work: work}
	for y := 0; y < info.height; y++ {
		if work != nil {
			if err := work.ctx.Err(); err != nil {
				return nil, err
			}
			if err := work.step(int64(info.width)/8 + 1); err != nil {
				return nil, err
			}
		}
		if !r.scanline(row) {
			return nil, r.err
		}
		for x, p := range row {
			at := y*out.Stride + 4*x
			if p[3] != 0 {
				scale := float32(math.Ldexp(1, int(p[3])-136))
				for c := 0; c < 3; c++ {
					v := min(max(mul32(float32(p[c]), scale), 0), 1)
					out.Pix[at+c] = byte(math.Round(float64(mul32(v, 255))))
				}
			}
			out.Pix[at+3] = 255
		}
	}
	return out, nil
}

func (i hdrInfo) config() image.Config {
	return image.Config{ColorModel: color.NRGBAModel, Width: i.width, Height: i.height}
}
