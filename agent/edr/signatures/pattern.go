package signatures

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const maxPatternTokens = 1024

// maxPatternStates bounds the failed states one start position remembers. It
// is a variable only so tests can reach the bound with small inputs: patterns
// read it when compiled, so it must not change while a database compiled under
// another value is still matching.
var maxPatternStates = 200000

// errPatternStates is one pattern's memory bound. Unlike the shared work budget
// it leaves every other rule sound, so Match keeps evaluating them.
var errPatternStates = fmt.Errorf("%w: pattern state limit", ErrMatchLimit)

// byteSet holds the bytes that can begin a match from a pattern node.
type byteSet [4]uint64

func (b *byteSet) add(c byte)      { b[c>>6] |= 1 << (c & 63) }
func (b *byteSet) has(c byte) bool { return b[c>>6]&(1<<(c&63)) != 0 }
func (b *byteSet) union(o *byteSet) {
	b[0], b[1], b[2], b[3] = b[0]|o[0], b[1]|o[1], b[2]|o[2], b[3]|o[3]
}

type token struct {
	literal      []byte
	value, mask  byte
	gap          bool
	min, max     int
	alternatives [][]token
	negative     bool
	// boundary is a zero-width word-boundary check (the fullword modifier).
	boundary byte
}

// Word boundaries of the fullword modifier, in plain and wide (UTF-16) form.
const (
	boundaryWordStart byte = 1 + iota
	boundaryWordEnd
	boundaryWideStart
	boundaryWideEnd
)

// boundaryHolds checks a word boundary at pos the reference's way: a plain
// boundary needs the neighbouring byte to be absent or not alphanumeric
// (ASCII); a wide one needs the neighbouring character not to be an
// alphanumeric byte followed by a zero byte, and a start within the first two
// bytes always holds.
func boundaryHolds(kind byte, data []byte, pos int) bool {
	switch kind {
	case boundaryWordStart:
		return pos == 0 || !alphaNumeric(data[pos-1])
	case boundaryWordEnd:
		return pos == len(data) || !alphaNumeric(data[pos])
	case boundaryWideStart:
		return pos <= 1 || !(alphaNumeric(data[pos-2]) && data[pos-1] == 0)
	case boundaryWideEnd:
		switch left := len(data) - pos; {
		case left == 0:
			return true
		case left == 1:
			return !alphaNumeric(data[pos])
		default:
			return !(alphaNumeric(data[pos]) && data[pos+1] == 0)
		}
	}
	return true
}

// widenBody is the reference's text rewrite for the wide modifier: every
// byte token gains a zero byte after it; gaps are copied unchanged, square
// ranges become curly ones, and grouping characters are copied.
func widenBody(body string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(body); {
		switch c := body[i]; c {
		case '*', '|', ')', '(', '!':
			b.WriteByte(c)
			i++
		case '{', '[':
			end := strings.IndexByte(body[i:], map[byte]byte{'{': '}', '[': ']'}[c])
			if end < 0 {
				return "", errors.New("unterminated wildcard range")
			}
			b.WriteByte('{')
			b.WriteString(body[i+1 : i+end])
			b.WriteByte('}')
			i += end + 1
		default:
			if i+1 >= len(body) {
				return "", errors.New("odd hexadecimal body length")
			}
			b.WriteString(body[i : i+2])
			b.WriteString("00")
			i += 2
		}
	}
	return b.String(), nil
}

type patternNode struct {
	token        token
	next         int
	alternatives []int
	// follow, on a gap node, holds the bytes that can begin its continuation;
	// nil when any byte (or none) can.
	follow *byteSet
}

// Only patterns with wildcards, alternatives, or boundary checks need a
// matching graph. Keeping its slice header and entry point off bodyPattern
// avoids paying for them on the much more common literal-only signature.
type patternGraph struct {
	nodes              []patternNode
	native             *nativeACPattern
	nativeError        string
	anchors            []patternAnchor // compilation-only; released after index publication
	start              int
	minWidth, maxWidth int
	proneFrom          int
}
type bodyPattern struct {
	offset offset
	graph  *patternGraph
	prefix []byte
	// The common literal-only pattern derives its one compile-time anchor from
	// prefix. Complex patterns keep anchors on graph until the index owns them.
	// anchored survives publication to preserve candidate filtering.
	fullword bool
	nocase   bool
	anchored bool
	// A single literal needs no two-node matching graph. The official feed
	// contains over a million of these; prefix retains the literal bytes.
	simple bool
	// nativeAC preserves the standalone matcher's routing even when graph
	// simplification turns its original syntax into one literal. Floating
	// offsets also route PE standalone rules to AC; other roots decide this
	// at rule evaluation because patterns are shared across target roots.
	nativeAC, nativeFloat bool
	// nativeFixed records the unmodified fixed-width hex/nibble syntax. It
	// distinguishes it from graphs simplified from gaps or alternatives.
	nativeFixed bool
	// holder is set on a placeholder subsignature, which has no matcher.
	holder *placeholder
}

func (p *bodyPattern) minWidth() int {
	if p.simple {
		return len(p.prefix)
	}
	if p.graph != nil {
		return p.graph.minWidth
	}
	return 0
}

func (p *bodyPattern) maxWidth() int {
	if p.simple {
		return len(p.prefix)
	}
	if p.graph != nil {
		return p.graph.maxWidth
	}
	return 0
}

// proneFrom is the smallest object length where a start without an anchor
// could still reach the state limit. A literal-only pattern has no such
// length; complex patterns retain their threshold in the graph.
func (p *bodyPattern) proneFrom() int {
	if p.graph != nil {
		return p.graph.proneFrom
	}
	return math.MaxInt
}

// patternAnchor is a literal a match contains, starting minPre to maxPre
// bytes after the match start (maxPre < 0: unbounded). A folded literal comes
// from a case-insensitive pattern: it is ASCII-lowercased and compared with
// lowercased data.
type patternAnchor struct {
	literal        []byte
	fold           bool
	minPre, maxPre int
}

func (p *bodyPattern) compiledAnchorCount() int {
	if p.graph != nil && p.graph.native != nil {
		return len(p.graph.native.anchors)
	}
	if p.simple {
		return 1
	}
	if p.graph != nil {
		return len(p.graph.anchors)
	}
	return 0
}

func (p *bodyPattern) compiledAnchor(i int) patternAnchor {
	if p.graph != nil && p.graph.native != nil {
		return p.graph.native.anchors[i]
	}
	if p.simple {
		// A single literal begins at the match start. The index, rather than
		// every compiled pattern, owns its optional folded copy.
		a := patternAnchor{literal: p.prefix}
		if p.nocase && bytes.ContainsFunc(a.literal, func(r rune) bool { return r < 0x80 && asciiLetter(byte(r)) }) {
			a.literal, a.fold = asciiLower(a.literal), true
		}
		return a
	}
	return p.graph.anchors[i]
}

func (p *bodyPattern) releaseCompiledAnchors() {
	if p.graph != nil {
		p.graph.anchors = nil
		if p.graph.native != nil {
			p.graph.native.anchors = nil
		}
	}
}

func compileBody(off, body string) (bodyPattern, error) {
	p := bodyPattern{}
	var err error
	p.offset, err = parseOffset(off)
	if err != nil {
		return p, err
	}
	mods := ""
	if i := strings.Index(body, "::"); i >= 0 {
		mods = body[i+2:] // an empty list means no modifiers
		body = body[:i]
	}
	wide, ascii := false, false
	for _, m := range mods {
		switch m {
		case 'i':
			p.nocase = true
		case 'f':
			p.fullword = true
		case 'w':
			wide = true
		case 'a':
			ascii = true
		default:
			return p, fmt.Errorf("unsupported body modifier %q", m)
		}
	}
	if len(body) == 0 || len(body) > 32768 {
		return p, errors.New("empty or oversized body pattern")
	}
	if mods != "" && strings.Contains(body, "!(") {
		return p, errors.New("negated alternatives take no body modifiers")
	}
	if p.fullword {
		// The reference brackets the body with word markers and reads square
		// ranges as curly ones, before widening.
		body = strings.NewReplacer("[", "{", "]", "}").Replace(body)
	}
	// A plain hex body needs neither token storage nor anchor selection. The
	// resulting simple pattern is identical to the one the general parser
	// produces: its literal is also its only anchor and prefix. Preserve the
	// general path when tests disable the simple-pattern optimization.
	if !p.fullword && !wide && !ascii && maxPatternStates > 0 && len(body)%2 == 0 && isHex(body) {
		literal := make([]byte, len(body)/2)
		_, _ = hex.Decode(literal, []byte(body)) // isHex established every digit.
		p.simple, p.anchored, p.prefix = true, true, literal
		return p, nil
	}
	form := func(wideForm bool) ([]token, error) {
		text := body
		if wideForm {
			var err error
			if text, err = widenBody(body); err != nil {
				return nil, err
			}
		}
		ts, err := parseTokens(text, 0)
		if err != nil || !p.fullword {
			return ts, err
		}
		start, end := boundaryWordStart, boundaryWordEnd
		if wideForm {
			start, end = boundaryWideStart, boundaryWideEnd
		}
		return append(append([]token{{boundary: start}}, ts...), token{boundary: end}), nil
	}
	tokens, err := form(wide)
	if err != nil {
		return p, err
	}
	if wide && ascii {
		// The plain form is a second pattern of the same subsignature: one
		// count of the starts where either matches.
		plain, err := form(false)
		if err != nil {
			return p, err
		}
		tokens = []token{{alternatives: [][]token{plain, tokens}}}
	}
	if len(tokens) > maxPatternTokens {
		return p, errors.New("body pattern token limit exceeded")
	}
	minWidth, maxWidth := tokenWidth(tokens)
	if minWidth <= 0 {
		return p, errors.New("body pattern must consume at least one byte")
	}
	var pre preAnchorStates
	anchors, pre := longestAnchors(tokens, 0, 0)
	if len(anchors) == 0 && !hasConstrainedByte(tokens) {
		return p, errors.New("unconstrained wildcard-only patterns are unsupported")
	}
	var bound, open int64
	stateWindows(tokens, 0, 0, &bound, &open)
	proneBoundary := proneFrom(bound, open)
	if proneBoundary == math.MaxInt && len(anchors) > 0 {
		// No start can be state-limited, so any literal every match contains
		// bounds the starts exactly: take the one expected to occur least.
		anchors = chooseAnchors(tokens, 0, 0, p.nocase)
	} else {
		// Otherwise the longest literal stays, because its span decides which
		// starts are tried and so where a limit can be observed; only starts
		// that cannot reach it may be skipped.
		proneBoundary = proneFrom(pre.bound, pre.open)
	}
	singleLiteral := len(tokens) == 1 && len(tokens[0].literal) > 0 && maxPatternStates > 0
	if !singleLiteral {
		for i, a := range anchors {
			if p.nocase && bytes.ContainsFunc(a.literal, func(r rune) bool { return r < 0x80 && asciiLetter(byte(r)) }) {
				anchors[i].literal, anchors[i].fold = asciiLower(a.literal), true
			}
		}
	}
	p.anchored = len(anchors) != 0
	first := tokens
	for len(first) > 0 && first[0].boundary != 0 {
		first = first[1:] // a boundary consumes nothing
	}
	if len(first) > 0 && len(first[0].literal) > 0 && !p.nocase {
		p.prefix = first[0].literal
	}
	if singleLiteral {
		p.simple = true
		p.prefix = tokens[0].literal
		return p, nil
	}
	p.graph = &patternGraph{nodes: make([]patternNode, 1, 1+sequenceNodeCount(tokens)), anchors: anchors, minWidth: minWidth, maxWidth: maxWidth, proneFrom: proneBoundary}
	p.graph.nodes[0].next = -1 // node zero accepts
	p.graph.start = p.compileSequence(tokens, 0)
	return p, nil
}

// Count the nodes compileSequence emits, including positive alternatives.
// Negative alternatives are evaluated directly from their tokens.
func sequenceNodeCount(ts []token) int {
	n := len(ts)
	for _, t := range ts {
		if !t.negative {
			for _, a := range t.alternatives {
				n += sequenceNodeCount(a)
			}
		}
	}
	return n
}
func parseTokens(body string, depth int) ([]token, error) {
	if depth > 8 {
		return nil, errors.New("body alternatives nesting limit exceeded")
	}
	var ts []token
	appendLiteral := func(b byte, at int) {
		if len(ts) > 0 && len(ts[len(ts)-1].literal) > 0 {
			ts[len(ts)-1].literal = append(ts[len(ts)-1].literal, b)
		} else {
			// Each exact byte is two hex digits. Reserve only this contiguous
			// literal run: whole-file feeds have many long literals, and growing
			// every one from a one-byte allocation dominates parse churn.
			end := at + 2
			for end+1 < len(body) && hexNibble(body[end]) >= 0 && hexNibble(body[end+1]) >= 0 {
				end += 2
			}
			literal := make([]byte, 1, (end-at)/2)
			literal[0] = b
			ts = append(ts, token{literal: literal})
		}
	}
	for i := 0; i < len(body); {
		if len(ts) > maxPatternTokens {
			return nil, errors.New("body token limit exceeded")
		}
		switch body[i] {
		case '*':
			ts = append(ts, token{gap: true, min: 0, max: -1})
			i++
		case '{', '[':
			endChar := byte('}')
			if body[i] == '[' {
				endChar = ']'
			}
			end := strings.IndexByte(body[i+1:], endChar)
			if end < 0 {
				return nil, errors.New("unterminated wildcard range")
			}
			end += i + 1
			lo, hi, e := parseGap(body[i+1 : end])
			if e != nil {
				return nil, e
			}
			if body[i] == '[' && (hi < 0 || hi > 32) {
				return nil, errors.New("square wildcard range exceeds 32 bytes")
			}
			ts = append(ts, token{gap: true, min: lo, max: hi})
			i = end + 1
		case '(', '!':
			negative := body[i] == '!'
			if negative {
				i++
				if i >= len(body) || body[i] != '(' {
					return nil, errors.New("invalid negated alternative")
				}
			}
			j := i + 1
			nest := 1
			for j < len(body) && nest > 0 {
				if body[j] == '(' {
					nest++
				}
				if body[j] == ')' {
					nest--
				}
				if nest > 0 {
					j++
				}
			}
			if nest != 0 {
				return nil, errors.New("unterminated body alternative")
			}
			inside := body[i+1 : j]
			alts := splitAlternatives(inside)
			tok := token{negative: negative}
			fixed := -1
			for _, alt := range alts {
				seq, e := parseTokens(alt, depth+1)
				if e != nil {
					return nil, e
				}
				lo, hi := tokenWidth(seq)
				if lo == 0 && negative {
					return nil, errors.New("empty negated body alternative")
				}
				if negative {
					if lo != hi {
						return nil, errors.New("negated alternatives must have fixed width")
					}
					for _, x := range seq {
						if len(x.literal) == 0 {
							return nil, errors.New("negated alternatives require literal bytes")
						}
					}
					if fixed >= 0 && fixed != lo {
						return nil, errors.New("negated alternatives have different widths")
					}
					fixed = lo
				}
				tok.alternatives = append(tok.alternatives, seq)
			}
			tok.min = fixed
			ts = append(ts, tok)
			i = j + 1
		default:
			if i+1 >= len(body) {
				return nil, errors.New("odd hexadecimal body length")
			}
			a, b := body[i], body[i+1]
			value, mask := byte(0), byte(0)
			for k, c := range []byte{a, b} {
				if c == '?' {
					continue
				}
				n := hexNibble(c)
				if n < 0 {
					return nil, fmt.Errorf("unsupported body syntax near %q", body[i:])
				}
				shift := uint(4 * (1 - k))
				value |= byte(n) << shift
				mask |= 15 << shift
			}
			if mask == 255 {
				appendLiteral(value, i)
			} else {
				ts = append(ts, token{value: value, mask: mask})
			}
			i += 2
		}
	}
	return ts, nil
}
func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}
func splitAlternatives(s string) []string {
	depth, start := 0, 0
	var result []string
	for i, c := range s {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
		}
		if c == '|' && depth == 0 {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	return append(result, s[start:])
}
func parseGap(s string) (int, int, error) {
	fields := strings.Split(s, "-")
	if len(fields) > 2 || s == "" || s == "-" {
		return 0, 0, errors.New("invalid wildcard range")
	}
	parse := func(s string) (int, error) {
		v, e := strconv.Atoi(s)
		if e != nil || v < 0 || v > 16777216 {
			return 0, errors.New("wildcard range is invalid or too large")
		}
		return v, nil
	}
	if len(fields) == 1 {
		v, e := parse(s)
		return v, v, e
	}
	lo, hi := 0, -1
	var e error
	if fields[0] != "" {
		lo, e = parse(fields[0])
		if e != nil {
			return 0, 0, e
		}
	}
	if fields[1] != "" {
		hi, e = parse(fields[1])
		if e != nil {
			return 0, 0, e
		}
		if hi < lo {
			return 0, 0, errors.New("inverted wildcard range")
		}
	}
	return lo, hi, nil
}
func tokenWidth(ts []token) (int, int) {
	lo, hi := 0, 0
	for _, t := range ts {
		a, b := 1, 1
		switch {
		case t.boundary != 0:
			a, b = 0, 0
		case len(t.literal) > 0:
			a, b = len(t.literal), len(t.literal)
		case t.gap:
			a, b = t.min, t.max
		case len(t.alternatives) > 0:
			a, b = 1<<30, 0
			for _, alt := range t.alternatives {
				x, y := tokenWidth(alt)
				a = min(a, x)
				if y < 0 || b < 0 {
					b = -1
				} else {
					b = max(b, y)
				}
			}
		}
		lo += a
		if hi < 0 || b < 0 {
			hi = -1
		} else {
			hi += b
		}
	}
	return lo, hi
}

// anchorsFor chooses the literals the candidate index uses for a pattern a
// state limit can reach (others use chooseAnchors): the longest top-level
// literal, or, when that is shorter, one literal from every branch of a
// top-level alternative (for example the plain and wide forms of an
// ASCII-or-wide literal), since every match passes through one branch. It
// returns nil when no literal is certain to occur.
func anchorsFor(ts []token) []patternAnchor {
	set, _ := longestAnchors(ts, 0, 0)
	return set
}

// preAnchorStates bounds, like stateWindows, the failed states of the nodes a
// start can expand when no anchor of a set occurs where that start allows it:
// every path to a match passes one of the anchors, so only the nodes before
// them (and an alternative whose branches hold them) can be reached.
type preAnchorStates struct{ bound, open int64 }

// longestAnchors is anchorsFor with the distances of the first token from the
// match start (lo, hi) and the pre-anchor states of the chosen set.
func longestAnchors(ts []token, lo, hi int) ([]patternAnchor, preAnchorStates) {
	var best []patternAnchor
	var bestPre, before preAnchorStates
	bestLen := 0
	for _, t := range ts {
		if len(t.literal) > bestLen {
			best, bestLen, bestPre = []patternAnchor{{literal: t.literal, minPre: lo, maxPre: hi}}, len(t.literal), before
		}
		if len(t.alternatives) > 0 && !t.negative {
			var set []patternAnchor
			shortest := 1 << 30
			pre := before
			addWindow(lo, hi, &pre.bound, &pre.open) // the alternative node itself
			for _, alternative := range t.alternatives {
				branch, branchPre := longestAnchors(alternative, lo, hi)
				if len(branch) == 0 {
					set = nil
					break
				}
				for _, a := range branch {
					shortest = min(shortest, len(a.literal))
					set = append(set, a)
				}
				pre.bound, pre.open = min(pre.bound+branchPre.bound, math.MaxInt64/4), pre.open+branchPre.open
			}
			if len(set) > 0 && shortest > bestLen {
				best, bestLen, bestPre = set, shortest, pre
			}
		}
		stateWindows([]token{t}, lo, hi, &before.bound, &before.open)
		a, b := tokenWidth([]token{t})
		lo += a
		if hi < 0 || b < 0 {
			hi = -1
		} else {
			hi += b
		}
	}
	return best, bestPre
}

// asciiLower returns a copy with A-Z folded to a-z, the only folding the
// case-insensitive modifier performs.
func asciiLower(b []byte) []byte {
	out := make([]byte, len(b))
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return out
}
func hasConstrainedByte(ts []token) bool {
	for _, t := range ts {
		if len(t.literal) > 0 || t.mask != 0 {
			return true
		}
		for _, a := range t.alternatives {
			if hasConstrainedByte(a) {
				return true
			}
		}
	}
	return false
}
func (p *bodyPattern) compileSequence(ts []token, next int) int {
	for i := len(ts) - 1; i >= 0; i-- {
		t := ts[i]
		node := patternNode{token: t, next: next}
		if len(t.alternatives) > 0 && !t.negative {
			for _, a := range t.alternatives {
				node.alternatives = append(node.alternatives, p.compileSequence(a, next))
			}
		}
		if t.gap {
			node.follow = p.firstBytes(next)
		}
		p.graph.nodes = append(p.graph.nodes, node)
		next = len(p.graph.nodes) - 1
	}
	return next
}

// firstBytes returns the bytes that can begin a match at node id, or nil when
// the node accepts without consuming, is a gap or a negated alternative, or can
// begin with any byte. A non-nil set always requires at least one byte.
func (p *bodyPattern) firstBytes(id int) *byteSet {
	if id == 0 {
		return nil
	}
	node := p.graph.nodes[id]
	t := node.token
	set := &byteSet{}
	switch {
	case t.boundary != 0:
		return p.firstBytes(node.next)
	case len(node.alternatives) > 0:
		for _, alt := range node.alternatives {
			first := p.firstBytes(alt)
			if first == nil {
				return nil
			}
			set.union(first)
		}
	case len(t.literal) > 0:
		c := t.literal[0]
		set.add(c)
		if p.nocase && asciiLetter(c) {
			set.add(c ^ 0x20)
		}
	case t.gap, t.negative:
		return nil
	default:
		if t.mask == 0 {
			return nil
		}
		for c := 0; c < 256; c++ {
			if byte(c)&t.mask == t.value {
				set.add(byte(c))
			}
		}
	}
	return set
}
func asciiLetter(b byte) bool  { return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' }
func alphaNumeric(b byte) bool { return asciiLetter(b) || b >= '0' && b <= '9' }
func equalBytes(a, b []byte, nocase bool) bool {
	if !nocase {
		return bytes.Equal(a, b)
	}
	for i, c := range a {
		x := b[i]
		if c == x {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if c != x {
			return false
		}
	}
	return true
}

type patternState struct{ node, pos int }

// literalHit records one search for the literal after a gap: the first
// occurrence at or after from starts at at (-1: none before the object end).
type literalHit struct{ from, at int }

// literalSearches holds, for one pattern on one object, the hits of the
// searches for each literal that follows a gap, by the literal's node, in
// order of from. Where a literal next occurs does not depend on the start
// being tried, so later starts reuse them instead of searching the object
// again; only the work charged changes. Keeping more than the latest hit
// matters: a start searches for a literal before its occurrence and, once
// the rest of the pattern fails, again after it, and each start would
// otherwise search most of the object again.
//
// Two hits with different answers never overlap (a search from inside [from,
// at] finds at), so hits with the same answer merge and a lookup is a binary
// search. A literal keeps at most maxLiteralHits of them, and one object's
// counts keep at most maxSearchHits beyond the first of each literal, so the
// memory stays bounded however many patterns are counted; past either bound
// only a hit after all others replaces the last one, as the latest hit always
// did.
type literalSearches map[int][]literalHit

const maxLiteralHits = 64

// maxSearchHits bounds the extra hits of one object (about 16 MiB). It is a
// variable only so tests can reach it with small inputs.
var maxSearchHits = 1 << 20

func (s *literalSearches) lookup(node, pos int) (int, bool) {
	if s == nil {
		return 0, false
	}
	hits := (*s)[node]
	i := sort.Search(len(hits), func(i int) bool { return hits[i].from > pos }) - 1
	if i < 0 || hits[i].at >= 0 && pos > hits[i].at {
		return 0, false
	}
	return hits[i].at, true
}

// record keeps one search; w counts the object's extra hits.
func (s *literalSearches) record(node, from, at int, w *matchWork) {
	if s == nil {
		return
	}
	if *s == nil {
		*s = literalSearches{}
	}
	hits := (*s)[node]
	i := sort.Search(len(hits), func(i int) bool { return hits[i].from > from })
	switch {
	case i > 0 && hits[i-1].at == at:
		return // already known from an earlier position
	case i < len(hits) && hits[i].at == at:
		hits[i].from = from
		return
	case len(hits) == 0 || len(hits) < maxLiteralHits && w.searchHits < maxSearchHits:
		if len(hits) > 0 {
			w.searchHits++
		}
		hits = slices.Insert(hits, i, literalHit{from, at})
	case i == len(hits):
		hits[i-1] = literalHit{from, at}
	}
	(*s)[node] = hits
}

func (p *bodyPattern) matchesAt(data []byte, start int, w *matchWork) (bool, error) {
	return p.matchesAtWith(data, start, w, nil)
}

// matchesAtWith is matchesAt sharing literal searches with the other starts of
// the same count (searches may be nil).
func (p *bodyPattern) matchesAtWith(data []byte, start int, w *matchWork, searches *literalSearches) (bool, error) {
	if p.simple {
		// Match the graph's budget charges exactly: one visit of the literal
		// node, its byte comparison when in bounds, then the accepting node.
		if err := w.step(1); err != nil {
			return false, err
		}
		if start < 0 || start > len(data) {
			return false, nil
		}
		end := start + len(p.prefix)
		if end > len(data) {
			return false, nil
		}
		if err := w.step(int64(len(p.prefix))); err != nil {
			return false, err
		}
		if !equalBytes(data[start:end], p.prefix, p.nocase) {
			return false, nil
		}
		if err := w.step(1); err != nil {
			return false, err
		}
		return true, nil
	}
	// Most starts fail before remembering anything, so the memo is allocated
	// on first use. gapFailed holds, for each unbounded gap that failed in
	// this start, the lowest position at which it did.
	var failed map[patternState]bool
	var gapFailed map[int]int
	var visit func(int, int) (bool, error)
	visit = func(id, pos int) (bool, error) {
		if err := w.step(1); err != nil {
			return false, err
		}
		if pos < 0 || pos > len(data) {
			return false, nil
		}
		if id == 0 {
			return true, nil
		}
		state := patternState{id, pos}
		if failed[state] {
			return false, nil
		}
		if len(failed) >= maxPatternStates {
			return false, errPatternStates
		}
		node := p.graph.nodes[id]
		t := node.token
		// Only a state that went on to visit a continuation is remembered as
		// failed. A node rejected by its own bytes has nothing below it to
		// re-explore, so the bounded memo holds only states whose failure did.
		var matched, expanded bool
		var err error
		switch {
		case t.boundary != 0:
			if boundaryHolds(t.boundary, data, pos) {
				expanded = true
				matched, err = visit(node.next, pos)
			}
		case len(node.alternatives) > 0:
			expanded = true
			for _, alt := range node.alternatives {
				matched, err = visit(alt, pos)
				if matched || err != nil {
					return matched, err
				}
			}
		case len(t.literal) > 0:
			end := pos + len(t.literal)
			if end <= len(data) {
				if err := w.step(int64(len(t.literal))); err != nil {
					return false, err
				}
			}
			if end <= len(data) && equalBytes(data[pos:end], t.literal, p.nocase) {
				expanded = true
				matched, err = visit(node.next, end)
			}
		case t.gap:
			expanded = true
			if from, ok := gapFailed[id]; ok && t.max < 0 && pos >= from {
				// From pos an unbounded gap reaches a subset of the
				// continuations it reached from an earlier position of this
				// start. Each of them failed there, and was remembered or was
				// rejected by its own bytes, so exploring them again would
				// fail without remembering anything new (nor reaching the
				// state bound, which only a new state can): only this state
				// is remembered, as before.
				break
			}
			end := len(data)
			if t.max >= 0 {
				end = min(end, pos+t.max)
			}
			for next := pos + t.min; next <= end; next++ {
				// A following literal lets us jump directly to its next occurrence.
				nextNode := p.graph.nodes[node.next]
				if !p.nocase && len(nextNode.token.literal) > 0 {
					needle := nextNode.token.literal
					bound := min(len(data), end+len(needle))
					if next > bound {
						break
					}
					at, known := searches.lookup(node.next, next)
					if known {
						if err := w.step(1); err != nil {
							return false, err
						}
					} else {
						window := data[next:bound]
						relative := bytes.Index(window, needle)
						searched := len(window) + len(needle)
						if relative >= 0 {
							searched = relative + len(needle)
						}
						if err := w.step(int64(max(1, searched))); err != nil {
							return false, err
						}
						// A match inside the window is the first at or after
						// next; no match up to the object end means none.
						at = -1
						if relative >= 0 {
							at = next + relative
						}
						if relative >= 0 || bound == len(data) {
							searches.record(node.next, next, at, w)
						}
					}
					if at < 0 || at+len(needle) > bound {
						break
					}
					next = at
				} else if node.follow != nil {
					// Other constrained continuations skip the positions whose
					// byte cannot begin them, charging the scan in bulk.
					from := next
					for next <= end && (next == len(data) || !node.follow.has(data[next])) {
						next++
					}
					if err := w.step(int64(max(1, next-from))); err != nil {
						return false, err
					}
					if next > end {
						break
					}
				}
				matched, err = visit(node.next, next)
				if matched || err != nil {
					return matched, err
				}
			}
			if t.max < 0 {
				if gapFailed == nil {
					gapFailed = make(map[int]int)
				}
				if from, ok := gapFailed[id]; !ok || pos < from {
					gapFailed[id] = pos
				}
			}
		case t.negative:
			if pos+t.min <= len(data) {
				allowed := true
				for _, seq := range t.alternatives {
					var literal []byte
					for _, x := range seq {
						literal = append(literal, x.literal...)
					}
					if err := w.step(int64(len(literal))); err != nil {
						return false, err
					}
					if equalBytes(data[pos:pos+t.min], literal, p.nocase) {
						allowed = false
						break
					}
				}
				if allowed {
					expanded = true
					matched, err = visit(node.next, pos+t.min)
				}
			}
		default:
			if pos < len(data) && data[pos]&t.mask == t.value {
				expanded = true
				matched, err = visit(node.next, pos+1)
			}
		}
		if !matched && err == nil && expanded {
			if failed == nil {
				failed = make(map[patternState]bool)
			}
			failed[state] = true
		}
		return matched, err
	}
	matched, err := visit(p.graph.start, start)
	if err == errPatternStates && w.forward {
		// The memory bound: decide the start by the forward search instead.
		return p.forwardMatchesAt(data, start, w)
	}
	return matched, err
}
func (p *bodyPattern) count(data []byte, meta *peMeta, w *matchWork, maxCount int) (int, error) {
	count, _, err := p.countWithOffset(data, meta, w, maxCount)
	return count, err
}

// countAt counts the matches that start exactly at the given sorted offsets
// and returns the first one.
func (p *bodyPattern) countAt(data []byte, starts []int, w *matchWork) (int, uint32, error) {
	n, first := 0, missingOffset
	for _, start := range starts {
		ok, err := p.matchesAt(data, start, w)
		if err != nil {
			return 0, missingOffset, err
		}
		if ok {
			if n == 0 {
				first = uint32(start)
			}
			n++
		}
	}
	return n, first, nil
}

// missingOffset is the bytecode match-offset sentinel for absent matches.
const missingOffset = uint32(0xfffffffe)

// patternCount is one object's progress counting one shared pattern. Rules
// sharing the pattern need different caps, so counting resumes where it
// stopped instead of starting again; a smaller cap reads the same prefix.
type patternCount struct {
	n        int    // matches found so far
	first    uint32 // earliest match start, or missingOffset
	last     int    // the latest match start, when n > 0
	lo       int    // the first possible start position
	next, hi int    // next start position to try and the last possible one
	done     bool   // every start position was tried
	limited  bool   // a start position exceeded the per-pattern state bound
	// windowed counting tries only the starts in the windows its anchor
	// occurrences allow (see startCount); winHi ends the current window.
	windowed bool
	winHi    int
	streams  []windowStream
	// searches lets the starts share where gap literals next occur.
	searches literalSearches
}

// newCount starts counting at the positions the offset allows within span,
// the start positions the anchor occurrences allow.
func (p *bodyPattern) newCount(data []byte, meta *peMeta, span startSpan) *patternCount {
	pc := &patternCount{first: missingOffset}
	lo, hi, ok := p.offset.bounds(len(data), meta)
	lo, hi = max(lo, span.lo), min(hi, len(data)-p.minWidth(), span.hi)
	if !ok || hi < lo {
		pc.done = true
	} else {
		pc.lo, pc.next, pc.hi = lo, lo, hi
	}
	return pc
}

// countMore continues counting in whichever way the count was started.
func (p *bodyPattern) countMore(data []byte, w *matchWork, pc *patternCount, want int) error {
	if pc.windowed {
		return p.countWindows(data, w, pc, want)
	}
	return p.countUntil(data, w, pc, want)
}

// countUntil continues counting matching start positions, including
// overlapping matches, until want matches are known or no start remains. A
// leading wildcard belongs to that start; an internal literal anchor is not
// the offset. errPatternStates marks the count limited; other errors abort.
func (p *bodyPattern) countUntil(data []byte, w *matchWork, pc *patternCount, want int) error {
	for ; pc.n < want && !pc.done && !pc.limited; pc.next++ {
		if pc.next > pc.hi {
			pc.done = true
			break
		}
		if err := w.step(1); err != nil {
			return err
		}
		if len(p.prefix) > 0 && !p.nocase {
			bound := min(len(data), pc.hi+len(p.prefix))
			window := data[pc.next:bound]
			relative := bytes.Index(window, p.prefix)
			searched := len(window) + len(p.prefix)
			if relative >= 0 {
				searched = relative + len(p.prefix)
			}
			if err := w.step(int64(max(1, searched))); err != nil {
				return err
			}
			if relative < 0 {
				pc.done = true
				break
			}
			pc.next += relative
		}
		matched, err := p.matchesAtWith(data, pc.next, w, &pc.searches)
		if errors.Is(err, errPatternStates) {
			pc.limited = true
			break
		}
		if err != nil {
			return err
		}
		if matched {
			if pc.n == 0 {
				pc.first = uint32(pc.next)
			}
			pc.n++
			pc.last = pc.next
		}
	}
	return nil
}

// value is the count a rule with this cap sees: the capped count when that
// many matches are known, otherwise the full count, or countStateLimited.
func (pc *patternCount) value(want int) (int, uint32) {
	switch {
	case pc.n >= want:
		return want, pc.first
	case pc.done:
		return pc.n, pc.first
	}
	return countStateLimited, missingOffset
}

// countWithOffset counts matching start positions and records their earliest
// start in the same traversal. Positive maxCount values cap the count; zero
// and negative values do not. Absence and errors use the bytecode
// match-offset sentinel. Errors discard partial counts and offsets so they
// cannot become callback results.
func (p *bodyPattern) countWithOffset(data []byte, meta *peMeta, w *matchWork, maxCount int) (int, uint32, error) {
	if maxCount <= 0 {
		maxCount = len(data) + 1
	}
	pc := p.newCount(data, meta, anySpan)
	if err := p.countUntil(data, w, pc, maxCount); err != nil {
		return 0, missingOffset, err
	}
	if pc.limited {
		return 0, missingOffset, errPatternStates
	}
	n, first := pc.value(maxCount)
	return n, first, nil
}

// longestPart is the byte length of the pattern's longest part as the
// reference's matcher splits it: parts end at '*' and at every gap except a
// fixed one of 1 to 127 bytes, which stays inside the part as wildcards; an
// alternative counts its longest branch.
func (p *bodyPattern) longestPart() int {
	if p.simple {
		return len(p.prefix)
	}
	longest, current := 0, 0
	for id := p.graph.start; id != 0; id = p.graph.nodes[id].next {
		t := p.graph.nodes[id].token
		switch {
		case t.boundary != 0:
		case t.gap && t.min == t.max && t.min > 0 && t.min < 128:
			current += t.min
		case t.gap:
			longest, current = max(longest, current), 0
		case len(t.literal) > 0:
			current += len(t.literal)
		case t.negative:
			current += t.min
		case len(t.alternatives) > 0:
			widest := 0
			for _, alt := range t.alternatives {
				_, hi := tokenWidth(alt)
				widest = max(widest, hi)
			}
			current += widest
		default:
			current++
		}
	}
	return max(longest, current)
}
