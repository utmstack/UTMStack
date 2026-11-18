package engine

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/mszip"
)

// cabMSZIPReader supplies the same 4096-byte stream reads as the reference
// adapter. A failure in a later record rejects that entire read, including
// bytes copied from earlier records during the call.
type cabMSZIPReader struct {
	ctx                  context.Context
	data                 []byte
	pos, blocks, reserve int
	input                *int
	extent               *int
	payload              []byte
	err                  error
}

type cabMSZIPInputError struct {
	frameStart uint64
	cause      error
}

func (e *cabMSZIPInputError) Error() string { return e.cause.Error() }
func (e *cabMSZIPInputError) Unwrap() error { return e.cause }

func (r *cabMSZIPReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.err != nil {
		if errors.Is(r.err, ErrLimit) {
			return 0, mszip.ErrLimit
		}
		return 0, r.err
	}
	n := 0
	for n < len(p) {
		if err := r.ctx.Err(); err != nil {
			r.err = err
			return 0, err
		}
		if len(r.payload) != 0 {
			i := copy(p[n:], r.payload)
			r.payload = r.payload[i:]
			n += i
			continue
		}
		if r.blocks == 0 {
			if n == 0 {
				return 0, io.EOF
			}
			return n, nil
		}
		if r.pos < 0 || r.reserve < 0 || len(r.data)-r.pos < 8+r.reserve {
			r.err = ErrMalformed
			return 0, r.err
		}
		header := r.data[r.pos : r.pos+8]
		size := int(binary.LittleEndian.Uint16(header[4:6]))
		output := binary.LittleEndian.Uint16(header[6:8])
		r.pos += 8 + r.reserve
		recordCABExtent(r.extent, r.pos)
		if size > len(r.data)-r.pos {
			r.err = ErrMalformed
			return 0, r.err
		}
		recordCABExtent(r.extent, r.pos+size)
		if output == 0 {
			r.err = ErrUnsupported
			return 0, r.err
		}
		cost := 8 + r.reserve + size
		if cost > *r.input {
			r.err = ErrLimit
			return 0, mszip.ErrLimit
		}
		*r.input -= cost
		r.payload = r.data[r.pos : r.pos+size]
		r.pos += size
		r.blocks--
	}
	return n, nil
}

func (s *scanState) cabMSZIP(data []byte, folder cabFolder, reserve int, input *int, requestedEnd uint64, maxOutput int64, extent *int) ([]byte, error) {
	r := &cabMSZIPReader{ctx: s.ctx, data: data, pos: folder.start, blocks: folder.blocks, reserve: reserve, input: input, extent: extent}
	target := min(requestedEnd, uint64(maxOutput))
	out, _, err := mszip.DecodeContext(s.ctx, r, int(target))
	if ctxErr := s.ctx.Err(); ctxErr != nil {
		return out, ctxErr
	}
	if errors.Is(err, mszip.ErrLimit) || errors.Is(r.err, ErrLimit) {
		return out, ErrLimit
	}
	if err != nil {
		var failed *mszip.InputError
		if errors.As(err, &failed) {
			cause := r.err
			if cause == nil {
				cause = ErrMalformed
			}
			return out, &cabMSZIPInputError{frameStart: uint64(failed.FrameStart), cause: cause}
		}
		if r.err != nil {
			return out, r.err
		}
		return out, fmt.Errorf("CAB MSZIP: %v: %w", err, ErrMalformed)
	}
	if requestedEnd > uint64(maxOutput) {
		return out, ErrLimit
	}
	return out, nil
}
