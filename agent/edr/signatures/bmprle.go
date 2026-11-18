package signatures

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

// bmpRLEInfo is the bounded subset of indexed RLE BMPs that ClamAV's image
// decoder hashes but x/image/bmp does not decode. Invalid or unsupported
// layouts remain unknown to the fuzzy matcher.
type bmpRLEInfo struct {
	width, height, bpp, offset int
	palette                    color.Palette
}

var errBMPRLE = errors.New("invalid or unsupported RLE BMP")

func bmpRLEHeader(data []byte) (bmpRLEInfo, error) {
	var info bmpRLEInfo
	if len(data) < 54 || data[0] != 'B' || data[1] != 'M' {
		return info, errBMPRLE
	}
	dib := uint64(binary.LittleEndian.Uint32(data[14:18]))
	if dib < 40 || dib > uint64(len(data)-14) {
		return info, errBMPRLE
	}
	info.width = int(int32(binary.LittleEndian.Uint32(data[18:22])))
	info.height = int(int32(binary.LittleEndian.Uint32(data[22:26])))
	info.bpp = int(binary.LittleEndian.Uint16(data[28:30]))
	compression := binary.LittleEndian.Uint32(data[30:34])
	info.offset = int(binary.LittleEndian.Uint32(data[10:14]))
	if info.width <= 0 || info.height <= 0 || binary.LittleEndian.Uint16(data[26:28]) != 1 ||
		(info.bpp != 4 || compression != 2) && (info.bpp != 8 || compression != 1) ||
		info.offset < 14+int(dib) || info.offset > len(data) {
		return bmpRLEInfo{}, errBMPRLE
	}
	colors := int(binary.LittleEndian.Uint32(data[46:50]))
	if colors == 0 {
		colors = 1 << info.bpp
	}
	if colors <= 0 || colors > 1<<info.bpp || colors > (info.offset-14-int(dib))/4 {
		return bmpRLEInfo{}, errBMPRLE
	}
	info.palette = make(color.Palette, 256)
	for i := range info.palette {
		info.palette[i] = color.NRGBA{A: 255}
	}
	for i := 0; i < colors; i++ {
		at := 14 + int(dib) + i*4
		info.palette[i] = color.NRGBA{R: data[at+2], G: data[at+1], B: data[at], A: 255}
	}
	return info, nil
}

func bmpRLECandidate(data []byte) bool {
	if len(data) < 34 || data[0] != 'B' || data[1] != 'M' {
		return false
	}
	compression := binary.LittleEndian.Uint32(data[30:34])
	return compression == 1 || compression == 2
}

func bmpRLEConfig(data []byte) (image.Config, error) {
	info, err := bmpRLEHeader(data)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{ColorModel: info.palette, Width: info.width, Height: info.height}, nil
}

func decodeBMPRLE(data []byte) (image.Image, error) {
	return decodeBMPRLEWork(data, nil)
}

func decodeBMPRLEWork(data []byte, work *matchWork) (image.Image, error) {
	info, err := bmpRLEHeader(data)
	if err != nil {
		return nil, err
	}
	img, _, err := decodeBMPRLEInfo(data, info, work)
	return img, err
}

// Return the exact input cursor as well as RGB pixels. ICO mask admission
// uses that cursor; the reference stops after the final EOL, before an EOB.
func decodeBMPRLEInfo(data []byte, info bmpRLEInfo, work *matchWork) (image.Image, int, error) {
	if info.width <= 0 || info.height <= 0 || info.offset < 0 || info.offset > len(data) ||
		info.width > imageHashMaxDimension || info.height > imageHashMaxDimension ||
		int64(info.width)*int64(info.height) > imageHashMaxPixels {
		return nil, 0, errBMPRLE
	}
	// ClamAV's image decoder paints pixels skipped by an RLE delta or row
	// ending black, even when palette entry zero is a different color.
	img := image.NewNRGBA(image.Rect(0, 0, info.width, info.height))
	b := data[info.offset:]
	x, row := 0, 0
	write := func(index byte) bool {
		if x >= info.width || row >= info.height || int(index) >= len(info.palette) {
			return false
		}
		c := info.palette[index].(color.NRGBA)
		at := (info.height-1-row)*img.Stride + 4*x
		img.Pix[at], img.Pix[at+1], img.Pix[at+2], img.Pix[at+3] = c.R, c.G, c.B, 255
		x++
		return true
	}
	charged := 0
	for at := 0; at+2 <= len(b); {
		if work != nil && at-charged >= 1024 {
			if err := work.step(int64(at-charged)/2 + 1); err != nil {
				return nil, 0, err
			}
			charged = at
		}
		count, value := int(b[at]), b[at+1]
		at += 2
		if count != 0 {
			pixels := count
			if info.bpp == 8 && pixels > info.width-x {
				// image 0.25.9 clamps an encoded RLE8 run at the
				// row boundary, but still advances its logical x by the
				// original count. Absolute runs and RLE4 reject overruns.
				pixels = info.width - x
				if pixels < 0 {
					pixels = 0
				}
			}
			for i := 0; i < pixels; i++ {
				index := value
				if info.bpp == 4 {
					index = value >> 4
					if i&1 != 0 {
						index = value & 15
					}
				}
				if !write(index) {
					return nil, 0, errBMPRLE
				}
			}
			x += count - pixels
			continue
		}
		switch value {
		case 0: // end of line
			x, row = 0, row+1
			if row == info.height {
				return img, info.offset + at, nil
			}
		case 1: // end of bitmap
			return img, info.offset + at, nil
		case 2: // delta
			if at+2 > len(b) {
				return nil, 0, errBMPRLE
			}
			xdelta, ydelta := int(b[at]), int(b[at+1])
			x += xdelta
			row += ydelta
			at += 2
			if (xdelta != 0 || ydelta != 0) && x > info.width || row >= info.height {
				return nil, 0, errBMPRLE
			}
		default: // literal pixels, padded to a 16-bit boundary
			n := int(value)
			bytesNeeded := n
			if info.bpp == 4 {
				bytesNeeded = (n + 1) / 2
			}
			padded := (bytesNeeded + 1) &^ 1
			if at+padded > len(b) {
				return nil, 0, errBMPRLE
			}
			for i := 0; i < n; i++ {
				var index byte
				if info.bpp == 4 {
					index = b[at+i/2] >> 4
					if i&1 != 0 {
						index = b[at+i/2] & 15
					}
				} else {
					index = b[at+i]
				}
				if !write(index) {
					return nil, 0, errBMPRLE
				}
			}
			at += padded
		}
	}
	return nil, 0, errBMPRLE
}
