package jsnorm

import (
	"bytes"
	"math"
	"strconv"
)

// The javascript view: each <script> block of an HTML object tokenized and
// written back in a canonical form, as the reference writes it. Its rules
// were learned from the reference's output on probe pages and random token
// soups (docs/2026-09-27-html-oracle/ in the workspace): the tokenizer
// carries its state across the block's segments; declared names are
// renamed n000, n001, ... per block (the bracket counts never go below
// zero); "a"+"b" read one after the other is joined; unescape("...") of the
// block's own text is folded when the next token arrives (that token is
// consumed); the output is lower-cased with one space only between two
// alphanumeric tokens.
//
// Decoder calls are processed in one forward walk over the final token list.
// The lexer keeps its state across strings. eval replaces its call, dF appends
// its output, and a recognized packer replaces its measured argument range.
// Function-end markers affect the decoder walk and output scope without adding
// output bytes. Decoded declarations are processed as the lexer emits them.
// Unproved argument shapes and exhausted component budgets remain incomplete.

type jsKind uint8

const (
	jsIdent jsKind = iota
	jsKeyword
	jsFunction
	jsNumber
	jsFloat
	jsString
	jsPunct
	jsScopeEnd // native dummy function-end token, used by the ordered decoder walk
)

type jsToken struct {
	kind  jsKind
	text  []byte // identifier, keyword, operator or string content
	num   int64
	float float64
	scope *jsScope
	// member marks a name right after '.': it is never renamed.
	member bool
	// folded marks a string made by unescape().
	folded      bool
	functionEnd bool
}

type jsFSM uint8

const (
	fsmBase jsFSM = iota
	fsmInsideVar
	fsmInsideInitializer
	fsmWaitFunctionName
	fsmWaitParameterList
	fsmInsideFunctionDecl
)

type jsScope struct {
	parent   *jsScope
	names    map[string]int
	fsm      jsFSM
	last     jsKind
	lastText string
	brackets int
	braces   int
}

var jsKeywords = func() map[string]bool {
	m := map[string]bool{}
	for _, k := range []string{"abstract", "boolean", "break", "byte", "case", "catch", "char", "class", "const",
		"continue", "debugger", "default", "delete", "do", "double", "else", "enum", "export", "extends", "false",
		"final", "finally", "float", "for", "function", "goto", "if", "implements", "import", "in", "instanceof",
		"int", "interface", "long", "native", "new", "null", "package", "private", "protected", "public", "return",
		"short", "static", "super", "switch", "synchronized", "this", "throw", "throws", "transient", "true", "try",
		"typeof", "var", "void", "volatile", "while", "with"} {
		m[k] = true
	}
	return m
}()

var jsOperators = func() map[string]bool {
	m := map[string]bool{}
	for _, o := range []string{">", ">>", ">>>", ">=", ">>=", ">>>=", "=", "==", "===", "<=", "<<=", "<", "<<", "/=",
		"|=", "*=", "~", "-=", "/", "+=", "%=", "|", "||", "*", "&=", "-", "--", "!=", "!==", "+", "++", "%", "^=",
		"?", "&", "&&", ":", "!", "^"} {
		m[o] = true
	}
	return m
}()

type jsLexState uint8

const (
	lexInitial jsLexState = iota
	lexBlockComment
	lexLineComment
	lexNumber
	lexDString
	lexSString
	lexIdent
)

// jsLexer reads script text. The block's own text and, after it, the text
// of decoded eval() strings are read by the same reader, which keeps its
// state between them (an unfinished name, an open /* comment).
type jsLexer struct {
	state   jsLexState
	partial []byte
	isFloat bool
	// lastString: the previous token read was a string; heldPlus: a '+'
	// after it waits to see whether a string follows ("a"+"b" is joined
	// only when both strings were read one after the other).
	lastString bool
	heldPlus   bool
}

// jsNormalizer normalizes one script block.
type jsNormalizer struct {
	lex    jsLexer
	tokens []jsToken
	global *jsScope
	scope  *jsScope
	uniq   int
	// parens counts the block's own parentheses (never below zero); the
	// missing ')' are added at the end.
	parens int
	// Decoded buffers use the shared lexer and their own local parse cursor.
	decoding           bool
	decoderIncomplete  bool
	packerIncomplete   bool
	packerDecodedBytes int
	packerDecodedCalls int
	decoderScope       *jsScope
	decodeParseScope   *jsScope
	decoderBytes       int
	decoderCalls       int
	decoderTokenWork   int
}

func newJSNormalizer() *jsNormalizer {
	g := &jsScope{names: map[string]int{}}
	return &jsNormalizer{global: g, scope: g}
}

// feed tokenizes one segment of the block's own text.
func (j *jsNormalizer) feed(seg []byte) { j.lexFeed(&j.lex, seg) }

func isPunct(t jsToken, p string) bool { return t.kind == jsPunct && string(t.text) == p }

// deliver takes a token from the reader and joins "a"+"b".
func (j *jsNormalizer) deliver(l *jsLexer, t jsToken) {
	if !j.decoding && t.kind == jsPunct {
		switch string(t.text) {
		case "(":
			j.parens++
		case ")":
			if j.parens > 0 {
				j.parens--
			}
		}
	}
	if l.heldPlus {
		l.heldPlus = false
		if n := len(j.tokens); t.kind == jsString && n > 0 && j.tokens[n-1].kind == jsString && !j.tokens[n-1].folded {
			last := &j.tokens[n-1]
			last.text = append(last.text, t.text...)
			l.lastString = true
			return
		}
		j.take(jsToken{kind: jsPunct, text: []byte("+")})
	}
	prevString := l.lastString
	l.lastString = t.kind == jsString
	if isPunct(t, "+") && prevString && j.lexLastWasString() {
		l.heldPlus = true
		return
	}
	j.take(t)
}

// lexLastWasString reports whether the token before the current one was a
// string read straight from the text.
func (j *jsNormalizer) lexLastWasString() bool {
	n := len(j.tokens)
	return n > 0 && j.tokens[n-1].kind == jsString && !j.tokens[n-1].folded
}

// take applies declaration rules and appends a token. In the block's own
// text, unescape("...") consumes the next token when folded. Decoded buffers
// retain explicit function-end markers for the ordered decoder walk.
func (j *jsNormalizer) take(t jsToken) {
	if j.decoding {
		j.declare(&j.decodeParseScope, &t)
		end := t.functionEnd
		t.functionEnd = false
		j.tokens = append(j.tokens, t)
		if t.kind == jsFunction || end {
			j.decoderScope = j.decodeParseScope
		}
		if end {
			j.tokens = append(j.tokens, jsToken{kind: jsScopeEnd})
		}
		return
	}
	if n := len(j.tokens); n >= 3 {
		u := j.tokens[n-3]
		if u.kind == jsIdent && !u.member && string(u.text) == "unescape" && isPunct(j.tokens[n-2], "(") &&
			j.tokens[n-1].kind == jsString && !j.tokens[n-1].folded {
			j.tokens = append(j.tokens[:n-3], jsToken{kind: jsString, text: jsUnescape(j.tokens[n-1].text), scope: u.scope, folded: true})
			return
		}
	}
	j.emit(t)
}

func isJSIdentByte(c byte) bool {
	return c == '$' || c == '_' || c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z'
}

// lexFeed tokenizes text with the reader l.
func (j *jsNormalizer) lexFeed(l *jsLexer, seg []byte) {
	for i := 0; i < len(seg); {
		c := seg[i]
		switch l.state {
		case lexBlockComment:
			if c == '*' && i+1 < len(seg) && seg[i+1] == '/' {
				l.state = lexInitial
				i += 2
				continue
			}
			i++
		case lexLineComment:
			for i < len(seg) && seg[i] != '\n' && seg[i] != ' ' {
				i++
			}
			// Processing a comment body clears its state even at the
			// buffer end; a bare // at the boundary has not entered here.
			l.state = lexInitial
		case lexDString, lexSString:
			q := byte('"')
			if l.state == lexSString {
				q = '\''
			}
			start := i
			for i < len(seg) && !(seg[i] == q && (i == 0 || seg[i-1] != '\\')) {
				i++
			}
			// The native reader normalizes each input-buffer piece before
			// retaining it. Escape sequences cannot join across buffers.
			l.partial = append(l.partial, decodeJSString(seg[start:i])...)
			if i < len(seg) {
				l.state = lexInitial
				j.deliver(l, jsToken{kind: jsString, text: l.partial})
				l.partial = nil
				i++
			}
		case lexNumber:
			switch {
			case c >= '0' && c <= '9':
				l.partial = append(l.partial, c)
			case c == '.' && !l.isFloat:
				l.isFloat = true
				l.partial = append(l.partial, c)
			case (c == 'e' || c == 'E') && l.isFloat && l.partial[len(l.partial)-1]|0x20 != 'e':
				// Any number of exponents, each after a character other
				// than e: "1.e5e5" is one number (strtod reads 1.e5).
				l.partial = append(l.partial, c)
				if i+1 < len(seg) && (seg[i+1] == '+' || seg[i+1] == '-') {
					i++
					l.partial = append(l.partial, seg[i])
				} else if i+1 == len(seg) {
					// A dangling exponent emits the number, then the native
					// reader revisits this final e as an identifier.
					j.emitNumber(l)
					continue
				}
			default:
				j.emitNumber(l)
				continue
			}
			i++
		case lexIdent:
			switch {
			case isJSIdentByte(c):
				l.partial = append(l.partial, c)
				i++
			case c == '\\':
				if i+1 < len(seg) && seg[i+1] == 'u' {
					l.partial = append(l.partial, '\\')
					i += 2
					continue
				}
				j.emitIdent(l)
				i++
				if i+1 == len(seg) {
					// Native parseId consumes an invalid lookahead byte
					// when that byte ends the current input buffer.
					i++
				}
			default:
				j.emitIdent(l)
			}
		default:
			i = j.initial(l, seg, i)
		}
	}
	if l.state == lexNumber {
		l.isFloat = false // does not carry into the next segment
	}
}

// initial handles a byte between tokens and returns the next position.
func (j *jsNormalizer) initial(l *jsLexer, seg []byte, i int) int {
	c := seg[i]
	switch {
	case c == '\t' || c == '\n' || c == 0x0b || c == '\r' || c == ' ':
		return i + 1
	case c <= 0x08 || c == 0x0c || c >= 0x0e && c <= 0x1f || c == '#' || c == '@' || c == '`' || c >= 0x7f:
		return i + 1
	case c == '/' && i+1 < len(seg) && seg[i+1] == '*':
		l.state = lexBlockComment
		return i + 2
	case c == '/' && i+1 < len(seg) && seg[i+1] == '/':
		l.state = lexLineComment
		return i + 2
	case bytes.IndexByte([]byte("/!%&*+-:<=>?^|~"), c) >= 0:
		n := 1
		for k := min(4, len(seg)-i); k > 1; k-- {
			if jsOperators[string(seg[i:i+k])] {
				n = k
				break
			}
		}
		j.deliver(l, jsToken{kind: jsPunct, text: append([]byte(nil), seg[i:i+n]...)})
		return i + n
	case c == '"':
		l.state, l.partial = lexDString, nil
		return i + 1
	case c == '\'':
		l.state, l.partial = lexSString, nil
		return i + 1
	case c >= '0' && c <= '9':
		l.state, l.partial, l.isFloat = lexNumber, []byte{c}, false
		return i + 1
	case isJSIdentByte(c) || c == '\\':
		l.state, l.partial = lexIdent, nil
		return i
	}
	j.deliver(l, jsToken{kind: jsPunct, text: []byte{c}})
	return i + 1
}

func (j *jsNormalizer) emitNumber(l *jsLexer) {
	text := string(l.partial)
	l.state, l.partial = lexInitial, nil
	if l.isFloat {
		f, _ := strconv.ParseFloat(trimFloat(text), 64)
		j.deliver(l, jsToken{kind: jsFloat, float: f})
		return
	}
	j.deliver(l, jsToken{kind: jsNumber, num: cAtol(text)})
}

// trimFloat keeps the longest prefix strtod reads.
func trimFloat(s string) string {
	for len(s) > 0 {
		if _, err := strconv.ParseFloat(s, 64); err == nil || isRangeErr(err) {
			return s
		}
		s = s[:len(s)-1]
	}
	return "0"
}

func isRangeErr(err error) bool {
	ne, ok := err.(*strconv.NumError)
	return ok && ne.Err == strconv.ErrRange
}

// cAtol converts a number as the reference does: the C library's 64-bit
// conversion of its leading digits (a number read as an integer may hold a
// '.' or an exponent), which stops at its largest value, kept to its low
// 32 bits as a signed number (4294967296 is 0, 2147483648 is -2147483648).
func cAtol(s string) int64 {
	n := 0
	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		n++
	}
	v, err := strconv.ParseInt(s[:n], 10, 64)
	if n > 0 && err != nil {
		v = math.MaxInt64
	}
	return int64(int32(v))
}

func (j *jsNormalizer) emitIdent(l *jsLexer) {
	// A name can be empty (a backslash not followed by u): the reference
	// still emits it, which affects the spacing of the output.
	text := l.partial
	l.state, l.partial = lexInitial, nil
	switch {
	case string(text) == "function":
		j.deliver(l, jsToken{kind: jsFunction, text: text})
	case jsKeywords[string(text)]:
		j.deliver(l, jsToken{kind: jsKeyword, text: text})
	default:
		j.deliver(l, jsToken{kind: jsIdent, text: text})
	}
}

// decodeJSString normalizes one native string-buffer piece.
func decodeJSString(raw []byte) []byte {
	var out []byte
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c != '\\' || i+1 >= len(raw) {
			out = append(out, c)
			continue
		}
		i++
		switch e := raw[i]; e {
		case '0':
			out = append(out, 1)
		case 'b':
			out = append(out, 8)
		case 't':
			out = append(out, 9)
		case 'n':
			out = append(out, 10)
		case 'v':
			out = append(out, 11)
		case 'f':
			out = append(out, 12)
		case 'r':
			out = append(out, 13)
		case 'x':
			if i+2 <= len(raw)-1 {
				out = append(out, byte(escapeHex(raw[i+1:i+3])))
				i += 2
			} else {
				out = append(out, '\\')
				i += 2
			}
		case 'u':
			if i+4 <= len(raw)-1 {
				v := escapeHex(raw[i+1 : i+5])
				if v == 0 {
					v = 1
				}
				out = appendUTF8(out, v)
				i += 4
			} else {
				out = append(out, '\\')
			}
		default:
			out = append(out, e)
		}
	}
	for k, c := range out {
		if c == 0 {
			out[k] = 1
		}
	}
	return out
}

// escapeHex is the value of an escape's hex digits as the reference reads
// them: a digit that is not hex counts as 0, except the last one, which makes
// every bit of the value 1.
func escapeHex(digits []byte) int {
	v := 0
	for _, c := range digits {
		v = v<<4 | max(0, hexDigit(c))
	}
	if hexDigit(digits[len(digits)-1]) < 0 {
		v = 1<<(4*len(digits)) - 1
	}
	return v
}

// appendUTF8 writes a code point below 0x10000 in UTF-8 form, surrogates
// included.
func appendUTF8(out []byte, v int) []byte {
	switch {
	case v < 0x80:
		return append(out, byte(v))
	case v < 0x800:
		return append(out, byte(0xc0|v>>6), byte(0x80|v&0x3f))
	}
	return append(out, byte(0xe0|v>>12), byte(0x80|v>>6&0x3f), byte(0x80|v&0x3f))
}

// emit applies the scope rules to a token of the block's own text and
// appends it.
func (j *jsNormalizer) emit(t jsToken) {
	j.declare(&j.scope, &t)
	j.tokens = append(j.tokens, t)
}

// declare advances the current scope's declaration state. Output resolves
// names through function transitions in the final token sequence.
func (j *jsNormalizer) declare(cur **jsScope, t *jsToken) {
	s := *cur
	t.scope = s
	names := s.names
	switch t.kind {
	case jsKeyword:
		if string(t.text) == "var" {
			s.fsm = fsmInsideVar
		}
	case jsFunction:
		n := &jsScope{parent: s, names: map[string]int{}, fsm: fsmWaitFunctionName, last: jsFunction}
		*cur = n
		t.scope = n
	case jsIdent:
		name := string(t.text)
		if s.last == jsPunct && s.lastText == "." {
			t.member = true
			break
		}
		switch s.fsm {
		case fsmInsideVar, fsmInsideFunctionDecl:
			names[name] = j.uniq
			j.uniq++
			s.fsm, s.brackets = fsmInsideInitializer, 0
		case fsmWaitFunctionName:
			names[name] = j.uniq
			j.uniq++
			s.fsm = fsmWaitParameterList
		}
	case jsPunct:
		switch string(t.text) {
		case "(":
			if s.fsm == fsmWaitFunctionName || s.fsm == fsmWaitParameterList {
				s.fsm = fsmInsideFunctionDecl
			}
		case ")":
			if s.fsm == fsmWaitParameterList {
				s.fsm = fsmBase
			}
		case "{":
			if s.fsm == fsmWaitFunctionName || s.fsm == fsmWaitParameterList || s.fsm == fsmInsideFunctionDecl {
				s.fsm = fsmBase
			}
			s.braces++
		case "}":
			// The bracket counts never go below zero, so a stray '}' in a
			// function also ends it.
			if s.braces > 0 {
				s.braces--
			}
			if s.braces == 0 && s.parent != nil {
				t.functionEnd = true
				*cur = s.parent
			}
		case "[":
			s.brackets++
		case "]":
			if s.brackets > 0 {
				s.brackets--
			}
		case ",":
			if s.fsm == fsmInsideInitializer && s.brackets == 0 && s.braces == 0 {
				s.fsm = fsmInsideVar
			}
		case ";":
			if s.brackets == 0 && s.braces == 0 {
				s.fsm = fsmBase
			}
		}
	}
	s.last, s.lastText = t.kind, ""
	if t.kind == jsPunct {
		s.lastText = string(t.text)
	}
}

// finish closes the block's unfinished strings and parentheses, then runs
// the ordered decoder walk. Other unfinished lexical state is retained.
func (j *jsNormalizer) finish() {
	l := &j.lex
	switch l.state {
	case lexDString, lexSString:
		j.deliver(l, jsToken{kind: jsString, text: l.partial})
		l.state, l.partial = lexInitial, nil
	case lexLineComment:
		l.state = lexInitial
	}
	for j.parens > 0 {
		j.deliver(l, jsToken{kind: jsPunct, text: []byte(")")})
	}
	if l.heldPlus {
		l.heldPlus = false
		j.take(jsToken{kind: jsPunct, text: []byte("+")})
	}
	if usesPackerDecoder(j.tokens) || hasJSCall(j.tokens, "dF") || hasJSCall(j.tokens, "eval") {
		j.decodeOrdered()
		return
	}
}

// jsUnescape decodes %uHHHH (to UTF-8) and %HH; a NUL becomes 0x01.
func jsUnescape(s []byte) []byte {
	var out []byte
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '%' && i+5 < len(s)+0 && i+5 <= len(s)-1 && s[i+1] == 'u' && allHex(s[i+2:i+6]):
			v := 0
			for _, c := range s[i+2 : i+6] {
				v = v<<4 | hexDigit(c)
			}
			if v == 0 {
				v = 1
			}
			out = appendUTF8(out, v)
			i += 5
		case s[i] == '%' && i+2 <= len(s)-1 && allHex(s[i+1:i+3]):
			v := byte(hexDigit(s[i+1])<<4 | hexDigit(s[i+2]))
			if v == 0 {
				v = 1
			}
			out = append(out, v)
			i += 2
		default:
			out = append(out, s[i])
		}
	}
	return out
}

func allHex(b []byte) bool {
	for _, c := range b {
		if hexDigit(c) < 0 {
			return false
		}
	}
	return true
}

// output writes the block: <script>, the tokens, then </script> unless the
// tokens already end with it.
func (j *jsNormalizer) output() []byte {
	out := []byte("<script>")
	var lastByte byte
	scope := j.global
	for _, t := range j.tokens {
		var text []byte
		var first, last byte
		switch t.kind {
		case jsScopeEnd:
			if scope.parent != nil {
				scope = scope.parent
			}
			continue
		case jsString:
			text = append(append([]byte{'"'}, t.text...), '"')
			first, last = '"', '"'
		case jsNumber:
			text = []byte(strconv.FormatInt(t.num, 10))
			first, last = '0', '0'
		case jsFloat:
			text = []byte(cFormatG(t.float))
			first, last = '0', '0'
		case jsIdent:
			text = j.nameInScope(t, scope)
			first, last = 'a', 'a'
			if t.member {
				// A member name is spaced by its own characters; an empty
				// one writes nothing and leaves the spacing as it was.
				if len(text) == 0 {
					continue
				}
				first, last = text[0], text[len(text)-1]
			}
		case jsFunction:
			if t.scope != nil {
				scope = t.scope
			}
			text = []byte("function")
			first, last = 'a', 'a'
		default:
			text = t.text
			first, last = text[0], text[len(text)-1]
		}
		if isAlnum(lastByte) && isAlnum(first) {
			out = append(out, ' ')
		}
		for _, c := range text {
			if isHTMLSpace(c) {
				c = ' '
			}
			out = append(out, lowerByte(c))
		}
		lastByte = last
		if t.functionEnd && scope.parent != nil {
			scope = scope.parent
		}
	}
	if !bytes.HasSuffix(out, []byte("</script>")) {
		out = append(out, "</script>"...)
	}
	return out
}

func isAlnum(c byte) bool { return c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z' }

// nameInScope resolves a declared identifier through the output scope chain.
func (j *jsNormalizer) nameInScope(t jsToken, scope *jsScope) []byte {
	if t.member {
		return t.text
	}
	for s := scope; s != nil; s = s.parent {
		if id, ok := s.names[string(t.text)]; ok && id >= 0 {
			return []byte("n" + pad3(id))
		}
	}
	return t.text
}

func pad3(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 3 {
		s = "0" + s
	}
	return s
}

// cFormatG prints a float as C's %g does.
func cFormatG(f float64) string {
	if math.IsInf(f, 0) {
		return "inf"
	}
	s := strconv.FormatFloat(f, 'g', 6, 64)
	// Go writes e+06 as e+06 too, but drops a zero exponent digit: C always
	// writes at least two exponent digits.
	if k := bytes.IndexAny([]byte(s), "e"); k >= 0 {
		exp := s[k+2:]
		if len(exp) < 2 {
			s = s[:k+2] + "0" + exp
		}
	}
	return s
}

// normalizeJS returns the javascript view of an object's script blocks: each
// block's output, separated by a newline. unsupported reports a block this
// engine does not reproduce: an unhandled packer shape, a decoder budget or
// argument violation.
func NormalizeBlocks(scripts [][][]byte) (out []byte, unsupported bool) {
	for _, segs := range scripts {
		j := newJSNormalizer()
		for _, s := range segs {
			j.feed(s)
		}
		j.finish()
		if j.decoderIncomplete || j.packerIncomplete {
			unsupported = true
		}
		if len(out) > 0 {
			out = append(out, '\n')
		}
		out = append(out, j.output()...)
	}
	return out, unsupported
}

func hasJSCall(ts []jsToken, name string) bool {
	for i, t := range ts {
		if t.kind == jsIdent && !t.member && string(t.text) == name && i+2 < len(ts) &&
			isPunct(ts[i+1], "(") && ts[i+2].kind == jsString {
			return true
		}
	}
	return false
}

func usesPackerDecoder(ts []jsToken) bool {
	for i, t := range ts {
		if t.kind == jsFunction && packerParams(ts[i+1:]) {
			return true
		}
	}
	return false
}

// packerParams reports a parameter list (p,a,c,k,e,r) or (p,a,c,k,e,d),
// optionally after a function name.
func packerParams(ts []jsToken) bool {
	if len(ts) > 0 && ts[0].kind == jsIdent {
		ts = ts[1:]
	}
	want := []string{"(", "p", ",", "a", ",", "c", ",", "k", ",", "e", ","}
	if len(ts) < len(want)+2 {
		return false
	}
	for i, w := range want {
		if string(ts[i].text) != w {
			return false
		}
	}
	last := string(ts[len(want)].text)
	return (last == "r" || last == "d") && string(ts[len(want)+1].text) == ")"
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}

func isHTMLSpace(c byte) bool { return c == ' ' || c >= 9 && c <= 13 }

func lowerByte(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
