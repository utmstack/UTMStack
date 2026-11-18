package signatures

import "math"

// stateWindows adds, for every node the tokens compile to, the number of
// positions relative to the match start at which matchesAt can visit it:
// bounded windows to bound (saturating) and unbounded ones to open. One start
// never remembers more failed states than these windows hold (a position is
// also never past the object end), so the total bounds its memo.
func stateWindows(ts []token, lo, hi int, bound, open *int64) {
	for _, t := range ts {
		addWindow(lo, hi, bound, open)
		if len(t.alternatives) > 0 && !t.negative {
			for _, alternative := range t.alternatives {
				stateWindows(alternative, lo, hi, bound, open)
			}
		}
		a, b := tokenWidth([]token{t})
		lo += a
		if hi < 0 || b < 0 {
			hi = -1
		} else {
			hi += b
		}
	}
}

// addWindow adds one node visited lo to hi bytes after the match start.
func addWindow(lo, hi int, bound, open *int64) {
	if hi < 0 {
		*open++
	} else {
		*bound = min(*bound+int64(hi-lo+1), math.MaxInt64/4)
	}
}

// proneFrom returns the smallest object length at which a start could hold
// maxPatternStates failed states: math.MaxInt when no length can, 0 when every
// length might. On an object of length L an unbounded window holds at most
// L+1 positions.
func proneFrom(bound, open int64) int {
	limit := int64(maxPatternStates)
	switch {
	case bound >= limit:
		return 0
	case open == 0:
		return math.MaxInt
	}
	need := (limit - bound + open - 1) / open // L+1 at which open*(L+1) reaches the rest
	return int(max(need-1, 0))
}

// chooseAnchors picks, like anchorsFor, either one top-level literal or one
// anchor from every branch of a top-level alternative, but by expected cost
// instead of length: the chance of an occurrence at a data position times the
// start positions each occurrence leaves to verify. Runs of zero, 0xff, int3
// and nop padding and common opcode bytes therefore lose to rarer literals.
// It returns nil only when anchorsFor does. lo and hi are the distances of the
// first token from the match start.
func chooseAnchors(ts []token, lo, hi int, nocase bool) []patternAnchor {
	set, _ := bestAnchors(ts, lo, hi, nocase)
	return set
}

func bestAnchors(ts []token, lo, hi int, nocase bool) ([]patternAnchor, float64) {
	var best []patternAnchor
	bestCost := math.Inf(1)
	for _, t := range ts {
		if len(t.literal) > 0 {
			a := patternAnchor{literal: t.literal, minPre: lo, maxPre: hi}
			if c := anchorCost(a, nocase); c < bestCost {
				best, bestCost = []patternAnchor{a}, c
			}
		}
		if len(t.alternatives) > 0 && !t.negative {
			var set []patternAnchor
			var costs []float64
			for _, alternative := range t.alternatives {
				branch, c := bestAnchors(alternative, lo, hi, nocase)
				if len(branch) == 0 {
					set = nil
					break
				}
				set, costs = append(set, branch...), append(costs, c)
			}
			if total := log2Sum(costs); len(set) > 0 && total < bestCost {
				best, bestCost = set, total
			}
		}
		a, b := tokenWidth([]token{t})
		lo += a
		if hi < 0 || b < 0 {
			hi = -1
		} else {
			hi += b
		}
	}
	return best, bestCost
}

// anchorCost estimates, as a base-2 logarithm, the verification work one
// anchor causes per data byte: its occurrence chance times the starts each
// occurrence leaves to try.
func anchorCost(a patternAnchor, nocase bool) float64 {
	starts := float64(1 << 24) // an unbounded distance leaves the whole span
	if a.maxPre >= 0 {
		starts = float64(a.maxPre-a.minPre) + 1
	}
	return math.Log2(3+starts) - literalBits(a.literal, nocase)
}

// log2Sum adds costs given as base-2 logarithms.
func log2Sum(costs []float64) float64 {
	top := math.Inf(-1)
	for _, c := range costs {
		top = max(top, c)
	}
	sum := 0.0
	for _, c := range costs {
		sum += math.Exp2(c - top)
	}
	return top + math.Log2(sum)
}

// literalBits is the estimated information content of a literal, in bits,
// under a first-order byte model: each byte costs its frequency, and a byte
// repeating its predecessor costs how often that byte repeats. With nocase an
// ASCII letter matches either case.
func literalBits(literal []byte, nocase bool) float64 {
	table := &byteBits.plain
	if nocase {
		table = &byteBits.folded
	}
	bits := 0.0
	for i, c := range literal {
		cost := table[c]
		if i > 0 && literal[i-1] == c {
			cost = min(cost, byteBits.repeat[c])
		}
		bits += cost
	}
	return bits
}

// byteBits holds the byte model in bits, derived once from the tables below.
var byteBits = func() (b struct{ plain, folded, repeat [256]float64 }) {
	for c := range 256 {
		b.plain[c] = float64(byteHalfBits[c]) / 2
		b.repeat[c] = float64(repeatHalfBits[c]) / 2
	}
	for c := range 256 {
		b.folded[c] = b.plain[c]
		if asciiLetter(byte(c)) {
			b.folded[c] = -math.Log2(math.Exp2(-b.plain[c|0x20]) + math.Exp2(-b.plain[c&^0x20]))
		}
	}
	return b
}()

// keyWindow returns where the lookup key of an anchor literal starts: its
// rarest window of min(4, len) bytes, the earliest on ties. literal is already
// lowercased when folded.
func keyWindow(literal []byte, fold bool) int32 {
	n := min(4, len(literal))
	best, bestBits := int32(0), -1.0
	for at := 0; at+n <= len(literal); at++ {
		if bits := literalBits(literal[at:at+n], fold); bits > bestBits {
			best, bestBits = int32(at), bits
		}
	}
	return best
}

// byteHalfBits is -2*log2 of each byte's frequency and repeatHalfBits the same
// for a byte following itself, rounded, as measured over about 31 MB of
// Windows executables, drivers, installers, disk images and other binaries.
// They only steer which literal a pattern is indexed by; results never depend
// on them.
var byteHalfBits = [256]uint8{
	5, 12, 13, 13, 14, 15, 15, 16, 13, 16, 16, 17, 16, 17, 14, 14,
	15, 17, 17, 17, 16, 17, 18, 17, 16, 18, 18, 18, 17, 18, 18, 16,
	13, 17, 18, 18, 13, 17, 18, 18, 16, 17, 18, 18, 17, 18, 16, 17,
	16, 16, 18, 17, 17, 18, 18, 18, 17, 17, 18, 18, 18, 17, 18, 18,
	14, 15, 16, 17, 15, 16, 17, 18, 12, 16, 18, 18, 15, 17, 18, 18,
	16, 18, 17, 17, 16, 17, 18, 17, 18, 19, 18, 18, 17, 17, 18, 16,
	17, 15, 17, 15, 16, 13, 15, 17, 16, 14, 18, 18, 15, 16, 14, 15,
	16, 18, 14, 15, 13, 15, 17, 18, 17, 17, 18, 18, 18, 18, 18, 18,
	15, 17, 18, 14, 16, 16, 18, 18, 17, 13, 18, 13, 18, 15, 19, 19,
	16, 16, 19, 19, 17, 19, 19, 18, 18, 19, 19, 19, 18, 19, 19, 18,
	17, 18, 19, 19, 18, 19, 19, 19, 18, 17, 17, 19, 19, 19, 19, 19,
	17, 19, 19, 19, 18, 18, 18, 19, 17, 17, 18, 18, 18, 18, 18, 18,
	16, 17, 18, 16, 17, 18, 17, 16, 18, 18, 19, 19, 15, 19, 19, 19,
	16, 17, 18, 18, 18, 19, 18, 19, 18, 19, 19, 18, 19, 19, 19, 18,
	16, 18, 18, 18, 18, 18, 18, 18, 14, 17, 19, 17, 17, 19, 19, 18,
	17, 18, 18, 18, 18, 19, 18, 18, 17, 16, 18, 17, 17, 17, 16, 11,
}

var repeatHalfBits = [256]uint8{
	1, 9, 9, 11, 11, 11, 12, 13, 10, 10, 8, 13, 10, 11, 17, 16,
	14, 14, 14, 14, 15, 14, 14, 15, 15, 14, 13, 12, 14, 14, 15, 17,
	8, 15, 14, 8, 12, 15, 14, 15, 16, 15, 15, 13, 13, 10, 13, 13,
	9, 11, 12, 13, 14, 12, 12, 14, 15, 14, 12, 15, 14, 14, 13, 15,
	14, 17, 16, 14, 16, 15, 13, 16, 14, 16, 15, 16, 16, 16, 16, 14,
	14, 13, 13, 13, 16, 14, 15, 15, 13, 11, 15, 14, 16, 14, 16, 10,
	16, 19, 14, 13, 11, 13, 11, 15, 16, 16, 15, 16, 7, 12, 12, 11,
	10, 16, 10, 10, 12, 17, 15, 13, 15, 16, 13, 15, 15, 15, 15, 14,
	14, 16, 15, 19, 15, 17, 15, 15, 13, 20, 15, 18, 15, 16, 15, 15,
	8, 16, 14, 15, 16, 15, 14, 15, 15, 14, 15, 15, 15, 15, 14, 15,
	14, 15, 15, 15, 15, 14, 14, 14, 15, 16, 14, 14, 14, 14, 14, 14,
	15, 14, 14, 14, 15, 14, 15, 15, 15, 16, 14, 14, 14, 14, 15, 15,
	17, 14, 16, 17, 17, 15, 16, 15, 16, 16, 15, 15, 1, 15, 15, 15,
	18, 16, 16, 16, 16, 14, 15, 15, 16, 15, 15, 15, 15, 14, 15, 15,
	17, 16, 15, 15, 11, 13, 13, 14, 15, 15, 13, 15, 16, 14, 11, 14,
	14, 13, 13, 14, 15, 14, 15, 14, 16, 16, 14, 13, 14, 14, 14, 2,
}
