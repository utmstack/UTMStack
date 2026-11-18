package engine

// cryptFF scans the decoded body of a CryptFF object as a new child layer.
// The format has a 16-byte header; every subsequent byte is XORed with 0xff.
func (s *scanState) cryptFF(data []byte, depth int) (string, error) {
	if len(data) <= 16 {
		return "", nil
	}
	body := data[16:]
	if int64(len(body)) > s.maxFile || int64(len(body)) > s.remaining {
		return "", ErrLimit
	}
	decoded := make([]byte, len(body))
	for i, b := range body {
		decoded[i] = b ^ 0xff
	}
	return s.visit(decoded, depth+1)
}
