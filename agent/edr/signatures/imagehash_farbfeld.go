package signatures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

var errFarbfeld = errors.New("invalid farbfeld image")

func farbfeldConfig(data []byte) (image.Config, error) {
	if len(data) < 16 || !bytes.Equal(data[:8], []byte("farbfeld")) {
		return image.Config{}, errFarbfeld
	}
	w, h := binary.BigEndian.Uint32(data[8:12]), binary.BigEndian.Uint32(data[12:16])
	if w == 0 || h == 0 || w > imageHashMaxDimension || h > imageHashMaxDimension ||
		uint64(w)*uint64(h) > imageHashMaxPixels {
		return image.Config{}, errFarbfeld
	}
	return image.Config{ColorModel: color.NRGBA64Model, Width: int(w), Height: int(h)}, nil
}

// Farbfeld stores row-major RGBA16 big-endian samples without compression.
func decodeFarbfeld(data []byte) (image.Image, error) {
	cfg, err := farbfeldConfig(data)
	if err != nil || (len(data)-16)/8 < cfg.Width*cfg.Height {
		return nil, errFarbfeld
	}
	img := image.NewNRGBA64(image.Rect(0, 0, cfg.Width, cfg.Height))
	copy(img.Pix, data[16:16+len(img.Pix)])
	return img, nil
}
