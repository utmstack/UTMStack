package signatures

import (
	"compress/zlib"
	"encoding/binary"
	"image"
	"io"
	"math"

	"golang.org/x/image/tiff/lzw"
)

// Floating TIFFs cannot be represented by the integer Go TIFF decoder. Decode
// their physical chunks directly, retaining the locked image-rs sample and
// predictor semantics. TIFF_SAMPLES_LICENSE.MIT covers the algorithm reference.
type tiffFloat struct{ *tiffChunks }

func readTIFFFloat(entries map[uint16][]byte, order binary.ByteOrder, p *tiffProfile,
	values func(uint16, uint32) ([]uint32, error), floatingFormat bool) (*tiffFloat, error) {
	chunks, err := readTIFFChunks(entries, order, p, values, floatingFormat)
	if err != nil || chunks.bits != 32 || chunks.photo != 2 || chunks.channels != 3 && chunks.channels != 4 || chunks.compression == 7 {
		return nil, errTIFFProfile
	}
	return &tiffFloat{chunks}, nil
}

func (f *tiffFloat) decode(p *tiffProfile) (image.Image, error) {
	if p.work.ctx.Err() != nil {
		return nil, errTIFFProfile
	}
	out := image.NewNRGBA(image.Rect(0, 0, f.cfg.Width, f.cfg.Height))
	rowBytes := f.blockWidth * f.channels * 4
	row := make([]byte, rowBytes)
	scratch := make([]byte, 32<<10)
	for block := range f.offsets {
		if err := p.work.step(1); err != nil {
			return nil, err
		}
		reader, close, err := f.reader(p, block)
		if err != nil {
			return nil, err
		}
		x0, y0 := (block%f.across)*f.blockWidth, (block/f.across)*f.blockHeight
		width, height := min(f.blockWidth, f.cfg.Width-x0), min(f.blockHeight, f.cfg.Height-y0)
		for y := 0; y < height; y++ {
			if p.work.ctx.Err() != nil || p.work.step(int64(rowBytes)/64+int64(width)+1) != nil {
				close()
				return nil, errTIFFProfile
			}
			used := width * f.channels * 4
			if f.predictor == 3 {
				// Byte-plane prediction mixes padding into useful samples.
				// Read and undo the complete padded row before selecting RGB.
				_, err = io.ReadFull(reader, row)
				if err == nil {
					for i := f.channels; i < len(row); i++ {
						row[i] += row[i-f.channels]
					}
				}
			} else {
				_, err = io.ReadFull(reader, row[:used])
				if err == nil && rowBytes > used {
					// The reference ignores EOF in discarded right padding.
					_, err = io.CopyBuffer(io.Discard, io.LimitReader(reader, int64(rowBytes-used)), scratch)
				}
				if err == nil && f.predictor == 2 {
					for i := f.channels * 4; i < used; i += 4 {
						v := f.order.Uint32(row[i:]) + f.order.Uint32(row[i-4*f.channels:])
						f.order.PutUint32(row[i:], v)
					}
				}
			}
			if err != nil {
				close()
				return nil, err
			}
			dst := out.Pix[(y0+y)*out.Stride+x0*4:]
			for x := 0; x < width; x++ {
				for c := 0; c < 3; c++ {
					index := x*f.channels + c
					var bits uint32
					if f.predictor == 3 {
						plane := rowBytes / 4
						bits = uint32(row[index])<<24 | uint32(row[plane+index])<<16 | uint32(row[2*plane+index])<<8 | uint32(row[3*plane+index])
					} else {
						bits = f.order.Uint32(row[index*4:])
					}
					dst[x*4+c] = tiffFloatRGB8(math.Float32frombits(bits))
				}
				dst[x*4+3] = 255
			}
		}
		close()
	}
	if p.work.ctx.Err() != nil {
		return nil, errTIFFProfile
	}
	return out, nil
}

func tiffFloatRGB8(v float32) byte {
	if !(v < 1) { // The image adapter maps NaN to the upper bound too.
		return 255
	}
	if v <= 0 {
		return 0
	}
	return byte(math.Round(float64(float32(v * 255))))
}

func (f *tiffFloat) reader(p *tiffProfile, block int) (io.Reader, func(), error) {
	offset, n := int64(f.offsets[block]), int64(f.lengths[block])
	if f.compression == 1 || f.compression == 8 || f.compression == 32946 {
		// Raw and flate2's streaming reader use the remaining physical file;
		// unlike LZW/PackBits their declared compressed length is advisory.
		n = int64(len(p.data)) - offset
	}
	input := io.NewSectionReader(p.reader(), offset, n)
	noop := func() {}
	switch f.compression {
	case 1:
		return input, noop, nil
	case 5:
		r := lzw.NewReader(input, lzw.MSB, 8)
		return r, func() { _ = r.Close() }, nil
	case 8, 32946:
		r, err := zlib.NewReader(input)
		if err != nil {
			return nil, noop, err
		}
		return tiffFloatZlib{r}, func() { _ = r.Close() }, nil
	case 32773:
		return &tiffPackBits{r: input}, noop, nil
	}
	return nil, noop, errTIFFProfile
}

// Rust's reader cannot return bytes together with a checksum failure. Go's
// zlib reader can, and io.ReadFull would discard the error when those bytes
// exactly satisfy the final row. A physically missing trailer is different:
// the reference accepts enough useful pixels without a complete checksum.
type tiffFloatZlib struct{ io.Reader }

func (r tiffFloatZlib) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if err == io.ErrUnexpectedEOF {
		return n, io.EOF
	}
	if err != nil && err != io.EOF {
		return 0, err
	}
	return n, err
}

// A PackBits packet can span rows; literal data is consumed only as needed.
// The physical reader charges even zero-output 0x80 packets and cancellation.
type tiffPackBits struct {
	r         io.Reader
	remaining int
	repeated  bool
	value     byte
}

func (r *tiffPackBits) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	for r.remaining == 0 {
		var h [1]byte
		if _, err := io.ReadFull(r.r, h[:]); err != nil {
			return 0, err
		}
		code := int(int8(h[0]))
		if code == -128 {
			continue
		}
		r.repeated = code < 0
		if r.repeated {
			if _, err := io.ReadFull(r.r, h[:]); err != nil {
				return 0, err
			}
			r.remaining, r.value = 1-code, h[0]
		} else {
			r.remaining = code + 1
		}
	}
	n := min(len(b), r.remaining)
	var err error
	if r.repeated {
		for i := range b[:n] {
			b[i] = r.value
		}
	} else {
		n, err = r.r.Read(b[:n])
	}
	r.remaining -= n
	return n, err
}
