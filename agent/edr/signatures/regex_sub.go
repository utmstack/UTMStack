package signatures

import (
	"bytes"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// A regex subsignature is [offset:]trigger/pattern/flags. The reference runs
// it once per object after content matching, when its trigger holds on the
// rule's earlier subsignature counts, on the whole object: from the offset's
// window start to the object's end (or to the window's end with the e flag).

// rxMaxObject is the largest object the reference runs regexes on; on larger
// ones every regex count is 0. It is a variable only for tests.
var rxMaxObject = 100 << 20

// regexSub is a compiled regex subsignature.
type regexSub struct {
	// source and flags are the pattern text and its flags, for diagnostics.
	source, flags string
	pattern       *rxPattern
	prefilter     *regexp.Regexp
	// first holds the bytes a match can start with (firstList lists them
	// when there are few); anyStart is set when a match can start anywhere,
	// including the subject end.
	first     byteSet
	firstList []byte
	anyStart  bool
	// required holds up to two literals every match contains.
	required []rxLiteral
	// A pattern starting with .* can only match at the call's start offset
	// (dotStar, with the s flag) or there and after a newline (lineStart):
	// the library then tries no other start.
	dotStar, lineStart bool
	offset             regexOffset
	global, rolling    bool
	encompass          bool
}

// regexOffset is the offset field: '*' (any), 'a' (absolute) or 'e' (from
// the end); shift is the window's extent after its start.
type regexOffset struct {
	kind         byte
	value, shift uint32
}

// The reference's special 32-bit offset values.
const (
	rxOffsetAny  = 0xFFFFFFFF
	rxOffsetNone = 0xFFFFFFFE
)

func parseRegexOffset(s string) (regexOffset, error) {
	o := regexOffset{kind: '*'}
	if s == "*" {
		return o, nil
	}
	num := func(t string) (uint32, error) {
		if t == "" || strings.Trim(t, "0123456789") != "" {
			return 0, errors.New("invalid regex offset")
		}
		v, err := strconv.ParseUint(t, 10, 32)
		return uint32(v), err
	}
	base, shift, hasShift := strings.Cut(s, ",")
	var err error
	if hasShift {
		if o.shift, err = num(shift); err != nil {
			return o, err
		}
	}
	switch {
	case strings.HasPrefix(base, "EOF-"):
		o.kind = 'e'
		o.value, err = num(base[4:])
	case base != "" && base[0] >= '0' && base[0] <= '9':
		o.kind = 'a'
		o.value, err = num(base)
	default:
		// Executable, version-information and macro offsets: none in the
		// feed; the subsignature stays unknown when reached.
		return o, rxUnsupported("regex offset " + base)
	}
	return o, err
}

// compileRegexSub compiles the offset, pattern and flags of a regex
// subsignature. dconf is the PCRE dynamic configuration: without its global
// bit the g flag is ignored.
func compileRegexSub(offset, pattern, flags string, dconf uint32) (*regexSub, error) {
	o, err := parseRegexOffset(offset)
	if err != nil {
		return nil, err
	}
	p, err := parseRegex(pattern, flags)
	if err != nil {
		return nil, err
	}
	r := &regexSub{source: pattern, flags: flags, pattern: p, offset: o, prefilter: compilePrefilter(p)}
	r.global = strings.Contains(flags, "g") && dconf&dconfPCREGlobal != 0
	r.rolling = strings.Contains(flags, "r")
	r.encompass = strings.Contains(flags, "e")
	var empty, looks bool
	r.first, empty, looks = firstBytes(p.root)
	// A lookaround before the first byte can do any amount of work at a
	// start that byte rules out, so every start is tried, as the library
	// may. Elsewhere a ruled-out start costs the library a few calls.
	r.anyStart = empty || looks
	for c := 0; c < 256; c++ {
		if r.first.has(byte(c)) && len(r.firstList) <= 3 {
			r.firstList = append(r.firstList, byte(c))
		}
	}
	if r.required = rxRequiredLiterals(p.root); len(r.required) > 2 {
		r.required = r.required[:2]
	}
	lead := p.root
	if lead.kind == rxConcat {
		lead = lead.subs[0]
	}
	if lead.kind == rxRepeat && lead.min == 0 && lead.max < 0 && lead.subs[0].kind == rxSet && lead.subs[0].dot {
		r.dotStar = lead.subs[0].set.full()
		r.lineStart = !r.dotStar
	}
	return r, nil
}

// firstBytes returns the bytes a match of n can start with, whether n can
// match the empty string, and whether a lookaround can run before n's first
// byte. Assertions are transparent: the bytes after them start the match.
func firstBytes(n *rxNode) (s byteSet, empty, looks bool) {
	switch n.kind {
	case rxEmpty, rxAssert:
		return s, true, false
	case rxLook:
		return s, true, true
	case rxSet:
		return n.set, false, false
	case rxString:
		return n.seq[0], false, false
	case rxConcat:
		for _, sub := range n.subs {
			f, e, l := firstBytes(sub)
			s.union(&f)
			looks = looks || l
			if !e {
				return s, false, looks
			}
		}
		return s, true, looks
	case rxAlt:
		for _, sub := range n.subs {
			f, e, l := firstBytes(sub)
			s.union(&f)
			empty, looks = empty || e, looks || l
		}
		return s, empty, looks
	case rxRepeat:
		f, e, l := firstBytes(n.subs[0])
		return f, e || n.min == 0, l
	case rxGroup, rxAtomic:
		return firstBytes(n.subs[0])
	}
	s.invert() // backreference: any text, possibly empty
	return s, true, false
}

// window computes the subject start, the maximum shift and whether the
// subsignature runs at all on an object of length size.
func (r *regexSub) window(size int) (start uint32, shift uint32, run bool) {
	switch r.offset.kind {
	case '*':
		return rxOffsetAny, 0, true
	case 'e':
		start = uint32(size) - r.offset.value
	default:
		start = r.offset.value
	}
	if start == rxOffsetNone {
		return 0, 0, false
	}
	return start, r.offset.shift, true
}

// run returns the count and first match offset on data, counting up to
// maxCount. errRegexLimit means the reference's library could have given up
// on one of the calls, so the count is unknown.
func (r *regexSub) run(data []byte, maxCount int, w *matchWork) (int, uint32, error) {
	if len(data) > rxMaxObject {
		return 0, missingOffset, nil
	}
	start32, shift, ok := r.window(len(data))
	if !ok {
		return 0, missingOffset, nil
	}
	anchored := r.pattern.anchored || !r.rolling && shift == 0 && start32 != rxOffsetAny
	if start32 == rxOffsetAny {
		start32 = 0
	}
	if int64(start32) >= int64(len(data)) {
		return 0, missingOffset, nil
	}
	start, end := int(start32), len(data)
	if r.encompass && shift > 0 {
		end = start + int(min(int64(shift), int64(len(data)-start)))
	}
	// A match starting more than shift bytes into the window is dropped
	// and ends the loop; searching past that point cannot add a count.
	last := end
	if shift > 0 {
		last = int(min(int64(end), int64(start)+int64(shift)))
	}
	// Text every match contains must be in the subject. Many regexes of
	// one object share such text, so each search is done once per object.
	for i := range r.required {
		l := &r.required[i]
		key := rxLiteralKey{string(l.text), l.fold, start, end}
		found, known := w.rxLiterals[key]
		if !known {
			// A fast byte search: charged per 64 bytes.
			if err := w.step(int64(end-start)/64 + 1); err != nil {
				return 0, missingOffset, err
			}
			found = l.contains(data[start:end])
			if w.rxLiterals == nil {
				w.rxLiterals = map[rxLiteralKey]bool{}
			}
			w.rxLiterals[key] = found
		}
		if !found {
			return 0, missingOffset, nil
		}
	}
	count, first := 0, uint32(missingOffset)
	m := newRxMatcher(r.pattern, data, start, end, w)
	next := start
	for {
		ms, me, found, err := r.find(m, next, last, anchored)
		if err == errRegexNoMatch {
			break
		}
		if err != nil {
			return count, first, err
		}
		if !found {
			break
		}
		if count == 0 {
			first = uint32(ms)
		}
		count++
		if !r.global || count >= maxCount {
			break
		}
		next = me
		if next >= end {
			break
		}
	}
	return count, first, nil
}

// find is one call of the library: the first match starting at or after
// next (exactly at next when anchored) and at or before last. The candidate
// starts are those whose byte can begin a match; each is tried with the
// backtracking matcher. When the matcher reaches its limits the prefilter
// decides: without a superset match from next there is no match whatever
// the library's limits; otherwise the call is unknown.
func (r *regexSub) find(m *rxMatcher, next, last int, anchored bool) (int, int, bool, error) {
	for pos := next; pos <= last; pos++ {
		switch {
		case anchored || r.dotStar:
			if pos != next {
				return 0, 0, false, nil
			}
		case r.lineStart:
			if pos != next && m.data[pos-1] != '\n' {
				continue
			}
		case r.anyStart:
		default:
			at, err := r.nextStart(m, pos, last)
			if err != nil || at < 0 {
				return 0, 0, false, err
			}
			pos = at
		}
		e, ok, err := m.matchAt(pos)
		if err == errRegexLimit {
			return 0, 0, false, r.limitReached(m, next, last, anchored)
		}
		if err != nil {
			return 0, 0, false, err
		}
		if ok {
			return pos, e, true, nil
		}
	}
	return 0, 0, false, nil
}

// nextStart returns the first position from pos to last whose byte can
// begin a match, or -1, charging the bytes it passes.
func (r *regexSub) nextStart(m *rxMatcher, pos, last int) (int, error) {
	end := min(last+1, m.end)
	if pos >= end {
		return -1, m.w.step(1)
	}
	window := m.data[pos:end]
	i, per := -1, 64 // a fast byte search is charged per 64 bytes
	if len(r.firstList) == 1 {
		i = bytes.IndexByte(window, r.firstList[0])
	} else {
		// bytes.IndexAny reads its set as text, so bytes above 0x7f would
		// not be found as themselves; a plain scan is used instead.
		per = 8
		for k, c := range window {
			if r.first.has(c) {
				i = k
				break
			}
		}
	}
	if i < 0 {
		return -1, m.w.step(int64(len(window)/per + 1))
	}
	return pos + i, m.w.step(int64(i/per + 1))
}

// limitReached decides a call on which the matcher reached its limits.
func (r *regexSub) limitReached(m *rxMatcher, next, last int, anchored bool) error {
	if r.prefilter == nil {
		return errRegexLimit
	}
	rd := &rxReader{data: m.data, pos: next, end: m.end}
	loc := r.prefilter.FindReaderIndex(rd)
	if err := m.w.step(int64(max(rd.read, 1))); err != nil {
		return err
	}
	if loc == nil || anchored && loc[0] != 0 || next+loc[0] > last {
		return errRegexNoMatch
	}
	return errRegexLimit
}

// rxLiteralKey identifies a required-text search on one object's window.
type rxLiteralKey struct {
	text       string
	fold       bool
	start, end int
}

// errRegexNoMatch: the prefilter proved there is no match.
var errRegexNoMatch = errors.New("regex cannot match")

// rxLiteral is text every match contains: bytes, compared ignoring ASCII
// case when fold is set.
type rxLiteral struct {
	text []byte
	fold bool
}

// rxRequiredLiterals returns the literal runs every match of n contains,
// longest first: runs of single bytes or letter case pairs in strings that
// every match passes through.
func rxRequiredLiterals(n *rxNode) []rxLiteral {
	var out []rxLiteral
	var walk func(*rxNode)
	walk = func(n *rxNode) {
		switch n.kind {
		case rxString:
			var cur rxLiteral
			flush := func() {
				if len(cur.text) >= 2 {
					out = append(out, cur)
				}
				cur = rxLiteral{}
			}
			for i := range n.seq {
				b, fold, ok := literalByte(&n.seq[i])
				if !ok {
					flush()
					continue
				}
				cur.text = append(cur.text, b)
				cur.fold = cur.fold || fold
			}
			flush()
		case rxConcat:
			for _, sub := range n.subs {
				walk(sub)
			}
		case rxGroup, rxAtomic:
			walk(n.subs[0])
		case rxRepeat:
			if n.min > 0 {
				walk(n.subs[0])
			}
		}
	}
	walk(n)
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].text) > len(out[j].text) })
	return out
}

// literalByte recognizes a set of one byte or of one ASCII letter in both
// cases.
func literalByte(s *byteSet) (byte, bool, bool) {
	var members []byte
	for c := 0; c < 256 && len(members) <= 2; c++ {
		if s.has(byte(c)) {
			members = append(members, byte(c))
		}
	}
	switch {
	case len(members) == 1:
		return members[0], false, true
	case len(members) == 2 && members[1] == members[0]+32 && members[0] >= 'A' && members[0] <= 'Z':
		return members[1], true, true
	}
	return 0, false, false
}

// contains reports whether data holds the literal.
func (l *rxLiteral) contains(data []byte) bool {
	if !l.fold {
		return bytes.Contains(data, l.text)
	}
	lo, up := l.text[0], l.text[0]
	if lo >= 'a' && lo <= 'z' {
		up = lo - 32
	}
	for i := 0; i+len(l.text) <= len(data); i++ {
		if c := data[i]; (c == lo || c == up) && equalBytes(data[i:i+len(l.text)], l.text, true) {
			return true
		}
	}
	return false
}
