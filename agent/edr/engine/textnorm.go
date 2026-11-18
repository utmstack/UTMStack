package engine

// normalizeText is the reference's normalization of text for its text rules
// (target 7): a run of white space (0x09-0x0D, 0x20) becomes one space, A-Z
// are lowered, 0x21-0x7F are kept, and every other byte is dropped without
// changing the white-space state. So UTF-16 text loses its NUL bytes and any
// byte of 0x80 or more disappears. With chunk > 0 the white-space state is
// reset every chunk input bytes, as the reference's file path does.
func normalizeText(data []byte, chunk int) []byte {
	out := make([]byte, 0, len(data))
	space := false
	for i, c := range data {
		if chunk > 0 && i > 0 && i%chunk == 0 {
			space = false
		}
		switch {
		case c >= 0x09 && c <= 0x0d || c == ' ':
			if !space {
				out = append(out, ' ')
				space = true
			}
		case c >= 'A' && c <= 'Z':
			out, space = append(out, c+'a'-'A'), false
		case c >= 0x21 && c <= 0x7f:
			out, space = append(out, c), false
		}
	}
	return out
}
