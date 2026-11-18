package signatures

import (
	"strings"
	"testing"
)

func bodyMatches(t *testing.T, body, data string) bool {
	t.Helper()
	d := compileTest(t, "ldb", "Body;Engine:51-255,Target:0;0;"+body+"\n")
	name, err := matchObject(t, d, data, MatchContext{Type: TypeAny})
	if err != nil {
		t.Fatalf("%s on %q: %v", body, data, err)
	}
	return name == "Body"
}

func TestWideModifierRewritesTheBody(t *testing.T) {
	cases := []struct {
		body, data string
		want       bool
	}{
		// Gaps are not doubled by widening.
		{"41{2}42::w", "A\x00xxB\x00", true},
		{"41{2}42::w", "A\x00xxxxB\x00", false},
		// A square range becomes a curly one, so it may exceed 32 bytes.
		{"41[32-128]42::w", "A\x00" + strings.Repeat("x", 40) + "B\x00", true},
		{"41[32-128]42::w", "A\x00" + strings.Repeat("x", 20) + "B\x00", false},
		// Alternatives are widened branch by branch.
		{"41(42|43)44::w", "A\x00C\x00D\x00", true},
		// Wide and ASCII forms together.
		{"414243::wa", "..ABC..", true},
		{"414243::wa", "..A\x00B\x00C\x00..", true},
		{"41??43::wa", "..A\x00x\x00C\x00..", true},
		// Word boundaries, wide form.
		{"4142::wf", "x\x00A\x00B\x00 \x00", false},
		{"4142::wf", " \x00A\x00B\x00 \x00", true},
		{"4142::wf", "A\x00B\x00", true},
		{"4142::wf", " \x00A\x00B\x00x\x00", false},
		{"4142::f", "xAB ", false},
		{"4142::f", " AB ", true},
		// An empty modifier list means no modifiers.
		{"4142::", "xxABxx", true},
	}
	for _, c := range cases {
		if got := bodyMatches(t, c.body, c.data); got != c.want {
			t.Errorf("%s on %q: %v, want %v", c.body, c.data, got, c.want)
		}
	}
}

func TestWideAndAsciiFormsShareOneCount(t *testing.T) {
	d := compileTest(t, "ldb", "Two;Engine:51-255,Target:0;0>1;414243::wa\n")
	if name, _ := matchObject(t, d, "ABC and A\x00B\x00C\x00", MatchContext{Type: TypeAny}); name != "Two" {
		t.Fatalf("one occurrence of each form must count twice: %q", name)
	}
	if name, _ := matchObject(t, d, "ABC only", MatchContext{Type: TypeAny}); name != "" {
		t.Fatalf("one occurrence counted twice: %q", name)
	}
}

func TestSingleAndEmptyAlternatives(t *testing.T) {
	cases := []struct {
		body, data string
		want       bool
	}{
		{"41!(00)42", "AxB", true},
		{"41!(00)42", "A\x00B", false},
		{"(41)42", "xABx", true},
		{"41(22|)42", "A\"B", true},
		{"41(22|)42", "AB", true},
		{"41(22|)42", "AxB", false},
	}
	for _, c := range cases {
		if got := bodyMatches(t, c.body, c.data); got != c.want {
			t.Errorf("%s on %q: %v, want %v", c.body, c.data, got, c.want)
		}
	}
	for _, body := range []string{"41!(42)::i", "41!(42)::w"} {
		if _, err := Compile(map[string][]byte{"a.ldb": []byte("Bad;Engine:51-255,Target:0;0;" + body + "\n")}); err == nil {
			t.Errorf("%s compiled: a negated alternative takes no modifiers", body)
		}
	}
}

func TestCountModifiersOnGroupsAndLenientExpressions(t *testing.T) {
	cases := []struct {
		expr   string
		counts []int
		want   logicState
	}{
		{"(0&1)>5", []int{3, 3}, logicTrue},
		{"(0&1)>5", []int{6, 0}, logicFalse},
		{"(0&1)=0", []int{1, 0}, logicTrue},
		{"((0|1)&2)>3,2", []int{2, 0, 2}, logicTrue},
		{"((0|1)&2)>3,3", []int{2, 0, 2}, logicFalse},
		{"0 & 1 > 200", []int{1, 201}, logicTrue},
		{"0 & 1 > 200", []int{1, 5}, logicFalse},
		{"0:0&1", []int{1, 1}, logicTrue},
		{"0=2,&1", []int{2, 1}, logicTrue},
	}
	for _, c := range cases {
		e, err := parseExpression(c.expr, len(c.counts))
		if err != nil {
			t.Errorf("%s: %v", c.expr, err)
			continue
		}
		if got := e.evaluate(c.counts).state; got != c.want {
			t.Errorf("%s with %v: %v, want %v", c.expr, c.counts, got, c.want)
		}
	}
}
