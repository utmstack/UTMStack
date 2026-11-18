package signatures

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// The prefilter proves "no match" in linear time: a Go regexp that matches
// wherever the pattern can (a superset). Go matches text, so each byte is
// read as one character: bytes below 0x80 as themselves and the others as
// private-use characters U+E080-U+E0FF, which never case-fold and are not
// word, digit or space characters. Constructs Go lacks are relaxed to
// something weaker: lookaround and word boundaries to nothing,
// backreferences to any text, counts above Go's limits to open repeats.
// Only the backtracking matcher decides a match.

// rxRune is the character a byte is read as.
func rxRune(b byte) rune {
	if b < 0x80 {
		return rune(b)
	}
	return 0xE000 + rune(b)
}

// rxReader reads data[pos:end] one byte per character.
type rxReader struct {
	data     []byte
	pos, end int
	read     int
}

func (r *rxReader) ReadRune() (rune, int, error) {
	if r.pos >= r.end {
		return 0, 0, io.EOF
	}
	b := r.data[r.pos]
	r.pos++
	r.read++
	return rxRune(b), 1, nil
}

// compilePrefilter returns the superset regexp, or nil when Go cannot
// express even a relaxed form.
func compilePrefilter(p *rxPattern) *regexp.Regexp {
	for _, relaxed := range []bool{false, true} {
		var b strings.Builder
		emitPrefilter(&b, p.root, relaxed)
		if re, err := regexp.Compile(b.String()); err == nil {
			return re
		}
	}
	return nil
}

func emitSet(b *strings.Builder, s *byteSet) {
	if s.full() {
		b.WriteString(`(?s:.)`)
		return
	}
	if s.empty() {
		b.WriteString(`[^\x{0}-\x{10FFFF}]`)
		return
	}
	b.WriteByte('[')
	// U+E000 is never read (bytes map below 0x80 or to U+E080-U+E0FF). In
	// every class of two or more bytes it keeps Go from turning a case pair
	// such as [Aa] into a case-folded letter: Go 1.25's alternation
	// factoring then loses the fold ("A1*|[Aa]" becomes "A(?:1*|)").
	b.WriteString(`\x{E000}`)
	for lo := 0; lo < 256; {
		if !s.has(byte(lo)) {
			lo++
			continue
		}
		hi := lo
		// A range stays on one side of 0x80, where the mapping jumps.
		for hi+1 < 256 && s.has(byte(hi+1)) && (hi+1 < 0x80) == (lo < 0x80) {
			hi++
		}
		fmt.Fprintf(b, `\x{%X}`, rxRune(byte(lo)))
		if hi > lo {
			fmt.Fprintf(b, `-\x{%X}`, rxRune(byte(hi)))
		}
		lo = hi + 1
	}
	b.WriteByte(']')
}

// hasRepeat reports a counted repeat inside n, which Go multiplies into the
// outer count.
func hasRepeat(n *rxNode) bool {
	if n.kind == rxRepeat && (n.min > 1 || n.max > 1) {
		return true
	}
	for _, s := range n.subs {
		if hasRepeat(s) {
			return true
		}
	}
	return false
}

func emitPrefilter(b *strings.Builder, n *rxNode, relaxed bool) {
	switch n.kind {
	case rxEmpty:
	case rxSet:
		emitSet(b, &n.set)
	case rxString:
		for i := range n.seq {
			emitSet(b, &n.seq[i])
		}
	case rxConcat:
		for _, s := range n.subs {
			emitPrefilter(b, s, relaxed)
		}
	case rxAlt:
		b.WriteString("(?:")
		for i, s := range n.subs {
			if i > 0 {
				b.WriteByte('|')
			}
			emitPrefilter(b, s, relaxed)
		}
		b.WriteByte(')')
	case rxGroup, rxAtomic:
		b.WriteString("(?:")
		emitPrefilter(b, n.subs[0], relaxed)
		b.WriteByte(')')
	case rxRepeat:
		b.WriteString("(?:")
		emitPrefilter(b, n.subs[0], relaxed)
		b.WriteByte(')')
		lo, hi := n.min, n.max
		counted := lo <= 1000 && hi <= 1000 && !(relaxed && hasRepeat(n.subs[0]))
		switch {
		case lo == 0 && hi == 1:
			b.WriteByte('?')
		case lo == 0 && (hi < 0 || !counted):
			b.WriteByte('*')
		case lo == 1 && hi < 0 || !counted:
			b.WriteByte('+')
		case hi < 0:
			fmt.Fprintf(b, "{%d,}", lo)
		case lo == hi:
			fmt.Fprintf(b, "{%d}", lo)
		default:
			fmt.Fprintf(b, "{%d,%d}", lo, hi)
		}
	case rxLook:
		// Nothing: an assertion only removes matches.
	case rxBackref:
		b.WriteString(`(?s:.*)`)
	case rxAssert:
		switch n.assert {
		case asSubjectStart:
			// The reader starts at the call's start offset.
			b.WriteString(`\A`)
		case asLineStart:
			b.WriteString(`(?m:^)`)
		case asSubjectEnd:
			b.WriteString(`\z`)
		case asEndOrFinalNL, asLineEnd:
			b.WriteString(`(?m:$)`)
		}
	}
}
