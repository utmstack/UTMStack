package bytecode

import "bytes"

// bytewiseReader prevents complete-stream codecs from reading beyond their
// end marker and hiding bytes the bytecode caller must still be able to read.
type bytewiseReader struct{ *bytes.Reader }

func (r bytewiseReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.Reader.Read(p)
}
