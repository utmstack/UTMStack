package engine

import "bytes"

// pdfSimpleObjectFlags decides the ClamAV PARSED flags for complete objects
// that cannot contain a stream or change a different object's flags. ClamAV
// begins each discovered object with flags zero, sets OBJ_DICT on a complete
// dictionary, and visits slash names in order. Its other name states and
// reference propagation are deliberately left unknown here.
func pdfSimpleObjectFlags(body []byte, kind byte, hasStream bool) (uint32, bool) {
	if bytes.Contains(body, []byte("endobj")) {
		return 0, false
	}
	var streamFlag uint32
	if hasStream {
		// Prove one ordinary stream delimiter pair, then restrict name
		// interpretation to the dictionary. Binary stream bytes are not names.
		body = bytes.TrimSpace(body)
		open := []byte("\nstream\n")
		close := []byte("\nendstream")
		i := bytes.Index(body, open)
		if i < 0 || !bytes.HasSuffix(body, close) ||
			bytes.Contains(body[:i], []byte("stream")) ||
			bytes.HasPrefix(body[i+len(open):len(body)-len(close)], []byte("stream\n")) ||
			bytes.Contains(body[i+len(open):len(body)-len(close)], open) ||
			bytes.Contains(body[i+len(open):len(body)-len(close)], []byte("endstream")) {
			return 0, false
		}
		body = body[:i]
		streamFlag = 1 << 0 // OBJ_STREAM
	} else if bytes.Contains(body, []byte("stream")) {
		return 0, false
	}
	if kind != 'd' {
		if bytes.Contains(body, []byte("<<")) {
			// ClamAV looks for the first dictionary token in an object's
			// raw bytes, even when those bytes are a PDF literal string.
			// An ordinary dictionary may follow the literal; the earlier
			// dictionary-like sequence wins if both are present.
			body = bytes.TrimSpace(body)
			if kind != 's' || hasStream || len(body) < 2 || body[0] != '(' {
				return 0, false
			}
			return pdfLeadingLiteralFlags(body)
		}
		return 0, true
	}
	body = bytes.TrimSpace(body)
	if len(body) < 4 || !bytes.HasPrefix(body, []byte("<<")) || !bytes.HasSuffix(body, []byte(">>")) {
		return 0, false
	}
	// The reference counts dictionary delimiters without excluding strings,
	// comments or arrays. Preserve that bounded shape before treating names
	// inside those values as ordinary slash names.
	depth := 0
	for i := 0; i+1 < len(body); i++ {
		switch {
		case body[i] == '<' && body[i+1] == '<':
			depth++
			i++
		case body[i] == '>' && body[i+1] == '>':
			depth--
			if depth < 0 || depth == 0 && i+2 != len(body) {
				return 0, false
			}
			i++
		}
	}
	if depth != 0 {
		return 0, false
	}
	inner := body[2 : len(body)-2]
	flags := streamFlag | 1<<1 // OBJ_STREAM and OBJ_DICT
	filterState, filterUsed, filterMatched := false, false, false
	for i := 0; i < len(inner); {
		c := inner[i]
		if c == '/' {
			i++
			start := i
			for i < len(inner) && !pdfWhite(inner[i]) && !pdfSimpleDelimiter(inner[i]) {
				if !pdfSimpleNameChar(inner[i]) {
					return 0, false
				}
				i++
			}
			if i == start {
				return 0, false
			}
			name := string(inner[start:i])
			if filterState && name != "Length" && name != "ASCIIHexDecode" && name != "AHx" && name != "FlateDecode" && name != "Fl" {
				return 0, false
			}
			switch name {
			case "Filter":
				if filterUsed {
					return 0, false
				}
				filterUsed, filterState = true, true
				flags |= 1 << 16 // OBJ_HASFILTERS
			case "ASCIIHexDecode", "AHx":
				if filterState {
					if filterMatched {
						return 0, false
					}
					filterMatched = true
					flags |= 1 << 3 // OBJ_FILTER_AH
				}
			case "FlateDecode", "Fl":
				if filterState {
					if filterMatched {
						return 0, false
					}
					filterMatched = true
					flags |= 1 << 5 // OBJ_FILTER_FLATE
				}
			case "Length":
				filterState = false
			case "Page":
				flags |= 1 << 23 // OBJ_PAGE
			case "Image":
				flags |= 1 << 18 // OBJ_IMAGE
			case "EmbeddedFile":
				flags |= 1 << 2
			case "Sig", "V", "R":
				flags |= 1 << 17 // OBJ_SIGNED
			case "JavaScript", "OpenAction", "Launch", "Contents", "URI", "Linearized", "S":
				return 0, false // parser state or cross-object effects
			}
			continue
		}
		if !pdfWhite(c) && !pdfSimpleValueChar(c) {
			return 0, false
		}
		i++
	}
	if filterUsed && !filterMatched {
		return 0, false
	}
	return flags, true
}

// pdfLeadingLiteralFlags accepts only a sequence of flat, unescaped PDF
// literal strings followed optionally by one ordinary dictionary. The
// reference's lexical object pass selects the first dictionary-like pair in
// the raw bytes, including pairs inside a literal. Its later pairs do not
// replace the first one's flags. Escaped or nested literals are kept unknown
// because their bounds are not established by this narrow profile.
func pdfLeadingLiteralFlags(body []byte) (uint32, bool) {
	var first []byte
	for len(body) != 0 {
		body = bytes.TrimSpace(body)
		if len(body) == 0 {
			break
		}
		if body[0] == '(' {
			close := bytes.IndexByte(body[1:], ')') + 1
			if close < 1 {
				return 0, false
			}
			literal := body[1:close]
			if bytes.IndexAny(literal, "\\()") >= 0 {
				return 0, false
			}
			if first == nil {
				if start := bytes.Index(literal, []byte("<<")); start >= 0 {
					end := bytes.Index(literal[start+2:], []byte(">>"))
					if end < 0 {
						return 0, false
					}
					first = literal[start : start+2+end+2]
				}
			}
			body = body[close+1:]
			continue
		}
		if !bytes.HasPrefix(body, []byte("<<")) || !bytes.HasSuffix(body, []byte(">>")) {
			return 0, false
		}
		if first == nil {
			first = body
		}
		break
	}
	if first == nil {
		return 0, false
	}
	return pdfSimpleObjectFlags(first, 'd', false)
}

func pdfSimpleNameChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}

func pdfSimpleValueChar(c byte) bool {
	return pdfSimpleNameChar(c) || c == '.' || c == '+' || c == '-' ||
		c == '(' || c == ')' || c == '[' || c == ']' || c == '<' || c == '>' ||
		c == '%'
}

func pdfSimpleDelimiter(c byte) bool {
	switch c {
	case '/', '(', ')', '[', ']', '<', '>', '%':
		return true
	default:
		return false
	}
}
