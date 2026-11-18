package signatures

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"strconv"
)

var errPNM = errors.New("invalid or unsupported Netpbm image")

type pnmHeader struct {
	kind, width, height, maxValue, raster int
}

func pnmSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

// pnmNumber consumes one bounded decimal header or ASCII-raster sample.
// Comments are allowed between tokens but never inside binary raster bytes.
func pnmNumber(data []byte, at *int) (int, error) {
	for *at < len(data) {
		if pnmSpace(data[*at]) {
			*at++
			continue
		}
		if data[*at] == '#' {
			for *at < len(data) && data[*at] != '\n' && data[*at] != '\r' {
				*at++
			}
			continue
		}
		break
	}
	start := *at
	for *at < len(data) && data[*at] >= '0' && data[*at] <= '9' && *at-start < 16 {
		*at++
	}
	if *at == start || *at-start >= 16 || *at < len(data) && !pnmSpace(data[*at]) && data[*at] != '#' {
		return 0, errPNM
	}
	n, err := strconv.Atoi(string(data[start:*at]))
	if err != nil {
		return 0, errPNM
	}
	return n, nil
}

func parsePNMHeader(data []byte) (pnmHeader, error) {
	var h pnmHeader
	if len(data) < 3 || data[0] != 'P' || data[1] < '1' || data[1] > '6' || !pnmSpace(data[2]) {
		return h, errPNM
	}
	h.kind = int(data[1] - '0')
	at := 2
	var err error
	if h.width, err = pnmNumber(data, &at); err != nil {
		return h, err
	}
	if h.height, err = pnmNumber(data, &at); err != nil {
		return h, err
	}
	h.maxValue = 1
	if h.kind != 1 && h.kind != 4 {
		if h.maxValue, err = pnmNumber(data, &at); err != nil {
			return h, err
		}
	}
	if h.width <= 0 || h.height <= 0 || h.maxValue <= 0 || h.maxValue > 65535 {
		return h, errPNM
	}
	if h.kind >= 4 {
		if at >= len(data) || !pnmSpace(data[at]) {
			return h, errPNM
		}
		at++
		if data[at-1] == '\r' && at < len(data) && data[at] == '\n' {
			at++
		}
	}
	h.raster = at
	return h, nil
}

func pnmConfig(data []byte) (image.Config, error) {
	h, err := parsePNMHeader(data)
	if err != nil {
		return image.Config{}, err
	}
	model := color.GrayModel
	if h.kind == 3 || h.kind == 6 {
		return image.Config{ColorModel: color.NRGBAModel, Width: h.width, Height: h.height}, nil
	}
	return image.Config{ColorModel: model, Width: h.width, Height: h.height}, nil
}

// pnmSample8 scales a declared sample to the 8-bit RGB/luma values used by
// the reference image decoder before hashing. The max-value multiplication
// is bounded by 65535*255 and cannot overflow int on supported platforms.
func pnmSample8(value, maxValue int) (uint8, error) {
	if value > maxValue {
		return 0, errPNM
	}
	return uint8((value*255 + maxValue/2) / maxValue), nil
}

func pnmBinarySample(data []byte, at *int, maxValue int) (uint8, error) {
	if maxValue < 256 {
		if *at >= len(data) {
			return 0, errPNM
		}
		value := int(data[*at])
		*at++
		return pnmSample8(value, maxValue)
	}
	if len(data)-*at < 2 {
		return 0, errPNM
	}
	value := int(binary.BigEndian.Uint16(data[*at:]))
	*at += 2
	return pnmSample8(value, maxValue)
}

// decodePNM covers the six classic Netpbm layouts, including 16-bit big-endian
// PGM/PPM samples. Invalid samples and truncated rasters remain undecided.
func decodePNM(data []byte) (image.Image, error) {
	h, err := parsePNMHeader(data)
	if err != nil || h.width > imageHashMaxDimension || h.height > imageHashMaxDimension ||
		int64(h.width)*int64(h.height) > imageHashMaxPixels {
		return nil, errPNM
	}
	n := h.width * h.height
	if h.kind == 3 || h.kind == 6 {
		img := image.NewNRGBA(image.Rect(0, 0, h.width, h.height))
		if h.kind == 6 {
			bytesPerSample := 1
			if h.maxValue >= 256 {
				bytesPerSample = 2
			}
			if n > (len(data)-h.raster)/(3*bytesPerSample) {
				return nil, errPNM
			}
			at := h.raster
			for i := 0; i < n; i++ {
				for c := 0; c < 3; c++ {
					v, err := pnmBinarySample(data, &at, h.maxValue)
					if err != nil {
						return nil, err
					}
					img.Pix[4*i+c] = v
				}
				img.Pix[4*i+3] = 255
			}
		} else {
			at := h.raster
			for i := 0; i < 3*n; i++ {
				v, err := pnmNumber(data, &at)
				if err != nil {
					return nil, errPNM
				}
				scaled, err := pnmSample8(v, h.maxValue)
				if err != nil {
					return nil, err
				}
				img.Pix[4*(i/3)+i%3] = scaled
				if i%3 == 2 {
					img.Pix[4*(i/3)+3] = 255
				}
			}
		}
		return img, nil
	}
	img := image.NewGray(image.Rect(0, 0, h.width, h.height))
	switch h.kind {
	case 4:
		stride := (h.width + 7) / 8
		if h.height > (len(data)-h.raster)/stride {
			return nil, errPNM
		}
		for y := 0; y < h.height; y++ {
			for x := 0; x < h.width; x++ {
				if data[h.raster+y*stride+x/8]&(0x80>>uint(x%8)) == 0 {
					img.Pix[y*h.width+x] = 255
				}
			}
		}
	case 5:
		bytesPerSample := 1
		if h.maxValue >= 256 {
			bytesPerSample = 2
		}
		if n > (len(data)-h.raster)/bytesPerSample {
			return nil, errPNM
		}
		if h.maxValue == 255 {
			copy(img.Pix, data[h.raster:h.raster+n])
		} else {
			at := h.raster
			for i := 0; i < n; i++ {
				v, err := pnmBinarySample(data, &at, h.maxValue)
				if err != nil {
					return nil, err
				}
				img.Pix[i] = v
			}
		}
	case 1, 2:
		at := h.raster
		for i := 0; i < n; i++ {
			v, err := pnmNumber(data, &at)
			if err != nil || v > h.maxValue {
				return nil, errPNM
			}
			if h.kind == 1 {
				v = 255 * (1 - v)
			} else {
				v = (v*255 + h.maxValue/2) / h.maxValue
			}
			img.Pix[i] = uint8(v)
		}
	}
	return img, nil
}
