package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// The reference matches HTML only after normalizing it into views: the page
// without comments, lower-cased and with collapsed whitespace
// (nocomment.html), the text outside tags (notags.html), the scripts (the
// javascript view, jsnorm.go) and each data: URI of an attribute (an
// "rfc2397" message). normalizeHTML produces them with the reference's
// byte-level rules, including its quirks, which were learned from its
// output: a chunked reader of 8,191 bytes, a character state machine, and
// entity decoding with the entity table in htmlentities.go.

// htmlChunk is the reader's chunk size.
const htmlChunk = 8191

type htmlState int

const (
	hsSkipWS htmlState = iota
	hsTrimWS
	hsNorm
	hs8Bit
	hsTag
	hsTagArg
	hsTagArgEqual
	hsTagArgVal
	hsComment
	hsProcessTag
	hsCharRef
	hsEntityRefDecode
	hsCharRefDecode
	hsRFC2397Type
	hsRFC2397Init
	hsRFC2397Data
	hsRFC2397Finish
	hsRFC2397Esc
	hsEscapeChar
	hsLookForScrEnc
	hsJSDecode
	hsJSDecodeLength
	hsJSDecodeDecrypt
	hsBad
)

const (
	inNone = iota
	inScript
	inStyle
)

// htmlViews are an HTML object's normalized views.
type htmlViews struct {
	noComment, noTags []byte
	// scripts holds each script block's raw segments, in order, for the
	// JavaScript normalizer.
	scripts [][][]byte
	style   []byte
	// rfc2397 keeps both the reference's exact mail-shaped representation
	// (for rules and hashes) and the fields needed to decode its body.
	rfc2397 []htmlDataPart
	// unsupported names a malformed or unmodeled representation; the
	// object is then incomplete even when independent views can detect.
	unsupported string
}

// htmlDataPart is produced only by the HTML normalizer. compatibility and
// body share storage; MIME decoding never has to reparse our own header.
type htmlDataPart struct {
	compatibility []byte
	contentType   string
	encoding      string
	body          []byte
}

type htmlAttr struct{ name, value []byte }

type htmlNormalizer struct {
	views                htmlViews
	state, next          htmlState
	inTag                int
	tsw, binary          bool
	tag, arg, val, ename []byte
	quoted               byte
	escape               bool
	value                int64
	hex                  bool
	mb, mb2              uint32
	length               int
	attrs                []htmlAttr
	// script state: whether a block is open, and in the current chunk its
	// start and end (-1 when not in this chunk).
	scriptActive   bool
	jsBegin, jsEnd int
	segments       [][]byte
	styleBegin     int
	rfc            []byte
	rfcOpen        bool
	rfcPart        htmlDataPart
	rfcBody        int
	styleSeen      bool
	// Microsoft Script Encoder state is carried between 8,191-byte chunks.
	screncRemaining uint32
	screncPosition  int
	screncActive    bool
	lookForScrenc   bool
	pendingScrenc   int
	savedNext       htmlState
	// chunk is the chunk being processed.
	chunk []byte
}

func isHTMLSpace(c byte) bool { return c == ' ' || c >= 9 && c <= 13 }

func lowerByte(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

func lowerBytes(b []byte) []byte {
	out := make([]byte, len(b))
	for i, c := range b {
		out[i] = lowerByte(c)
	}
	return out
}

// appendLimited appends c unless b already holds 1,024 bytes.
func appendLimited(b []byte, c ...byte) []byte {
	for _, x := range c {
		if len(b) < 1024 {
			b = append(b, x)
		}
	}
	return b
}

// u16norm writes a decoded character: nothing for 0, the byte itself below
// 0xFF, a dot for the three ideographic and full-width stops, otherwise
// "&#x" and four hex digits.
func u16norm(u uint16) []byte {
	switch {
	case u == 0:
		return nil
	case u < 0xff:
		return []byte{byte(u)}
	case u == 0x3002 || u == 0xff0e || u == 0xfe52:
		return []byte{'.'}
	}
	return []byte(fmt.Sprintf("&#x%04x;", u))
}

// readHTMLChunk returns the chunk starting at off and the offset after it,
// as the reference's reader does: at most 8,191 bytes without NUL bytes
// (from a NUL on, the non-NUL bytes after it are gathered up to that size),
// then cut back after the last whitespace byte when the input goes on with
// a non-space byte, moving the offset back by the bytes cut.
func readHTMLChunk(data []byte, off int) ([]byte, int) {
	n := min(len(data)-off, htmlChunk)
	window := data[off : off+n]
	var chunk []byte
	if k := bytes.IndexByte(window, 0); k < 0 {
		chunk = append([]byte(nil), window...)
		off += n
	} else {
		chunk = append([]byte(nil), window[:k]...)
		j := off + k
		for j < len(data) && len(chunk) < htmlChunk {
			if data[j] != 0 {
				chunk = append(chunk, data[j])
			}
			j++
		}
		off = j
	}
	if off < len(data) && !isHTMLSpace(data[off]) {
		for i := len(chunk) - 1; i >= 1; i-- {
			if isHTMLSpace(chunk[i]) {
				off -= len(chunk) - (i + 1)
				chunk = chunk[:i+1]
				break
			}
		}
	}
	return chunk, off
}

// normalizeHTML returns the normalized views of an HTML object.
func normalizeHTML(data []byte) htmlViews {
	h := &htmlNormalizer{state: hsNorm, next: hsBad, jsBegin: -1, jsEnd: -1, pendingScrenc: -1}
	h.views.noComment = []byte{}
	h.views.noTags = []byte{}
	for off := 0; off < len(data) && h.state != hsBad; {
		var chunk []byte
		chunk, off = readHTMLChunk(data, off)
		if len(chunk) == 0 {
			break
		}
		if h.inTag == inStyle {
			h.styleBegin = 0
		}
		if h.screncActive {
			h.state, h.next = hsJSDecodeDecrypt, hsBad
		} else if h.lookForScrenc && h.state != hsLookForScrEnc {
			h.savedNext, h.next, h.state = h.next, h.state, hsLookForScrEnc
		}
		h.run(chunk)
		if h.next == h.state {
			h.next = hsBad
		}
	}
	h.finish()
	return h.views
}

func (h *htmlNormalizer) out(c ...byte)  { h.views.noComment = append(h.views.noComment, c...) }
func (h *htmlNormalizer) text(c ...byte) { h.views.noTags = append(h.views.noTags, c...) }

// textSpace writes one space to the text view for a run of separators
// outside scripts and styles.
func (h *htmlNormalizer) textSpace() {
	if h.inTag == inNone && !h.tsw {
		h.text(' ')
		h.tsw = true
	}
}

func (h *htmlNormalizer) addAttr() {
	// The value is a C string: it ends at its first NUL byte. A value that
	// begins with a quote then loses its first and last bytes, whatever the
	// last one is.
	v := h.val
	if k := bytes.IndexByte(v, 0); k >= 0 {
		v = v[:k]
	}
	if len(v) > 0 && v[0] == '"' {
		v = v[1:max(1, len(v)-1)]
	}
	h.attrs = append(h.attrs, htmlAttr{append([]byte(nil), h.arg...), append([]byte(nil), v...)})
}

func (h *htmlNormalizer) attr(name string) ([]byte, bool) {
	for _, a := range h.attrs {
		if string(a.name) == name {
			return a.value, true
		}
	}
	return nil, false
}

// run processes one chunk.
func (h *htmlNormalizer) run(buf []byte) {
	h.chunk = buf
	start := 0
	for pass := 0; pass < 2; pass++ {
		i := start
		for i < len(buf) && isHTMLSpace(buf[i]) {
			i++
		}
		for i < len(buf) && h.state != hsBad {
			c := buf[i]
			if !h.binary {
				if c == '\n' {
					buf[i], c = ' ', ' '
				} else if c == '\r' {
					i++
					continue
				}
			}
			i = h.step(buf, i, c)
			buf = h.chunk // Script Encoder may compact this chunk
		}
		if h.scriptActive {
			// The block goes on in the next chunk.
			if scriptStart := max(h.jsBegin, 0); scriptStart < len(buf) {
				h.segments = append(h.segments, append([]byte(nil), buf[scriptStart:]...))
			}
			h.jsBegin = -1
		}
		if h.inTag == inStyle {
			h.styleSeen = true
			if h.styleBegin < len(buf) {
				h.views.style = append(h.views.style, buf[h.styleBegin:]...)
			}
		}
		if h.pendingScrenc < 0 {
			break
		}
		start, h.pendingScrenc = h.pendingScrenc, -1
		h.state, h.next = hsJSDecodeLength, hsBad
		buf = h.chunk
		if start >= len(buf) {
			break
		}
	}
}

func (h *htmlNormalizer) endScript() {
	h.views.scripts = append(h.views.scripts, h.segments)
	h.segments = nil
	h.scriptActive = false
}

// step processes the byte at i in the current state and returns the next
// position (i itself when the byte is processed again in a new state).
func (h *htmlNormalizer) step(buf []byte, i int, c byte) int {
	switch h.state {
	case hsSkipWS:
		if isHTMLSpace(c) {
			return i + 1
		}
		h.state, h.next = h.next, hsBad
	case hsTrimWS:
		if isHTMLSpace(c) {
			return i + 1
		}
		if h.inTag == inNone {
			h.out(' ')
		}
		h.state, h.next = h.next, hsBad
	case hsNorm:
		switch {
		case c == '<':
			h.out('<')
			h.textSpace()
			h.state, h.next = hsSkipWS, hsTag
			h.tag = h.tag[:0]
			return i + 1
		case isHTMLSpace(c):
			h.textSpace()
			h.state, h.next = hsTrimWS, hsNorm
		case c == '&':
			h.textSpace()
			h.state, h.next = hsCharRef, hsNorm
			return i + 1
		case c >= 0x80:
			h.state, h.next, h.mb = hs8Bit, hsNorm, uint32(c)
			return i + 1
		default:
			x := lowerByte(c)
			if h.inTag != inNone && x == '\'' {
				x = '"'
			}
			h.out(x)
			if h.inTag == inNone {
				if c < 0x20 {
					h.textSpace()
				} else {
					h.text(x)
					h.tsw = false
				}
			}
			return i + 1
		}
	case hsJSDecode:
		if len(buf)-i >= 4 && bytes.Equal(buf[i:i+4], []byte("#@~^")) {
			buf[i], buf[i+1] = '/', '/'
			h.state, h.next = hsJSDecodeLength, hsBad
			return i + 4
		}
		h.out(lowerByte(c))
		return i + 1
	case hsJSDecodeLength:
		if len(buf)-i < 8 {
			h.state, h.next = hsNorm, hsBad
			return i
		}
		var size [4]byte
		if n, err := base64.StdEncoding.Decode(size[:], buf[i:i+8]); err != nil || n != 4 {
			h.views.unsupported = "script-encoder"
			h.state, h.next = hsNorm, hsBad
			return i
		}
		h.screncRemaining = binary.LittleEndian.Uint32(size[:])
		if h.screncRemaining > maxHTMLNormalize {
			h.views.unsupported = "script-encoder-limit"
			h.state, h.next = hsNorm, hsBad
			return i
		}
		h.screncPosition = 0
		h.screncActive = true
		buf[i+7] = '\n'
		h.state, h.next = hsJSDecodeDecrypt, hsBad
		return i + 8
	case hsJSDecodeDecrypt:
		h.chunk = h.decodeScriptChunk(buf, i)
		h.state, h.next = hsNorm, hsBad
		return i
	case hsLookForScrEnc:
		h.lookForScrenc = true
		if pos := bytes.Index(buf[i:], []byte("#@~^")); pos >= 0 {
			pos += i
			buf[pos], buf[pos+1] = '/', '/'
			h.pendingScrenc = pos + 4
		}
		h.state, h.next = h.next, h.savedNext
		return i
	case hs8Bit:
		if c >= 0x80 && h.mb < 0x10000 {
			h.mb = h.mb<<8 | uint32(c)
			return i + 1
		}
		switch {
		case h.mb == 0xE38082 || h.mb == 0xEFBC8E || h.mb == 0xEFB992:
			h.out('.')
			h.text('.')
		case h.mb == 0xA1 && (c == 'C' || c == 'D' || c == 'O'):
			h.out('.')
			h.text('.')
			h.mb = 0
			return i + 1
		default:
			c0, c1, c2 := byte(h.mb>>16), byte(h.mb>>8), byte(h.mb)
			if c0 != 0 {
				h.out(c0)
				h.text(c0)
			}
			if c0 != 0 || c1 != 0 {
				h.out(c1)
				h.text(c1)
			}
			h.out(c2)
			h.text(c1) // the reference writes the middle byte again
		}
		h.mb = 0
		h.state, h.next = h.next, hsNorm
	case hsTag:
		switch {
		case len(h.tag) == 0 && c == '!':
			if h.inTag != inNone {
				h.out('!')
			} else if n := len(h.views.noComment); n > 0 {
				h.views.noComment = h.views.noComment[:n-1]
			}
			h.state = hsComment
			return i + 1
		case c == '>':
			h.out('>')
			h.state, h.next = hsSkipWS, hsProcessTag
			return i + 1
		case isHTMLSpace(c):
			h.state = hsSkipWS
			h.arg = h.arg[:0]
			if h.inTag == inNone {
				h.next = hsTagArg
			} else {
				h.next = hsProcessTag
			}
		default:
			x := lowerByte(c)
			h.out(x)
			if h.inTag != inNone && len(h.tag) == 0 && c != '/' {
				h.state = hsNorm
			}
			h.tag = appendLimited(h.tag, x)
			return i + 1
		}
	case hsTagArg:
		switch {
		case c == '=':
			h.startValue()
			return i + 1
		case isHTMLSpace(c):
			h.state, h.next = hsSkipWS, hsTagArgEqual
			return i + 1
		case c == '>':
			h.out('>')
			if len(h.arg) > 0 {
				h.val = h.val[:0]
				h.addAttr()
			}
			h.state = hsProcessTag
			return i + 1
		default:
			if len(h.arg) == 0 {
				h.out(' ')
			}
			x := lowerByte(c)
			h.out(x)
			h.arg = appendLimited(h.arg, x)
			return i + 1
		}
	case hsTagArgEqual:
		if c == '=' {
			h.startValue()
			return i + 1
		}
		if len(h.arg) > 0 {
			h.val = h.val[:0]
			h.addAttr()
		}
		h.arg = h.arg[:0]
		h.state = hsTagArg
	case hsTagArgVal:
		return h.tagArgVal(buf, i, c)
	case hsComment:
		if h.inTag != inNone && !isHTMLSpace(c) {
			x := lowerByte(c)
			if x == '\'' {
				x = '"'
			}
			h.out(x)
		}
		if c == '>' {
			h.state, h.next = hsSkipWS, hsNorm
		}
		return i + 1
	case hsProcessTag:
		h.processTag(i)
	case hsCharRef:
		if c == '#' {
			h.value, h.hex = 0, false
			h.state = hsCharRefDecode
			return i + 1
		}
		h.state = hsEntityRefDecode
	case hsEntityRefDecode:
		switch {
		case c == ';':
			var o []byte
			if code, ok := htmlEntities[string(h.ename)]; ok {
				o = u16norm(code)
			} else {
				o = append(append([]byte{'&'}, lowerBytes(h.ename)...), ';')
			}
			h.entityOut(o)
			h.ename = h.ename[:0]
			h.state = h.next
			return i + 1
		case (c|0x20 >= 'a' && c|0x20 <= 'z' || c >= '0' && c <= '9' || c == '_' || c == ':' || c == '-') && len(h.ename) < 1024:
			h.ename = append(h.ename, c)
			return i + 1
		default:
			h.entityOut(append([]byte{'&'}, lowerBytes(h.ename)...))
			h.ename = h.ename[:0]
			h.state = h.next
		}
	case hsCharRefDecode:
		return h.charRefDecode(i, c)
	case hsRFC2397Type:
		return h.rfcType(buf, i, c)
	case hsRFC2397Init:
		t := append([]byte(nil), h.val...)
		h.rfc = append([]byte("From html-normalise\nContent-type: "), t...)
		h.rfc = append(h.rfc, '\n')
		if bytes.Contains(t, []byte(";base64")) {
			h.rfc = append(h.rfc, "Content-transfer-encoding: base64\n"...)
		}
		h.rfc = append(h.rfc, '\n')
		contentType, encoding := string(t), ""
		if strings.HasSuffix(contentType, ";base64") {
			contentType = strings.TrimSuffix(contentType, ";base64")
			encoding = "base64"
		}
		// A base64=value parameter stays an ordinary MIME parameter even
		// though the compatibility header above contains a transfer header.
		if strings.HasPrefix(contentType, ";") {
			contentType = "text/plain" + contentType
		}
		h.rfcPart = htmlDataPart{contentType: contentType, encoding: encoding}
		h.rfcBody = len(h.rfc)
		h.rfcOpen = true
		h.binary = true
		h.state = hsRFC2397Data
	case hsRFC2397Data:
		switch {
		case c == '&':
			h.state, h.next = hsCharRef, hsRFC2397Data
		case c == '%':
			h.length, h.value = 0, 0
			h.state, h.next = hsEscapeChar, hsRFC2397Esc
		case h.quoted != 0 && c == h.quoted && !h.escape:
			h.state = hsRFC2397Finish
		case h.quoted == 0 && (isHTMLSpace(c) || c == '>'):
			h.state = hsRFC2397Finish
		default:
			h.rfc = append(h.rfc, c)
		}
		return i + 1
	case hsEscapeChar:
		if h.value > (1<<63-1)/16 {
			h.state = h.next
			return i + 1
		}
		h.value *= 16
		h.length++
		if d := hexDigit(c); d >= 0 {
			h.value += int64(d)
		} else {
			h.state = h.next
		}
		if h.length == 2 {
			h.state = h.next
		}
		return i + 1
	case hsRFC2397Esc:
		switch h.length {
		case 2:
			h.rfc = append(h.rfc, byte(h.value))
		case 1:
			h.rfc = append(h.rfc, '%', byte(h.value)+'0')
		default:
			h.rfc = append(h.rfc, '%')
		}
		h.state = hsRFC2397Data
	case hsRFC2397Finish:
		h.closeRFC()
		h.state, h.next = hsSkipWS, hsTagArg
		h.quoted, h.escape, h.binary = 0, false, false
	}
	return i
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c|0x20 >= 'a' && c|0x20 <= 'f':
		return int(c|0x20-'a') + 10
	}
	return -1
}

func (h *htmlNormalizer) closeRFC() {
	if h.rfcOpen {
		h.rfcPart.compatibility = h.rfc
		h.rfcPart.body = h.rfc[h.rfcBody:]
		h.views.rfc2397 = append(h.views.rfc2397, h.rfcPart)
		h.rfcPart = htmlDataPart{}
		h.rfc, h.rfcOpen = nil, false
	}
}

// entityOut writes decoded entity text to the page view, and to the value
// being collected when the entity is in an attribute value.
func (h *htmlNormalizer) entityOut(o []byte) {
	h.out(o...)
	if h.next == hsTagArgVal {
		h.val = appendLimited(h.val, o...)
	}
}

func (h *htmlNormalizer) startValue() {
	h.out('=')
	h.state, h.next = hsSkipWS, hsTagArgVal
	h.escape, h.quoted = false, 0
	h.val = h.val[:0]
}

func (h *htmlNormalizer) charRefDecode(i int, c byte) int {
	const maxInt64 = 1<<63 - 1
	switch {
	case h.value == 0 && (c == 'x' || c == 'X'):
		h.hex = true
		return i + 1
	case c == ';':
		if h.next == hsTagArgVal {
			h.val = appendLimited(h.val, byte(h.value))
		}
		if h.value < 0x80 {
			h.out(lowerByte(byte(h.value)))
		} else {
			h.out(u16norm(uint16(h.value))...)
		}
		h.state = h.next
		return i + 1
	case c >= '0' && c <= '9' || h.hex && hexDigit(c) >= 0:
		switch {
		case h.hex && h.value < maxInt64/16:
			h.value *= 16
		case h.value < maxInt64/10:
			h.value *= 10
		default:
			h.out(byte(h.value))
			h.state = h.next
			return i + 1
		}
		d := int64(hexDigit(c))
		if h.value > maxInt64-d {
			h.out(byte(h.value))
			h.state = h.next
			return i + 1
		}
		h.value += d
		return i + 1
	}
	h.out(byte(h.value))
	h.state = h.next
	return i
}

func (h *htmlNormalizer) tagArgVal(buf []byte, i int, c byte) int {
	next := i
	skipEscape := false
	switch {
	case string(h.val) == "data:" || string(h.val) == "\"data:":
		if n := len(h.views.noComment); n > 0 {
			h.views.noComment = h.views.noComment[:n-1]
		}
		if h.quoted != 0 {
			h.out('"')
		}
		h.val = h.val[:0]
		h.state, h.next = hsRFC2397Type, hsTagArg
	case c == '&':
		h.state, h.next = hsCharRef, hsTagArgVal
		next = i + 1
	case c == '\'' || c == '"':
		switch {
		case len(h.val) == 0:
			h.quoted = c
			h.out('"')
			h.val = appendLimited(h.val, '"')
		case h.quoted == c && !h.escape:
			h.out('"')
			h.val = appendLimited(h.val, '"')
			h.addAttr()
			h.state, h.next = hsSkipWS, hsTagArg
			h.arg = h.arg[:0]
		default:
			h.out('"')
			h.val = appendLimited(h.val, '"')
		}
		next = i + 1
	case isHTMLSpace(c) || c == '>':
		if h.quoted == 0 {
			h.addAttr()
			h.state, h.next = hsSkipWS, hsTagArg
			h.arg = h.arg[:0]
		} else {
			h.out(c)
			if c == '>' {
				h.val = appendLimited(h.val, '>')
			} else {
				h.val = appendLimited(h.val, ' ')
			}
			h.state, h.next = hsSkipWS, hsTagArgVal
			h.quoted, h.escape = 0, false
			next = i + 1
		}
	default:
		if h.mb2 != 0 && (c < 0x80 || h.mb2 >= 0x10000) {
			switch {
			case h.mb2 == 0xE38082 || h.mb2 == 0xEFBC8E || h.mb2 == 0xEFB992:
				h.out('.')
				h.val = appendLimited(h.val, '.')
			case h.mb2 == 0xA1 && (c == 'C' || c == 'D' || c == 'O'):
				h.out('.')
				h.val = appendLimited(h.val, '.')
				h.mb2 = 0
				return i + 1
			default:
				c0, c1, c2 := byte(h.mb2>>16), byte(h.mb2>>8), byte(h.mb2)
				if c0 != 0 {
					h.out(c0)
					h.val = appendLimited(h.val, c0)
				}
				if c0 != 0 || c1 != 0 {
					h.out(c1)
					h.val = appendLimited(h.val, c1)
				}
				h.out(c2)
				h.val = appendLimited(h.val, c2)
			}
			h.mb2 = 0
		}
		if c >= 0x80 {
			h.mb2 = h.mb2<<8 | uint32(c)
		} else {
			h.out(lowerByte(c))
			h.val = appendLimited(h.val, c)
		}
		next = i + 1
	}
	if !skipEscape {
		h.escape = next < len(buf) && buf[next] == '\\'
	}
	return next
}

func (h *htmlNormalizer) rfcType(buf []byte, i int, c byte) int {
	next := i
	switch {
	case (c == '\'' || c == '"') && h.quoted == c && !h.escape:
		h.state, h.next = hsSkipWS, hsTagArg
		h.arg = h.arg[:0]
		next = i + 1
	case c == '\'' || c == '"':
		h.val = appendLimited(h.val, '"')
		next = i + 1
	case (isHTMLSpace(c) || c == '>') && h.quoted == 0:
		h.state, h.next = hsSkipWS, hsTagArg
		h.arg = h.arg[:0]
	case isHTMLSpace(c) || c == '>':
		if c == '>' {
			h.val = appendLimited(h.val, '>')
		} else {
			h.val = appendLimited(h.val, ' ')
		}
		h.quoted = 0
		h.state, h.next = hsSkipWS, hsRFC2397Type
		next = i + 1
	case c == ',':
		h.state = hsRFC2397Init
		next = i + 1
	default:
		h.val = appendLimited(h.val, lowerByte(c))
		next = i + 1
	}
	h.escape = next < len(buf) && buf[next] == '\\'
	return next
}

// processTag acts on a complete tag at position i (the first byte after
// it), which it does not consume.
func (h *htmlNormalizer) processTag(i int) {
	h.state, h.next = hsSkipWS, hsNorm
	tag := string(h.tag)
	switch {
	case len(tag) > 0 && tag[0] == '/':
		switch {
		case tag == "/script":
			h.inTag = inNone
			if h.scriptActive {
				// The block ends here, with its closing tag text.
				start := max(h.jsBegin, 0)
				if start < i {
					h.segments = append(h.segments, append([]byte(nil), h.chunk[start:i]...))
				}
				h.jsBegin = -1
				h.endScript()
			}
		case tag == "/style" && h.inTag == inStyle:
			if i-8 < h.styleBegin {
				h.state = hsBad
				return
			}
			h.views.style = append(h.views.style, h.chunk[h.styleBegin:i-8]...)
			h.inTag = inNone
		}
	case tag == "script":
		if lang, ok := h.attr("language"); ok {
			l := string(lowerBytes(lang))
			if l == "jscript.encode" {
				h.scriptEncoderTag("javascript")
				h.state, h.next = hsSkipWS, hsJSDecode
			} else if l == "vbscript.encode" {
				h.scriptEncoderTag("vbscript")
				h.state, h.next = hsSkipWS, hsJSDecode
			}
		}
		h.inTag = inScript
		if !h.scriptActive {
			h.scriptActive = true
			h.jsBegin = i
		}
	case tag == "style":
		h.inTag = inStyle
		h.styleBegin = i
	case tag == "%@":
		if lang, ok := h.attr("language"); ok {
			l := string(lowerBytes(lang))
			if l == "jscript.encode" || l == "vbscript.encode" {
				h.savedNext, h.next, h.state = h.next, h.state, hsLookForScrEnc
				h.lookForScrenc = false
			}
		}
	case tag == "a" || tag == "img":
		name := "href"
		if tag == "img" {
			name = "src"
		}
		if v, ok := h.attr(name); ok && len(v) > 0 {
			h.text(v...)
			h.text(' ')
			h.tsw = true
		}
	}
	h.attrs = h.attrs[:0]
}

// scriptEncoderTag is the second tag written to nocomment.html when the
// language attribute chooses Microsoft's script encoder. The first matching
// language attribute is replaced, and every attribute keeps its input order.
func (h *htmlNormalizer) scriptEncoderTag(language string) {
	h.out([]byte("<script")...)
	replaced := false
	for _, a := range h.attrs {
		h.out(' ')
		h.out(a.name...)
		value := a.value
		if !replaced && string(a.name) == "language" {
			value = []byte(language)
			replaced = true
		}
		if value != nil {
			h.out('=', '"')
			for _, c := range value {
				h.out(lowerByte(c))
			}
			h.out('"')
		}
	}
	h.out('>')
}

func (h *htmlNormalizer) finish() {
	if h.screncActive {
		h.views.unsupported = "script-encoder-truncated"
	}
	if len(h.ename) > 0 {
		if code, ok := htmlEntities[string(h.ename)]; ok {
			h.out(u16norm(code)...)
		} else {
			h.out('&')
			h.out(lowerBytes(h.ename)...)
		}
	}
	if h.scriptActive {
		h.endScript()
	}
}

// htmlUTF16ToASCII converts HTML typed as UTF-16 the way the reference does
// before normalizing it: each byte pair (low, high) becomes the single byte
// (low + 16*high) & 0xFF, and a trailing odd byte is dropped.
func htmlUTF16ToASCII(data []byte) []byte {
	out := make([]byte, len(data)/2)
	for i := range out {
		out[i] = data[2*i] + 16*data[2*i+1]
	}
	return out
}
