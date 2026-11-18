package signatures

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

// bmpPlainInfo covers uncompressed Windows and OS/2 BMPs. The reference
// accepts core, V2 and V3 headers and gaps before pixel data that x/image/bmp
// rejects. Keep its palette and pixel reads bounded before constructing an
// image for fuzzy hashing.
type bmpPlainInfo struct {
	width, height, offset, bpp int
	topDown                    bool
	palette                    [256][3]byte
}

var errBMPPlain = errors.New("invalid or unsupported uncompressed BMP")

func bmpPlainCandidate(data []byte) bool {
	if len(data) < 26 || data[0] != 'B' || data[1] != 'M' {
		return false
	}
	dib := binary.LittleEndian.Uint32(data[14:18])
	if dib == 12 {
		return true
	}
	if len(data) < 34 {
		return false
	}
	switch dib {
	case 40, 52, 56, 108, 124:
		return binary.LittleEndian.Uint32(data[30:34]) == 0
	}
	return false
}

func bmpPlainHeader(data []byte) (bmpPlainInfo, error) {
	var info bmpPlainInfo
	if !bmpPlainCandidate(data) {
		return info, errBMPPlain
	}
	dib := int(binary.LittleEndian.Uint32(data[14:18]))
	if dib > len(data)-14 {
		return info, errBMPPlain
	}
	info.offset = int(binary.LittleEndian.Uint32(data[10:14]))
	if info.offset > len(data) {
		return info, errBMPPlain
	}
	paletteEntryBytes := 4
	paletteStart := 14 + dib
	colorsUsed := 0
	if dib == 12 {
		info.width = int(binary.LittleEndian.Uint16(data[18:20]))
		info.height = int(binary.LittleEndian.Uint16(data[20:22]))
		info.bpp = int(binary.LittleEndian.Uint16(data[24:26]))
		if binary.LittleEndian.Uint16(data[22:24]) != 1 {
			return bmpPlainInfo{}, errBMPPlain
		}
		paletteEntryBytes = 3
	} else {
		info.width = int(int32(binary.LittleEndian.Uint32(data[18:22])))
		signedHeight := int32(binary.LittleEndian.Uint32(data[22:26]))
		if signedHeight == -1<<31 || binary.LittleEndian.Uint16(data[26:28]) != 1 {
			return bmpPlainInfo{}, errBMPPlain
		}
		info.topDown = signedHeight < 0
		info.height = int(signedHeight)
		if info.topDown {
			info.height = -info.height
		}
		info.bpp = int(binary.LittleEndian.Uint16(data[28:30]))
		colorsUsed = int(binary.LittleEndian.Uint32(data[46:50]))
	}
	if info.width <= 0 || info.height <= 0 {
		return bmpPlainInfo{}, errBMPPlain
	}
	switch info.bpp {
	case 1, 2, 4, 8:
		if colorsUsed == 0 {
			colorsUsed = 1 << info.bpp
		}
		if colorsUsed > 1<<info.bpp || colorsUsed*paletteEntryBytes > len(data)-paletteStart {
			return bmpPlainInfo{}, errBMPPlain
		}
		for i := 0; i < colorsUsed; i++ {
			p := paletteStart + i*paletteEntryBytes
			info.palette[i] = [3]byte{data[p+2], data[p+1], data[p]}
		}
	case 24, 32:
	default:
		return bmpPlainInfo{}, errBMPPlain
	}
	return info, nil
}

func bmpPlainConfig(data []byte) (image.Config, error) {
	info, err := bmpPlainHeader(data)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{ColorModel: color.NRGBAModel, Width: info.width, Height: info.height}, nil
}

func decodeBMPPlain(data []byte) (image.Image, error) {
	info, err := bmpPlainHeader(data)
	if err != nil {
		return nil, err
	}
	return decodeBMPPlainInfo(data, info)
}

func decodeBMPPlainInfo(data []byte, info bmpPlainInfo) (image.Image, error) {
	if info.width <= 0 || info.height <= 0 || info.offset < 0 || info.offset > len(data) ||
		info.width > imageHashMaxDimension || info.height > imageHashMaxDimension ||
		int64(info.width)*int64(info.height) > imageHashMaxPixels {
		return nil, errBMPPlain
	}
	rowBytes := (info.width*info.bpp + 31) / 32 * 4
	if rowBytes > (len(data)-info.offset)/info.height {
		return nil, errBMPPlain
	}
	img := image.NewNRGBA(image.Rect(0, 0, info.width, info.height))
	for sourceY := 0; sourceY < info.height; sourceY++ {
		y := info.height - 1 - sourceY
		if info.topDown {
			y = sourceY
		}
		row := data[info.offset+sourceY*rowBytes:]
		for x := 0; x < info.width; x++ {
			var r, g, b byte
			switch info.bpp {
			case 1, 2, 4, 8:
				bit := x * info.bpp
				index := int(row[bit/8]>>(8-info.bpp-bit%8)) & ((1 << info.bpp) - 1)
				r, g, b = info.palette[index][0], info.palette[index][1], info.palette[index][2]
			case 24:
				p := 3 * x
				r, g, b = row[p+2], row[p+1], row[p]
			case 32:
				p := 4 * x
				r, g, b = row[p+2], row[p+1], row[p]
			}
			p := img.Pix[y*img.Stride+4*x:]
			p[0], p[1], p[2], p[3] = r, g, b, 255
		}
	}
	return img, nil
}
