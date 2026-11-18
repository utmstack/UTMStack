package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"
	"strings"
)

// typingWindow is how much of an object typing reads.
const typingWindow = 1024

// DetermineType types an object from its first KiB the way the reference
// does: the magic list, then text classification, a tar or Office check,
// then type patterns for text and binary results, then two UTF-16 passes that
// only look for HTML. complete is always true: every encoding the last pass
// meets is converted as the reference converts it.
func (d *Database) DetermineType(data []byte) (FileType, bool) {
	buf := data[:min(len(data), typingWindow)]
	t, ok := d.compareMagic(buf)
	if !ok {
		t = textType(buf)
	}
	switch {
	case t == TypeBinaryData:
		if tar, ok := isTar(buf); ok {
			return tar, true
		}
	case t == TypeZIP && len(buf) > 2*(30+5):
		t = ooxmlType(data)
	}
	if t < TypeTextASCII || t > TypeBinaryData {
		return t, true
	}
	if found := d.typePatternPass(buf, t); found >= TypeTextASCII && !droppedSFX(found) {
		return found, true
	}
	if d.typePatternPass(utf16ToASCII(buf), TypeTextASCII) == TypeHTML {
		return TypeHTMLUTF16, true
	}
	if d.dconf.Phishing&dconfPhishingEntconv != 0 {
		converted, ok := entconvToASCII(buf)
		if !ok {
			return t, false
		}
		if converted != nil && d.typePatternPass(converted, TypeAny) == TypeHTML {
			return TypeHTML, true
		}
	}
	return t, true
}

// droppedSFX are self-extractor types that typing never returns; a raw scan
// finds and extracts them instead.
func droppedSFX(t FileType) bool {
	switch t {
	case TypeZIPSFX, TypeARJSFX, TypeRARSFX, TypeEGGSFX, TypeCABSFX, Type7ZSFX:
		return true
	}
	return false
}

// typeMatch is one occurrence of a type pattern.
type typeMatch struct {
	start, index int
}

// typePatternPass runs the type patterns over buf, a typing window, as the
// running type ftype. Only absolute and any-position offsets apply here; the
// result rule is typeResult.
func (d *Database) typePatternPass(buf []byte, ftype FileType) FileType {
	w := &matchWork{ctx: context.Background(), limit: typingWorkLimit}
	var matches []typeMatch
	for i := range d.ftypes.patterns {
		fp := &d.ftypes.patterns[i]
		if fp.rtype != TypeAny && fp.rtype != ftype {
			continue
		}
		p := &d.patterns[fp.id]
		lo, hi := 0, len(buf)
		switch p.offset.kind {
		case offsetAbsolute:
			lo, hi = int(p.offset.value), int(int64(p.offset.value)+int64(p.offset.shift))
		case offsetAny:
		default:
			continue // relative offsets need information typing does not compute
		}
		hi = min(hi, len(buf)-p.minWidth())
		for s := lo; s <= hi; s++ {
			if prefix := p.prefix; len(prefix) > 0 {
				next := bytes.Index(buf[s:min(len(buf), hi+len(prefix))], prefix)
				if next < 0 {
					break
				}
				s += next
			}
			matched, err := p.matchesAt(buf, s, w)
			if err != nil {
				break // a pattern past its work share stops; others still run
			}
			if matched {
				matches = append(matches, typeMatch{s, i})
			}
		}
	}
	sort.Slice(matches, func(a, b int) bool {
		if matches[a].start != matches[b].start {
			return matches[a].start < matches[b].start
		}
		return matches[a].index < matches[b].index
	})
	return foldTypeResults(0, matches, d.ftypes.patterns)
}

// typingWorkLimit bounds the type patterns' work on one typing window.
const typingWorkLimit = 50_000_000

// typeResult applies one match to the running result: IGNORED ends the pass
// (ok false); otherwise the match replaces a lower result, and a
// self-extractor type (from SFX on) or MSEXE replaces any result.
func typeResult(result, typ FileType) (FileType, bool) {
	if typ == TypeIgnored {
		return TypeIgnored, false
	}
	if typ > result || typ >= TypeSFX || typ == TypeMSEXE {
		return typ, true
	}
	return result, true
}

// foldTypeResults applies the matches in buffer order.
func foldTypeResults(result FileType, matches []typeMatch, patterns []ftmPattern) FileType {
	for _, m := range matches {
		var ok bool
		if result, ok = typeResult(result, patterns[m.index].typ); !ok {
			return result
		}
	}
	return result
}

// utf16ToASCII is the reference's quick UTF-16 decode for typing: each byte
// pair (low byte at the even offset) becomes ((high << 4) + low) & 0xFF, which
// is right for UTF-16LE ASCII text and garbage otherwise.
func utf16ToASCII(buf []byte) []byte {
	n := len(buf) &^ 1
	out := make([]byte, n/2)
	for i := 0; i < n; i += 2 {
		out[i/2] = byte(uint(buf[i+1])<<4 + uint(buf[i]))
	}
	return out
}

type textEncoding uint8

const (
	encodingNone textEncoding = iota
	encodingUTF16BE
	encodingUTF16LE
	encodingUCS4BE
	encodingUCS4LE
	// UCS-4 in the 2143 or 3412 byte order: the reference converts it as
	// big-endian UCS-4 (checked on generated files: "00 00 FF FE" then
	// big-endian markup is typed HTML, a "FE FF 00 00" mark stops the
	// conversion at once).
	encodingUCS4Unusual
)

// detectMultiByteEncoding finds a UTF-16 or UCS-4 byte-order mark, or the
// leading '<' of markup written in one of them.
func detectMultiByteEncoding(b []byte) textEncoding {
	has := func(p string) bool { return strings.HasPrefix(string(b), p) }
	switch {
	case has("\x00\x00\xfe\xff"), has("\x00\x00\x00<"):
		return encodingUCS4BE
	case has("\xff\xfe\x00\x00"), has("<\x00\x00\x00"):
		return encodingUCS4LE
	case has("\x00\x00\xff\xfe"), has("\xfe\xff\x00\x00"), has("\x00\x00<\x00"), has("\x00<\x00\x00"):
		return encodingUCS4Unusual
	case has("\xfe\xff"), has("\x00<\x00?"):
		return encodingUTF16BE
	case has("\xff\xfe"), has("<\x00?\x00"):
		return encodingUTF16LE
	}
	return encodingNone
}

// entconvToASCII converts a UTF-16 or UCS-4 typing window to UTF-16BE code
// units and turns each unit (high, low) into ((high << 4) + low) & 0xFF,
// dropping zero bytes, as the reference's second HTML typing pass does.
// Conversion stops at the first invalid sequence or incomplete trailing
// unit. It returns nil when no such encoding is detected; the bool is always
// true (every detected encoding is converted).
func entconvToASCII(buf []byte) ([]byte, bool) {
	var units []uint16
	switch detectMultiByteEncoding(buf) {
	case encodingNone:
		return nil, true
	case encodingUTF16BE, encodingUTF16LE:
		order := binary.ByteOrder(binary.BigEndian)
		if detectMultiByteEncoding(buf) == encodingUTF16LE {
			order = binary.LittleEndian
		}
		for i := 0; i+1 < len(buf); i += 2 {
			u := order.Uint16(buf[i:])
			if u >= 0xd800 && u <= 0xdbff {
				if i+3 >= len(buf) {
					break
				}
				if v := order.Uint16(buf[i+2:]); v < 0xdc00 || v > 0xdfff {
					break
				} else {
					units = append(units, u, v)
					i += 2
					continue
				}
			}
			if u >= 0xdc00 && u <= 0xdfff {
				break
			}
			units = append(units, u)
		}
	default:
		order := binary.ByteOrder(binary.BigEndian)
		if detectMultiByteEncoding(buf) == encodingUCS4LE {
			order = binary.LittleEndian
		}
		for i := 0; i+3 < len(buf); i += 4 {
			c := order.Uint32(buf[i:])
			if c > 0x10ffff || c >= 0xd800 && c <= 0xdfff {
				break
			}
			if c > 0xffff {
				c -= 0x10000
				units = append(units, uint16(0xd800+c>>10), uint16(0xdc00+c&0x3ff))
				continue
			}
			units = append(units, uint16(c))
		}
	}
	out := make([]byte, 0, len(units))
	for _, u := range units {
		if c := byte(uint(u>>8)<<4 + uint(u&0xff)); c != 0 {
			out = append(out, c)
		}
	}
	return out, true
}

// Office Open XML and similar zip formats are typed by member names: the
// local headers in the first part of the archive, then, when their names
// look like such a document, the central directory.
var (
	ooxmlTyped = []struct {
		name string
		typ  FileType
	}{{"xl/", TypeOOXMLXL}, {"ppt/", TypeOOXMLPPT}, {"word/", TypeOOXMLWord}, {"Contents/content.hpf", TypeOOXMLHWP}}
	ooxmlLikely = []string{"BinData", "mimetype", "Contents", "docProps/", "customXml/", "version.xml",
		"settings.xml", "_rels/.rels", "[ContentTypes].xml", "[Content_Types].xml", "Preview/PrvText.txt",
		"META-INF/container.xml"}
)

const (
	ooxmlNameMax   = len("META-INF/container.xml")
	ooxmlScanLimit = 32 * typingWindow
)

// ooxmlType checks up to three local headers: a typed name returns its type;
// after the third, when any name was a likely one, the central directory
// decides. Otherwise the object stays a zip.
func ooxmlType(data []byte) FileType {
	region := data[:min(len(data), ooxmlScanLimit)]
	readable, likely := 0, false
	for pos := 0; pos < len(region); {
		i := bytes.Index(region[pos:], []byte("PK\x03\x04"))
		if i < 0 {
			break
		}
		name := pos + i + 30
		if len(region)-name <= ooxmlNameMax {
			break
		}
		for _, t := range ooxmlTyped {
			if bytes.HasPrefix(region[name:], []byte(t.name)) {
				return t.typ
			}
		}
		for _, n := range ooxmlLikely {
			likely = likely || bytes.HasPrefix(region[name:], []byte(n))
		}
		if readable++; readable > 2 {
			if likely {
				if t := centralOOXMLType(data); t != TypeAny {
					return t
				}
			}
			break
		}
		pos = name
	}
	return TypeZIP
}

// centralOOXMLType returns the type of the first central-directory entry
// whose name starts with a typed prefix (Contents/content.hpf must be the
// whole name), or ANY.
func centralOOXMLType(data []byte) FileType {
	eocd := bytes.LastIndex(data[max(0, len(data)-(22+65535)):], []byte("PK\x05\x06"))
	if eocd < 0 {
		return TypeAny
	}
	eocd += max(0, len(data)-(22+65535))
	if eocd+22 > len(data) {
		return TypeAny
	}
	le := binary.LittleEndian
	entries := int(le.Uint16(data[eocd+10:]))
	pos := int(le.Uint32(data[eocd+16:]))
	for n := 0; n < entries && pos >= 0 && pos+46 <= len(data); n++ {
		if string(data[pos:pos+4]) != "PK\x01\x02" {
			break
		}
		nameLen := int(le.Uint16(data[pos+28:]))
		extra := int(le.Uint16(data[pos+30:]))
		comment := int(le.Uint16(data[pos+32:]))
		if pos+46+nameLen > len(data) {
			break
		}
		name := data[pos+46 : pos+46+nameLen]
		for _, t := range ooxmlTyped {
			if t.typ == TypeOOXMLHWP && string(name) == t.name || t.typ != TypeOOXMLHWP && bytes.HasPrefix(name, []byte(t.name)) {
				return t.typ
			}
		}
		pos += 46 + nameLen + extra + comment
	}
	return TypeAny
}
