package signatures

import (
	"errors"
	"math/bits"
	"strconv"
	"strings"
)

type numberRange struct{ min, max int64 }

func (r numberRange) contains(v int64) bool { return v >= r.min && v <= r.max }

// parseRange reads a logical attribute range as the reference does: a '-' is
// required and each side is digits or empty (0), read as an unsigned 32-bit
// value. A minimum above the maximum is valid and contains nothing.
func parseRange(s string) (*numberRange, error) {
	lo, hi, ok := strings.Cut(s, "-")
	if !ok {
		return nil, errors.New("invalid numeric range: a range requires '-'")
	}
	if !isNumber(lo) || !isNumber(hi) {
		return nil, errors.New("invalid numeric range: only digits are allowed")
	}
	return &numberRange{int64(uint32(cAtoi(lo))), int64(uint32(cAtoi(hi)))}, nil
}

// maxSubsignatures is the reference limit on logical subsignatures.
const maxSubsignatures = 64

type expressionKey struct {
	text  string
	count int
}

// Logical expressions are immutable after parsing. Official feeds repeat the
// same trees across many rules; sharing them does not share per-object counts.
func (d *Database) sharedExpression(text string, count int) (*expression, error) {
	key := expressionKey{text, count}
	if expr := d.expressionIDs[key]; expr != nil {
		return expr, nil
	}
	expr, err := parseExpression(text, count)
	if err == nil && len(d.expressionIDs) < 65536 {
		key.text = strings.Clone(text) // do not retain the complete database line
		d.expressionIDs[key] = expr
	}
	return expr, err
}

// addLogicalRule compiles one logical rule, selecting it as the reference
// does first: the PUA and ignore-list filters, the expression syntax and its
// subsignature count, the attribute block, the Engine window and the target.
// Only a selected rule is checked against what this engine evaluates, so an
// unselected rule may use syntax this engine lacks. forBytecode applies the
// rules for a bytecode trigger, whose program may use subsignatures the
// expression does not name; entry is then the rule name.
func (d *Database) addLogicalRule(line, entry string, forBytecode bool) error {
	p := ldbTokenize(line, maxSubsignatures+4)
	if len(p) < 4 {
		return errors.New("logical signature requires 1-64 subsignatures")
	}
	if forBytecode {
		entry = p[0]
	}
	if d.ignore.ignored(p[0], entry) {
		return skip(SkipIgnoreList)
	}
	highest, err := lsigHighestIndex(p[2])
	if err != nil {
		return err
	}
	count := highest + 1
	switch {
	case forBytecode && count > len(p)-3:
		return errors.New("logical expression names more subsignatures than the bytecode trigger has")
	case forBytecode:
		count = len(p) - 3
	case count != len(p)-3:
		return errors.New("logical expression and subsignature count do not match")
	}
	if count > maxSubsignatures {
		return errors.New("logical signature requires 1-64 subsignatures")
	}
	t, err := parseTargetDescription(p[1])
	if err != nil {
		return err
	}
	level := uint32(d.functionality())
	if t.engine != nil && (t.engine.min > int64(level) || t.engine.max < int64(level)) {
		return skip(SkipLevel)
	}
	if t.target == nil {
		return errors.New("logical signature requires an explicit target")
	}
	if *t.target >= maxTargets {
		return skip(SkipTarget)
	}
	if t.iconGroup && *t.target != 1 {
		return errors.New("icon group descriptors require target 1")
	}
	if (t.entry != nil || t.sections != nil) && *t.target != 1 && *t.target != 6 && *t.target != 9 {
		return errors.New("executable descriptors require an executable target")
	}
	// The reference selects the rule; everything below is this engine's subset.
	if e := validName(p[0]); e != nil {
		return e
	}
	if t.dropped {
		return errors.New("more than ten logical target descriptors are unsupported")
	}
	seen := map[string]bool{}
	for _, name := range t.names {
		if seen[name] {
			return errors.New("invalid or duplicate logical target descriptor")
		}
		seen[name] = true
	}
	// A copy, so the rule does not keep its whole line alive.
	r := rule{name: strings.Clone(p[0]), ruleCold: &emptyRuleCold}
	if t.size != nil || t.entry != nil || t.sections != nil ||
		t.container != nil || len(t.intermediates) != 0 {
		r.ruleCold = &ruleCold{size: t.size, entry: t.entry, sections: t.sections,
			container: t.container, intermediates: t.intermediates}
	}
	if r.target, err = supportedTarget(*t.target); err != nil {
		return err
	}
	if r.expr, err = d.sharedExpression(p[2], len(p)-3); err != nil {
		return err
	}
	switch {
	case t.iconGroup:
		r.ownCold()
		groups := [2]string{strings.Clone(t.iconGroups[0]), strings.Clone(t.iconGroups[1])}
		r.icons = &groups
	case t.handlerType:
		r.ownCold()
		if t.handler == TypePDF || t.handler == TypeGraphics || t.handler == TypeScript || t.handler == TypeZIP {
			handler := t.handler
			r.handler = &handler
		} else {
			r.unknown = UnknownHandlerType
		}
	}
	subsignatures := make([]subsignature, 0, len(p)-3)
	var triggers []*expression
	for i, sub := range p[3:] {
		holder, isPlaceholder, err := parsePlaceholder(sub, i)
		if err != nil {
			return err
		}
		if isPlaceholder {
			subsignatures = append(subsignatures, subsignature{offset: "*", body: sub, holder: holder})
			if holder.trigger != nil {
				triggers = append(triggers, holder.trigger)
			}
			continue
		}
		off, body := "*", sub
		if i := strings.IndexByte(sub, ':'); i >= 0 && (i+1 == len(sub) || sub[i+1] != ':') {
			off, body = sub[:i], sub[i+1:]
		}
		subsignatures = append(subsignatures, subsignature{offset: off, body: body})
	}
	if r.patterns, err = d.compilePatterns(r.target, subsignatures); err != nil {
		return err
	}
	for _, sub := range subsignatures {
		if sub.holder == nil || sub.holder.bcomp == nil {
			continue
		}
		ref := sub.holder.ref
		if subsignatures[ref].holder != nil {
			// ClamAV only permits an ordinary body subsignature as the
			// trigger. Keep invalid or unmodelled forms incomplete.
			continue
		}
		r.ownCold().byteCompareRefs |= uint64(1) << uint(ref)
	}
	if len(triggers) > 0 {
		// Regex triggers may compare counts the expression itself does not.
		r.ownCold()
		r.caps = make([]int, len(r.patterns))
		for i := range r.caps {
			r.caps[i] = r.expr.countLimit(i)
			for _, trigger := range triggers {
				r.caps[i] = max(r.caps[i], trigger.countLimit(i))
			}
		}
	}
	d.rules = append(d.rules, r)
	if reason := ruleUnknownReason(&r, d.patterns); reason != "" && !forBytecode {
		d.unknown(reason, 1)
	}
	return nil
}

type expression struct {
	op                     byte
	index, value, distinct int
	left, right            *expression
	mask                   uint64
}

// logicState is three-valued: a rule whose outcome depends on a subsignature
// count that could not be computed is neither matched nor unmatched.
type logicState int8

const (
	logicFalse logicState = iota
	logicTrue
	logicUnknown
)

// expressionResult follows the reference evaluator: count and ids accumulate
// only from true subexpressions and feed the count modifiers of groups.
type expressionResult struct {
	state logicState
	count int
	ids   uint64
}

// evaluate treats a negative count as unknown. It is sound: a true or false
// result holds for every value the unknown counts could have.
func (e *expression) evaluate(counts []int) expressionResult {
	switch e.op {
	case 's':
		n := counts[e.index]
		switch {
		case n < 0:
			return expressionResult{state: logicUnknown}
		case n > 0:
			return expressionResult{logicTrue, n, uint64(1) << uint(e.index)}
		}
		return expressionResult{}
	case '&', '|':
		l, r := e.left.evaluate(counts), e.right.evaluate(counts)
		state := logicUnknown
		switch {
		case e.op == '&' && (l.state == logicFalse || r.state == logicFalse),
			e.op == '|' && l.state == logicFalse && r.state == logicFalse:
			state = logicFalse
		case e.op == '&' && l.state == logicTrue && r.state == logicTrue,
			e.op == '|' && (l.state == logicTrue || r.state == logicTrue):
			state = logicTrue
		}
		if state != logicTrue {
			return expressionResult{state: state}
		}
		result := expressionResult{state: logicTrue}
		if l.state == logicTrue {
			result.count, result.ids = l.count, l.ids
		}
		if r.state == logicTrue {
			result.count += r.count
			result.ids |= r.ids
		}
		return result
	default:
		// Only subsignatures and alternations of them carry count modifiers, so
		// every count that can reach this comparison is in its mask.
		for i := range counts {
			if e.mask&(uint64(1)<<uint(i)) != 0 && counts[i] < 0 {
				return expressionResult{state: logicUnknown}
			}
		}
		total, ids := 0, uint64(0)
		if e.left.op == 's' {
			// A single subsignature compares its count even when it is zero.
			total, ids = counts[e.left.index], e.mask
		} else if inner := e.left.evaluate(counts); inner.state == logicTrue {
			total, ids = inner.count, inner.ids
		}
		matched := false
		switch e.op {
		case '=':
			matched = total == e.value
		case '>':
			matched = total > e.value
		case '<':
			matched = total < e.value
		}
		if !matched || bits.OnesCount64(ids) < e.distinct {
			return expressionResult{}
		}
		if e.left.op == 's' {
			return expressionResult{logicTrue, total, ids}
		}
		return expressionResult{logicTrue, total, 0}
	}
}

// A cap of comparison+1 is sufficient for every supported count comparison.
// Boolean-only subsignatures stop after the first occurrence.
func (e *expression) countLimit(id int) int {
	if e == nil {
		return 1
	}
	cap := 1
	if e.op == '=' || e.op == '<' || e.op == '>' {
		if e.mask&(uint64(1)<<uint(id)) != 0 {
			cap = e.value + 1
		}
	}
	if e.left != nil {
		cap = max(cap, e.left.countLimit(id))
	}
	if e.right != nil {
		cap = max(cap, e.right.countLimit(id))
	}
	return cap
}
func (e *expression) countable() bool {
	if e.op == 's' {
		return true
	}
	return e.op == '|' && e.left.countable() && e.right.countable()
}

// expressionParser reproduces how the reference evaluator splits a logical
// expression, which gives & and | equal precedence. An unparenthesized range is
// split at its FIRST operator outside groups, so 0&1|2 is 0&(1|2). A group
// that spans an operand is split at the LAST operator directly inside it, so
// (0|1&2) is (0|1)&2, and its count modifier applies to that split. Syntax the
// reference rejects, or reads leniently (spaces, stray characters), is rejected.
type expressionParser struct {
	s     string
	count int
	close []int // index of the ')' matching each '('
}

func parseExpression(s string, count int) (*expression, error) {
	if len(s) > 4096 {
		return nil, errors.New("logical expression length limit exceeded")
	}
	p := expressionParser{s: s, count: count, close: make([]int, len(s))}
	var open []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			if open = append(open, i); len(open) > 64 {
				return nil, errors.New("logical expression nesting limit exceeded")
			}
		case ')':
			if len(open) == 0 {
				return nil, errors.New("unbalanced logical expression group")
			}
			p.close[open[len(open)-1]] = i
			open = open[:len(open)-1]
		}
	}
	if len(open) != 0 {
		return nil, errors.New("unclosed logical expression group")
	}
	return p.parse(0, len(s))
}

// operator returns the first (or last) & or | in [lo,hi) outside any group, or -1.
func (p *expressionParser) operator(lo, hi int, first bool) int {
	found := -1
	for i := lo; i < hi; i++ {
		switch p.s[i] {
		case '(':
			i = p.close[i]
		case '&', '|':
			if first {
				return i
			}
			found = i
		}
	}
	return found
}
func (p *expressionParser) parse(lo, hi int) (*expression, error) {
	if lo >= hi {
		return nil, errors.New("truncated logical expression")
	}
	if k := p.operator(lo, hi, true); k >= 0 {
		return p.binary(lo, k, hi)
	}
	if p.s[lo] != '(' {
		return p.subsignature(lo, hi)
	}
	end := p.close[lo]
	k := p.operator(lo+1, end, false)
	if k < 0 {
		// The reference strips one outer pair here; a modifier after a group
		// without an operator, such as (0)>1, is a syntax error there.
		if end != hi-1 {
			return nil, errors.New("count modifiers require an operator inside the group")
		}
		return p.parse(lo+1, end)
	}
	e, err := p.binary(lo+1, k, end)
	if err != nil || end == hi-1 {
		return e, err
	}
	return p.modifier(e, end+1, hi, true)
}
func (p *expressionParser) binary(lo, k, hi int) (*expression, error) {
	if k == lo || k == hi-1 {
		return nil, errors.New("logical operator is missing an argument")
	}
	a, err := p.parse(lo, k)
	if err != nil {
		return nil, err
	}
	b, err := p.parse(k+1, hi)
	if err != nil {
		return nil, err
	}
	return &expression{op: p.s[k], left: a, right: b, mask: a.mask | b.mask}, nil
}

// subsignature reads a leaf the reference's way: its index with sscanf %u
// (leading white space skipped, trailing text ignored); a count modifier is
// the last '=', '<' or '>' in the leaf, its value read the same way. A
// single subsignature ignores a ",Y" distinct count, as the reference does.
func (p *expressionParser) subsignature(lo, hi int) (*expression, error) {
	if strings.ContainsAny(p.s[lo:hi], "()") {
		return nil, errors.New("unsupported or trailing logical expression syntax")
	}
	mod := strings.LastIndexAny(p.s[lo:hi], "=<>")
	end := hi
	if mod >= 0 {
		mod += lo
		end = mod
	}
	index, _, err := p.number(lo, end)
	if err != nil {
		return nil, err
	}
	if index >= p.count {
		return nil, errors.New("logical subsignature index is out of range")
	}
	e := &expression{op: 's', index: index, mask: uint64(1) << uint(index)}
	if mod < 0 {
		return e, nil
	}
	n, _, err := p.number(mod+1, hi)
	if err != nil {
		return nil, err
	}
	return &expression{op: p.s[mod], value: n, left: e, mask: e.mask}, nil
}

// modifier parses a group's =N, >N or <N with an optional ,Y
// distinct-subsignature minimum. The group's count is the sum over its true
// sides when the group holds. Trailing text is rejected.
func (p *expressionParser) modifier(e *expression, pos, hi int, group bool) (*expression, error) {
	if strings.IndexByte("=<>", p.s[pos]) < 0 {
		return nil, errors.New("unsupported or trailing logical expression syntax")
	}
	op := p.s[pos]
	n, pos, err := p.digits(pos+1, hi)
	if err != nil {
		return nil, err
	}
	distinct := 0
	if pos < hi && p.s[pos] == ',' {
		if distinct, pos, err = p.digits(pos+1, hi); err != nil {
			return nil, err
		}
		if distinct > p.count {
			return nil, errors.New("logical distinct count exceeds subsignature count")
		}
	}
	if pos != hi {
		return nil, errors.New("unsupported or trailing logical expression syntax")
	}
	return &expression{op: op, value: n, distinct: distinct, left: e, mask: e.mask}, nil
}

// digits reads a plain decimal number at pos.
func (p *expressionParser) digits(pos, hi int) (int, int, error) {
	if pos >= hi || p.s[pos] < '0' || p.s[pos] > '9' {
		return 0, pos, errors.New("expected a logical expression number")
	}
	return p.number(pos, hi)
}

// number is sscanf %u over [pos, hi): white space, an optional '+', then at
// least one digit. It returns the value and the position after the digits.
func (p *expressionParser) number(pos, hi int) (int, int, error) {
	for pos < hi && cSpace(p.s[pos]) {
		pos++
	}
	if pos < hi && p.s[pos] == '+' {
		pos++
	}
	start := pos
	for pos < hi && p.s[pos] >= '0' && p.s[pos] <= '9' {
		pos++
	}
	if start == pos {
		return 0, pos, errors.New("expected a logical expression number")
	}
	n, e := strconv.Atoi(p.s[start:pos])
	if e != nil || n > 1000000 {
		return 0, pos, errors.New("logical expression number exceeds limit")
	}
	return n, pos, nil
}
