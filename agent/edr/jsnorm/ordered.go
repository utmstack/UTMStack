package jsnorm

import "bytes"

const maxDecoderTokenWork = 8 << 20

// decodeOrdered follows the reference's one forward token walk. Function-end
// markers matter to this walk even though they produce no output bytes. The
// same walk handles nested decoders introduced by decoded strings.
func (j *jsNormalizer) decodeOrdered() {
	j.decoderScope = j.scope
	original := j.tokens
	j.tokens = nil
	for _, t := range original {
		end := t.functionEnd
		t.functionEnd = false
		j.tokens = append(j.tokens, t)
		if end {
			j.tokens = append(j.tokens, jsToken{kind: jsScopeEnd})
		}
	}
	functions := 0
	for i := 0; i < len(j.tokens); i++ {
		if !j.chargeDecoderWork(1) {
			return
		}
		at := i
		t := j.tokens[i]
		switch {
		case (t.kind == jsFunction || t.kind == jsScopeEnd) && i+13 < len(j.tokens):
			i++
			if j.tokens[i].kind == jsIdent {
				i++
			}
			if t.kind == jsFunction && packerParams(j.tokens[at+1:]) {
				functions++
				if functions > maxPackerCalls {
					j.packerIncomplete = true
					return
				}
				if !j.chargeDecoderWork(2 * (len(j.tokens) - at)) {
					return
				}
				calls, ok := packerCalls(j.tokens, at, maxPackerCalls-j.packerDecodedCalls)
				if !ok {
					j.packerIncomplete = true
					break
				}
				if len(calls) == 0 {
					break
				}
				var decoded []byte
				for _, call := range calls {
					part, valid := unpackDeanEdwards(call.payload, call.base, call.dictionary)
					if !valid || len(part) > maxPackerOutput-j.packerDecodedBytes-len(decoded) {
						ok = false
						break
					}
					decoded = append(decoded, part...)
				}
				if !ok {
					j.packerIncomplete = true
					break
				}
				j.packerDecodedCalls += len(calls)
				j.packerDecodedBytes += len(decoded)
				if !j.replaceDecoded(calls[0].start, calls[len(calls)-1].end, decoded, false) {
					return
				}
			}
		case t.kind == jsIdent && !t.member && i+3 < len(j.tokens) && isPunct(j.tokens[i+1], "(") && j.tokens[i+2].kind == jsString:
			name := string(t.text)
			if name != "eval" && name != "dF" {
				break
			}
			if !isPunct(j.tokens[i+3], ")") {
				if name == "dF" {
					j.decoderIncomplete = true
				}
				break
			}
			text := cstring(j.tokens[i+2].text)
			appendOutput := name == "dF"
			if appendOutput {
				if len(text) == 0 {
					break
				}
				if len(text) > maxPackerOutput-j.decoderBytes {
					j.decoderIncomplete = true
					return
				}
				key := text[len(text)-1] - '0'
				text = jsUnescape(text[:len(text)-1])
				for k := range text {
					text[k] -= key
				}
				text = cstring(jsUnescape(cstring(text)))
			}
			if !j.replaceDecoded(i, i+4, text, appendOutput) {
				return
			}
		}
		if i < len(j.tokens) {
			switch token := j.tokens[i]; token.kind {
			case jsFunction:
				if token.scope != nil {
					j.decoderScope = token.scope
				}
			case jsScopeEnd:
				if j.decoderScope.parent != nil {
					j.decoderScope = j.decoderScope.parent
				}
			}
		}
	}
}

// replaceDecoded reads through the same lexer, then either replaces an argument
// range or appends dF output. A single aggregate budget covers all decoder kinds.
func (j *jsNormalizer) replaceDecoded(start, end int, text []byte, appendOutput bool) bool {
	if j.decoderCalls >= 4096 || len(text) > maxPackerOutput-j.decoderBytes {
		j.decoderIncomplete = true
		return false
	}
	j.decoderCalls++
	j.decoderBytes += len(text)
	if end < len(j.tokens) && isPunct(j.tokens[end], ";") {
		end++
	}
	if !j.chargeDecoderWork(2 * len(j.tokens)) {
		return false
	}
	parent := j.tokens
	before := append([]jsToken(nil), parent[:start]...)
	after := append([]jsToken(nil), parent[end:]...)
	j.tokens = nil
	j.decodeParseScope = j.decoderScope
	j.decoding = true
	j.lexFeed(&j.lex, text)
	if j.lex.heldPlus {
		// Each decoded buffer has a fresh token list in the reference.
		// A trailing plus is already a token, not a deferred fold into
		// a later buffer's first string.
		j.lex.heldPlus = false
		j.take(jsToken{kind: jsPunct, text: []byte("+")})
	}
	j.decoding = false
	embedded := j.tokens
	j.tokens = append(before, after...)
	if appendOutput {
		j.tokens = append(j.tokens, embedded...)
	} else {
		j.tokens = append(before, embedded...)
		j.tokens = append(j.tokens, after...)
	}
	return true
}

func cstring(data []byte) []byte {
	if n := bytes.IndexByte(data, 0); n >= 0 {
		return data[:n]
	}
	return data
}

// Charge token traversal and range-copy work before a repeated scan or copy.
// Small decoder calls cannot turn a large token list into unbounded quadratic
// processing; a work-budget refusal remains incomplete.
func (j *jsNormalizer) chargeDecoderWork(n int) bool {
	if n < 0 || n > maxDecoderTokenWork-j.decoderTokenWork {
		j.decoderIncomplete = true
		return false
	}
	j.decoderTokenWork += n
	return true
}
