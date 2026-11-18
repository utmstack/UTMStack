package jpeg154

import "io"

// ComponentImage contains interleaved JPEG samples in the input colorspace.
// TIFF's locked zune-jpeg 0.4.21 profile retains YCbCr/CMYK channel values
// rather than converting to RGB or inverting Adobe CMYK.
type ComponentImage struct {
	Width, Height, Channels int
	Pix                     []byte
}

// ComponentConfig describes a raw TIFF JPEG frame without allocating planes.
type ComponentConfig struct {
	Width, Height, Channels int
}

// DecodeTIFFConfig applies the raw component parser's independent sampling
// rules. A normal image.YCbCr configuration assumes equal chroma sampling.
func DecodeTIFFConfig(r io.Reader) (ComponentConfig, error) {
	d := decoder{rawComponents: true}
	if _, err := d.decode(r, true); err != nil {
		return ComponentConfig{}, err
	}
	if d.width <= 0 || d.height <= 0 || d.nComp == 0 {
		return ComponentConfig{}, FormatError("missing component frame")
	}
	return ComponentConfig{Width: d.width, Height: d.height, Channels: d.nComp}, nil
}

// DecodeTIFFComponents preserves the older TIFF dependency's component
// sampling. Dimensions are capped before plane allocation at 16,384 per axis
// and 40 Mi pixels. Callers may apply tighter bounds with DecodeTIFFConfig. A nil
// step skips accounting; a callback charges scan/reconstruction/output work.
func DecodeTIFFComponents(r io.Reader, step func(int64) error) (*ComponentImage, error) {
	d := decoder{rawComponents: true, componentStep: step}
	if _, err := d.decode(referenceJPEGEnd(r), false); err != nil {
		return nil, err
	}
	if d.width <= 0 || d.height <= 0 || d.nComp < 1 || d.nComp > 4 {
		return nil, FormatError("invalid component image")
	}
	out := &ComponentImage{Width: d.width, Height: d.height, Channels: d.nComp}
	out.Pix = make([]byte, d.width*d.height*d.nComp)
	paddedWidth := ((d.width + 8*d.maxH - 1) / (8 * d.maxH)) * 8 * d.maxH
	for y := 0; y < d.height; y++ {
		if step != nil {
			if err := step(int64(d.width)*int64(d.nComp) + 1); err != nil {
				return nil, err
			}
		}
		for x := 0; x < d.width; x++ {
			index := (y*d.width + x) * d.nComp
			for c := 0; c < d.nComp; c++ {
				h, v := d.maxH/d.comp[c].h, d.maxV/d.comp[c].v
				// Conversion consumes one common padded output stride, even
				// when an earlier component has a wider sampling buffer.
				width := d.componentStride[c] * h
				position := (y%(8*d.maxV))*paddedWidth + x
				sx, sy := position%width, position/width+(y/(8*d.maxV))*8*d.maxV
				out.Pix[index+c] = sampleTIFFComponent(d.componentPix[c], d.componentStride[c], sx, sy, h, v)
			}
		}
	}
	return out, nil
}

// 0.4.21 performs vertical interpolation before horizontal interpolation.
// Its penultimate padded output column weights the previous input three
// times, a distinct boundary from the newer standalone JPEG profile.
func sampleTIFFComponent(plane []byte, stride, x, y, h, v int) byte {
	if h > 2 || v > 2 {
		// The older decoder expands a generic component row linearly into
		// h*v repetitions before splitting it into output rows. For 4:1:0
		// this differs from a conventional spatial nearest-neighbor filter.
		return plane[(y/v)*stride+((y%v)*stride*h+x)/(h*v)]
	}
	width, height := stride*h, len(plane)/stride*v
	near := func(x, y int) int {
		x = min(max(x, 0), width-1)
		y = min(max(y, 0), height-1)
		return int(plane[(y/v)*stride+x/h])
	}
	neighbor := func(n int) int {
		if n&1 == 0 {
			return n - 1
		}
		return n + 1
	}
	vertical := func(x int) int {
		n := near(x, y)
		if v == 2 {
			n = (3*n + near(x, neighbor(y)) + 2) >> 2
		}
		return n
	}
	n := vertical(x)
	if h == 2 {
		other := vertical(neighbor(x))
		if x == width-2 {
			n = (3*other + n + 2) >> 2
		} else {
			n = (3*n + other + 2) >> 2
		}
	}
	return byte(n)
}
