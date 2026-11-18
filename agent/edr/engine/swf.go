package engine

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

// SWF structure follows Adobe's SWF 19 specification, chapters 2, 5, 13 and 15:
// https://open-flash.github.io/mirrors/swf-spec-19.pdf
// This extracts content; it does not execute or emulate ActionScript.
func (s *scanState) swf(data []byte, depth int) (string, error) {
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if len(data) < 8 || string(data[1:3]) != "WS" || data[3] == 0 {
		return "", ErrMalformed
	}
	declared := uint64(binary.LittleEndian.Uint32(data[4:8]))
	if declared < 13 {
		return "", ErrMalformed
	}
	switch data[0] {
	case 'Z':
		if data[3] < 13 || len(data) < 17 {
			return "", ErrMalformed
		}
		limit := min(s.maxFile, s.remaining)
		if depth >= s.maxDepth || s.files >= s.maxFiles || limit < 8 || declared > uint64(limit) {
			return "", ErrLimit
		}
		packed := uint64(binary.LittleEndian.Uint32(data[8:12]))
		// Publishers use both compressed-data-only and properties-inclusive
		// sizes. Both refer to the same complete immutable input, never a slice
		// that could hide appended content from scanning.
		if packed != uint64(len(data)-17) && packed != uint64(len(data)-12) {
			return "", ErrMalformed
		}
		if binary.LittleEndian.Uint32(data[13:17]) > maxCodecDictionary {
			return "", ErrLimit
		}
		header := make([]byte, 13)
		copy(header, data[12:17])
		binary.LittleEndian.PutUint64(header[5:], declared-8)
		input := bytes.NewReader(data[17:])
		r, err := (lzma.ReaderConfig{DictCap: maxCodecDictionary}).NewReader(contextReader{s.ctx, io.MultiReader(bytes.NewReader(header), input)})
		if err != nil {
			return "", ErrMalformed
		}
		body, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, limit-8+1))
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if int64(len(body)) > limit-8 {
			s.spend(len(body))
			return "", ErrLimit
		}
		decoded := append([]byte(nil), data[:8]...)
		decoded[0] = 'F'
		decoded = append(decoded, body...)
		sig, scanErr := s.visit(decoded, depth+1)
		if sig != "" {
			return sig, nil
		}
		if readErr != nil || input.Len() != 0 {
			return "", ErrMalformed
		}
		return "", scanErr
	case 'C':
		if data[3] < 6 {
			return "", ErrMalformed
		}
		limit := min(s.maxFile, s.remaining)
		if depth >= s.maxDepth || s.files >= s.maxFiles || limit < 8 || declared > uint64(limit) {
			return "", ErrLimit
		}
		compressed := bytes.NewReader(data[8:])
		r, err := zlib.NewReader(compressed)
		if err != nil {
			return "", ErrMalformed
		}
		body, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, limit-8+1))
		_ = r.Close()
		if s.ctx.Err() != nil {
			return "", s.ctx.Err()
		}
		if int64(len(body)) > limit-8 {
			s.spend(len(body))
			return "", ErrLimit
		}
		decoded := make([]byte, 8, 8+len(body))
		copy(decoded, data[:8])
		decoded[0] = 'F'
		decoded = append(decoded, body...)
		// Matching the normalized object first preserves detections even if its
		// checksum, declared length, or a later tag is malformed.
		sig, scanErr := s.visit(decoded, depth+1)
		if sig != "" {
			return sig, nil
		}
		if readErr != nil || compressed.Len() != 0 {
			return "", ErrMalformed
		}
		return "", scanErr
	case 'F':
		if declared != uint64(len(data)) {
			return "", ErrMalformed
		}
	default:
		return "", ErrMalformed
	}
	if len(data) < 9 {
		return "", ErrMalformed
	}
	bits := int(data[8] >> 3)
	rectBytes := (5 + 4*bits + 7) / 8
	start := 8 + rectBytes + 4 // RECT, frame rate, frame count
	if start > len(data) {
		return "", ErrMalformed
	}
	return s.swfTags(data[start:], depth)
}

func (s *scanState) swfTags(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	for pos := 0; pos < len(data); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if s.files >= s.maxFiles {
			return "", ErrLimit
		}
		s.files++
		if len(data)-pos < 2 {
			return "", ErrMalformed
		}
		header := binary.LittleEndian.Uint16(data[pos : pos+2])
		pos += 2
		code, size := header>>6, uint64(header&63)
		if size == 63 {
			if len(data)-pos < 4 {
				return "", ErrMalformed
			}
			size = uint64(binary.LittleEndian.Uint32(data[pos : pos+4]))
			pos += 4
			if size > 0x7fffffff {
				return "", ErrMalformed
			}
		}
		if size > uint64(len(data)-pos) {
			return "", ErrMalformed
		}
		payload := data[pos : pos+int(size)]
		pos += int(size)
		if code == 0 {
			if size != 0 || pos != len(data) {
				return "", ErrMalformed
			}
			return "", incomplete
		}
		if len(payload) > 0 {
			if sig, err := s.visit(payload, depth+1); sig != "" {
				return sig, nil
			} else if err != nil {
				incomplete = err
			}
		}
		var child []byte
		switch code {
		case 1: // ShowFrame
			if size != 0 {
				incomplete = ErrMalformed
			}
		case 9: // SetBackgroundColor
			if size != 3 {
				incomplete = ErrMalformed
			}
		case 12: // DoAction: action records were scanned above.
		case 59: // DoInitAction: skip sprite identifier.
			if len(payload) < 2 {
				incomplete = ErrMalformed
			} else {
				child = payload[2:]
			}
		case 69: // FileAttributes
			if size != 4 {
				incomplete = ErrMalformed
			}
		case 77: // Metadata (NUL-terminated XML)
			if len(payload) == 0 || payload[len(payload)-1] != 0 {
				incomplete = ErrMalformed
			} else {
				child = payload[:len(payload)-1]
			}
		case 82: // DoABC: flags, NUL-terminated name, ABC bytecode.
			if len(payload) < 5 {
				incomplete = ErrMalformed
				break
			}
			n := bytes.IndexByte(payload[4:], 0)
			if n < 0 {
				incomplete = ErrMalformed
			} else {
				child = payload[5+n:]
			}
		case 87: // DefineBinaryData: character id, reserved uint32, arbitrary data.
			if len(payload) < 6 || binary.LittleEndian.Uint32(payload[2:6]) != 0 {
				incomplete = ErrMalformed
			} else {
				child = payload[6:]
			}
		case 39: // DefineSprite embeds a tag sequence after id and frame count.
			if len(payload) < 4 {
				incomplete = ErrMalformed
				break
			}
			if sig, err := s.swfTags(payload[4:], depth+1); sig != "" {
				return sig, nil
			} else if err != nil {
				incomplete = err
			}
		default:
			// Other tags receive raw matching, but may contain encoded images,
			// fonts, sound or video. Do not imply those codecs were decoded.
			incomplete = ErrUnsupported
		}
		if len(child) > 0 {
			if sig, err := s.visit(child, depth+1); sig != "" {
				return sig, nil
			} else if err != nil {
				incomplete = err
			}
		}
	}
	return "", ErrMalformed // an End tag is mandatory
}
