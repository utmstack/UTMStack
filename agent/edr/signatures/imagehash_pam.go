package signatures

import (
	"bytes"
	"image"
	"image/color"
	"strconv"
	"strings"
)

const pamHeaderLimit = 64 << 10

type pamHeader struct {
	width, height, depth, maxValue, raster int
	tupleType                              string
}

func pamNumber(field string) (int, error) {
	if len(field) == 0 || len(field) > 15 {
		return 0, errPNM
	}
	for i := range field {
		if field[i] < '0' || field[i] > '9' {
			return 0, errPNM
		}
	}
	n, err := strconv.Atoi(field)
	if err != nil {
		return 0, errPNM
	}
	return n, nil
}

// parsePAMHeader accepts the reference-tested visual tuple layouts. Unknown
// fields or tuples are undecided, since their pixel meaning may differ.
func parsePAMHeader(data []byte) (pamHeader, error) {
	var h pamHeader
	if !bytes.HasPrefix(data, []byte("P7\n")) {
		return h, errPNM
	}
	seen := map[string]bool{}
	at := 3
	end := min(len(data), pamHeaderLimit)
	for at < end {
		n := bytes.IndexByte(data[at:end], '\n')
		if n < 0 || n > 4096 {
			return h, errPNM
		}
		line := data[at : at+n]
		at += n + 1
		if len(line) == 0 || pnmSpace(line[0]) {
			return h, errPNM
		}
		if line[0] == '#' {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) == 0 {
			continue
		}
		key := fields[0]
		if seen[key] {
			return h, errPNM
		}
		seen[key] = true
		if key == "ENDHDR" {
			if len(fields) != 1 || !seen["WIDTH"] || !seen["HEIGHT"] || !seen["DEPTH"] || !seen["MAXVAL"] {
				return h, errPNM
			}
			h.raster = at
			break
		}
		if len(fields) != 2 {
			return h, errPNM
		}
		if key == "TUPLTYPE" {
			h.tupleType = fields[1]
			continue
		}
		value, err := pamNumber(fields[1])
		if err != nil {
			return h, err
		}
		switch key {
		case "WIDTH":
			h.width = value
		case "HEIGHT":
			h.height = value
		case "DEPTH":
			h.depth = value
		case "MAXVAL":
			h.maxValue = value
		default:
			return h, errPNM
		}
	}
	if h.raster == 0 || h.width <= 0 || h.height <= 0 || h.depth < 1 || h.depth > 4 ||
		h.maxValue < 1 || h.maxValue > 65535 {
		return h, errPNM
	}
	switch h.tupleType {
	case "": // image-rs infers gray/gray-alpha/RGB/RGB-alpha from depth
	case "BLACKANDWHITE":
		if h.depth != 1 || h.maxValue != 1 {
			return h, errPNM
		}
	case "BLACKANDWHITE_ALPHA":
		if h.depth != 2 || h.maxValue != 1 {
			return h, errPNM
		}
	case "GRAYSCALE":
		if h.depth != 1 {
			return h, errPNM
		}
	case "GRAYSCALE_ALPHA":
		if h.depth != 2 {
			return h, errPNM
		}
	case "RGB":
		if h.depth != 3 {
			return h, errPNM
		}
	case "RGB_ALPHA":
		if h.depth != 4 {
			return h, errPNM
		}
	default:
		return h, errPNM
	}
	return h, nil
}

func pamConfig(data []byte) (image.Config, error) {
	h, err := parsePAMHeader(data)
	if err != nil {
		return image.Config{}, err
	}
	model := color.GrayModel
	if h.depth >= 3 {
		model = color.NRGBAModel
	}
	return image.Config{ColorModel: model, Width: h.width, Height: h.height}, nil
}

// decodePAM ignores the alpha plane for hashing, like the pinned reference's
// RGB conversion, but validates every plane's sample and the complete raster.
func decodePAM(data []byte) (image.Image, error) {
	h, err := parsePAMHeader(data)
	if err != nil || h.width > imageHashMaxDimension || h.height > imageHashMaxDimension ||
		int64(h.width)*int64(h.height) > imageHashMaxPixels {
		return nil, errPNM
	}
	n := h.width * h.height
	bytesPerSample := 1
	if h.tupleType != "" && h.maxValue >= 256 {
		bytesPerSample = 2
	}
	if n > (len(data)-h.raster)/(h.depth*bytesPerSample) {
		return nil, errPNM
	}
	at := h.raster
	readSample := func() (uint8, error) {
		if bytesPerSample == 1 && h.maxValue >= 256 {
			// The pinned image-rs decoder's untyped PAM path consumes one
			// byte per plane even when MaxVal declares a 16-bit sample.
			// Its scaler still uses the declared MaxVal.
			value := int(data[at])
			at++
			return pnmSample8(value, h.maxValue)
		}
		return pnmBinarySample(data, &at, h.maxValue)
	}
	if h.depth <= 2 {
		img := image.NewGray(image.Rect(0, 0, h.width, h.height))
		for i := 0; i < n; i++ {
			for c := 0; c < h.depth; c++ {
				v, err := readSample()
				if err != nil {
					return nil, err
				}
				if c == 0 {
					img.Pix[i] = v
				}
			}
		}
		return img, nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, h.width, h.height))
	for i := 0; i < n; i++ {
		for c := 0; c < h.depth; c++ {
			v, err := readSample()
			if err != nil {
				return nil, err
			}
			if c < 3 {
				img.Pix[4*i+c] = v
			}
		}
		img.Pix[4*i+3] = 255
	}
	return img, nil
}
