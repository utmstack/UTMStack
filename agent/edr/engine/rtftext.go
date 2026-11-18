package engine

import (
	"bytes"
	"context"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
	"golang.org/x/text/encoding/charmap"
)

// RTF text is an additional defensive representation, not an extracted file
// and not a complete renderer. In particular, it never receives file hashes,
// EOF/executable offsets, type recognition, or recursive container decoding.
// The existing raw/object scans remain unchanged. Text/control semantics come
// from MSFT-RTF 1.9.1 (Unicode RTF and RTF syntax) and MS-OXRTFEX 2.2.3.3.
func (s *scanState) rtfText(data []byte, depth int) (string, error) {
	if depth+1 > s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	text, decodeErr := decodeRTFText(s.ctx, data, min(s.maxFile, s.remaining))
	if len(text) < minAnalyzedObject {
		s.spend(len(text))
		return "", decodeErr
	}
	if done, err := s.admitObject(text, depth+1, 0); done {
		s.spend(len(text)) // discarded decoded work, not an admitted object
		return "", err
	}
	s.layers.Push(signatures.Layer{Type: signatures.TypeTextUTF8, Size: int64(len(text)), Normalized: true})
	defer s.layers.Pop()
	res, err := s.db.MatchObject(s.ctx, text, signatures.MatchContext{Type: signatures.TypeAny, Mode: signatures.BufferMode, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth + 1), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
	if res.Name != "" {
		return res.Name, nil
	}
	if decodeErr != nil {
		return "", decodeErr
	}
	return "", err
}

const (
	maxRTFTextGroups = 256
	maxRTFTextFonts  = 256
)

type rtfTextGroup struct {
	skip, fontTable, separate bool
	uc, font, page            int
}

type rtfTextFont struct{ page, charset int }

type rtfTextDecoder struct {
	limit   int64
	out     []byte
	err     error
	pending rune // high UTF-16 surrogate, if any
}

func (d *rtfTextDecoder) note(err error) {
	if d.err == nil {
		d.err = err
	}
}

func (d *rtfTextDecoder) rune(r rune) error {
	n := utf8.RuneLen(r)
	if n < 0 {
		r, n = utf8.RuneError, 3
		d.note(ErrMalformed)
	}
	if int64(len(d.out))+int64(n) > d.limit {
		return ErrLimit
	}
	if cap(d.out)-len(d.out) < n {
		capacity := min(d.limit, max(int64(4096), int64(cap(d.out))*2))
		b := make([]byte, len(d.out), int(capacity))
		copy(b, d.out)
		d.out = b
	}
	d.out = utf8.AppendRune(d.out, r)
	return nil
}

func (d *rtfTextDecoder) flushSurrogate() error {
	if d.pending == 0 {
		return nil
	}
	d.pending = 0
	d.note(ErrMalformed)
	return d.rune(utf8.RuneError)
}

func (d *rtfTextDecoder) unicode(value int) error {
	r := rune(uint16(value))
	if d.pending != 0 {
		if r >= 0xdc00 && r <= 0xdfff {
			r = utf16.DecodeRune(d.pending, r)
			d.pending = 0
			return d.rune(r)
		}
		if err := d.flushSurrogate(); err != nil {
			return err
		}
	}
	if r >= 0xd800 && r <= 0xdbff {
		d.pending = r
		return nil
	}
	if r >= 0xdc00 && r <= 0xdfff {
		d.note(ErrMalformed)
		r = utf8.RuneError
	}
	return d.rune(r)
}

func (d *rtfTextDecoder) literal(value byte, page int, font rtfTextFont, known bool) error {
	if err := d.flushSurrogate(); err != nil {
		return err
	}
	if font.page != 0 {
		page = font.page
	}
	// Unsupported font/DBCS interpretations must not manufacture ASCII runs
	// from trail bytes or Symbol-font glyph indices. Keep a separator instead.
	if !known || page != 1252 || font.charset != 0 && font.charset != 1 {
		d.note(ErrUnsupported)
		return d.rune(0)
	}
	r := charmap.Windows1252.DecodeByte(value)
	if r == utf8.RuneError {
		d.note(ErrMalformed)
	}
	return d.rune(r)
}

// decodeRTFText uses a fixed group bound and bounded font map. All emitted
// bytes fit limit before allocation; cancellation is checked during the walk.
// Unknown starred destinations, binary/object/image data and metadata never
// become body text. Alternate Unicode/HTML representations and non-1252 font
// encodings are incomplete rather than guessed. Unicode escapes are decoded
// independently of a font's legacy byte encoding.
func decodeRTFText(ctx context.Context, data []byte, limit int64) ([]byte, error) {
	d := rtfTextDecoder{limit: max(0, limit)}
	if !bytes.HasPrefix(data, []byte(`{\rtf1`)) || len(data) == 6 || isRTFLetter(data[6]) || data[6] >= '0' && data[6] <= '9' {
		return nil, ErrMalformed
	}
	groups := make([]rtfTextGroup, 0, maxRTFTextGroups)
	fonts := make(map[int]rtfTextFont)
	state := rtfTextGroup{uc: 1, page: 1252}
	fallback, defaultFont, nextCheck := 0, 0, 0
	for pos := 0; pos < len(data); {
		if pos >= nextCheck {
			if err := ctx.Err(); err != nil {
				return d.out, err
			}
			nextCheck = pos + 4096
		}
		c := data[pos]
		pos++
		switch c {
		case '{':
			fallback = 0 // group delimiters terminate Unicode fallback
			if len(groups) >= maxRTFTextGroups {
				return d.out, ErrLimit
			}
			groups = append(groups, state)
		case '}':
			fallback = 0
			if len(groups) == 0 {
				return d.out, ErrMalformed
			}
			parent := groups[len(groups)-1]
			if state.separate && !parent.separate && !state.skip {
				if err := d.flushSurrogate(); err != nil {
					return d.out, err
				}
				if err := d.rune(0); err != nil {
					return d.out, err
				}
			}
			state, groups = parent, groups[:len(groups)-1]
			if len(groups) == 0 {
				if len(bytes.TrimSpace(data[pos:])) != 0 {
					return d.out, ErrMalformed
				}
				if err := d.flushSurrogate(); err != nil {
					return d.out, err
				}
				return d.out, d.err
			}
		case '\r', '\n': // lexical line breaks are not text or fallback
		case '\\':
			if pos >= len(data) {
				return d.out, ErrMalformed
			}
			c = data[pos]
			pos++
			if !isRTFLetter(c) {
				symbol := c
				if c == '\'' {
					if len(data)-pos < 2 {
						return d.out, ErrMalformed
					}
					hi, a := hexNibble(data[pos])
					lo, b := hexNibble(data[pos+1])
					if !a || !b {
						return d.out, ErrMalformed
					}
					c, pos = hi<<4|lo, pos+2
				} else if c == '*' && fallback == 0 {
					state.skip = true
					continue
				} else if c != '\\' && c != '{' && c != '}' && c != '~' && c != '-' && c != '_' && c != '\r' && c != '\n' {
					if fallback > 0 {
						fallback--
					}
					continue
				}
				if fallback > 0 {
					fallback--
					continue
				}
				if state.skip {
					continue
				}
				var err error
				if symbol == '\r' || symbol == '\n' {
					err = d.flushSurrogate()
					if err == nil {
						err = d.rune('\r')
					}
					if err == nil {
						err = d.rune('\n')
					}
				} else if symbol == '~' || symbol == '-' || symbol == '_' {
					r := map[byte]rune{'~': '\u00a0', '-': '\u00ad', '_': '\u2011'}[symbol]
					err = d.flushSurrogate()
					if err == nil {
						err = d.rune(r)
					}
				} else {
					font, known := fonts[state.font]
					err = d.literal(c, state.page, font, known || len(fonts) == 0)
				}
				if err != nil {
					return d.out, err
				}
				continue
			}
			start := pos - 1
			for pos < len(data) && isRTFLetter(data[pos]) && pos-start <= 32 {
				pos++
			}
			if pos-start > 32 {
				return d.out, ErrMalformed
			}
			word := string(data[start:pos])
			value, parameter, negative := int64(0), false, false
			if pos < len(data) && data[pos] == '-' {
				negative, pos = true, pos+1
			}
			maxValue := int64(2147483647)
			if negative {
				maxValue++
			}
			for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
				parameter = true
				value = value*10 + int64(data[pos]-'0')
				pos++
				if value > maxValue {
					return d.out, ErrMalformed
				}
			}
			if negative {
				if !parameter {
					return d.out, ErrMalformed
				}
				value = -value
			}
			if pos < len(data) && data[pos] == ' ' {
				pos++ // one delimiter, never part of Unicode fallback
			}
			if word == "bin" {
				if !parameter || value < 0 || value > int64(len(data)-pos) {
					return d.out, ErrMalformed
				}
				pos += int(value)
			}
			if fallback > 0 {
				fallback-- // a whole control word or bin block is one unit
				continue
			}
			switch word {
			case "fonttbl":
				state.skip, state.fontTable = true, true
			case "f", "af":
				if !parameter || value < 0 {
					return d.out, ErrMalformed
				}
				state.font = int(value)
				if state.fontTable {
					if _, ok := fonts[state.font]; !ok && len(fonts) >= maxRTFTextFonts {
						return d.out, ErrLimit
					}
					if _, ok := fonts[state.font]; !ok {
						fonts[state.font] = rtfTextFont{}
					}
				}
			case "fcharset", "cpg":
				if state.fontTable {
					if !parameter || value < 0 || value > 65535 || word == "fcharset" && value > 255 {
						return d.out, ErrMalformed
					}
					if _, ok := fonts[state.font]; !ok && len(fonts) >= maxRTFTextFonts {
						return d.out, ErrLimit
					}
					font := fonts[state.font]
					if word == "cpg" {
						font.page = int(value)
					} else {
						font.charset = int(value)
					}
					fonts[state.font] = font
				}
			case "deff":
				if !parameter || value < 0 {
					return d.out, ErrMalformed
				}
				defaultFont, state.font = int(value), int(value)
			case "plain":
				state.font = defaultFont
			case "ansi":
				state.page = 1252
			case "ansicpg":
				if !parameter || value < 1 || value > 65535 {
					return d.out, ErrMalformed
				}
				state.page = int(value)
			case "mac", "pc", "pca":
				state.page = -1
			case "uc":
				if !parameter || value < 0 || value > 32767 {
					return d.out, ErrMalformed
				}
				state.uc = int(value)
			case "upr", "fromhtml":
				if !state.skip {
					d.note(ErrUnsupported)
					state.skip = true
				}
			case "ud":
				// Unicode alternatives need selection/correlation with their
				// ANSI branch. Do not silently discard a recognized alternative.
				d.note(ErrUnsupported)
				state.skip = true
			default:
				if rtfNonTextDestination(word) {
					state.skip = true
				}
			}
			if state.skip {
				continue
			}
			// Headers, notes and shape text are independent text runs. They
			// remain inspectable but cannot join unrelated main-body runs.
			switch word {
			case "header", "headerl", "headerr", "headerf", "footer", "footerl", "footerr", "footerf", "footnote", "annotation", "shptxt":
				if err := d.flushSurrogate(); err != nil {
					return d.out, err
				}
				if err := d.rune(0); err != nil {
					return d.out, err
				}
				state.separate = true
			}
			if word == "u" {
				if !parameter || value < -32768 || value > 32767 {
					return d.out, ErrMalformed
				}
				if err := d.unicode(int(value)); err != nil {
					return d.out, err
				}
				fallback = state.uc
				continue
			}
			text := rtfTextControl(word)
			if word == "bin" {
				d.note(ErrUnsupported)
				text = "\x00" // binary data cannot join unrelated body runs
			}
			switch word {
			case "chdate", "chdpl", "chdpa", "chpgn", "chatn", "chftn", "chftnsep", "chftnsepc":
				d.note(ErrUnsupported)
				text = "\x00" // a dynamic field is not a zero-width format
			}
			if text != "" {
				if err := d.flushSurrogate(); err != nil {
					return d.out, err
				}
				for _, r := range text {
					if err := d.rune(r); err != nil {
						return d.out, err
					}
				}
			}
		default:
			if fallback > 0 {
				fallback--
				continue
			}
			if !state.skip {
				font, known := fonts[state.font]
				if err := d.literal(c, state.page, font, known || len(fonts) == 0); err != nil {
					return d.out, err
				}
			}
		}
	}
	return d.out, ErrMalformed
}

func rtfNonTextDestination(word string) bool {
	switch word {
	case "colortbl", "stylesheet", "info", "filetbl", "revtbl", "rsidtbl", "listtable", "listoverridetable", "listpicture", "generator", "pict", "shppict", "nonshppict", "object", "objdata", "objclass", "objname", "objtime", "fldinst", "datafield", "themedata", "colorschememapping", "latentstyles", "xmlnstbl", "datastore", "docvar", "bkmkstart", "bkmkend", "pnseclvl", "private", "userprops":
		return true
	}
	return false
}

func rtfTextControl(word string) string {
	switch word {
	case "par", "line", "page", "sect", "row":
		return "\r\n"
	case "tab", "cell":
		return "\t"
	case "emdash":
		return "\u2014"
	case "endash":
		return "\u2013"
	case "bullet":
		return "\u2022"
	case "lquote":
		return "\u2018"
	case "rquote":
		return "\u2019"
	case "ldblquote":
		return "\u201c"
	case "rdblquote":
		return "\u201d"
	case "emspace":
		return "\u2003"
	case "enspace":
		return "\u2002"
	case "qmspace":
		return "\u2005"
	case "zwj":
		return "\u200d"
	case "zwnj":
		return "\u200c"
	case "zwbo":
		return "\u200b"
	case "zwnbo":
		return "\u2060"
	case "ltrmark":
		return "\u200e"
	case "rtlmark":
		return "\u200f"
	}
	return ""
}
