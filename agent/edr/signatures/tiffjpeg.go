package signatures

import (
	"bytes"
	"encoding/binary"
	"image"
	"math"

	"github.com/utmstack/UTMStack/agent/edr/jpeg154"
)

type tiffJPEG struct {
	*tiffChunks
	tables    []byte
	hasTables bool
}

func readTIFFJPEG(entries map[uint16][]byte, order binary.ByteOrder, p *tiffProfile,
	values func(uint16, uint32) ([]uint32, error), floatingFormat bool) (*tiffJPEG, error) {
	c, err := readTIFFChunks(entries, order, p, values, floatingFormat)
	if err != nil || c.compression != 7 || (c.photo == 0 || c.photo == 1) && c.channels != 1 ||
		c.bits == 1 && c.predictor != 1 {
		return nil, errTIFFProfile
	}
	j := &tiffJPEG{tiffChunks: c}
	if entry := entries[347]; entry != nil {
		kind, n := order.Uint16(entry[2:]), uint64(order.Uint32(entry[4:]))
		if kind != 1 && kind != 7 || n < 2 || n > 8<<20 {
			return nil, errTIFFProfile
		}
		raw := entry[8:12]
		if n > 4 {
			off := uint64(order.Uint32(raw))
			if off > uint64(len(p.data)) || n > uint64(len(p.data))-off {
				return nil, errTIFFProfile
			}
			raw = p.data[off : off+n]
		}
		j.tables, j.hasTables = raw[:n], true
	}
	return j, nil
}

func (j *tiffJPEG) decode(p *tiffProfile) (image.Image, error) {
	if p.work.ctx.Err() != nil {
		return nil, errTIFFProfile
	}
	out := image.NewNRGBA(image.Rect(0, 0, j.cfg.Width, j.cfg.Height))
	rowBytes := (j.blockWidth*j.channels*j.bits + 7) / 8
	row := make([]byte, rowBytes)
	for block, at := range j.offsets {
		n := int64(j.lengths[block])
		if n > 16<<20 || j.hasTables && n < 2 || p.work.ctx.Err() != nil || p.work.step((n+int64(len(j.tables)))/64+1) != nil {
			return nil, errTIFFProfile
		}
		// Original physical bounds were checked before any codec view existed.
		jpeg := p.data[int(at) : int64(at)+n]
		if j.hasTables {
			merged := make([]byte, len(j.tables)-2+len(jpeg)-2)
			copy(merged, j.tables[:len(j.tables)-2])
			copy(merged[len(j.tables)-2:], jpeg[2:])
			jpeg = merged
		}
		cfg, err := jpeg154.DecodeTIFFConfig(bytes.NewReader(jpeg))
		if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > imageHashMaxDimension || cfg.Height > imageHashMaxDimension || int64(cfg.Width)*int64(cfg.Height) > 16<<20 {
			return nil, errTIFFProfile
		}
		components, err := jpeg154.DecodeTIFFComponents(bytes.NewReader(jpeg), func(n int64) error {
			if err := p.work.ctx.Err(); err != nil {
				return err
			}
			return p.work.step(n)
		})
		if err != nil {
			return nil, err
		}
		x0, y0 := (block%j.across)*j.blockWidth, (block/j.across)*j.blockHeight
		width, height := min(j.blockWidth, j.cfg.Width-x0), min(j.blockHeight, j.cfg.Height-y0)
		position := 0
		for y := 0; y < height; y++ {
			if p.work.ctx.Err() != nil || p.work.step(int64(rowBytes)/64+int64(width)*int64(j.channels)+1) != nil {
				return nil, errTIFFProfile
			}
			used := (width*j.channels*j.bits + 7) / 8
			needed := used
			if j.predictor == 3 {
				needed = rowBytes
			}
			if position > len(components.Pix) || needed > len(components.Pix)-position {
				return nil, errTIFFProfile
			}
			copy(row[:needed], components.Pix[position:position+needed])
			position += rowBytes
			if j.predictor == 3 {
				for i := j.channels; i < rowBytes; i++ {
					row[i] += row[i-j.channels]
				}
			} else if j.predictor == 2 {
				for i := j.channels; i < width*j.channels; i++ {
					switch j.bits {
					case 8:
						row[i] += row[i-j.channels]
					case 16:
						j.order.PutUint16(row[i*2:], j.order.Uint16(row[i*2:])+j.order.Uint16(row[(i-j.channels)*2:]))
					case 32:
						j.order.PutUint32(row[i*4:], j.order.Uint32(row[i*4:])+j.order.Uint32(row[(i-j.channels)*4:]))
					}
				}
			}
			dst := out.Pix[(y0+y)*out.Stride+x0*4:]
			for x := 0; x < width; x++ {
				var values [4]uint32
				for c := 0; c < j.channels; c++ {
					index := x*j.channels + c
					switch j.bits {
					case 1:
						values[c] = uint32(row[index/8] >> (7 - index%8) & 1)
					case 8:
						values[c] = uint32(row[index])
					case 16:
						values[c] = uint32(j.order.Uint16(row[index*2:]))
					case 32:
						if j.predictor == 3 {
							plane := rowBytes / 4
							values[c] = uint32(row[index])<<24 | uint32(row[plane+index])<<16 | uint32(row[2*plane+index])<<8 | uint32(row[3*plane+index])
						} else {
							values[c] = j.order.Uint32(row[index*4:])
						}
					}
				}
				var rgb [3]byte
				if j.photo == 5 {
					maximum := float32(255)
					if j.bits == 16 {
						maximum = 65535
					}
					factor := float32(1) - float32(values[3])/maximum
					for c := range rgb {
						v := uint32((maximum - float32(values[c])) * factor)
						if j.bits == 16 {
							v = (v + 128) / 257
						}
						rgb[c] = byte(v)
					}
				} else {
					for c := range rgb {
						v := values[c]
						if j.photo == 0 || j.photo == 1 {
							v = values[0]
						}
						if j.photo == 0 {
							v = (uint32(1)<<j.bits - 1) - v
						}
						switch j.bits {
						case 1:
							rgb[c] = byte(v * 255)
						case 8:
							rgb[c] = byte(v)
						case 16:
							rgb[c] = byte((v + 128) / 257)
						case 32:
							rgb[c] = tiffFloatRGB8(math.Float32frombits(v))
						}
					}
				}
				copy(dst[x*4:x*4+3], rgb[:])
				dst[x*4+3] = 255
			}
		}
	}
	if err := p.work.ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
