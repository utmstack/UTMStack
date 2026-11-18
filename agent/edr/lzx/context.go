package lzx

import "context"

// DecodeContext decodes bounded output, checking cancellation between 32 KiB
// frames. On failure it returns decoded frames so callers can charge expansion
// work even when they discard the partial output.
func DecodeContext(ctx context.Context, src []byte, windowBits, outLen, resetFrames int) ([]byte, error) {
	return decodeContext(ctx, src, windowBits, outLen, resetFrames, nil)
}

// DecodeValidatedContext returns only completed, translated frames after an
// error, alongside the total bytes produced, including the unfinished frame.
// Callers can scan that prefix while still charging all expansion work.
func DecodeValidatedContext(ctx context.Context, src []byte, windowBits, outLen, resetFrames int) (out []byte, produced int, err error) {
	valid := 0
	out, err = decodeContext(ctx, src, windowBits, outLen, resetFrames, &valid)
	produced = len(out)
	if err != nil {
		out = out[:valid]
	}
	return
}

func decodeContext(ctx context.Context, src []byte, windowBits, outLen, resetFrames int, valid *int) ([]byte, error) {
	if outLen < 0 {
		return nil, ErrData
	}
	d, err := NewDecoder(src, windowBits, resetFrames)
	if err != nil {
		return nil, err
	}
	for len(d.out) < outLen {
		if err := ctx.Err(); err != nil {
			return d.out, err
		}
		if err := d.frame(min(frameSize, outLen-len(d.out))); err != nil {
			return d.out, err
		}
		if valid != nil {
			*valid = len(d.out)
		}
	}
	return d.out, ctx.Err()
}
