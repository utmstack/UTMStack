package sevenzip

import "bytes"

// decode runs one coder on its inputs; a closed done stops the slow coders.
func decode(c Coder, in [][]byte, size uint64, done <-chan struct{}) ([]byte, error) {
	id := c.ID
	one := func() ([]byte, error) {
		if len(in) != 1 {
			return nil, ErrFormat
		}
		return in[0], nil
	}
	filter := func(f func([]byte)) ([]byte, error) {
		src, err := one()
		if err != nil {
			return nil, err
		}
		if uint64(len(src)) != size {
			return nil, ErrData
		}
		if len(c.Props) != 0 && !bytes.Equal(c.Props, []byte{0, 0, 0, 0}) {
			return nil, ErrUnsupported // a start offset
		}
		out := append([]byte(nil), src...)
		f(out)
		return out, nil
	}
	switch {
	case bytes.Equal(id, []byte{0}):
		src, err := one()
		if err != nil {
			return nil, err
		}
		if uint64(len(src)) < size {
			return nil, ErrData
		}
		return src[:size], nil
	case bytes.Equal(id, []byte{3, 1, 1}):
		src, err := one()
		if err != nil {
			return nil, err
		}
		return decodeLZMA(c.Props, src, size, done)
	case bytes.Equal(id, []byte{0x21}):
		src, err := one()
		if err != nil {
			return nil, err
		}
		return decodeLZMA2(c.Props, src, size, done)
	case bytes.Equal(id, []byte{3, 3, 1, 3}):
		return filter(x86Decode)
	case bytes.Equal(id, []byte{3, 3, 5, 1}):
		return filter(armDecode)
	case bytes.Equal(id, []byte{3, 3, 1, 0x1b}):
		if len(in) != 4 {
			return nil, ErrFormat
		}
		return bcj2Decode(in[0], in[1], in[2], in[3], size)
	case bytes.Equal(id, []byte{6, 0xf1, 7, 1}):
		return nil, ErrEncrypted
	case bytes.Equal(id, []byte{3, 4, 1}):
		src, err := one()
		if err != nil {
			return nil, err
		}
		return decodePPMd(c.Props, src, size, done)
	}
	return nil, ErrUnsupported
}

// stopped reports whether done is closed.
func stopped(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
