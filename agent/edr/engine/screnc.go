package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// decodeScrEnc unwraps one Microsoft Script Encoder envelope. The length is
// the count of encoded symbols, not the count of decoded bytes: CR/LF consume
// neither, @ escapes consume two and high-bit pairs consume one. The reference
// scans partially decoded streams; callers receive partial bytes with an error.
func decodeScrEnc(ctx context.Context, data []byte) ([]byte, error) {
	start := bytes.Index(data, []byte("#@~^"))
	if start < 0 || len(data)-start < 4+8 {
		return nil, unknownRepresentation(formatReason(signatures.TypeScrEnc))
	}
	header := data[start+4 : start+12]
	var lengthBytes [4]byte
	if n, err := base64.StdEncoding.Decode(lengthBytes[:], header); err != nil || n != 4 {
		return nil, unknownRepresentation(formatReason(signatures.TypeScrEnc))
	}
	remaining := binary.LittleEndian.Uint32(lengthBytes[:])
	if remaining > maxHTMLNormalize {
		return nil, ErrLimit
	}
	i := start + 12
	out := make([]byte, 0, min(int(remaining), 4096))
	position := 0
	lastCheck := i
	for remaining > 0 {
		if i-lastCheck >= 4096 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			lastCheck = i
		}
		if i >= len(data) || data[i] == 0 {
			return out, unknownRepresentation(formatReason(signatures.TypeScrEnc))
		}
		c := data[i]
		i++
		if c == '\n' || c == '\r' {
			continue
		}
		if c >= 0x80 {
			if i >= len(data) || data[i] == 0 {
				return out, unknownRepresentation(formatReason(signatures.TypeScrEnc))
			}
			out = append(out, c, data[i])
			i++
			remaining--
			if len(out) > maxHTMLNormalize-17 {
				return nil, ErrLimit
			}
			continue
		}
		v := screncTables[screncOrder[position]][c]
		if v == 0xff {
			if remaining < 2 || i >= len(data) {
				return out, unknownRepresentation(formatReason(signatures.TypeScrEnc))
			}
			switch data[i] {
			case '!':
				v = '<'
			case '#':
				v = '\r'
			case '$':
				v = '@'
			case '&':
				v = '\n'
			case '*':
				v = '>'
			}
			i++
			remaining--
		}
		out = append(out, v)
		position = (position + 1) & 63
		remaining--
		if len(out) > maxHTMLNormalize-17 {
			return nil, ErrLimit
		}
	}
	if len(data)-i < 12 {
		// The reference leaves a too-short checksum/terminator suffix in
		// its in-place buffer, so the generated HTML includes those bytes.
		suffix := data[i:]
		if end := bytes.IndexByte(suffix, 0); end >= 0 {
			suffix = suffix[:end]
		}
		if len(out)+len(suffix) > maxHTMLNormalize-17 {
			return nil, ErrLimit
		}
		return append(out, suffix...), unknownRepresentation(formatReason(signatures.TypeScrEnc))
	}
	if !bytes.Equal(data[i+8:i+12], []byte("^#~@")) {
		return out, unknownRepresentation(formatReason(signatures.TypeScrEnc))
	}
	return out, nil
}

func (s *scanState) screnc(data []byte, depth int) (string, error) {
	decoded, err := decodeScrEnc(s.ctx, data)
	if decoded == nil {
		return "", err
	}
	wrapper := make([]byte, 0, len(decoded)+17)
	wrapper = append(wrapper, "<script>"...)
	wrapper = append(wrapper, decoded...)
	wrapper = append(wrapper, "</script>"...)
	if sig, childErr := s.childBytes(wrapper, depth); sig != "" || childErr != nil {
		return sig, childErr
	}
	return "", err
}

// decodeScriptChunk performs the reference's in-place HTML Script Encoder
// transform on the current chunk. A completed envelope removes its twelve
// checksum/terminator bytes when they fit in this chunk; an incomplete body
// is carried into the next chunk with its table position intact.
func (h *htmlNormalizer) decodeScriptChunk(buf []byte, from int) []byte {
	read, write := from, from
	for h.screncRemaining > 0 && read < len(buf) {
		c := buf[read]
		if c == '\n' || c == '\r' {
			read++
			continue
		}
		if c >= 0x80 {
			if read+1 >= len(buf) {
				h.views.unsupported = "script-encoder-boundary"
				break
			}
			buf[write], buf[write+1] = c, buf[read+1]
			write += 2
			read += 2
			h.screncRemaining--
			continue
		}
		value := screncTables[screncOrder[h.screncPosition]][c]
		read++
		if value == 0xff {
			if read >= len(buf) || h.screncRemaining < 2 {
				h.views.unsupported = "script-encoder-boundary"
				break
			}
			switch buf[read] {
			case '!':
				value = '<'
			case '#':
				value = '\r'
			case '$':
				value = '@'
			case '&':
				value = '\n'
			case '*':
				value = '>'
			}
			read++
			h.screncRemaining--
		}
		buf[write] = value
		write++
		h.screncPosition = (h.screncPosition + 1) & 63
		h.screncRemaining--
	}
	if h.screncRemaining > 0 {
		return buf[:write]
	}
	h.screncActive = false
	if len(buf)-read >= 12 {
		read += 12
	}
	copy(buf[write:], buf[read:])
	return buf[:write+len(buf)-read]
}
