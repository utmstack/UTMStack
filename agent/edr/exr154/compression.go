package exr154

import (
	"bytes"
	"compress/zlib"
	"io"
)

type chargedReader struct {
	r io.Reader
	l Limits
}

func (r chargedReader) Read(b []byte) (int, error) {
	if err := r.l.charge(int64(len(b))/64 + 1); err != nil {
		return 0, err
	}
	return r.r.Read(b)
}
func inflate(data []byte, max int, l Limits) ([]byte, error) {
	z, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, ErrFormat
	}
	defer z.Close()
	out, err := io.ReadAll(io.LimitReader(chargedReader{z, l}, int64(max)+1))
	if err != nil {
		return nil, err
	}
	if len(out) > max {
		return nil, ErrLimit
	}
	return out, nil
}
func unpack(data []byte, h header, w, ht, expected int, l Limits) ([]byte, error) {
	// Every codec may store an uncompressed block when compression grows it.
	if len(data) == expected {
		return data, nil
	}
	switch h.compression {
	case 0:
		return nil, ErrFormat
	case 1:
		out := make([]byte, 0, expected)
		c := cursor{data: data}
		for c.pos < len(data) && len(out) < expected && c.err == nil {
			if err := l.charge(1); err != nil {
				return nil, err
			}
			v := c.take(1)
			if v == nil {
				break
			}
			n := int(int8(v[0]))
			if n < 0 {
				b := c.take(-n)
				if len(b) > expected-len(out) {
					return nil, ErrFormat
				}
				out = append(out, b...)
			} else {
				b := c.take(1)
				if b == nil {
					break
				}
				n++
				if n > expected-len(out) {
					return nil, ErrFormat
				}
				for i := 0; i < n; i++ {
					out = append(out, b[0])
				}
			}
		}
		if c.err != nil || len(out) != expected {
			return nil, ErrFormat
		}
		return undoBytePrediction(out, l)
	case 2, 3:
		out, err := inflate(data, expected, l)
		if err != nil {
			return nil, err
		}
		if len(out) != expected {
			return nil, ErrFormat
		}
		return undoBytePrediction(out, l)
	case 4:
		return unpackPIZ(data, h, w, ht, expected, l)
	case 5:
		return unpackPXR(data, h, w, ht, expected, l)
	case 6, 7:
		return unpackB44(data, h, w, ht, expected, l)
	default:
		return nil, ErrFormat
	}
}
func undoBytePrediction(data []byte, l Limits) ([]byte, error) {
	for i := 1; i < len(data); i++ {
		if i&4095 == 0 {
			if err := l.charge(513); err != nil {
				return nil, err
			}
		}
		data[i] = data[i-1] + data[i] - 128
	}
	out := make([]byte, len(data))
	split := (len(data) + 1) / 2
	for i := range out {
		if i&4095 == 0 {
			if err := l.charge(513); err != nil {
				return nil, err
			}
		}
		if i&1 == 0 {
			out[i] = data[i/2]
		} else {
			out[i] = data[split+i/2]
		}
	}
	return out, nil
}
func unpackPXR(data []byte, h header, w, ht, expected int, l Limits) ([]byte, error) {
	b, err := inflate(data, expected, l)
	if err != nil {
		return nil, err
	}
	c := cursor{data: b}
	out := make([]byte, 0, expected)
	for y := 0; y < ht; y++ {
		for _, ch := range h.channels {
			n := ch.size
			if ch.kind == 2 {
				n = 3
			}
			planes := c.take(n * w)
			if c.err != nil {
				return nil, c.err
			}
			var previous uint32
			for x := 0; x < w; x++ {
				var difference uint32
				for p := 0; p < n; p++ {
					difference = difference<<8 | uint32(planes[p*w+x])
				}
				if ch.kind == 2 {
					difference <<= 8
				}
				previous += difference
				if ch.size == 2 {
					out = le.AppendUint16(out, uint16(previous))
				} else {
					out = le.AppendUint32(out, previous)
				}
			}
		}
	}
	return out, nil
}
