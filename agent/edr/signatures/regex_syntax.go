package signatures

import (
	"errors"
	"strconv"
	"strings"
)

// Regex subsignatures are matched by the reference with an 8-bit regular
// expression library: no Unicode, the C-locale character tables, a line-feed
// newline, and only the options the subsignature's flags set. This file
// parses that syntax into rxNode trees; regex_match.go matches them.
//
// Constructs the feed does not use and whose meaning depends on the
// library's version or build (recursion, conditionals, control verbs, \R,
// \X, \p, quoting, the x option, variable-length lookbehind, {,m}) are
// reported as unsupported: such a subsignature stays unknown when reached.

// errRegexUnsupported marks a valid pattern this engine does not evaluate.
type errRegexUnsupported struct{ what string }

func (e *errRegexUnsupported) Error() string { return "unsupported regex construct: " + e.what }

func rxUnsupported(what string) error { return &errRegexUnsupported{what} }

func (s *byteSet) addRange(lo, hi int) {
	for c := lo; c <= hi; c++ {
		s.add(byte(c))
	}
}
func (s *byteSet) invert() {
	for i := range s {
		s[i] = ^s[i]
	}
}
func (s *byteSet) full() bool { return s[0]&s[1]&s[2]&s[3] == ^uint64(0) }
func (s *byteSet) empty() bool {
	return s[0]|s[1]|s[2]|s[3] == 0
}

// fold adds the other case of every ASCII letter in the set: with the
// C-locale tables only A-Z and a-z have another case.
func (s *byteSet) fold() {
	for c := 'A'; c <= 'Z'; c++ {
		if s.has(byte(c)) || s.has(byte(c+32)) {
			s.add(byte(c))
			s.add(byte(c + 32))
		}
	}
}

// The C-locale classes.
func isRxDigit(b byte) bool { return b >= '0' && b <= '9' }
func isRxSpace(b byte) bool { return b == ' ' || b >= 9 && b <= 13 }
func isRxWord(b byte) bool {
	return b == '_' || isRxDigit(b) || b|0x20 >= 'a' && b|0x20 <= 'z'
}

func classSet(pred func(byte) bool, negate bool) byteSet {
	var s byteSet
	for c := 0; c < 256; c++ {
		if pred(byte(c)) != negate {
			s.add(byte(c))
		}
	}
	return s
}

type rxKind uint8

const (
	rxEmpty   rxKind = iota
	rxSet            // one byte of set
	rxString         // a fixed sequence of byte sets
	rxConcat         // subs in order
	rxAlt            // subs in priority order
	rxRepeat         // subs[0] from min to max times
	rxGroup          // subs[0], captured as group cap when cap > 0
	rxAtomic         // subs[0] matched once, without backtracking into it
	rxLook           // assertion on subs[0] (behind, negate)
	rxBackref        // the text of group cap
	rxAssert         // an anchor or word boundary
)

type rxAssertKind uint8

const (
	asSubjectStart    rxAssertKind = iota // \A; ^ without m
	asLineStart                           // ^ with m
	asSubjectEnd                          // \z; $ with E and without m
	asEndOrFinalNL                        // \Z; $ without m or E
	asLineEnd                             // $ with m
	asWordBoundary                        // \b
	asNotWordBoundary                     // \B
)

type rxNode struct {
	kind rxKind
	set  byteSet // rxSet
	// dot marks the set of '.', which the library treats apart from an
	// equal class (see regexSub.dotStar).
	dot  bool
	seq  []byteSet // rxString
	subs []*rxNode
	// rxRepeat: max < 0 is unbounded.
	min, max   int
	greedy     bool
	possessive bool
	// rxGroup: the capture index (0 for none); rxBackref: the group read.
	cap      int
	caseless bool // rxBackref
	behind   bool // rxLook
	negate   bool // rxLook
	// length is a lookbehind alternative's fixed length.
	lengths []int
	assert  rxAssertKind
}

// rxFlags are the options in effect at one point of the pattern.
type rxFlags struct {
	caseless, multiline, dotall, ungreedy, dollarEndOnly bool
}

// rxPattern is a parsed pattern.
type rxPattern struct {
	root     *rxNode
	groups   int
	anchored bool // the A flag
}

type rxParser struct {
	s      string
	i      int
	groups int
	names  map[string]int
	// namedRefs are backreference nodes waiting for their group's number.
	namedRefs map[*rxNode]string
	depth     int
}

// parseRegex parses pattern as compiled with the flags of a subsignature
// (the compile options among "ismxAEU"; engine flags are ignored here).
func parseRegex(pattern, flags string) (*rxPattern, error) {
	var f rxFlags
	anchored := false
	for i := 0; i < len(flags); i++ {
		switch flags[i] {
		case 'i':
			f.caseless = true
		case 's':
			f.dotall = true
		case 'm':
			f.multiline = true
		case 'E':
			f.dollarEndOnly = true
		case 'U':
			f.ungreedy = true
		case 'A':
			anchored = true
		case 'x':
			return nil, rxUnsupported("extended syntax option")
		}
	}
	p := &rxParser{s: pattern, names: map[string]int{}, namedRefs: map[*rxNode]string{}}
	root, err := p.alternation(&f)
	if err != nil {
		return nil, err
	}
	if p.i < len(p.s) {
		return nil, errors.New("unmatched closing parenthesis")
	}
	for n, name := range p.namedRefs {
		g, ok := p.names[name]
		if !ok {
			return nil, errors.New("reference to non-existent subpattern")
		}
		n.cap = g
	}
	var check func(*rxNode) error
	check = func(n *rxNode) error {
		if n.kind == rxBackref && (n.cap < 1 || n.cap > p.groups) {
			return errors.New("reference to non-existent subpattern")
		}
		for _, s := range n.subs {
			if err := check(s); err != nil {
				return err
			}
		}
		return nil
	}
	if err := check(root); err != nil {
		return nil, err
	}
	return &rxPattern{root: root, groups: p.groups, anchored: anchored}, nil
}

func (p *rxParser) more() bool { return p.i < len(p.s) }
func (p *rxParser) peek() byte { return p.s[p.i] }

// alternation parses branches up to a ')' or the end. Option changes made in
// one branch stay in effect for the later branches of the same group.
func (p *rxParser) alternation(f *rxFlags) (*rxNode, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 250 {
		return nil, errors.New("parentheses are too deeply nested")
	}
	var branches []*rxNode
	for {
		b, err := p.sequence(f)
		if err != nil {
			return nil, err
		}
		branches = append(branches, b)
		if !p.more() || p.peek() != '|' {
			break
		}
		p.i++
	}
	if len(branches) == 1 {
		return branches[0], nil
	}
	return &rxNode{kind: rxAlt, subs: branches}, nil
}

func (p *rxParser) sequence(f *rxFlags) (*rxNode, error) {
	var items []*rxNode
	for p.more() {
		c := p.peek()
		if c == '|' || c == ')' {
			break
		}
		atom, repeatable, err := p.atom(f)
		if err != nil {
			return nil, err
		}
		if atom == nil {
			continue // an option setting or a comment
		}
		q, err := p.quantifier(f)
		if err != nil {
			return nil, err
		}
		if q != nil {
			if !repeatable {
				return nil, rxUnsupported("quantified assertion")
			}
			q.subs = []*rxNode{atom}
			atom = q
			if next, _, _ := p.quantifierAhead(); next {
				return nil, errors.New("quantifier does not follow a repeatable item")
			}
		}
		items = append(items, atom)
	}
	return concat(items), nil
}

// concat joins items, merging runs of single bytes into strings.
func concat(items []*rxNode) *rxNode {
	var out []*rxNode
	for _, it := range items {
		if it.kind == rxEmpty {
			continue
		}
		if it.kind == rxSet && len(out) > 0 {
			last := out[len(out)-1]
			switch last.kind {
			case rxSet:
				out[len(out)-1] = &rxNode{kind: rxString, seq: []byteSet{last.set, it.set}}
				continue
			case rxString:
				last.seq = append(last.seq, it.set)
				continue
			}
		}
		out = append(out, it)
	}
	switch len(out) {
	case 0:
		return &rxNode{kind: rxEmpty}
	case 1:
		return out[0]
	}
	return &rxNode{kind: rxConcat, subs: out}
}

// quantifierAhead reports whether a quantifier starts at the current
// position, and its text length.
func (p *rxParser) quantifierAhead() (bool, int, error) {
	if !p.more() {
		return false, 0, nil
	}
	switch p.peek() {
	case '*', '+', '?':
		return true, 1, nil
	case '{':
		_, _, n, err := p.braces(p.i)
		return n > 0, n, err
	}
	return false, 0, nil
}

// braces reads a counted quantifier at i: {n}, {n,} or {n,m}. n is 0 when
// the text is not a quantifier (a literal brace). Forms whose reading depends
// on the library version ({,m}, spaces) are unsupported.
func (p *rxParser) braces(i int) (lo, hi, n int, err error) {
	end := strings.IndexByte(p.s[i:], '}')
	if end < 0 {
		return 0, 0, 0, nil
	}
	body := p.s[i+1 : i+end]
	a, b, comma := strings.Cut(body, ",")
	digits := func(s string) bool {
		for j := 0; j < len(s); j++ {
			if !isRxDigit(s[j]) {
				return false
			}
		}
		return s != ""
	}
	switch {
	case digits(a) && (!comma || b == "" || digits(b)):
	case strings.ContainsAny(body, " \t") && strings.Trim(strings.ReplaceAll(body, ",", ""), " \t0123456789") == "":
		return 0, 0, 0, rxUnsupported("spaces in a counted quantifier")
	case a == "" && comma && digits(b):
		return 0, 0, 0, rxUnsupported("{,m} quantifier")
	default:
		return 0, 0, 0, nil
	}
	lo, err1 := strconv.Atoi(a)
	hi = lo
	if comma {
		hi = -1
		if b != "" {
			var err2 error
			if hi, err2 = strconv.Atoi(b); err2 != nil {
				return 0, 0, 0, errors.New("number too big in quantifier")
			}
		}
	}
	if err1 != nil || lo > 65535 || hi > 65535 {
		return 0, 0, 0, errors.New("number too big in quantifier")
	}
	if hi >= 0 && hi < lo {
		return 0, 0, 0, errors.New("numbers out of order in quantifier")
	}
	return lo, hi, end + 1, nil
}

func (p *rxParser) quantifier(f *rxFlags) (*rxNode, error) {
	ok, n, err := p.quantifierAhead()
	if err != nil || !ok {
		return nil, err
	}
	q := &rxNode{kind: rxRepeat}
	switch p.peek() {
	case '*':
		q.min, q.max = 0, -1
	case '+':
		q.min, q.max = 1, -1
	case '?':
		q.min, q.max = 0, 1
	case '{':
		q.min, q.max, _, _ = p.braces(p.i)
	}
	p.i += n
	lazy := false
	if p.more() {
		switch p.peek() {
		case '?':
			lazy = true
			p.i++
		case '+':
			q.possessive = true
			p.i++
		}
	}
	q.greedy = lazy == f.ungreedy
	if q.possessive {
		q.greedy = true
	}
	return q, nil
}

// atom parses one item. repeatable is false for assertions. A nil node is
// an option setting or comment, which produces nothing.
func (p *rxParser) atom(f *rxFlags) (n *rxNode, repeatable bool, err error) {
	c := p.peek()
	switch c {
	case '(':
		return p.group(f)
	case '[':
		if p.i+1 < len(p.s) && strings.IndexByte(":.=", p.s[p.i+1]) >= 0 {
			if _, ok := p.posixSyntax(p.i + 1); ok {
				return nil, false, errors.New("POSIX named classes are supported only within a class")
			}
		}
		p.i++
		set, err := p.class(f)
		if err != nil {
			return nil, false, err
		}
		return &rxNode{kind: rxSet, set: set}, true, nil
	case '.':
		p.i++
		set := classSet(func(b byte) bool { return b != '\n' }, false)
		if f.dotall {
			set = classSet(func(byte) bool { return true }, false)
		}
		return &rxNode{kind: rxSet, set: set, dot: true}, true, nil
	case '^':
		p.i++
		if f.multiline {
			return &rxNode{kind: rxAssert, assert: asLineStart}, false, nil
		}
		return &rxNode{kind: rxAssert, assert: asSubjectStart}, false, nil
	case '$':
		p.i++
		switch {
		case f.multiline:
			return &rxNode{kind: rxAssert, assert: asLineEnd}, false, nil
		case f.dollarEndOnly:
			return &rxNode{kind: rxAssert, assert: asSubjectEnd}, false, nil
		}
		return &rxNode{kind: rxAssert, assert: asEndOrFinalNL}, false, nil
	case '\\':
		return p.escape(f)
	case '*', '+', '?':
		return nil, false, errors.New("quantifier does not follow a repeatable item")
	case '{':
		if ok, _, err := p.quantifierAhead(); err != nil {
			return nil, false, err
		} else if ok {
			return nil, false, errors.New("quantifier does not follow a repeatable item")
		}
	}
	p.i++
	return p.literal(c, f), true, nil
}

func (p *rxParser) literal(c byte, f *rxFlags) *rxNode {
	n := &rxNode{kind: rxSet}
	n.set.add(c)
	if f.caseless {
		n.set.fold()
	}
	return n
}

func (p *rxParser) group(f *rxFlags) (*rxNode, bool, error) {
	p.i++ // (
	rest := p.s[p.i:]
	inner := *f
	body := func(wrap *rxNode) (*rxNode, bool, error) {
		sub, err := p.alternation(&inner)
		if err != nil {
			return nil, false, err
		}
		if !p.more() || p.peek() != ')' {
			return nil, false, errors.New("missing closing parenthesis")
		}
		p.i++
		wrap.subs = []*rxNode{sub}
		return wrap, wrap.kind != rxLook, nil
	}
	switch {
	case strings.HasPrefix(rest, "*"):
		return nil, false, rxUnsupported("control verb")
	case !strings.HasPrefix(rest, "?"):
		p.groups++
		return body(&rxNode{kind: rxGroup, cap: p.groups})
	}
	rest = rest[1:]
	switch {
	case strings.HasPrefix(rest, "#"):
		end := strings.IndexByte(rest, ')')
		if end < 0 {
			return nil, false, errors.New("missing ) after comment")
		}
		p.i += 1 + end + 1
		return nil, false, nil
	case strings.HasPrefix(rest, ":"):
		p.i += 2
		return body(&rxNode{kind: rxGroup})
	case strings.HasPrefix(rest, ">"):
		p.i += 2
		return body(&rxNode{kind: rxAtomic})
	case strings.HasPrefix(rest, "="), strings.HasPrefix(rest, "!"):
		p.i += 2
		return body(&rxNode{kind: rxLook, negate: rest[0] == '!'})
	case strings.HasPrefix(rest, "<="), strings.HasPrefix(rest, "<!"):
		p.i += 3
		n, _, err := body(&rxNode{kind: rxLook, behind: true, negate: rest[1] == '!'})
		if err != nil {
			return nil, false, err
		}
		if err := setLookbehindLengths(n); err != nil {
			return nil, false, err
		}
		return n, false, nil
	case strings.HasPrefix(rest, "P<"), strings.HasPrefix(rest, "<"), strings.HasPrefix(rest, "'"):
		skip, close := 2, byte('>')
		switch rest[0] {
		case 'P':
			skip = 3
		case '\'':
			close = '\''
		}
		name, err := p.groupName(p.i+skip, close)
		if err != nil {
			return nil, false, err
		}
		p.i += skip + len(name) + 1
		if _, dup := p.names[name]; dup {
			return nil, false, errors.New("two named subpatterns have the same name")
		}
		p.groups++
		p.names[name] = p.groups
		return body(&rxNode{kind: rxGroup, cap: p.groups})
	case strings.HasPrefix(rest, "P="):
		name, err := p.groupName(p.i+3, ')')
		if err != nil {
			return nil, false, err
		}
		p.i += 3 + len(name) + 1
		n := &rxNode{kind: rxBackref, caseless: f.caseless}
		p.namedRefs[n] = name
		return n, true, nil
	}
	// Option settings: (?imsU-imsU) or (?imsU-imsU:...).
	j := 0
	on := true
	opts := *f
	for ; j < len(rest); j++ {
		switch rest[j] {
		case 'i':
			opts.caseless = on
		case 'm':
			opts.multiline = on
		case 's':
			opts.dotall = on
		case 'U':
			opts.ungreedy = on
		case '-':
			if !on {
				return nil, false, errors.New("invalid option setting")
			}
			on = false
		case ')':
			p.i += 1 + j + 1
			*f = opts // the rest of the enclosing group
			return nil, false, nil
		case ':':
			p.i += 1 + j + 1
			inner = opts
			return body(&rxNode{kind: rxGroup})
		default:
			return nil, false, rxUnsupported("group syntax (?" + string(rest[j]))
		}
	}
	return nil, false, errors.New("missing closing parenthesis")
}

func (p *rxParser) groupName(i int, close byte) (string, error) {
	end := strings.IndexByte(p.s[i:], close)
	if end <= 0 || end > 32 {
		return "", errors.New("invalid group name")
	}
	name := p.s[i : i+end]
	for j := 0; j < len(name); j++ {
		c := name[j]
		if !(c == '_' || c|0x20 >= 'a' && c|0x20 <= 'z' || j > 0 && isRxDigit(c)) {
			return "", errors.New("invalid group name")
		}
	}
	return name, nil
}

// setLookbehindLengths requires each alternative of a lookbehind to have one
// fixed length, which every library version accepts the same way.
func setLookbehindLengths(n *rxNode) error {
	alts := []*rxNode{n.subs[0]}
	if n.subs[0].kind == rxAlt {
		alts = n.subs[0].subs
	}
	for _, a := range alts {
		l, ok := fixedLength(a)
		if !ok {
			return rxUnsupported("variable-length lookbehind")
		}
		n.lengths = append(n.lengths, l)
	}
	return nil
}

func fixedLength(n *rxNode) (int, bool) {
	switch n.kind {
	case rxEmpty, rxAssert, rxLook:
		return 0, true
	case rxSet:
		return 1, true
	case rxString:
		return len(n.seq), true
	case rxConcat:
		t := 0
		for _, s := range n.subs {
			l, ok := fixedLength(s)
			if !ok {
				return 0, false
			}
			t += l
		}
		return t, true
	case rxAlt:
		first := -1
		for _, s := range n.subs {
			l, ok := fixedLength(s)
			if !ok || first >= 0 && l != first {
				return 0, false
			}
			first = l
		}
		return first, true
	case rxGroup, rxAtomic:
		return fixedLength(n.subs[0])
	case rxRepeat:
		if n.min != n.max {
			return 0, false
		}
		l, ok := fixedLength(n.subs[0])
		return l * n.min, ok
	}
	return 0, false
}

// escape parses a backslash sequence outside a class.
func (p *rxParser) escape(f *rxFlags) (*rxNode, bool, error) {
	if p.i+1 >= len(p.s) {
		return nil, false, errors.New("\\ at end of pattern")
	}
	c := p.s[p.i+1]
	switch c {
	case 'A':
		p.i += 2
		return &rxNode{kind: rxAssert, assert: asSubjectStart}, false, nil
	case 'z':
		p.i += 2
		return &rxNode{kind: rxAssert, assert: asSubjectEnd}, false, nil
	case 'Z':
		p.i += 2
		return &rxNode{kind: rxAssert, assert: asEndOrFinalNL}, false, nil
	case 'b':
		p.i += 2
		return &rxNode{kind: rxAssert, assert: asWordBoundary}, false, nil
	case 'B':
		p.i += 2
		return &rxNode{kind: rxAssert, assert: asNotWordBoundary}, false, nil
	case 'g':
		return p.gReference(f)
	case 'k':
		if p.i+2 < len(p.s) {
			close := map[byte]byte{'<': '>', '\'': '\'', '{': '}'}[p.s[p.i+2]]
			if close != 0 {
				name, err := p.groupName(p.i+3, close)
				if err != nil {
					return nil, false, err
				}
				p.i += 3 + len(name) + 1
				n := &rxNode{kind: rxBackref, caseless: f.caseless}
				p.namedRefs[n] = name
				return n, true, nil
			}
		}
		return nil, false, errors.New("\\k is not followed by a name")
	case 'N':
		p.i += 2
		return &rxNode{kind: rxSet, set: classSet(func(b byte) bool { return b != '\n' }, false)}, true, nil
	}
	if c >= '1' && c <= '9' {
		// A decimal number is a backreference when it is below 10, starts
		// with 8 or 9, or names a group already opened; otherwise up to
		// three octal digits are a byte value.
		j := p.i + 1
		for j < len(p.s) && isRxDigit(p.s[j]) {
			j++
		}
		num, err := strconv.Atoi(p.s[p.i+1 : j])
		if err != nil || num < 10 || c >= '8' || num <= p.groups {
			if err != nil {
				return nil, false, errors.New("number too big")
			}
			p.i = j
			return &rxNode{kind: rxBackref, cap: num, caseless: f.caseless}, true, nil
		}
	}
	set, isSet, value, n, err := p.escapeValue(p.i+1, false)
	if err != nil {
		return nil, false, err
	}
	p.i += 1 + n
	if isSet {
		return &rxNode{kind: rxSet, set: set}, true, nil
	}
	return p.literal(value, f), true, nil
}

func (p *rxParser) gReference(f *rxFlags) (*rxNode, bool, error) {
	i := p.i + 2
	text, braced := "", false
	switch {
	case i < len(p.s) && p.s[i] == '{':
		end := strings.IndexByte(p.s[i:], '}')
		if end < 0 {
			return nil, false, errors.New("\\g is not followed by a group")
		}
		text, braced = p.s[i+1:i+end], true
		i += end + 1
	default:
		j := i
		if j < len(p.s) && (p.s[j] == '-' || p.s[j] == '+') {
			j++
		}
		for j < len(p.s) && isRxDigit(p.s[j]) {
			j++
		}
		text = p.s[i:j]
		i = j
	}
	if strings.HasPrefix(text, "+") {
		return nil, false, rxUnsupported("\\g+ reference")
	}
	n := &rxNode{kind: rxBackref, caseless: f.caseless}
	if num, err := strconv.Atoi(text); err == nil {
		if num < 0 {
			num = p.groups + 1 + num
		}
		if num <= 0 {
			return nil, false, errors.New("invalid \\g reference")
		}
		n.cap = num
	} else if braced {
		p.namedRefs[n] = text
	} else {
		return nil, false, rxUnsupported("\\g subroutine call")
	}
	p.i = i
	return n, true, nil
}

// escapeValue reads the escape starting at s[i] (after the backslash): a
// class escape (isSet) or a byte value, and the number of bytes read.
// inClass applies the class rules: \b is a backspace and a digit always
// starts an octal value (\8 and \9 are the digits themselves).
func (p *rxParser) escapeValue(i int, inClass bool) (set byteSet, isSet bool, value byte, n int, err error) {
	c := p.s[i]
	switch c {
	case 'd', 'D':
		return classSet(isRxDigit, c == 'D'), true, 0, 1, nil
	case 's', 'S':
		return classSet(isRxSpace, c == 'S'), true, 0, 1, nil
	case 'w', 'W':
		return classSet(isRxWord, c == 'W'), true, 0, 1, nil
	case 'h', 'H':
		return classSet(func(b byte) bool { return b == 9 || b == ' ' || b == 0xa0 }, c == 'H'), true, 0, 1, nil
	case 'v', 'V':
		return classSet(func(b byte) bool { return b >= 10 && b <= 13 || b == 0x85 }, c == 'V'), true, 0, 1, nil
	case 'a':
		return set, false, 7, 1, nil
	case 'e':
		return set, false, 0x1b, 1, nil
	case 'f':
		return set, false, 0x0c, 1, nil
	case 'n':
		return set, false, '\n', 1, nil
	case 'r':
		return set, false, '\r', 1, nil
	case 't':
		return set, false, '\t', 1, nil
	case 'b':
		if inClass {
			return set, false, 8, 1, nil
		}
	case 'c':
		if i+1 >= len(p.s) || p.s[i+1] < 32 || p.s[i+1] > 126 {
			return set, false, 0, 0, errors.New("\\c must be followed by a printable character")
		}
		x := p.s[i+1]
		if x >= 'a' && x <= 'z' {
			x -= 32
		}
		return set, false, x ^ 0x40, 2, nil
	case 'x':
		if i+1 < len(p.s) && p.s[i+1] == '{' {
			end := strings.IndexByte(p.s[i+1:], '}')
			if end < 0 {
				return set, false, 0, 0, errors.New("missing } after \\x{")
			}
			v, e := strconv.ParseUint(p.s[i+2:i+1+end], 16, 32)
			if e != nil || v > 255 {
				return set, false, 0, 0, errors.New("character code point value in \\x{} is too large")
			}
			return set, false, byte(v), 1 + end + 1, nil
		}
		v, j := 0, i+1
		for ; j < len(p.s) && j < i+3 && hexNibble(p.s[j]) >= 0; j++ {
			v = v*16 + hexNibble(p.s[j])
		}
		return set, false, byte(v), j - i, nil
	case 'o':
		if i+1 >= len(p.s) || p.s[i+1] != '{' {
			return set, false, 0, 0, errors.New("missing { after \\o")
		}
		end := strings.IndexByte(p.s[i+1:], '}')
		if end < 0 {
			return set, false, 0, 0, errors.New("missing } after \\o{")
		}
		v, e := strconv.ParseUint(p.s[i+2:i+1+end], 8, 32)
		if e != nil || v > 255 {
			return set, false, 0, 0, errors.New("octal value is too large")
		}
		return set, false, byte(v), 1 + end + 1, nil
	case '8', '9':
		if inClass {
			return set, false, c, 1, nil
		}
	}
	if c >= '0' && c <= '7' {
		v, j := 0, i
		for ; j < len(p.s) && j < i+3 && p.s[j] >= '0' && p.s[j] <= '7'; j++ {
			v = v*8 + int(p.s[j]-'0')
		}
		if v > 255 {
			return set, false, 0, 0, errors.New("octal value is too large")
		}
		return set, false, byte(v), j - i, nil
	}
	switch {
	case c == 'Q' || c == 'E':
		return set, false, 0, 0, rxUnsupported("\\Q...\\E quoting")
	case c == 'R' || c == 'X' || c == 'p' || c == 'P' || c == 'C' || c == 'K' || c == 'G':
		return set, false, 0, 0, rxUnsupported("\\" + string(c))
	case c|0x20 >= 'a' && c|0x20 <= 'z' || isRxDigit(c):
		return set, false, 0, 0, errors.New("unrecognized character follows \\")
	}
	return set, false, c, 1, nil
}

var posixClasses = map[string]func(byte) bool{
	"alnum":  func(b byte) bool { return isRxDigit(b) || b|0x20 >= 'a' && b|0x20 <= 'z' },
	"alpha":  func(b byte) bool { return b|0x20 >= 'a' && b|0x20 <= 'z' },
	"ascii":  func(b byte) bool { return b < 128 },
	"blank":  func(b byte) bool { return b == ' ' || b == '\t' },
	"cntrl":  func(b byte) bool { return b < 32 || b == 127 },
	"digit":  isRxDigit,
	"graph":  func(b byte) bool { return b > 32 && b < 127 },
	"lower":  func(b byte) bool { return b >= 'a' && b <= 'z' },
	"print":  func(b byte) bool { return b >= 32 && b < 127 },
	"punct":  func(b byte) bool { return b > 32 && b < 127 && !isRxWord(b) || b == '_' },
	"space":  isRxSpace,
	"upper":  func(b byte) bool { return b >= 'A' && b <= 'Z' },
	"word":   isRxWord,
	"xdigit": func(b byte) bool { return hexNibble(b) >= 0 },
}

// posixSyntax reports whether a POSIX bracket item starts at i (':', '.' or
// '=' after a '['), as the library decides it: its terminator followed by
// ']' comes before a ']' or another '[' with the same character; "\]" and
// "\\" are skipped. It returns the terminator's position.
func (p *rxParser) posixSyntax(i int) (int, bool) {
	term := p.s[i]
	for j := i + 1; len(p.s)-j >= 2; j++ {
		switch {
		case p.s[j] == '\\' && (p.s[j+1] == ']' || p.s[j+1] == '\\'):
			j++
		case p.s[j] == '[' && p.s[j+1] == term || p.s[j] == ']':
			return 0, false
		case p.s[j] == term && p.s[j+1] == ']':
			return j, true
		}
	}
	return 0, false
}

// class parses a character class after its '['.
func (p *rxParser) class(f *rxFlags) (byteSet, error) {
	var set byteSet
	negate := false
	if p.more() && p.peek() == '^' {
		negate = true
		p.i++
	}
	first := true
	// item reads one class member: a single byte, or a set (class escape or
	// POSIX class).
	item := func() (byte, *byteSet, error) {
		c := p.peek()
		if c == '[' && p.i+1 < len(p.s) && strings.IndexByte(":.=", p.s[p.i+1]) >= 0 {
			if term, ok := p.posixSyntax(p.i + 1); ok {
				if p.s[p.i+1] != ':' {
					return 0, nil, errors.New("POSIX collating elements are not supported")
				}
				end := term - (p.i + 2)
				name := p.s[p.i+2 : term]
				if name == "<" || name == ">" {
					return 0, nil, rxUnsupported("[[:<:]] word boundaries")
				}
				neg := strings.HasPrefix(name, "^")
				name = strings.TrimPrefix(name, "^")
				pred, ok := posixClasses[name]
				if !ok {
					return 0, nil, errors.New("unknown POSIX class name")
				}
				if f.caseless && (name == "lower" || name == "upper") {
					pred = posixClasses["alpha"]
				}
				s := classSet(pred, neg)
				p.i += 2 + end + 2
				return 0, &s, nil
			}
		}
		if c == '\\' {
			if p.i+1 >= len(p.s) {
				return 0, nil, errors.New("\\ at end of pattern")
			}
			switch p.s[p.i+1] {
			case 'N', 'B', 'A', 'z', 'Z', 'G', 'X', 'R', 'g', 'k':
				return 0, nil, errors.New("escape sequence is invalid in character class")
			}
			s, isSet, v, n, err := p.escapeValue(p.i+1, true)
			if err != nil {
				return 0, nil, err
			}
			p.i += 1 + n
			if isSet {
				return 0, &s, nil
			}
			return v, nil, nil
		}
		p.i++
		return c, nil, nil
	}
	for {
		if !p.more() {
			return set, errors.New("missing terminating ] for character class")
		}
		if p.peek() == ']' && !first {
			p.i++
			break
		}
		first = false
		lo, loSet, err := item()
		if err != nil {
			return set, err
		}
		rangeAhead := p.i+1 < len(p.s) && p.peek() == '-' && p.s[p.i+1] != ']'
		if loSet != nil {
			if rangeAhead {
				return set, errors.New("invalid range in character class")
			}
			set.union(loSet)
			continue
		}
		if !rangeAhead {
			set.add(lo)
			continue
		}
		p.i++ // -
		hi, hiSet, err := item()
		if err != nil {
			return set, err
		}
		if hiSet != nil {
			return set, errors.New("invalid range in character class")
		}
		if hi < lo {
			return set, errors.New("range out of order in character class")
		}
		set.addRange(int(lo), int(hi))
	}
	if f.caseless {
		set.fold()
	}
	if negate {
		set.invert()
	}
	return set, nil
}
