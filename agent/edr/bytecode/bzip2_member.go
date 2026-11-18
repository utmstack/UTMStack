package bytecode

import (
	"context"
	"fmt"
	"io"
)

// Bzip2MemberReader decodes one bzip2 member using the same Go decoder as the
// bytecode pipe API. Publication follows ClamAV 1.5.4's file scanner: 8192-byte
// input/output packets, with the current unflushed packet discarded on error.
// Caller read sizes do not alter which damaged-stream bytes are published.
type Bzip2MemberReader struct {
	ctx                 context.Context
	input               []byte
	at, end             int
	decoder             streamBzip2
	buffer              [8192]byte
	used, ready, pos    int
	produced, delivered int64
	terminal            error
	closed              bool
}

func NewBzip2MemberReader(ctx context.Context, input []byte) *Bzip2MemberReader {
	return &Bzip2MemberReader{ctx: ctx, input: input}
}

// CompressedBytes includes the last byte containing the member's end marker
// and CRC, but excludes trailing bytes and subsequent members.
func (r *Bzip2MemberReader) CompressedBytes() int { return r.at }

// DecodedBytes includes bytes discarded after an error or buffered ahead of
// the caller. The scanner charges these bytes to its expansion budget too.
func (r *Bzip2MemberReader) DecodedBytes() int64 { return r.produced }

func (r *Bzip2MemberReader) DeliveredBytes() int64 { return r.delivered }

func (r *Bzip2MemberReader) Close() error {
	r.closed = true
	r.decoder.tt = nil
	r.input = nil
	return nil
}

func (r *Bzip2MemberReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	for {
		if err := r.ctx.Err(); err != nil {
			r.decoder.tt = nil
			return 0, err
		}
		if r.pos < r.ready {
			n := copy(p, r.buffer[r.pos:r.ready])
			r.pos += n
			r.delivered += int64(n)
			return n, nil
		}
		r.ready, r.pos = 0, 0
		if r.terminal != nil {
			return 0, r.terminal
		}
		if r.at == len(r.input) {
			r.terminal = io.ErrUnexpectedEOF
			r.used = 0
			r.decoder.tt = nil
			continue
		}
		if r.at == r.end {
			r.end = min(len(r.input), r.at+len(r.buffer))
		}
		n, m, status, err := r.decoder.process(r.input[r.at:r.end], r.buffer[r.used:],
			func(uint64) error { return r.ctx.Err() })
		r.at += n
		r.used += m
		r.produced += int64(m)
		if err != nil || status < 0 {
			if err == nil {
				err = fmt.Errorf("bzip2: invalid member (status %d)", status)
			}
			r.terminal = err
			r.used = 0
			r.decoder.tt = nil
			continue
		}
		if status == bzipEnd || r.used == len(r.buffer) {
			r.ready, r.used = r.used, 0
			if status == bzipEnd {
				r.terminal = io.EOF
				r.decoder.tt = nil
			}
			continue
		}
		if n == 0 && m == 0 && r.at < r.end {
			r.terminal = io.ErrNoProgress
			r.used = 0
			r.decoder.tt = nil
		}
	}
}
