package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

// DDS/DXT policy follows image-rs 0.25.9. Alpha is deliberately discarded
// by fuzzy hashing; its encoded bytes still have to be physically present.
var errDDS = errors.New("invalid or unsupported DDS image")

type ddsInfo struct {
	width, height, offset, blockBytes int
	dxt1                              bool
}

func readDDSHeader(data []byte) (ddsInfo, error) {
	if len(data) < 128 || !bytes.Equal(data[:4], []byte("DDS ")) {
		return ddsInfo{}, errDDS
	}
	le := binary.LittleEndian
	flags := le.Uint32(data[8:12])
	const required = 0x1007
	const allowed = 0x1 | 0x2 | 0x4 | 0x8 | 0x1000 | 0x20000 | 0x80000 | 0x00800000
	if le.Uint32(data[4:8]) != 124 || flags&required != required || flags&^uint32(allowed) != 0 || le.Uint32(data[76:80]) != 32 || le.Uint32(data[80:84])&4 == 0 {
		return ddsInfo{}, errDDS
	}
	i := ddsInfo{width: int(le.Uint32(data[16:20])), height: int(le.Uint32(data[12:16])), offset: 128, blockBytes: 16}
	switch string(data[84:88]) {
	case "DXT1":
		i.dxt1 = true
		i.blockBytes = 8
	case "DXT3", "DXT5":
	case "DX10":
		if len(data) < 148 {
			return ddsInfo{}, errDDS
		}
		format, dimension, misc, array, alpha := le.Uint32(data[128:132]), le.Uint32(data[132:136]), le.Uint32(data[136:140]), le.Uint32(data[140:144]), le.Uint32(data[144:148])
		if dimension < 2 || dimension > 4 || misc != 0 && misc != 4 || dimension == 4 && array != 1 || alpha > 4 {
			return ddsInfo{}, errDDS
		}
		switch {
		case format >= 70 && format <= 72:
			i.dxt1 = true
			i.blockBytes = 8
		case format >= 73 && format <= 78:
		default:
			return ddsInfo{}, errDDS
		}
		i.offset = 148
	default:
		return ddsInfo{}, errDDS
	}
	if i.width <= 0 || i.height <= 0 || i.width%4 != 0 || i.height%4 != 0 {
		return ddsInfo{}, errDDS
	}
	return i, nil
}

func decodeDDS(data []byte, i ddsInfo, work *matchWork) (image.Image, error) {
	if i.width <= 0 || i.height <= 0 || i.width%4 != 0 || i.height%4 != 0 || i.blockBytes != 8 && i.blockBytes != 16 || i.width > imageHashMaxDimension || i.height > imageHashMaxDimension || int64(i.width)*int64(i.height) > imageHashMaxPixels || i.offset < 0 || i.offset > len(data) {
		return nil, errDDS
	}
	wide, high := i.width/4, i.height/4
	if wide > (len(data)-i.offset)/i.blockBytes/high {
		return nil, errDDS
	}
	out := image.NewNRGBA(image.Rect(0, 0, i.width, i.height))
	pos := i.offset
	for by := 0; by < high; by++ {
		if work != nil {
			if err := work.ctx.Err(); err != nil {
				return nil, err
			}
			if err := work.step(int64(wide) + 1); err != nil {
				return nil, err
			}
		}
		for bx := 0; bx < wide; bx++ {
			block := data[pos : pos+i.blockBytes]
			pos += i.blockBytes
			if !i.dxt1 {
				block = block[8:]
			}
			c0, c1 := binary.LittleEndian.Uint16(block[:2]), binary.LittleEndian.Uint16(block[2:4])
			var table [4][3]byte
			for n, value := range []uint16{c0, c1} {
				table[n] = [3]byte{byte(uint32(value>>11&31) * 255 / 31), byte(uint32(value>>5&63) * 255 / 63), byte(uint32(value&31) * 255 / 31)}
			}
			for c := 0; c < 3; c++ {
				a, b := uint16(table[0][c]), uint16(table[1][c])
				if c0 > c1 || !i.dxt1 {
					table[2][c] = byte((2*a + b + 1) / 3)
					table[3][c] = byte((a + 2*b + 1) / 3)
				} else {
					table[2][c] = byte((a + b + 1) / 2)
				}
			}
			indices := binary.LittleEndian.Uint32(block[4:8])
			for p := 0; p < 16; p++ {
				color := table[indices>>uint(2*p)&3]
				at := (by*4+p/4)*out.Stride + (bx*4+p%4)*4
				copy(out.Pix[at:at+3], color[:])
				out.Pix[at+3] = 255
			}
		}
	}
	return out, nil
}

func (i ddsInfo) config() image.Config {
	return image.Config{ColorModel: color.NRGBAModel, Width: i.width, Height: i.height}
}
