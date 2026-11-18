package signatures

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"math/bits"
)

// bmpBitfieldInfo describes 16-bit RGB555 and 16/32-bit BI_BITFIELDS BMPs.
// These are hashed by ClamAV's image decoder but unsupported by x/image/bmp.
type bmpBitfieldInfo struct {
	width, height, offset, bpp int
	topDown                    bool
	fields                     [4]bmpChannel
}

type bmpChannel struct{ shift, width uint }

var errBMPBitfields = errors.New("invalid or unsupported bitfield BMP")

func bmpBitfieldCandidate(data []byte) bool {
	if len(data) < 34 || data[0] != 'B' || data[1] != 'M' {
		return false
	}
	bpp := binary.LittleEndian.Uint16(data[28:30])
	compression := binary.LittleEndian.Uint32(data[30:34])
	return compression == 0 && bpp == 16 || compression == 3 && (bpp == 16 || bpp == 32)
}

func bmpChannelFromMask(mask uint32, bpp uint) (bmpChannel, error) {
	if mask == 0 {
		return bmpChannel{}, nil
	}
	shift := uint(bits.TrailingZeros32(mask))
	width := uint(bits.OnesCount32(mask))
	if uint64(mask)>>shift != (uint64(1)<<width)-1 || shift+width > bpp {
		return bmpChannel{}, errBMPBitfields
	}
	if width > 8 {
		shift += width - 8
		width = 8
	}
	return bmpChannel{shift: shift, width: width}, nil
}

func (c bmpChannel) read(value uint32) byte {
	if c.width == 0 {
		return 0
	}
	n := (value >> c.shift) & ((1 << c.width) - 1)
	switch c.width {
	case 7:
		return byte((n << 1) | (n >> 6))
	case 8:
		return byte(n)
	default:
		max := uint32((1 << c.width) - 1)
		return byte((n*255 + max/2) / max)
	}
}

func bmpBitfieldHeader(data []byte) (bmpBitfieldInfo, error) {
	var info bmpBitfieldInfo
	if !bmpBitfieldCandidate(data) || len(data) < 54 {
		return info, errBMPBitfields
	}
	dib := int(binary.LittleEndian.Uint32(data[14:18]))
	switch dib {
	case 40, 52, 56, 108, 124:
	default:
		return info, errBMPBitfields
	}
	if dib > len(data)-14 {
		return info, errBMPBitfields
	}
	info.width = int(int32(binary.LittleEndian.Uint32(data[18:22])))
	signedHeight := int32(binary.LittleEndian.Uint32(data[22:26]))
	if signedHeight == -1<<31 {
		return info, errBMPBitfields
	}
	info.topDown = signedHeight < 0
	info.height = int(signedHeight)
	if info.topDown {
		info.height = -info.height
	}
	info.offset = int(binary.LittleEndian.Uint32(data[10:14]))
	info.bpp = int(binary.LittleEndian.Uint16(data[28:30]))
	if info.width <= 0 || info.height <= 0 || binary.LittleEndian.Uint16(data[26:28]) != 1 ||
		info.offset < 14+dib || info.offset > len(data) {
		return bmpBitfieldInfo{}, errBMPBitfields
	}
	compression := binary.LittleEndian.Uint32(data[30:34])
	masks := [4]uint32{0x7c00, 0x03e0, 0x001f, 0}
	if compression == 3 {
		maskEnd := 66
		if dib >= 56 {
			maskEnd = 70
		}
		if maskEnd > len(data) || info.offset < maskEnd {
			return bmpBitfieldInfo{}, errBMPBitfields
		}
		for i := range masks[:3] {
			masks[i] = binary.LittleEndian.Uint32(data[54+i*4 : 58+i*4])
		}
		if dib >= 56 {
			masks[3] = binary.LittleEndian.Uint32(data[66:70])
		}
	}
	for i, mask := range masks {
		field, err := bmpChannelFromMask(mask, uint(info.bpp))
		if err != nil || i < 3 && field.width == 0 {
			return bmpBitfieldInfo{}, errBMPBitfields
		}
		info.fields[i] = field
	}
	return info, nil
}

func bmpBitfieldConfig(data []byte) (image.Config, error) {
	info, err := bmpBitfieldHeader(data)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{ColorModel: color.NRGBAModel, Width: info.width, Height: info.height}, nil
}

func decodeBMPBitfields(data []byte) (image.Image, error) {
	info, err := bmpBitfieldHeader(data)
	if err != nil {
		return nil, err
	}
	return decodeBMPBitfieldsInfo(data, info)
}

func decodeBMPBitfieldsInfo(data []byte, info bmpBitfieldInfo) (image.Image, error) {
	if info.width <= 0 || info.height <= 0 || info.offset < 0 || info.offset > len(data) ||
		info.width > imageHashMaxDimension || info.height > imageHashMaxDimension ||
		int64(info.width)*int64(info.height) > imageHashMaxPixels {
		return nil, errBMPBitfields
	}
	bytesPerPixel := info.bpp / 8
	rowBytes := (info.width*bytesPerPixel + 3) &^ 3
	if rowBytes > (len(data)-info.offset)/info.height {
		return nil, errBMPBitfields
	}
	img := image.NewNRGBA(image.Rect(0, 0, info.width, info.height))
	for sourceY := 0; sourceY < info.height; sourceY++ {
		y := info.height - 1 - sourceY
		if info.topDown {
			y = sourceY
		}
		row := data[info.offset+sourceY*rowBytes:]
		for x := 0; x < info.width; x++ {
			at := x * bytesPerPixel
			var value uint32
			if bytesPerPixel == 2 {
				value = uint32(binary.LittleEndian.Uint16(row[at : at+2]))
			} else {
				value = binary.LittleEndian.Uint32(row[at : at+4])
			}
			pix := img.Pix[y*img.Stride+4*x:]
			pix[0], pix[1], pix[2], pix[3] = info.fields[0].read(value), info.fields[1].read(value), info.fields[2].read(value), 255
		}
	}
	return img, nil
}
