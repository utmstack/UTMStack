package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

var errQOI = errors.New("invalid QOI image")

var qoiEnd = []byte{0, 0, 0, 0, 0, 0, 0, 1}

func qoiConfig(data []byte) (image.Config, error) {
	if len(data) < 22 || !bytes.Equal(data[:4], []byte("qoif")) ||
		(data[12] != 3 && data[12] != 4) || data[13] > 1 {
		return image.Config{}, errQOI
	}
	w, h := binary.BigEndian.Uint32(data[4:8]), binary.BigEndian.Uint32(data[8:12])
	if w == 0 || h == 0 || w > imageHashMaxDimension || h > imageHashMaxDimension ||
		uint64(w)*uint64(h) > imageHashMaxPixels {
		return image.Config{}, errQOI
	}
	return image.Config{ColorModel: color.NRGBAModel, Width: int(w), Height: int(h)}, nil
}

// decodeQOI implements the six QOI chunk forms while bounding both the
// decoded raster and every chunk read. The alpha channel is retained so an
// index lookup has the right state even though fuzzy hashing ignores alpha.
func decodeQOI(data []byte) (image.Image, error) {
	cfg, err := qoiConfig(data)
	if err != nil || !bytes.Equal(data[len(data)-8:], qoiEnd) {
		return nil, errQOI
	}
	img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	var index [64][4]byte
	px := [4]byte{0, 0, 0, 255}
	at, end, run := 14, len(data)-8, 0
	for i := 0; i < cfg.Width*cfg.Height; i++ {
		if run > 0 {
			run--
		} else {
			if at >= end {
				return nil, errQOI
			}
			b := data[at]
			at++
			switch {
			case b == 0xfe:
				if end-at < 3 {
					return nil, errQOI
				}
				copy(px[:3], data[at:at+3])
				at += 3
			case b == 0xff:
				if end-at < 4 {
					return nil, errQOI
				}
				copy(px[:], data[at:at+4])
				at += 4
			case b&0xc0 == 0x00:
				px = index[b&0x3f]
			case b&0xc0 == 0x40:
				px[0] += ((b >> 4) & 3) - 2
				px[1] += ((b >> 2) & 3) - 2
				px[2] += (b & 3) - 2
			case b&0xc0 == 0x80:
				if at >= end {
					return nil, errQOI
				}
				b2 := data[at]
				at++
				dg := int(b&0x3f) - 32
				px[0] += byte(dg + int(b2>>4) - 8)
				px[1] += byte(dg)
				px[2] += byte(dg + int(b2&0x0f) - 8)
			default:
				run = int(b & 0x3f)
				if i+run >= cfg.Width*cfg.Height {
					return nil, errQOI
				}
			}
			index[(int(px[0])*3+int(px[1])*5+int(px[2])*7+int(px[3])*11)%64] = px
		}
		copy(img.Pix[i*4:i*4+4], px[:])
	}
	if at != end || run != 0 {
		return nil, errQOI
	}
	return img, nil
}
