package signatures

import (
	"context"
	"math/bits"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// sig-2: the reference evaluator gives & and | equal precedence. Outside a
// group it splits at the FIRST operator (0&1|2 is 0&(1|2)); a group spanning an
// operand is split at the LAST operator directly inside it ((0|1&2) is
// (0|1)&2). Subsignatures 0-3 are "aaa", "bbb", "ccc" and "ddd".
func TestLogicalOperatorGroupingMatchesReference(t *testing.T) {
	const subsigs = ";616161;626262;636363;646464"
	for _, tc := range []struct {
		expr, data string
		match      bool
	}{
		{"0&1|2", "ccc", false}, // 0&(1|2)
		{"0&1|2", "aaa ccc", true},
		{"0&1|2", "aaa bbb", true},
		{"0&1|2", "aaa", false},
		{"(0|1&2)", "aaa", false}, // (0|1)&2
		{"(0|1&2)", "bbb ccc", true},
		{"(0|1&2)", "aaa ccc", true},
		{"(0&1|2)", "ccc", true}, // (0&1)|2
		{"(0&1|2)", "aaa", false},
		{"0&1|2&3", "ccc ddd", false}, // 0&(1|(2&3))
		{"0&1|2&3", "aaa ccc ddd", true},
		{"0&1|2&3", "aaa ddd", false},
		{"0|1&2", "aaa", true}, // 0|(1&2)
		{"0|1&2", "bbb", false},
		{"(0|1)&2|3", "ddd", false}, // (0|1)&(2|3)
		{"(0|1)&2|3", "bbb ddd", true},
		{"((0|1)&2|3)", "ddd", true}, // ((0|1)&2)|3
		{"(0|1)>1&2|3", "aaa aaa ddd", true},
		{"(0|1)>1&2|3", "aaa ddd", false},
	} {
		// The reference requires exactly the subsignatures the expression names.
		highest, err := lsigHighestIndex(tc.expr)
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		used := strings.Join(strings.Split(subsigs, ";")[:highest+2], ";")
		d, err := Compile(map[string][]byte{"t.ldb": []byte("P;Target:0;" + tc.expr + used + "\n")})
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		got, err := d.Match(context.Background(), []byte(tc.data))
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		if (got == "P") != tc.match {
			t.Errorf("expression %q on %q = %q; reference match = %v", tc.expr, tc.data, got, tc.match)
		}
	}
}

// Syntax the reference rejects at load time must not load here either. The
// reference reads leaf numbers leniently (the official feed holds "0 & 1 >
// 200"); after a group's count modifier this engine rejects trailing text.
func TestLogicalReferenceSyntaxErrorsRejected(t *testing.T) {
	for _, expr := range []string{"(0)>1", "((0|1))>1", "(0|1)2", "0|", "|0", "()", "0>", "(0|1)>1,", "(0|1)>1x", "0>1,2="} {
		if _, err := Compile(map[string][]byte{"t.ldb": []byte("P;Target:0;" + expr + ";616161;626262\n")}); err == nil {
			t.Errorf("accepted %q", expr)
		}
	}
}

// sig-7: for a single subsignature the reference reads only the number after
// the operator; a trailing ",Y" is ignored. Groups still enforce Y.
func TestSingleSubsignatureIgnoresDistinctCount(t *testing.T) {
	for _, tc := range []struct {
		expr, data, want string
	}{
		{"0<3,1", "zzz", "P"}, // count 0 < 3
		{"0=0,1", "zzz", "P"}, // count 0 == 0
		{"0=0,5", "zzz", "P"},
		{"0>1,1", "aaaaaa", "P"},
		{"0<3,1", "aaaaaa", ""}, // four overlapping matches
		{"(0|1)>1,2", "aaaaaa", ""},
		{"(0|1)>1,2", "aaaa bbb", "P"},
	} {
		subsigs := ";616161"
		if highest, _ := lsigHighestIndex(tc.expr); highest == 1 {
			subsigs += ";626262"
		}
		d, err := Compile(map[string][]byte{"t.ldb": []byte("P;Target:0;" + tc.expr + subsigs + "\n")})
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		got, err := d.Match(context.Background(), []byte(tc.data))
		if err != nil || got != tc.want {
			t.Errorf("expression %q on %q = %q, %v; want %q", tc.expr, tc.data, got, err, tc.want)
		}
	}
}

// referenceLogic is an independent model, for differential tests only, of how
// the reference evaluator splits and evaluates a logical expression. It keeps
// positions in the whole string because the reference reads numbers and the
// character after a group past the end of the range being evaluated.
type referenceLogic struct {
	s      string
	counts []int
}
type referenceResult struct {
	ret, count int // ret: -1 syntax error, 0 false, 1 true
	ids        uint64
}

func (m referenceLogic) number(pos int) (int, int, bool) {
	start := pos
	for pos < len(m.s) && m.s[pos] >= '0' && m.s[pos] <= '9' {
		pos++
	}
	if start == pos || pos-start > 7 {
		return 0, pos, false
	}
	n, _ := strconv.Atoi(m.s[start:pos])
	return n, pos, true
}
func referenceCompare(op byte, value, limit int) bool {
	switch op {
	case '=':
		return value == limit
	case '<':
		return value < limit
	case '>':
		return value > limit
	}
	return false
}
func (m referenceLogic) eval(lo, hi int) referenceResult {
	bad := referenceResult{ret: -1}
	n := hi - lo
	pth, opoff, op1off, modoff, blkend, modval1, modval2 := 0, 0, 0, 0, 0, 0, 0
	var op, op1, mod, blkmod byte
	for i := 0; i < n; i++ {
		switch c := m.s[lo+i]; c {
		case '(':
			pth++
		case ')':
			if pth == 0 {
				return bad
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
			if next := lo + i + 1; next < len(m.s) && strings.IndexByte("><=", m.s[next]) >= 0 {
				blkmod = m.s[next]
				v, p, ok := m.number(next + 1)
				if !ok {
					return bad
				}
				modval1 = v
				if p < len(m.s) && m.s[p] == ',' {
					if v2, _, ok := m.number(p + 1); ok {
						modval2 = v2
					}
				}
				for i += 2; i+1 < n && (m.s[lo+i+1] >= '0' && m.s[lo+i+1] <= '9' || m.s[lo+i+1] == ','); i++ {
				}
			}
			if i+1 == n {
				break
			}
			blkmod = 0
		}
	}
	if pth != 0 {
		return bad
	}
	if op == 0 && op1 == 0 {
		if lo < len(m.s) && m.s[lo] == '(' {
			return m.eval(lo+1, hi-1)
		}
		id, _, ok := m.number(lo)
		if !ok || id >= len(m.counts) {
			return bad
		}
		val := m.counts[id]
		if mod != 0 {
			limit, _, ok := m.number(lo + modoff + 1)
			if !ok {
				return bad
			}
			if !referenceCompare(mod, val, limit) {
				return referenceResult{}
			}
			return referenceResult{1, val, 1 << id}
		}
		if val > 0 {
			return referenceResult{1, val, 1 << id}
		}
		return referenceResult{}
	}
	lstart, rend := lo, hi
	if op == 0 {
		op, opoff, lstart, rend = op1, op1off, lo+1, lo+blkend
	}
	if opoff == 0 || opoff+1 == n {
		return bad
	}
	l, r := m.eval(lstart, lo+opoff), m.eval(lo+opoff+1, rend)
	if l.ret < 0 || r.ret < 0 {
		return bad
	}
	ret := l.ret == 1 && r.ret == 1
	if op == '|' {
		ret = l.ret == 1 || r.ret == 1
	}
	total, ids := 0, uint64(0)
	if ret {
		total, ids = l.count+r.count, l.ids|r.ids
	}
	if blkmod == 0 {
		if !ret {
			return referenceResult{}
		}
		return referenceResult{1, total, ids}
	}
	if !referenceCompare(blkmod, total, modval1) || modval2 != 0 && bits.OnesCount64(ids) < modval2 {
		return referenceResult{}
	}
	return referenceResult{1, total, 0}
}

func randomLogic(r *rand.Rand, depth int) string {
	modifier := func() string {
		s := string("=<>"[r.Intn(3)]) + strconv.Itoa(r.Intn(4))
		if r.Intn(3) == 0 {
			s += "," + strconv.Itoa(r.Intn(3))
		}
		return s
	}
	chain := func() string {
		s := randomLogic(r, depth+1)
		for n := r.Intn(4); n > 0; n-- {
			s += string("&|"[r.Intn(2)]) + randomLogic(r, depth+1)
		}
		return s
	}
	switch k := r.Intn(10); {
	case depth > 3 || k < 4:
		s := strconv.Itoa(r.Intn(4))
		if r.Intn(3) == 0 {
			s += modifier()
		}
		return s
	case k < 7:
		s := "(" + chain() + ")"
		if r.Intn(3) == 0 {
			s += modifier()
		}
		return s
	}
	return chain()
}

// Every expression this compiler accepts must be accepted by the reference and
// evaluate identically on every count vector; malformed variants included.
func TestLogicalExpressionsDifferentialAgainstReferenceModel(t *testing.T) {
	r := rand.New(rand.NewSource(20260923))
	accepted, rejected := 0, 0
	for iteration := 0; iteration < 20000; iteration++ {
		expr := randomLogic(r, 0)
		if r.Intn(4) == 0 && len(expr) > 1 {
			i := r.Intn(len(expr))
			if r.Intn(2) == 0 {
				expr = expr[:i] + expr[i+1:]
			} else {
				expr = expr[:i] + string("()&|<>=,0"[r.Intn(9)]) + expr[i:]
			}
		}
		e, err := parseExpression(expr, 4)
		for trial := 0; trial < 16; trial++ {
			counts := []int{r.Intn(4), r.Intn(4), r.Intn(4), r.Intn(4)}
			want := referenceLogic{expr, counts}.eval(0, len(expr))
			if err != nil {
				if want.ret >= 0 && trial == 0 {
					rejected++
				}
				continue
			}
			if want.ret < 0 {
				t.Fatalf("accepted %q, which the reference rejects", expr)
			}
			if got := e.evaluate(counts).state == logicTrue; got != (want.ret == 1) {
				t.Fatalf("%q with counts %v: got %v, reference %v", expr, counts, got, want.ret == 1)
			}
		}
		if err == nil {
			accepted++
		}
	}
	t.Logf("%d accepted and compared; %d reference-valid expressions rejected as unsupported", accepted, rejected)
	if accepted < 5000 {
		t.Fatalf("differential coverage too small: %d accepted", accepted)
	}
}
