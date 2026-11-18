package exr154

import _ "embed"

//go:embed b44_log.bin
var b44Log []byte

func unpackB44(data []byte, h header, w, ht, expected int, l Limits) ([]byte, error) {
	c := cursor{data: data}
	planes := make([][]byte, len(h.channels))
	for ci, ch := range h.channels {
		if err := l.charge(int64(w)*int64(ht)/8 + 1); err != nil {
			return nil, err
		}
		if ch.kind != 1 {
			b := c.take(w * ht * 4)
			if c.err != nil {
				return nil, c.err
			}
			planes[ci] = b
			continue
		}
		plane := make([]byte, w*ht*2)
		planes[ci] = plane
		for y := 0; y < ht; y += 4 {
			if err := l.charge(int64(w)/8 + 1); err != nil {
				return nil, err
			}
			for x := 0; x < w; x += 4 {
				b := c.take(3)
				if b == nil {
					return nil, c.err
				}
				var s [16]uint16
				initial := uint16(b[0])<<8 | uint16(b[1])
				if b[2] >= 52 {
					for i := range s {
						s[i] = initial
					}
				} else {
					tail := c.take(11)
					if tail == nil {
						return nil, c.err
					}
					var block [14]byte
					copy(block[:3], b)
					copy(block[3:], tail)
					shift := uint(block[2] >> 2)
					bias := uint32(32) << shift
					s[0] = initial
					nodes := [15]int{4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14, 3, 7, 11, 15}
					parents := [15]int{0, 4, 8, 0, 4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14}
					bit := 22
					for i, to := range nodes {
						var delta uint32
						for n := 0; n < 6; n++ {
							delta = delta<<1 | uint32(block[bit/8]>>uint(7-bit%8)&1)
							bit++
						}
						s[to] = uint16(uint32(s[parents[i]]) + (delta << shift) - bias)
					}
				}
				for i, v := range s {
					if v&0x8000 != 0 {
						v &= 0x7fff
					} else {
						v = ^v
					}
					if ch.linear {
						v = le.Uint16(b44Log[int(v)*2:])
					}
					s[i] = v
				}
				for yy := 0; yy < 4 && y+yy < ht; yy++ {
					for xx := 0; xx < 4 && x+xx < w; xx++ {
						le.PutUint16(plane[((y+yy)*w+x+xx)*2:], s[yy*4+xx])
					}
				}
			}
		}
	}
	out := make([]byte, 0, expected)
	for y := 0; y < ht; y++ {
		for ci, ch := range h.channels {
			line := w * ch.size
			out = append(out, planes[ci][y*line:(y+1)*line]...)
		}
	}
	return out, nil
}
