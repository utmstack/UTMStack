package signatures

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Rule selection follows the reference loader at FunctionalityLevel. A rule it
// does not load (outside its functionality window, named by an ignore list, a
// potentially-unwanted-application database, an attribute or target the
// reference does not know, or a database format switched off by the dynamic
// configuration) is skipped and counted in Report.Skipped by one of these
// reasons. A malformed selection field stays an error, as in the reference.
const (
	SkipLevel            = "level"
	SkipIgnoreList       = "ignore-list"
	SkipPUA              = "pua"
	SkipUnknownAttribute = "unknown-attribute"
	SkipTarget           = "target"
	SkipDynamicConfig    = "dconf"
	// The reference loader ignores a catalog whose signature, trust chain or
	// member structure fails verification, then loads the remaining sources.
	SkipInvalidCatalog = "invalid-catalog"
	// SkipLegacyFormat is a legacy .db line whose pattern starts with '=',
	// which the reference ignores.
	SkipLegacyFormat = "legacy-format"
	// A .pwdb entry with a numeric but unimplemented storage code is ignored
	// by the reference loader.
	SkipPasswordStorage = "password-storage"
)

// maxTargets is the reference's number of target roots; higher targets are
// skipped rather than rejected.
const maxTargets = 15

type skipError struct{ reason string }

func (e *skipError) Error() string { return "rule not selected: " + e.reason }

func skip(reason string) error { return &skipError{reason} }

func skipReason(err error) (string, bool) {
	var s *skipError
	if errors.As(err, &s) {
		return s.reason, true
	}
	return "", false
}

// cAtoi is the C atoi of the Windows runtime the endpoint replaces: leading
// whitespace, an optional sign and leading digits, 0 when there are none, and
// saturation at the int32 limits on overflow.
func cAtoi(s string) int32 {
	s = strings.TrimLeft(s, " \t\n\v\f\r")
	negative := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		negative = s[0] == '-'
		s = s[1:]
	}
	var v int64
	for i := 0; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		if v = v*10 + int64(s[i]-'0'); v > math.MaxInt32+1 {
			v = math.MaxInt32 + 1
		}
	}
	if negative {
		v = -v
	}
	return int32(max(min(v, math.MaxInt32), math.MinInt32))
}

// isNumber accepts only ASCII digits; like the reference, the empty string
// qualifies and then reads as 0.
func isNumber(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// cStrtokenize splits at every delimiter, keeping empty fields, into at most
// limit fields; text after the limit-th field is dropped, as the reference
// tokenizer does. dropped reports whether anything was.
func cStrtokenize(s string, delim byte, limit int) []string {
	fields, _ := cStrtokenizeDropped(s, delim, limit)
	return fields
}
func cStrtokenizeDropped(s string, delim byte, limit int) ([]string, bool) {
	fields := strings.SplitN(s, string(delim), limit+1)
	if len(fields) > limit {
		return fields[:limit], true
	}
	return fields, false
}

// cStrtok returns field number field of line, where runs of the delimiter
// separate fields; ok is false when the field is missing or empty.
func cStrtok(line string, field int, delim byte) (string, bool) {
	i := 0
	for counter := 0; i < len(line) && counter != field; i++ {
		if line[i] == delim {
			counter++
			for i+1 < len(line) && line[i+1] == delim {
				i++
			}
		}
	}
	if i >= len(line) {
		return "", false
	}
	j := strings.IndexByte(line[i:], delim)
	if j < 0 {
		j = len(line) - i
	}
	if j == 0 {
		return "", false
	}
	return line[i : i+j], true
}

// cSpace is the C isspace set.
func cSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'
}

// scanUnsigned reads a C sscanf %u conversion at pos: optional whitespace and
// sign, then at least one digit. Values beyond 32 bits are rejected instead of
// reproducing the C library's undefined behavior.
func scanUnsigned(s string, pos int) (uint32, bool) {
	for pos < len(s) && cSpace(s[pos]) {
		pos++
	}
	negative := false
	if pos < len(s) && (s[pos] == '+' || s[pos] == '-') {
		negative = s[pos] == '-'
		pos++
	}
	start := pos
	var v uint64
	for pos < len(s) && s[pos] >= '0' && s[pos] <= '9' {
		if v = v*10 + uint64(s[pos]-'0'); v > math.MaxUint32 {
			return 0, false
		}
		pos++
	}
	if pos == start {
		return 0, false
	}
	if negative {
		return uint32(-v), true
	}
	return uint32(v), true
}

// fieldWindow is the window check of .ndb, .cdb, .ftm and .crb rules: digit-only
// fields (empty reads as 0, so an empty maximum excludes every real level),
// inclusive at both ends. The minimum decides before the maximum is read.
func (d *Database) fieldWindow(minField string, maxField *string) error {
	level := uint32(d.functionality())
	if !isNumber(minField) {
		return errors.New("invalid minimum functionality level")
	}
	if uint32(cAtoi(minField)) > level {
		return skip(SkipLevel)
	}
	if maxField != nil {
		if !isNumber(*maxField) {
			return errors.New("invalid maximum functionality level")
		}
		if uint32(cAtoi(*maxField)) < level {
			return skip(SkipLevel)
		}
	}
	return nil
}

// hashWindow is the hash-file check, which runs before any other validation:
// both fields are read with atoi and never rejected, the minimum compares as
// an unsigned value and the maximum after an unsigned cast. It returns the
// minimum, which wildcard-size rules must also satisfy.
func (d *Database) hashWindow(p []string) (uint32, error) {
	if len(p) <= 3 {
		return 0, nil
	}
	level := uint32(d.functionality())
	minimum := uint32(cAtoi(p[3]))
	if len(p) > 5 {
		return 0, errors.New("invalid hash signature fields")
	}
	if level < minimum {
		return 0, skip(SkipLevel)
	}
	if len(p) == 5 && level > uint32(cAtoi(p[4])) {
		return 0, skip(SkipLevel)
	}
	return minimum, nil
}

// parseHashSize applies the reference size rule after the window: '*' needs a
// declared minimum level of at least 73; otherwise a decimal 0 < size <
// 0xffffffff. A negative size is rejected rather than wrapped.
func parseHashSize(s string, fields int, minimum uint32) (int64, error) {
	if s == "*" {
		if fields < 4 || minimum < 73 {
			return 0, errors.New("wildcard-size hash signatures require a minimum functionality level of at least 73")
		}
		return -1, nil
	}
	t := strings.TrimLeft(s, " \t\n\v\f\r")
	t = strings.TrimPrefix(t, "+")
	n, err := strconv.ParseUint(t, 10, 64)
	if err != nil || t == "" || t[0] < '0' || t[0] > '9' || n == 0 || n >= 0xffffffff {
		return 0, fmt.Errorf("invalid signature size %q", s)
	}
	return int64(n), nil
}

// phishingLevel applies the .pdb/.wdb level suffix check to a whole line. The
// text after the last ':' is a level only when it contains '-' and the text on
// each side of the first '-' is digits or empty: an empty minimum is 0, an
// empty maximum is unbounded and both ends are inclusive. A selected line
// loses its suffix; anything else keeps it as part of the pattern.
func phishingLevel(line string, level int) (string, error) {
	colon := strings.LastIndexByte(line, ':')
	if colon < 0 {
		return line, nil
	}
	field := line[colon+1:]
	dash := strings.IndexByte(field, '-')
	if dash < 0 || !isNumber(field[:dash]) || !isNumber(field[dash+1:]) {
		return line, nil
	}
	minimum := int64(cAtoi(field[:dash]))
	maximum := int64(math.MaxInt32)
	if field[dash+1:] != "" {
		maximum = int64(cAtoi(field[dash+1:]))
	}
	if minimum > int64(level) || maximum < int64(level) {
		return "", skip(SkipLevel)
	}
	return line[:colon], nil
}

// bytecodeWindow reads only the header fields the reference checks before it
// validates anything else in a bytecode file: the magic, the format level (6
// or 7, otherwise the file is skipped) and the functionality window, where 0
// means no bound. Fields after the format are read optimistically, as there:
// a failed number leaves its position unchanged and reads as 0.
func (d *Database) bytecodeWindow(b []byte) error {
	h, err := readBytecodeHeader(b)
	if err != nil {
		return err
	}
	if h.format != 6 && h.format != 7 {
		return skip(SkipLevel)
	}
	level := uint32(d.functionality())
	if h.minimum != 0 && h.minimum > level || h.maximum != 0 && h.maximum < level {
		return skip(SkipLevel)
	}
	return nil
}

// bytecodeKind is a bytecode file's kind from its header, or 0.
func bytecodeKind(b []byte) uint64 {
	h, _ := readBytecodeHeader(b)
	return h.kind
}

type bytecodeHeader struct {
	format           uint32
	kind             uint64
	minimum, maximum uint32
}

// readBytecodeHeader reads the header fields the reference checks before it
// validates a bytecode file.
func readBytecodeHeader(b []byte) (bytecodeHeader, error) {
	var h bytecodeHeader
	// The reference reads the header from its first line buffer.
	line := b[:min(len(b), 8191)]
	if i := strings.IndexByte(string(line), '\n'); i >= 0 {
		line = line[:i]
	}
	s := strings.TrimRight(string(line), "\r")
	if !strings.HasPrefix(s, "ClamBC") {
		return h, errors.New("missing signature bytecode magic")
	}
	pos := len("ClamBC")
	ok := true
	read := func() (uint64, bool) {
		if pos >= len(s) || s[pos] < 0x60 || s[pos]-0x60 > 16 {
			ok = false
			return 0, false
		}
		n := int(s[pos] - 0x60)
		if pos+n+1 > len(s) {
			ok = false
			return 0, false
		}
		var v uint64
		for i := 0; i < n; i++ {
			c := s[pos+1+i]
			if c&0xf0 != 0x60 {
				ok = false
				return 0, false
			}
			v |= uint64(c&15) << uint(4*i)
		}
		pos += n + 1
		return v, true
	}
	number := func() uint64 {
		v, _ := read()
		return v
	}
	text := func() {
		if pos >= len(s) || s[pos] != '|' {
			ok = false
			return
		}
		pos++
		// Only this length read decides whether the data is read.
		n, good := read()
		if !good || n == 0 {
			return
		}
		if uint64(len(s)-pos) < 2*n {
			ok = false
			return
		}
		last := byte(0)
		for i := pos; i < pos+int(2*n); i += 2 {
			if s[i]&0xf0 != 0x60 || s[i+1]&0xf0 != 0x60 {
				ok = false // an invalid pair leaves the position unchanged
				return
			}
			last = s[i]&15 | s[i+1]&15<<4
		}
		pos += int(2 * n)
		if last != 0 {
			ok = false // strings end in NUL
		}
	}
	h.format = uint32(number())
	if !ok {
		return h, errors.New("malformed signature bytecode format level")
	}
	if h.format != 6 && h.format != 7 {
		return h, nil
	}
	number() // timestamp
	text()   // author
	number() // target exclusion
	h.kind = number()
	h.minimum, h.maximum = uint32(number()), uint32(number())
	return h, nil
}

// ldbTokenize splits a logical signature at ';' like the reference: after the
// first two fields, an unescaped '/' toggles a regular-expression section in
// which ';' does not separate fields. At most limit fields are returned.
func ldbTokenize(s string, limit int) []string {
	var fields []string
	within := false
	i := 0
	for len(fields) < limit {
		start := i
		for i < len(s) {
			if !within && s[i] == ';' {
				break
			}
			if len(fields) >= 2 && i > 0 && s[i-1] != '\\' && s[i] == '/' {
				within = !within
			}
			i++
		}
		fields = append(fields, s[start:i])
		if i >= len(s) {
			break
		}
		i++
	}
	return fields
}

// maxLogicalExpression bounds the expression the reference-compatible
// checker walks; its recursion is quadratic in the expression length.
const maxLogicalExpression = 4096

// lsigHighestIndex reproduces the reference's parse-only logical expression
// check, which runs before the rule's window: it returns the highest
// subsignature index named, or an error where the reference rejects the
// syntax. It is more lenient than parseExpression, which accepts only what
// the evaluator implements; that stricter parser runs once the rule is
// selected.
func lsigHighestIndex(expr string) (int, error) {
	if len(expr) > maxLogicalExpression {
		return 0, errors.New("logical expression length limit exceeded")
	}
	return chklsig(expr, 0, len(expr))
}

func chklsig(s string, lo, end int) (int, error) {
	// at reads like C: the ranges are windows into one NUL-terminated string.
	at := func(i int) byte {
		if i < len(s) {
			return s[i]
		}
		return 0
	}
	length := end - lo
	pth, op, op1 := 0, byte(0), byte(0)
	opoff, op1off, blkend, modoff := 0, 0, 0, 0
	mod := byte(0)
	rend := end
	for i := 0; i < length; i++ {
		switch c := at(lo + i); c {
		case '(':
			pth++
		case ')':
			if pth == 0 {
				return 0, errors.New("logical expression group is not opened")
			}
			pth--
			mod, modoff = c, i
		case '>', '<', '=':
			mod, modoff = c, i
		case '&', '|':
			if pth == 0 {
				op, opoff = c, i
			} else if pth == 1 {
				op1, op1off = c, i
			}
		}
		if op != 0 {
			break
		}
		if op1 != 0 && pth == 0 {
			blkend = i
			if next := at(lo + i + 1); next == '>' || next == '<' || next == '=' {
				if _, ok := scanUnsigned(s, lo+i+2); !ok {
					return 0, errors.New("logical expression count modifier lacks a number")
				}
				for i += 2; i+1 < length && (at(lo+i+1) >= '0' && at(lo+i+1) <= '9' || at(lo+i+1) == ','); i++ {
				}
			}
			if lo+i+1 == rend {
				break
			}
		}
	}
	if pth != 0 {
		return 0, errors.New("logical expression group is not closed")
	}
	if op == 0 && op1 == 0 {
		if at(lo) == '(' {
			return chklsig(s, lo+1, end-1)
		}
		id, ok := scanUnsigned(s, lo)
		if !ok || id > math.MaxInt32 {
			return 0, errors.New("logical expression subsignature index is invalid")
		}
		if mod != 0 {
			if _, ok := scanUnsigned(s, lo+modoff+1); !ok {
				return 0, errors.New("logical expression count modifier lacks a number")
			}
		}
		return int(id), nil
	}
	lstart := lo
	if op == 0 {
		opoff, lstart, rend = op1off, lo+1, lo+blkend
	}
	if opoff == 0 {
		return 0, errors.New("logical operator is missing its left argument")
	}
	if opoff+1 == length {
		return 0, errors.New("logical operator is missing its right argument")
	}
	left, err := chklsig(s, lstart, lo+opoff)
	if err != nil {
		return 0, err
	}
	right, err := chklsig(s, lo+opoff+1, rend)
	if err != nil {
		return 0, err
	}
	return max(left, right), nil
}

// targetDescription is a logical rule's attribute block as the reference
// reads it. Repeated attributes keep the last value there; names records the
// original order so this engine can still reject repetition once selected.
type targetDescription struct {
	target                        *uint32
	engine, size, entry, sections *numberRange
	iconGroup, handlerType        bool
	handler                       FileType
	iconGroups                    [2]string // IconGroup1 and IconGroup2; "" is any
	container                     *FileType
	intermediates                 []FileType
	names                         []string
	dropped                       bool // attributes after the tenth, which the reference ignores
}

// parseTargetDescription reproduces the reference attribute pass, which runs
// before the Engine window: a token without ':' or a misplaced or malformed
// range is an error; an unknown attribute name or file type skips the rule at
// once; attributes other than Target and Engine require Engine with a
// minimum of at least 51.
func parseTargetDescription(s string) (targetDescription, error) {
	var t targetDescription
	tokens, dropped := cStrtokenizeDropped(s, ',', 10)
	t.dropped = dropped
	extended := false
	for i, token := range tokens {
		name, value, ok := strings.Cut(token, ":")
		if !ok {
			return t, errors.New("invalid or duplicate logical target descriptor")
		}
		switch name {
		case "Target":
			if !isNumber(value) {
				return t, errors.New("invalid logical signature target")
			}
			v := uint32(cAtoi(value))
			t.target = &v
		case "Engine", "FileSize", "EntryPoint", "NumberOfSections":
			if name == "Engine" && i != 0 {
				return t, errors.New("engine descriptor must be first")
			}
			r, err := parseRange(value)
			if err != nil {
				return t, err
			}
			switch name {
			case "Engine":
				t.engine = r
			case "FileSize":
				t.size = r
			case "EntryPoint":
				t.entry = r
			default:
				t.sections = r
			}
		case "IconGroup1", "IconGroup2":
			t.iconGroup = true
			t.iconGroups[name[9]-'1'] = value
		case "HandlerType":
			typ, ok := ParseFileType(value)
			if !ok {
				return t, skip(SkipUnknownAttribute)
			}
			t.handlerType = true
			t.handler = typ
		case "Container":
			typ, ok := ParseFileType(value)
			if !ok {
				return t, skip(SkipUnknownAttribute)
			}
			t.container = &typ
		case "Intermediates":
			t.intermediates = t.intermediates[:0]
			for _, v := range cStrtokenize(value, '>', 16) {
				typ, ok := ParseFileType(v)
				if !ok {
					return t, skip(SkipUnknownAttribute)
				}
				t.intermediates = append(t.intermediates, typ)
			}
		default:
			return t, skip(SkipUnknownAttribute)
		}
		if name != "Target" && name != "Engine" {
			extended = true
		}
		t.names = append(t.names, name)
	}
	if extended && (t.engine == nil || t.engine.min < 51) {
		return t, errors.New("logical signature attributes beyond Target require Engine with a minimum of at least 51")
	}
	return t, nil
}

// selectUnsupported reports a rule of a format this engine does not load.
func (d *Database) selectUnsupported(ext, line, entry string) error {
	return fmt.Errorf("unsupported signature database format %q", ext)
}
