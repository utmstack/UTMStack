package jsnorm

import "bytes"

const (
	maxPackerOutput = 20 << 20
	maxPackerCalls  = 128
)

type packerInvocation struct {
	start, end, resume  int
	payload, dictionary []byte
	base                int64
}

// packerCalls follows the native named-call walk. Each decoded payload is
// concatenated without a separator, and the entire first-to-last argument
// span is replaced. The walk skips one token after each invocation; adjacent
// calls therefore differ from calls separated by a semicolon.
func packerCalls(tokens []jsToken, function, limit int) (calls []packerInvocation, ok bool) {
	i := function + 1
	name := ""
	if i < len(tokens) && tokens[i].kind == jsIdent {
		name = string(tokens[i].text)
		i++
	}
	if i+12 >= len(tokens) || !isPunct(tokens[i], "(") {
		return
	}
	for i < len(tokens) && !isPunct(tokens[i], ")") {
		i++
	}
	if i+1 >= len(tokens) || !isPunct(tokens[i+1], "{") {
		return
	}
	braces := 1
	i += 2
	for i < len(tokens) && braces > 0 {
		if isPunct(tokens[i], "{") {
			braces++
		} else if isPunct(tokens[i], "}") {
			braces--
		}
		i++
	}
	if braces != 0 {
		return
	}
	if name == "" {
		for i < len(tokens) && !isPunct(tokens[i], "(") {
			i++
		}
		if i == len(tokens) {
			return nil, true
		}
		call, valid, complete := packerArguments(tokens, i)
		if !complete {
			return nil, true
		}
		if !valid || limit < 1 {
			return nil, false
		}
		return []packerInvocation{call}, true
	}
	for i+1 < len(tokens) {
		if tokens[i].kind == jsIdent && !tokens[i].member && string(tokens[i].text) == name && isPunct(tokens[i+1], "(") {
			call, valid, complete := packerArguments(tokens, i+1)
			if !complete {
				return calls, true
			}
			if !valid || len(calls) >= limit {
				return nil, false
			}
			calls = append(calls, call)
			i = call.resume
		} else {
			i++
		}
	}
	return calls, true
}

func packerArguments(tokens []jsToken, call int) (invocation packerInvocation, ok, complete bool) {
	var args [6]int
	i := call + 1
	for a := range args {
		if i >= len(tokens) {
			return
		}
		args[a] = i
		if a == 5 {
			break
		}
		for i < len(tokens) && !isPunct(tokens[i], ",") {
			i++
		}
		if i >= len(tokens) {
			return
		}
		i++
	}
	// The reference leaves the original tokens when its walk cannot collect
	// all six argument starts. A collected but unproved shape stays unknown.
	complete = true
	if args[5]+1 >= len(tokens) ||
		tokens[args[0]].kind != jsString || tokens[args[1]].kind != jsNumber ||
		tokens[args[3]].kind != jsString {
		return
	}
	last := tokens[args[5]]
	switch {
	case isPunct(last, "{") && isPunct(tokens[args[5]+1], "}"):
	case isPunct(last, "[") && isPunct(tokens[args[5]+1], "]"):
	case (last.kind == jsNumber || last.kind == jsIdent || last.kind == jsKeyword || last.kind == jsString) &&
		isPunct(tokens[args[5]+1], ")"):
	default:
		return
	}
	resume := args[5] + 1
	for resume < len(tokens) && !isPunct(tokens[resume], ")") {
		resume++
	}
	resume += 2 // native call walk increments once more after consuming ')'
	return packerInvocation{start: args[0], end: args[5] + 2, resume: resume,
		payload: tokens[args[0]].text, base: tokens[args[1]].num,
		dictionary: tokens[args[3]].text}, true, true
}

func packerDigit(c byte) (uint32, bool) {
	switch {
	case c >= '0' && c <= '9':
		return uint32(c - '0'), true
	case c >= 'a' && c <= 'z':
		return uint32(c-'a') + 10, true
	case c >= 'A' && c <= 'Z':
		return uint32(c-'A') + 36, true
	}
	return 0, false
}

func unpackDeanEdwards(p []byte, base int64, dictionary []byte) ([]byte, bool) {
	if len(p) > maxPackerOutput || len(dictionary) > maxPackerOutput {
		return nil, false
	}
	if bytes.Count(dictionary, []byte{'|'}) > 65535 {
		return nil, false
	}
	words := bytes.Split(dictionary, []byte{'|'})
	out := make([]byte, 0, len(p))
	for i := 0; i < len(p); {
		if p[i] == '\\' && i+1 < len(p) && (p[i+1] == '\'' || p[i+1] == '"') {
			i++
		}
		if _, digit := packerDigit(p[i]); !digit {
			out = append(out, p[i])
			i++
			continue
		}
		start := i
		var value uint32
		for i < len(p) {
			digit, ok := packerDigit(p[i])
			if !ok {
				break
			}
			value = value*uint32(base) + digit
			i++
		}
		if int(value) < len(words) && len(words[value]) > 0 {
			out = append(out, words[value]...)
		} else {
			out = append(out, p[start:i]...)
		}
		if len(out) > maxPackerOutput {
			return nil, false
		}
	}
	return out, true
}
