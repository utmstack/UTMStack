package signatures

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Shared physical strip/tile geometry; sample interpretation belongs to the
// individual codec. All source offsets/counts remain tied to the real input.
type tiffChunks struct {
	cfg                              image.Config
	order                            binary.ByteOrder
	channels, predictor, compression int
	bits, photo                      int
	blockWidth, blockHeight, across  int
	offsets, lengths                 []uint32
}

func readTIFFChunks(entries map[uint16][]byte, order binary.ByteOrder, p *tiffProfile,
	values func(uint16, uint32) ([]uint32, error), floatingFormat bool) (*tiffChunks, error) {
	scalar := func(tag uint16, fallback uint32) (int, error) {
		v, err := values(tag, fallback)
		if err != nil || len(v) != 1 {
			return 0, errTIFFProfile
		}
		return int(v[0]), nil
	}
	f := &tiffChunks{order: order}
	w, err := scalar(256, 0)
	if err != nil {
		return nil, err
	}
	h, err := scalar(257, 0)
	if err != nil || w <= 0 || h <= 0 || w > imageHashMaxDimension || h > imageHashMaxDimension || int64(w)*int64(h) > imageHashMaxPixels {
		return nil, errTIFFProfile
	}
	f.cfg = image.Config{Width: w, Height: h, ColorModel: color.NRGBAModel}
	if f.channels, err = scalar(277, 1); err != nil || f.channels < 1 || f.channels > 4 {
		return nil, errTIFFProfile
	}
	bits, err := values(258, 1)
	if err != nil || len(bits) != 1 && len(bits) != f.channels {
		return nil, errTIFFProfile
	}
	f.bits = int(bits[0])
	if f.photo, err = scalar(262, 0); err != nil {
		return nil, err
	}
	if f.predictor, err = scalar(317, 1); err != nil || f.predictor < 1 || f.predictor > 3 || f.predictor == 3 && !floatingFormat {
		return nil, errTIFFProfile
	}
	f.compression, err = scalar(259, 1)
	if err != nil {
		return nil, err
	}
	switch f.compression {
	case 1, 5, 7, 8, 32946, 32773:
	default:
		// The pinned TIFF dependency has no ZSTD feature enabled. Do not
		// turn its unsupported-compression rejection into a confident hash.
		return nil, errTIFFProfile
	}
	array := func(tag uint16) ([]uint32, error) {
		e := entries[tag]
		if e == nil {
			return nil, errTIFFProfile
		}
		n := uint64(order.Uint32(e[4:]))
		if n == 0 || n > 1<<20 || p.work.step(int64(n)+1) != nil {
			return nil, errTIFFProfile
		}
		var size uint64
		switch order.Uint16(e[2:]) {
		case 1:
			size = 1
		case 3:
			size = 2
		case 4:
			size = 4
		default:
			return nil, errTIFFProfile
		}
		raw := e[8:12]
		if n*size > 4 {
			off := uint64(order.Uint32(raw))
			if off > uint64(len(p.data)) || n*size > uint64(len(p.data))-off {
				return nil, errTIFFProfile
			}
			raw = p.data[off : off+n*size]
		}
		out := make([]uint32, n)
		for i := range out {
			switch size {
			case 1:
				out[i] = uint32(raw[i])
			case 2:
				out[i] = uint32(order.Uint16(raw[2*i:]))
			case 4:
				out[i] = order.Uint32(raw[4*i:])
			}
		}
		return out, nil
	}
	strip := entries[273] != nil || entries[279] != nil
	tile := entries[324] != nil || entries[325] != nil
	if strip == tile {
		return nil, errTIFFProfile
	}
	offsetTag, countTag := uint16(273), uint16(279)
	f.blockWidth, f.blockHeight = w, h
	if tile {
		if f.blockWidth, err = scalar(322, 0); err != nil {
			return nil, err
		}
		if f.blockHeight, err = scalar(323, 0); err != nil {
			return nil, err
		}
		offsetTag, countTag = 324, 325
	} else if f.blockHeight, err = scalar(278, uint32(h)); err != nil {
		return nil, err
	}
	if f.blockWidth <= 0 || f.blockHeight <= 0 || (int64(f.blockWidth)*int64(f.channels)*int64(f.bits)+7)/8 > 8<<20 {
		return nil, errTIFFProfile
	}
	f.across = (w-1)/f.blockWidth + 1
	count := int64(f.across) * int64((h-1)/f.blockHeight+1)
	if count > 1<<20 {
		return nil, errTIFFProfile
	}
	if f.offsets, err = array(offsetTag); err != nil {
		return nil, err
	}
	if f.lengths, err = array(countTag); err != nil || int64(len(f.offsets)) != count || len(f.lengths) != len(f.offsets) {
		return nil, errTIFFProfile
	}
	return f, nil
}
