package signatures

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

// nativeACPattern retains the part boundaries that a general matching graph
// deliberately simplifies away. A short exact curly gap can expand to byte
// wildcards, while a ranged or large gap owns independent predecessor state.
// Variants (wide and ASCII) share one logical count but have separate history.
type nativeACPattern struct {
	variants [][]nativeACPart
	anchors  []patternAnchor
}

type nativeACSymbol struct {
	value, mask byte
	class       byte // B, L or W; zero for ordinary bytes/alternatives
	negative    bool
	alts        [][]nativeACSymbol
}

type nativeACChar struct {
	value, mask byte
	min, max    int
	set         bool
}

type nativeACPart struct {
	symbols              []nativeACSymbol
	literal              []byte
	key                  [3]byte
	keyAt, depth         int
	prefixMin, prefixMax int
	minWidth, maxWidth   int
	minGap, maxGap       int // maxGap == 0 is native's unbounded distance
	left, right          nativeACChar
	leftClass            byte
	rightClass           byte
	leftNeg, rightNeg    bool
	nocase, wide         bool
	modified             bool
	identity             [16]byte
}

func compileNativeAC(body string) (*nativeACPattern, error) {
	text, mods, _ := strings.Cut(body, "::")
	nocase, wide, ascii, word := false, false, false, false
	for _, m := range mods {
		switch m {
		case 'i':
			nocase = true
		case 'w':
			wide = true
		case 'a':
			ascii = true
		case 'f':
			word = true
		default:
			return nil, errors.New("native AC modifier")
		}
	}
	if word {
		text = "(W)" + strings.NewReplacer("[", "{", "]", "}").Replace(text) + "(W)"
	}
	forms := []struct {
		text string
		wide bool
	}{{text, false}}
	if wide {
		widened, err := widenNativeAC(text)
		if err != nil {
			return nil, err
		}
		forms = []struct {
			text string
			wide bool
		}{{widened, true}}
		if ascii {
			forms = append(forms, struct {
				text string
				wide bool
			}{text, false})
		}
	}
	out := &nativeACPattern{}
	for _, form := range forms {
		texts, gaps, err := splitNativeACParts(form.text)
		if err != nil {
			return nil, err
		}
		var parts []nativeACPart
		for i, text := range texts {
			part, err := compileNativeACPart(text, nocase, form.wide, mods != "")
			if err != nil {
				return nil, err
			}
			if i > 0 {
				part.minGap, part.maxGap = gaps[i-1][0], gaps[i-1][1]
			}
			parts = append(parts, part)
		}
		out.variants = append(out.variants, parts)
	}
	// A long static run in any required part is a necessary anchor. Keep
	// its distance from the full start as well: file-type recognition and
	// fallback graph consumers use the same index to clip their searches.
	for _, parts := range out.variants {
		var key []byte
		minPre, maxPre, beforeMin, beforeMax := 0, 0, 0, 0
		for i := range parts {
			part := &parts[i]
			if i > 0 {
				beforeMin += part.minGap
				if beforeMax < 0 || part.maxGap == 0 {
					beforeMax = -1
				} else {
					beforeMax += part.maxGap
				}
			}
			if len(part.literal) > len(key) {
				key = part.literal
				minPre, maxPre = beforeMin, beforeMax
			}
			run := 0
			keep := func(end int) {
				if end-run > len(key) {
					key = make([]byte, end-run)
					for j := run; j < end; j++ {
						key[j-run] = part.symbols[j].value
					}
					lo, hi := nativeACWidth(part.symbols[:run])
					minPre, maxPre = beforeMin+lo, -1
					if beforeMax >= 0 {
						maxPre = beforeMax + hi
					}
				}
			}
			for j, s := range part.symbols {
				if s.mask == 255 && s.class == 0 && len(s.alts) == 0 {
				} else {
					keep(j)
					run = j + 1
				}
			}
			keep(len(part.symbols))
			beforeMin += part.minWidth
			if beforeMax >= 0 {
				beforeMax += part.maxWidth
			}
		}
		if len(key) == 0 {
			key = append([]byte(nil), parts[0].key[:parts[0].depth]...)
		}
		if parts[0].nocase {
			key = asciiLower(key)
		}
		out.anchors = append(out.anchors, patternAnchor{literal: key, fold: parts[0].nocase, minPre: minPre, maxPre: maxPre})
	}
	return out, nil
}

func widenNativeAC(text string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(text); {
		switch text[i] {
		case '*', '|', ')', '!':
			out.WriteByte(text[i])
			i++
		case '(':
			out.WriteByte('(')
			i++
			if i+1 < len(text) && strings.ContainsRune("BLW", rune(text[i])) && text[i+1] == ')' {
				out.WriteByte(text[i])
				out.WriteByte(')')
				i += 2
			}
		case '{', '[':
			end := byte('}')
			if text[i] == '[' {
				end = ']'
			}
			j := strings.IndexByte(text[i+1:], end)
			if j < 0 {
				return "", errors.New("native AC wide range")
			}
			j += i + 1
			out.WriteByte('{')
			out.WriteString(text[i+1 : j])
			out.WriteByte('}')
			i = j + 1
		default:
			if i+1 >= len(text) {
				return "", errors.New("native AC wide byte")
			}
			out.WriteString(text[i : i+2])
			out.WriteString("00")
			i += 2
		}
	}
	return out.String(), nil
}

func splitNativeACParts(text string) ([]string, [][2]int, error) {
	// The reference expands the first curly item only while it is a short
	// exact positive gap. Once a range/large/zero gap is reached, every later
	// curly item is a part boundary, including later short exact gaps.
	for {
		i := strings.IndexByte(text, '{')
		if i < 0 {
			break
		}
		j := strings.IndexByte(text[i+1:], '}')
		if j < 0 {
			return nil, nil, errors.New("native AC range")
		}
		j += i + 1
		item := text[i+1 : j]
		n, err := strconv.Atoi(item)
		if err != nil || !isNumber(item) || n <= 0 || n >= 128 {
			break
		}
		text = text[:i] + strings.Repeat("??", n) + text[j+1:]
	}
	var parts []string
	var gaps [][2]int
	start, nest := 0, 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(':
			nest++
		case ')':
			nest--
		case '*', '{':
			if nest != 0 || i == start {
				return nil, nil, errors.New("native AC empty/nested part")
			}
			parts = append(parts, text[start:i])
			gap := [2]int{}
			if text[i] == '{' {
				j := strings.IndexByte(text[i+1:], '}')
				if j < 0 {
					return nil, nil, errors.New("native AC range")
				}
				j += i + 1
				lo, hi, err := parseGap(text[i+1 : j])
				if err != nil {
					return nil, nil, err
				}
				if hi < 0 {
					hi = 0
				}
				gap = [2]int{lo, hi}
				i = j
			}
			gaps = append(gaps, gap)
			start = i + 1
		}
	}
	if start == len(text) || nest != 0 {
		return nil, nil, errors.New("native AC incomplete part")
	}
	return append(parts, text[start:]), gaps, nil
}

func parseNativeACSymbols(text string, depth int) ([]nativeACSymbol, error) {
	if depth > 8 {
		return nil, errors.New("native AC nesting")
	}
	var out []nativeACSymbol
	for i := 0; i < len(text); {
		negative := text[i] == '!'
		if negative {
			i++
		}
		if i < len(text) && text[i] == '(' {
			start, nest := i+1, 1
			i++
			for i < len(text) && nest != 0 {
				if text[i] == '(' {
					nest++
				}
				if text[i] == ')' {
					nest--
				}
				i++
			}
			if nest != 0 {
				return nil, errors.New("native AC alternative")
			}
			inner := text[start : i-1]
			s := nativeACSymbol{negative: negative}
			if len(inner) == 1 && strings.ContainsRune("BLW", rune(inner[0])) {
				s.class = inner[0]
			} else {
				for _, branch := range splitAlternatives(inner) {
					syms, err := parseNativeACSymbols(branch, depth+1)
					if err != nil {
						return nil, err
					}
					s.alts = append(s.alts, syms)
				}
				if negative {
					width := -1
					for _, alt := range s.alts {
						lo, hi := nativeACWidth(alt)
						if lo != hi || width >= 0 && width != lo {
							return nil, errors.New("native AC negative width")
						}
						for _, x := range alt {
							if x.mask != 255 || len(x.alts) != 0 || x.class != 0 {
								return nil, errors.New("native AC negative byte")
							}
						}
						width = lo
					}
				}
			}
			out = append(out, s)
			continue
		}
		if negative || i+1 >= len(text) {
			return nil, errors.New("native AC byte")
		}
		s := nativeACSymbol{}
		for j := 0; j < 2; j++ {
			c := text[i+j]
			if c == '?' {
				continue
			}
			n := hexNibble(c)
			if n < 0 {
				return nil, errors.New("native AC hexadecimal byte")
			}
			shift := uint(4 * (1 - j))
			s.value |= byte(n) << shift
			s.mask |= 15 << shift
		}
		out = append(out, s)
		i += 2
	}
	return out, nil
}

func nativeACWidth(symbols []nativeACSymbol) (int, int) {
	lo, hi := 0, 0
	for _, s := range symbols {
		minw, maxw := 1, 1
		if len(s.alts) != 0 {
			minw, maxw = int(^uint(0)>>1), 0
			for _, alt := range s.alts {
				a, b := nativeACWidth(alt)
				minw = min(minw, a)
				maxw = max(maxw, b)
			}
		} else if s.class == 'L' && !s.negative {
			maxw = 2
		}
		lo += minw
		hi += maxw
	}
	return lo, hi
}

func compileNativeACPart(text string, nocase, wide, modified bool) (nativeACPart, error) {
	p := nativeACPart{nocase: nocase, wide: wide, modified: modified}
	// Square ranges check one adjacent character outside the core match;
	// they do not move its reported start/end or the multipart predecessor.
	for pass := 0; pass < 2; pass++ {
		i := strings.IndexByte(text, '[')
		if i < 0 {
			break
		}
		j := strings.IndexByte(text[i+1:], ']')
		if j < 0 {
			return p, errors.New("native AC character range")
		}
		j += i + 1
		lo, hi, err := parseGap(text[i+1 : j])
		if err != nil || hi < 0 || hi > 32 {
			return p, errors.New("native AC character distance")
		}
		var ch string
		left := i == 2 && pass == 0
		if left {
			ch, text = text[:i], text[j+1:]
		} else if len(text[j+1:]) == 2 {
			ch, text = text[j+1:], text[:i]
		} else {
			return p, errors.New("native AC character anchor")
		}
		s, err := parseNativeACSymbols(ch, 0)
		if err != nil || len(s) != 1 || len(s[0].alts) != 0 || s[0].class != 0 {
			return p, errors.New("native AC character")
		}
		c := nativeACChar{s[0].value, s[0].mask, lo, hi, s[0].mask != 0}
		if left {
			p.left = c
		} else {
			p.right = c
		}
	}
	syms, err := parseNativeACSymbols(text, 0)
	if err != nil {
		return p, err
	}
	if len(syms) > 0 && syms[0].class != 0 {
		p.leftClass, p.leftNeg = syms[0].class, syms[0].negative
		syms = syms[1:]
	}
	if len(syms) > 0 && syms[len(syms)-1].class != 0 {
		p.rightClass, p.rightNeg = syms[len(syms)-1].class, syms[len(syms)-1].negative
		syms = syms[:len(syms)-1]
	}
	if len(syms) < 2 {
		return p, errors.New("native AC minimum depth")
	}
	p.symbols = syms
	p.minWidth, p.maxWidth = nativeACWidth(syms)
	static := func(i int) bool { s := syms[i]; return s.mask == 255 && s.class == 0 && len(s.alts) == 0 }
	nonzero := func(i int) bool { return syms[i].value != 0 || nocase && static(i) }
	zero, wild := true, false
	for i := 0; i < min(3, len(syms)); i++ {
		if !static(i) {
			wild = true
			break
		}
		zero = zero && !nonzero(i)
	}
	pos := 0
	if wild || zero {
		best, nz, nzpos := 0, 0, 0
		for i := 0; i < len(syms)-1; i++ {
			for j := i; j < min(i+3, len(syms)); j++ {
				if !static(j) {
					break
				}
				if j-i+1 >= best {
					best, pos = j-i+1, i
				}
				if nonzero(pos) || pos+1 < len(syms) && nonzero(pos+1) {
					if best >= 3 {
						break
					}
					if best >= 2 && best > nz {
						nz, nzpos = best, pos
					}
				}
			}
			if best >= 3 && (nonzero(pos) || nonzero(pos+1)) {
				break
			}
		}
		if nz != 0 && pos+1 < len(syms) && !nonzero(pos) && !nonzero(pos+1) {
			best, pos = nz, nzpos
		}
		if best < 2 {
			return p, errors.New("native AC static key")
		}
	}
	p.keyAt = pos
	p.prefixMin, p.prefixMax = nativeACWidth(syms[:pos])
	for p.depth < 3 && pos+p.depth < len(syms) && static(pos+p.depth) {
		p.key[p.depth] = syms[pos+p.depth].value
		p.depth++
	}
	if p.depth < 2 {
		return p, errors.New("native AC static key")
	}
	allLiteral := true
	for i := range syms {
		allLiteral = allLiteral && static(i)
	}
	if allLiteral {
		p.literal = make([]byte, len(syms))
		for i, s := range syms {
			p.literal[i] = s.value
		}
		p.symbols = nil
	}
	p.identity = p.fingerprint()
	return p, nil
}

func (p *nativeACPart) fingerprint() [16]byte {
	if p.literal != nil {
		return p.literalFingerprint()
	}
	var buf []byte
	var symbols func([]nativeACSymbol)
	put := func(n int) { var b [4]byte; binary.LittleEndian.PutUint32(b[:], uint32(n)); buf = append(buf, b[:]...) }
	symbols = func(ss []nativeACSymbol) {
		put(len(ss))
		for _, s := range ss {
			v := s.value
			if p.nocase && asciiLetter(v) {
				v |= 32
			}
			buf = append(buf, v, s.mask, s.class)
			if s.negative {
				buf = append(buf, 1)
			} else {
				buf = append(buf, 0)
			}
			put(len(s.alts))
			if len(s.alts) != 0 {
				// Modifiers switch fixed alternatives to native's generic
				// representation; literal-only bodies have no such distinction.
				if p.modified {
					buf = append(buf, 1)
				} else {
					buf = append(buf, 0)
				}
			}
			for _, alt := range s.alts {
				symbols(alt)
			}
		}
	}
	symbols(p.symbols)
	buf = append(buf, p.leftClass, p.rightClass)
	for _, b := range []bool{p.leftNeg, p.rightNeg, p.nocase} {
		if b {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}
	}
	for _, c := range []nativeACChar{p.left, p.right} {
		buf = append(buf, c.value, c.mask)
	}
	h := sha256.Sum256(buf)
	var out [16]byte
	copy(out[:], h[:16])
	return out
}

func (p *nativeACPart) literalFingerprint() [16]byte {
	h := sha256.New()
	var count [4]byte
	binary.LittleEndian.PutUint32(count[:], uint32(len(p.literal)))
	_, _ = h.Write(count[:])
	var block [512]byte
	n := 0
	for _, v := range p.literal {
		if p.nocase && asciiLetter(v) {
			v |= 32
		}
		block[n], block[n+1] = v, 255
		n += 8
		if n == len(block) {
			_, _ = h.Write(block[:])
			n = 0
			clear(block[:])
		}
	}
	if n > 0 {
		_, _ = h.Write(block[:n])
	}
	footer := []byte{p.leftClass, p.rightClass, 0, 0, 0, p.left.value, p.left.mask, p.right.value, p.right.mask}
	if p.leftNeg {
		footer[2] = 1
	}
	if p.rightNeg {
		footer[3] = 1
	}
	if p.nocase {
		footer[4] = 1
	}
	_, _ = h.Write(footer)
	var digest [32]byte
	h.Sum(digest[:0])
	var out [16]byte
	copy(out[:], digest[:16])
	return out
}
