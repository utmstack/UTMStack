package engine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

// zipLZMA handles APPNOTE 5.8's method-14 framing: SDK version (two
// bytes), property length (two bytes), five LZMA properties, then the raw
// range-coded stream. Reconstruct only the fixed classic LZMA header for
// the existing decoder; no untrusted size causes an input copy/allocation.
func (s *scanState) zipLZMA(packed []byte, flags uint16, size uint64, depth int) (string, error) {
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	if len(packed) < 9 {
		return "", ErrMalformed
	}
	if binary.LittleEndian.Uint16(packed[2:]) != 5 {
		return "", ErrUnsupported
	}
	if packed[4] >= 225 {
		return "", ErrMalformed
	}
	// Read the dictionary before asking the codec to construct its state.
	// ReaderConfig is also an upper bound in the pinned v0.5.17 codec.
	if binary.LittleEndian.Uint32(packed[5:]) > maxCodecDictionary {
		return "", ErrLimit
	}
	var header [13]byte
	copy(header[:5], packed[4:9])
	if flags&2 != 0 {
		// Method 14's bit 1 means an end-of-stream marker is present.
		binary.LittleEndian.PutUint64(header[5:], ^uint64(0))
	} else {
		limit := min(s.maxFile, s.remaining)
		if limit < 0 || size > uint64(limit) {
			return "", ErrLimit
		}
		binary.LittleEndian.PutUint64(header[5:], size)
	}
	input := io.MultiReader(bytes.NewReader(header[:]), bytes.NewReader(packed[9:]))
	r, err := (lzma.ReaderConfig{DictCap: maxCodecDictionary}).NewReader(contextReader{s.ctx, input})
	if err != nil {
		if s.ctx.Err() != nil {
			return "", s.ctx.Err()
		}
		var dictionary *lzma.ErrDictSize
		if errors.As(err, &dictionary) {
			return "", ErrLimit
		}
		return "", ErrMalformed
	}
	return s.child(r, depth)
}
