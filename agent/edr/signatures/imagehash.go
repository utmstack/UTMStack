package signatures

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"
	"math"
	"sort"

	"github.com/utmstack/UTMStack/agent/edr/dct32"
	"github.com/utmstack/UTMStack/agent/edr/exr154"
	"github.com/utmstack/UTMStack/agent/edr/jpeg154"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

// Image fuzzy hash subsignatures (fuzzy_img#hash) match objects typed as
// images whose perceptual hash equals the rule's. The reference computes a
// pHash-like hash with its own image library: the image decoded, reduced to
// grey, resized to 32x32 with a Lanczos3 filter, transformed with a
// two-dimensional cosine transform, and the 8x8 lowest frequencies compared
// with their median. The fixed f32 transform follows the reference's plan.
// Decoder or resize differences can still make pixel values uncertain;
// coefficient movement bounds preserve that uncertainty. A rule is decided
// only when the certain bits settle it; otherwise
// the object is incomplete. Checked against the reference's own hashes of
// 369 checked-in PNG, GIF and JPEG examples are all decoded with identical
// reference hashes and no differing certain bits. The 202 BMP/TIFF
// format examples all match exactly. After keeping the reference's f32
// operations separate, a 1,000-image BMP/TIFF rounding corpus, a 1,000-image
// no-resize BMP corpus, a 3,000-image varied BMP/TIFF/PNG probe and a
// 1,000-image GIF probe, a 1,000-image 257–512-pixel BMP/TIFF/PNG probe,
// and 112 opaque/transparent 257–512-pixel GIF probes all have identical
// computed hashes. Six P1–P6 Netpbm golden controls and 54 varied-size
// pinned-reference probes have identical hashes, as do 224 PGM/PPM probes
// spanning declared maxima 1–65535. Nine P7/PAM golden controls and 336
// varied tuple, depth and max-value hashes also match exactly. QOI and farbfeld
// decoders match 36 further pinned-reference hashes.
// TIFF CMYK8/16 uses raw sample-preserving Go codec views followed by the
// reference's f32 color conversion. Palette, separated-plane and unsupported
// sample/color profiles stay undecided instead of producing incompatible hashes.
// The bounded TIFF RGB/RGBA32F reader decodes physical raw/LZW/Deflate/PackBits
// chunks, including integer and byte-plane predictors, before RGB8 conversion.
// JPEG-compressed TIFF uses the older locked raw-component profile, including
// shared tables and qualified progressive/subsampled layouts, in Go.
// Radiance HDR (raw and both RLE families) and DDS DXT1/3/5, including
// compatible DX10 layouts, use bounded Go readers. These paths require an
// image file type from the database or handler; they do not add default
// file-type magic absent from the pinned ClamAV comparator.
// The EXR profile reads the first supported non-deep RGB layer, crops to the
// display window, and decodes raw/RLE/ZIP/PIZ/PXR24/B44 scanlines and tiles.
// File-type selection remains database/handler controlled, as for HDR/DDS.
// Bounded PNG- and DIB-backed ICOs, including reference directory selection
// and validated RGB/bitfield layouts, match reference hashes on their certain bits;
// other icon layouts remain undecided.
// Simple and VP8X
// lossless WebP profiles, with or without alpha, match 120 further
// pinned-reference hashes. Bounded simple and extended lossy VP8 profiles,
// with or without alpha, use the reference's chroma interpolation. The
// 1–32 by 1–32 tiny VP8 profile evaluates the qualified resize math profiles;
// disagreements leave fuzzy rules undecidable; other WebP layouts remain incomplete.
// Ordinary JFIF and Adobe
// CMYK JPEGs through 2048 pixels
// and one million total pixels, including restart intervals, can make exact
// decisions unless a floating-point boundary is close. Other JPEG layouts
// and larger images retain decoder uncertainty.

const (
	// imageHashJPEGError is the fallback per-pixel difference allowance for
	// JPEG layouts outside the exact-decision profile.
	imageHashJPEGError = 1.0
	// imageHashMaxPixels bounds the images decoded.
	imageHashMaxPixels = 40 << 20
	// A very narrow image could otherwise allocate a huge 32-row temporary.
	imageHashMaxDimension = 1 << 14
)

// imageHash is one object's hash: uncertain marks the bits that could
// differ from the reference's; ok is false when the object is not an image
// this engine decodes.
type imageHash struct {
	hash, uncertain uint64
	ok              bool
}

// fuzzyImageHash computes the hash of an image object.
func fuzzyImageHash(data []byte, w *matchWork) imageHash {
	if w.ctx.Err() != nil {
		return imageHash{}
	}
	var decode func([]byte) (image.Image, error)
	lossy := false
	gifFile := false
	webpFile := false
	webpLossy := false
	webpAlpha := false
	icoFile := false
	var ico *icoImage
	var hdr *hdrInfo
	var dds *ddsInfo
	var exr *exr154.Decoder
	var tif *tiffProfile

	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		decode = decodePNGReferenceTerminal

	case len(data) >= 4 && bytes.Equal(data[:4], []byte{0, 0, 1, 0}):
		selected, ok := selectedICOImage(data)
		if !ok {
			return imageHash{}
		}
		var err error
		ico, err = readICOImage(selected)
		if err != nil {
			return imageHash{}
		}
		decode = func([]byte) (image.Image, error) { return ico.decode(w) }
		icoFile = true

	case bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")):
		decode = func(b []byte) (image.Image, error) { return gif.Decode(bytes.NewReader(b)) }
		gifFile = true

	case bytes.HasPrefix(data, []byte("BM")):
		if bmpRLECandidate(data) {
			decode = func(b []byte) (image.Image, error) { return decodeBMPRLEWork(b, w) }
		} else if bmpBitfieldCandidate(data) {
			decode = decodeBMPBitfields
		} else if bmpPlainCandidate(data) {
			decode = decodeBMPPlain
		} else {
			decode = func(b []byte) (image.Image, error) { return bmp.Decode(bytes.NewReader(b)) }
		}

	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		decode = func(b []byte) (image.Image, error) { return jpeg154.Decode(bytes.NewReader(b)) }
		lossy = true

	case bytes.HasPrefix(data, []byte("II*\x00")) || bytes.HasPrefix(data, []byte("MM\x00*")):
		var err error
		tif, err = readTIFFProfile(data, w)
		if err != nil {
			return imageHash{}
		}
		decode = func([]byte) (image.Image, error) { return tif.decode() }

	case len(data) >= 3 && data[0] == 'P' && data[1] >= '1' && data[1] <= '6' && pnmSpace(data[2]):
		decode = decodePNM

	case bytes.HasPrefix(data, []byte("P7\n")):
		decode = decodePAM

	case bytes.HasPrefix(data, []byte("qoif")):
		decode = decodeQOI

	case bytes.HasPrefix(data, []byte("farbfeld")):
		decode = decodeFarbfeld

	case bytes.HasPrefix(data, []byte("#?RADIANCE")):
		info, err := readHDRHeader(data)
		if err != nil {
			return imageHash{}
		}
		hdr = &info
		decode = func(b []byte) (image.Image, error) { return decodeHDR(b, info, w) }

	case bytes.HasPrefix(data, []byte("DDS ")):
		info, err := readDDSHeader(data)
		if err != nil {
			return imageHash{}
		}
		dds = &info
		decode = func(b []byte) (image.Image, error) { return decodeDDS(b, info, w) }

	case bytes.HasPrefix(data, []byte{0x76, 0x2f, 0x31, 1}):
		var err error
		exr, err = exr154.Read(data, exr154.Limits{MaxDimension: imageHashMaxDimension, MaxPixels: imageHashMaxPixels, Step: func(n int64) error {
			if err := w.ctx.Err(); err != nil {
				return err
			}
			return w.step(n)
		}})
		if err != nil {
			return imageHash{}
		}
		decode = func([]byte) (image.Image, error) { return exr.Decode() }

	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		decode = func(b []byte) (image.Image, error) { return webp.Decode(bytes.NewReader(b)) }
		webpFile = true

	default:
		return imageHash{}
	}
	if webpFile {
		webpAlpha = extendedWebPAlphaLossy(data)
		webpLossy = webpLossyProfile(data)
		if !webpLossy && !webpLosslessProfile(data) {
			return imageHash{}
		}
	}
	var cfg image.Config
	var err error
	switch {
	case lossy:
		cfg, err = jpeg154.DecodeConfig(bytes.NewReader(data))
	case data[0] == 'G':
		cfg, err = gif.DecodeConfig(bytes.NewReader(data))
	case data[0] == 'B':
		if bmpRLECandidate(data) {
			cfg, err = bmpRLEConfig(data)
		} else if bmpBitfieldCandidate(data) {
			cfg, err = bmpBitfieldConfig(data)
		} else if bmpPlainCandidate(data) {
			cfg, err = bmpPlainConfig(data)
		} else {
			cfg, err = bmp.DecodeConfig(bytes.NewReader(data))
		}
	case data[0] == 'I' || data[0] == 'M':
		cfg, err = tif.config()
	case data[0] == 'P':
		if data[1] == '7' {
			cfg, err = pamConfig(data)
		} else {
			cfg, err = pnmConfig(data)
		}
	case data[0] == 'q':
		cfg, err = qoiConfig(data)
	case data[0] == 'f':
		cfg, err = farbfeldConfig(data)
	case webpFile:
		cfg, err = webp.DecodeConfig(bytes.NewReader(data))
	case icoFile:
		cfg = ico.config
	case hdr != nil:
		cfg = hdr.config()
	case dds != nil:
		cfg = dds.config()
	case exr != nil:
		cfg = exr.Config()
	default:
		cfg, err = png.DecodeConfig(bytes.NewReader(data))
	}
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > imageHashMaxDimension ||
		cfg.Height > imageHashMaxDimension || int64(cfg.Width)*int64(cfg.Height) > imageHashMaxPixels {
		return imageHash{}
	}
	if webpFile && (cfg.Width > 512 || cfg.Height > 512) {
		return imageHash{}
	}
	// Decoding and resizing touch every pixel a few times.
	if w.step(int64(len(data))/64+3*int64(cfg.Width)*int64(cfg.Height)/8+1) != nil {
		return imageHash{}
	}
	img, err := decode(data)
	if err != nil || w.ctx.Err() != nil || img.Bounds().Dx() != cfg.Width || img.Bounds().Dy() != cfg.Height {
		return imageHash{}
	}
	if gifFile {
		var ok bool
		img, ok = gifKeepTransparentColor(data, img)
		if !ok {
			return imageHash{}
		}
	}
	if webpLossy {
		if webpAlpha {
			// Decode the complete ALPH wrapper first so malformed alpha data
			// cannot turn a damaged image into a confident VP8-only verdict.
			if w.step(int64(cfg.Width)*int64(cfg.Height)/4+1) != nil {
				return imageHash{}
			}
			vp8Only := webpVP8Only(data)
			if vp8Only == nil {
				return imageHash{}
			}
			img, err = webp.Decode(bytes.NewReader(vp8Only))
			if err != nil || img.Bounds().Dx() != cfg.Width || img.Bounds().Dy() != cfg.Height {
				return imageHash{}
			}
		}
		ycbcr, ok := img.(*image.YCbCr)
		if !ok || ycbcr.SubsampleRatio != image.YCbCrSubsampleRatio420 {
			return imageHash{}
		}
		img = webpFancyRGB(ycbcr)
	}
	if lossy {
		if ycbcr, ok := img.(*image.YCbCr); ok {
			img = &referenceJPEGImage{ycbcr}
		} else if cmyk, ok := img.(*image.CMYK); ok {
			img = &referenceCMYKImage{cmyk}
		}
	}
	jpegError := 0.0
	if lossy {
		jpegError = imageHashJPEGError
		if referenceJPEGProfile(data) {
			switch img.(type) {
			case *referenceJPEGImage, *referenceCMYKImage, *image.Gray:
				jpegError = 0
			}
		}
	}
	h, u := perceptualHash(img, jpegError)
	if w.ctx.Err() != nil {
		return imageHash{}
	}
	return imageHash{hash: h, uncertain: u, ok: true}
}

// The pinned ClamAV image decoder accepts a zero-length terminal non-IDAT
// chunk even when it is not named IEND or its CRC is damaged or incomplete.
// Go's PNG decoder rejects it. Substitute only that final chunk's type and
// checksum through the reader, without copying or changing the original
// object seen by other signature matchers. A terminal IDAT is not enough for
// the reference, and other malformed chunks still fail the ordinary decoder.
func decodePNGReferenceTerminal(data []byte) (image.Image, error) {
	end := []byte{'I', 'E', 'N', 'D', 0xae, 0x42, 0x60, 0x82}
	for kept := 4; kept >= 0; kept-- {
		if len(data) < 8+kept {
			continue
		}
		start := len(data) - 8 - kept
		if !bytes.Equal(data[start:start+4], []byte{0, 0, 0, 0}) ||
			bytes.Equal(data[start+4:start+8], []byte("IDAT")) {
			continue
		}
		if kept == 4 && bytes.Equal(data[start+4:], end) {
			return png.Decode(bytes.NewReader(data))
		}
		return png.Decode(io.MultiReader(bytes.NewReader(data[:start+4]), bytes.NewReader(end)))
	}
	return png.Decode(bytes.NewReader(data))
}

// gifKeepTransparentColor restores the first frame's original color at its
// transparent index. Go's decoder replaces that palette entry with black;
// the reference keeps its RGB color and drops alpha for hashing. An unknown
// palette cannot yield a certain hash.
func gifKeepTransparentColor(data []byte, img image.Image) (image.Image, bool) {
	p, ok := img.(*image.Paletted)
	if !ok {
		return img, true
	}
	transparent := false
	for _, c := range p.Palette {
		if _, _, _, a := c.RGBA(); a == 0 {
			transparent = true
			break
		}
	}
	if !transparent {
		return img, true
	}
	table, ok := gifFirstColorTable(data)
	if !ok {
		return nil, false
	}
	pal := make(color.Palette, len(p.Palette))
	copy(pal, p.Palette)
	for i, c := range pal {
		if _, _, _, a := c.RGBA(); a == 0 {
			if 3*(i+1) > len(table) {
				return nil, false
			}
			pal[i] = color.NRGBA{table[3*i], table[3*i+1], table[3*i+2], 0}
		}
	}
	q := *p
	q.Palette = pal
	return &q, true
}

// gifFirstColorTable reads the palette actually used by the first image.
// A GIF may have a global palette, a local one, or both; graphic-control and
// other extension blocks can precede the image descriptor.
func gifFirstColorTable(data []byte) ([]byte, bool) {
	if len(data) < 13 {
		return nil, false
	}
	palette := func(at int, packed byte) ([]byte, int, bool) {
		n := 3 * (2 << (packed & 7))
		if at > len(data) || n > len(data)-at {
			return nil, 0, false
		}
		return data[at : at+n], at + n, true
	}
	at := 13
	var global []byte
	if data[10]&0x80 != 0 {
		var ok bool
		global, at, ok = palette(at, data[10])
		if !ok {
			return nil, false
		}
	}
	for at < len(data) {
		switch data[at] {
		case 0x21: // extension: label, then size-prefixed sub-blocks
			if at+2 > len(data) {
				return nil, false
			}
			at += 2
			for {
				if at >= len(data) {
					return nil, false
				}
				n := int(data[at])
				at++
				if n == 0 {
					break
				}
				if n > len(data)-at {
					return nil, false
				}
				at += n
			}
		case 0x2c: // image descriptor, then optional local palette
			if len(data)-at < 10 {
				return nil, false
			}
			packed := data[at+9]
			if packed&0x80 != 0 {
				local, _, ok := palette(at+10, packed)
				return local, ok
			}
			return global, global != nil
		default:
			return nil, false
		}
	}
	return nil, false
}

// rgb8 returns a pixel's 8-bit red, green and blue: alpha is ignored and a
// transparent pixel keeps its color; 16-bit values are rounded.
func rgb8(c color.Color) (float64, float64, float64) {
	r16 := func(v uint16) float64 { return math.Round(float64(v) / 257) }
	switch c := c.(type) {
	case color.NRGBA:
		return float64(c.R), float64(c.G), float64(c.B)
	case color.NRGBA64:
		return r16(c.R), r16(c.G), r16(c.B)
	case color.RGBA64:
		return r16(c.R), r16(c.G), r16(c.B)
	case color.Gray16:
		v := r16(c.Y)
		return v, v, v
	}
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return float64(n.R), float64(n.G), float64(n.B)
}

// lumaProducts keeps the three f32 multiplications separate from the sums.
// The reference rounds each product to f32 before adding; fusing a multiply
// and add changes the 8-bit result at half-integer boundaries.
//
//go:noinline
func lumaProducts(r, g, b float32) (float32, float32, float32) {
	return float32(299.0/1000.0) * r, float32(587.0/1000.0) * g, float32(114.0/1000.0) * b
}

// grey is the reference's 8-bit luma calculation.
func grey(r, g, b float64) float32 {
	rp, gp, bp := lumaProducts(float32(r), float32(g), float32(b))
	v := rp + gp + bp
	return float32(math.Round(float64(v)))
}

// greyPlane converts an image to grey values, reading the common layouts
// directly (the generic per-pixel path is several times slower).
func greyPlane(img image.Image) [][]float32 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([][]float32, h)
	for y := range out {
		out[y] = make([]float32, w)
	}
	switch m := img.(type) {
	case *referenceCMYKImage:
		for y := 0; y < h; y++ {
			row := m.Pix[y*m.Stride : y*m.Stride+4*w]
			for x := 0; x < w; x++ {
				p := row[4*x:]
				k := 255 - p[3]
				r := referenceCMYKProduct(255-p[0], k)
				g := referenceCMYKProduct(255-p[1], k)
				bl := referenceCMYKProduct(255-p[2], k)
				out[y][x] = grey(float64(r), float64(g), float64(bl))
			}
		}
	case *referenceJPEGImage:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				px, py := b.Min.X+x, b.Min.Y+y
				yv := m.Y[m.YOffset(px, py)]
				cb := referenceJPEGChroma(m.YCbCr, px, py, true)
				cr := referenceJPEGChroma(m.YCbCr, px, py, false)
				r, g, bl := referenceJPEGToRGB(yv, cb, cr)
				out[y][x] = grey(float64(r), float64(g), float64(bl))
			}
		}
	case *image.NRGBA:
		for y := 0; y < h; y++ {
			row := m.Pix[y*m.Stride : y*m.Stride+4*w]
			for x := 0; x < w; x++ {
				out[y][x] = grey(float64(row[4*x]), float64(row[4*x+1]), float64(row[4*x+2]))
			}
		}
	case *image.Gray:
		for y := 0; y < h; y++ {
			row := m.Pix[y*m.Stride : y*m.Stride+w]
			for x := 0; x < w; x++ {
				out[y][x] = float32(row[x])
			}
		}
	case *image.Paletted:
		var table [256]float32
		for i, c := range m.Palette {
			table[i] = grey(rgb8(c))
		}
		for y := 0; y < h; y++ {
			row := m.Pix[y*m.Stride : y*m.Stride+w]
			for x := 0; x < w; x++ {
				out[y][x] = table[row[x]]
			}
		}
	case *image.YCbCr:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c := m.YCbCrAt(b.Min.X+x, b.Min.Y+y)
				r, g, bl := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
				out[y][x] = grey(float64(r), float64(g), float64(bl))
			}
		}
	default:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out[y][x] = grey(rgb8(img.At(b.Min.X+x, b.Min.Y+y)))
			}
		}
	}
	return out
}

// cosTable[k][i] is cos(pi/32 (i+1/2) k).
var cosTable = func() (t [32][32]float64) {
	for k := 0; k < 32; k++ {
		for i := 0; i < 32; i++ {
			t[k][i] = math.Cos(math.Pi / 32 * (float64(i) + 0.5) * float64(k))
		}
	}
	return t
}()

// perceptualHash evaluates the owned Mac, Linux and pinned Windows resize
// math profiles. Pixels are transformed independently; the union of differing
// hash bits remains uncertain.
// Decoder uncertainty retains its separate DCT movement bound. This
// replaces the empirical half-integer margin; additional native math/decoder
// profiles still require qualification before claiming complete image parity.
func perceptualHash(img image.Image, pixelError float64) (uint64, uint64) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	grey := greyPlane(img)
	t, alternate, windows := grey, grey, grey
	if w != 32 || h != 32 {
		t = resample32(resample32(grey, h, true), w, false)
		alternate = resample32Profile(resample32Profile(grey, h, true, resizeSine32), w, false, resizeSine32)
		windows = resample32Profile(resample32Profile(grey, h, true, resizeWindowsSine32), w, false, resizeWindowsSine32)
	}
	var px, alternatePixels, windowsPixels [32][32]uint8
	var errPix [32][32]float64
	boundary, windowsBoundary := false, false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			v := math.Min(math.Max(float64(t[y][x]), 0), 255)
			v2 := math.Min(math.Max(float64(alternate[y][x]), 0), 255)
			v3 := math.Min(math.Max(float64(windows[y][x]), 0), 255)
			linuxError := math.Abs(math.Round(v) - math.Round(v2))
			windowsError := math.Abs(math.Round(v) - math.Round(v3))
			errPix[y][x] = math.Max(linuxError, windowsError)
			boundary = boundary || linuxError != 0
			windowsBoundary = windowsBoundary || windowsError != 0
			errPix[y][x] += pixelError
			px[y][x] = uint8(math.Round(v))
			alternatePixels[y][x] = uint8(math.Round(v2))
			windowsPixels[y][x] = uint8(math.Round(v3))
		}
	}
	if pixelError == 0 {
		hash := resizedPixelHash(&px)
		var uncertain uint64
		if boundary {
			uncertain |= hash ^ resizedPixelHash(&alternatePixels)
		}
		if windowsBoundary {
			uncertain |= hash ^ resizedPixelHash(&windowsPixels)
		}
		return hash, uncertain
	}
	coef := dct32.Coefficients(&px)
	sorted := coef
	sort.Slice(sorted[:], func(i, j int) bool { return sorted[i] < sorted[j] })
	median := (sorted[31] + sorted[32]) / 2
	var hash uint64
	for i, c := range coef {
		if c > median {
			hash |= uint64(1) << (63 - i)
		}
	}
	// Retain pixel/resize uncertainty. The transform is normalized by
	// 4/255; the cosine weights bound a pixel perturbation's contribution.
	var moves [64]float64
	maxMove := 0.0
	for ky := 0; ky < 8; ky++ {
		for kx := 0; kx < 8; kx++ {
			e := 0.0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					e += errPix[y][x] * math.Abs(cosTable[ky][y]) * math.Abs(cosTable[kx][x])
				}
			}
			moves[8*ky+kx] = e * 4 / 255
			maxMove = math.Max(maxMove, moves[8*ky+kx])
		}
	}
	var uncertain uint64
	for i, c := range coef {
		// The added guard covers f32 transform roundoff when comparing
		// two different pixel populations, not ambiguity in a fixed input.
		if math.Abs(float64(c)-float64(median)) <= moves[i]+maxMove+0.1 {
			uncertain |= uint64(1) << (63 - i)
		}
	}
	return hash, uncertain
}

// With exact decoded pixels each qualified math profile yields a definite
// hash. Comparing those hashes avoids introducing uncertainty when different
// rounded pixels nevertheless produce the same signature decision.
func resizedPixelHash(px *[32][32]uint8) uint64 {
	flat := true
	for y := 0; y < 32 && flat; y++ {
		for x := 0; x < 32; x++ {
			if px[y][x] != px[0][0] {
				flat = false
				break
			}
		}
	}
	if flat {
		if px[0][0] > 0 {
			return 1 << 63
		}
		return 0
	}
	coef := dct32.Coefficients(px)
	ordered := coef
	sort.Slice(ordered[:], func(i, j int) bool { return ordered[i] < ordered[j] })
	median := (ordered[31] + ordered[32]) / 2
	var hash uint64
	for i, value := range coef {
		if value > median {
			hash |= uint64(1) << (63 - i)
		}
	}
	return hash
}
