package signatures

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// Regex subsignatures run with the reference's meaning inside rules
// (parity/pcre.md): the trigger, the offset's window and anchoring, the g
// count, the flags, and the 8-bit library's classes.
func TestRegexSubsignatureIsEvaluated(t *testing.T) {
	for _, tt := range []struct {
		rule, data, want string
	}{
		{"R;Engine:81-255,Target:0;0&1;414243;0/x+y/i", "ABC and XXY", "R"},
		{"R;Engine:81-255,Target:0;0&1;414243;0/x+y/i", "ABC and xz", ""},
		{"R;Engine:81-255,Target:0;0&1;414243;0/x+y/", "nothing ABC here but xy", "R"},
		{"R;Engine:81-255,Target:0;0&1;414243;0/x+y/", "no marker but xy", ""},
		// The trigger may need a higher count than the rule expression does.
		{"R;Engine:81-255,Target:0;0&1;414243;0>2/x/", "ABC ABC x", ""},
		{"R;Engine:81-255,Target:0;0&1;414243;0>2/x/", "ABC ABC ABC x", "R"},
		// A regex may hold the field separators ';' and ':'.
		{"R;Engine:81-255,Target:0;0&1;414243;0/a;b:c/", "ABC a;b:c", "R"},
		{"R;Engine:81-255,Target:0;0&1;414243;0/a;b:c/", "ABC a;b", ""},
		// EOF-1 anchors at the last byte.
		{"R;Engine:81-255,Target:0;1;414243;EOF-1:0/\\./", "ABC.", "R"},
		{"R;Engine:81-255,Target:0;1;414243;EOF-1:0/\\./", "ABC.x", ""},
		// An absolute offset anchors there; 'r' lets it roll.
		{"R;Engine:81-255,Target:0;1;414243;3:0/x/", "ABCx", "R"},
		{"R;Engine:81-255,Target:0;1;414243;3:0/x/", "ABC x", ""},
		{"R;Engine:81-255,Target:0;1;414243;3:0/x/r", "ABC x", "R"},
		// A window: the start must fall inside it, the match may run past it.
		{"R;Engine:81-255,Target:0;1;414243;3,2:0/xyz/", "ABC  xyz", "R"},
		{"R;Engine:81-255,Target:0;1;414243;3,2:0/xyz/", "ABC   xyz", ""},
		// With e the subject ends at the window's end.
		{"R;Engine:81-255,Target:0;1;414243;3,4:0/xyz/e", "ABC xyz", "R"},
		{"R;Engine:81-255,Target:0;1;414243;3,4:0/xyz/e", "ABC  xyz", ""},
		// g counts non-overlapping matches; without g the count is at most 1.
		{"R;Engine:81-255,Target:0;1>2;414243;0/ab/g", "ABC abab ab", "R"},
		{"R;Engine:81-255,Target:0;1>2;414243;0/ab/g", "ABC abab", ""},
		{"R;Engine:81-255,Target:0;1>2;414243;0/aba/g", "ABC ababababa", ""},
		{"R;Engine:81-255,Target:0;1>1;414243;0/ab/", "ABC ab ab ab", ""},
		// The library's classes: \s holds the vertical tab, bytes are bytes.
		{"R;Engine:81-255,Target:0;1;414243;0/a\\sb/", "ABC a\x0bb", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/\\xe9t/", "ABC \xe9t", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/\\xe9t/", "ABC \xc3\xa9t", ""},
		{"R;Engine:81-255,Target:0;1;414243;0/k\\w/i", "ABC K_", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/k/i", "ABC \xe2\x84\xaa", ""},
		// $ matches before a final newline.
		{"R;Engine:81-255,Target:0;1;414243;0/end$/", "ABC end\n", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/end$/", "ABC end\n\n", ""},
		{"R;Engine:81-255,Target:0;1;414243;0/end$/E", "ABC end\n", ""},
		// Backreferences and lookaround.
		{"R;Engine:81-255,Target:0;1;414243;0/(\\w)\\1\\1/", "ABC zzz", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/(?P<q>['\"]).*(?P=q)/", "ABC 'x'", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/(?P<q>['\"]).*(?P=q)/", "ABC 'x\"", ""},
		{"R;Engine:81-255,Target:0;1;414243;0/foo(?!bar)/", "ABC foobar", ""},
		{"R;Engine:81-255,Target:0;1;414243;0/foo(?!bar)/", "ABC foobaz", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/(?<=x)y/", "ABC xy", "R"},
		{"R;Engine:81-255,Target:0;1;414243;0/(?<=x)y/", "ABC zy", ""},
	} {
		d := compileTest(t, "ldb", tt.rule+"\n")
		expectOutcome(t, d, tt.data, anyType, tt.want, "")
	}
}

// Where the reference's library could give up (100,000 backtracking calls
// in one start), the count is unknown: a lazy gap over more than 100,000
// bytes gives up in the reference, so this engine does not detect.
func TestRegexBeyondTheLibraryLimitsIsUnknown(t *testing.T) {
	d := compileTest(t, "ldb", "R;Engine:81-255,Target:0;1;414243;0/a.*?b/s\n")
	expectOutcome(t, d, "ABC a"+strings.Repeat("x", 1000)+"b", anyType, "R", "")
	expectOutcome(t, d, "ABC a"+strings.Repeat("x", 150000)+"b", anyType, "", UnknownPCRELimit)
	// Without a b the regex cannot match, whatever the library's limits.
	expectOutcome(t, d, "ABC a"+strings.Repeat("x", 150000), anyType, "", "")
	// Group repeats nest frames: 1,500 iterations of a capturing group are
	// past what the depth limit allows the library.
	d = compileTest(t, "ldb", "R;Engine:81-255,Target:0;1;414243;0/(ab)+c/\n")
	expectOutcome(t, d, "ABC "+strings.Repeat("ab", 100)+"c", anyType, "R", "")
	expectOutcome(t, d, "ABC "+strings.Repeat("ab", 1500)+"c", anyType, "", UnknownPCRELimit)
}

// The reference runs no regex on an object larger than 100 MiB: every
// regex count is 0 there.
func TestRegexSkipsObjectsOverTheSizeLimit(t *testing.T) {
	defer func(n int) { rxMaxObject = n }(rxMaxObject)
	rxMaxObject = 64
	d := compileTest(t, "ldb", "R;Engine:81-255,Target:0;1;414243;0/xyz/\n")
	expectOutcome(t, d, "ABC xyz", anyType, "R", "")
	expectOutcome(t, d, "ABC xyz"+strings.Repeat(" ", 64), anyType, "", "")
}

// A regex whose trigger is unknown decides the rule only when it cannot
// match: a match leaves the count unknown.
func TestRegexWithAnUnknownTrigger(t *testing.T) {
	// Subsignature 1 uses recursion (unknown when reached); it triggers 2.
	d := compileTest(t, "ldb", "R;Engine:81-255,Target:0;2;414243;0/x(?R)/;1/yz/\n")
	expectOutcome(t, d, "no marker, x and yz", anyType, "", "")
	expectOutcome(t, d, "ABC but no regex text", anyType, "", "")
	name, err := matchObject(t, d, "ABC has x and yz", anyType)
	if name != "" || unknownReason(err) != UnknownPCRE {
		t.Fatalf("unknown trigger with a match: %q, %v", name, err)
	}
}

// Macro module text is matched in buffer mode, where the reference cuts the
// subject and anchors every relative offset at 0; regexes stay unknown there.
// In buffer mode (macro module text) a regex applies to the whole buffer and
// its offsets resolve as in a file, as the reference does (checked with
// marker rules on a macro module: absolute 52 and EOF-57 match a marker at
// 52 of 109 bytes, 51 and EOF-58 do not, a rolling search from 0 does).
func TestRegexInBufferMode(t *testing.T) {
	cases := []struct {
		rule, want string
		unknown    bool
	}{
		{"R;Engine:81-255,Target:0;0&1;414243;0/xyz/\n", "R", false},
		{"R;Engine:81-255,Target:0;0&1;414243;0/nothere/\n", "", false},
		{"R;Engine:81-255,Target:0;0&1;414243;4:0/xyz/\n", "R", false},
		{"R;Engine:81-255,Target:0;0&1;414243;3:0/xyz/\n", "", false},
		{"R;Engine:81-255,Target:0;0&1;414243;EOF-3:0/xyz/\n", "R", false},
		{"R;Engine:81-255,Target:0;0&1;414243;EOF-4:0/xyz/\n", "", false},
		{"R;Engine:81-255,Target:0;0&1;414243;0:0/xyz/r\n", "R", false},
	}
	for _, c := range cases {
		d := compileTest(t, "ldb", c.rule)
		res, err := d.MatchObject(context.Background(), []byte("ABC xyz"), MatchContext{Type: TypeAny, Mode: BufferMode})
		if res.Name != c.want || (unknownReason(err) == UnknownPCRE) != c.unknown {
			t.Errorf("%s: %q, %v", c.rule, res.Name, err)
		}
	}
}

// The text every match must contain is found before any matching starts.
func TestRegexRequiredLiterals(t *testing.T) {
	for _, tt := range []struct {
		pattern, flags string
		want           []string
	}{
		{`\x3b[\s\t]*filename[\s\t]*=[\s\t]*\x22DHL_DOCUMENTS\.jar\x22`, "", []string{"\"DHL_DOCUMENTS.jar\"", "filename"}},
		{`abc(?:def)+x?yz`, "", []string{"abc", "def", "yz"}},
		{`(?:abc|abd)ef`, "", []string{"ef"}},
		{`abc(?:def)*gh`, "i", []string{"abc", "gh"}},
		{`a(?=bcd)`, "", nil},
		{`(ab)\1`, "", []string{"ab"}},
	} {
		p, err := parseRegex(tt.pattern, tt.flags)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, l := range rxRequiredLiterals(p.root) {
			got = append(got, string(l.text))
		}
		if strings.Join(got, "|") != strings.Join(tt.want, "|") {
			t.Errorf("/%s/%s: %q, want %q", tt.pattern, tt.flags, got, tt.want)
		}
	}
	l := rxLiteral{text: []byte("filename"), fold: true}
	if !l.contains([]byte("x; FileName=")) || l.contains([]byte("x; FileNam=")) {
		t.Error("case-insensitive required text")
	}
}

// Many regexes that need the same absent text cost one search of the
// object, not one per regex (172 official mail rules share "filename"), and
// a first byte above 0x7f is found as itself.
func TestRegexRequiredTextIsSearchedOncePerObject(t *testing.T) {
	compile := func(n int) *Database {
		var rules strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&rules, "R%d;Engine:81-255,Target:0;0&1;414243;0/\\x3b[\\s\\t]*filename[\\s\\t]*=\\d{%d}/\n", i, i+1)
		}
		return compileTest(t, "ldb", rules.String())
	}
	data := []byte("ABC " + strings.Repeat("x", 1<<20))
	_, one, _ := compile(1).matchCounted(context.Background(), data, 1<<40)
	d := compile(200)
	name, steps, err := d.matchCounted(context.Background(), data, 1<<40)
	if name != "" || err != nil || steps-one > 100_000 {
		t.Fatalf("(%q, %v) in %d steps, %d with one rule", name, err, steps, one)
	}
	data = append(data, "; filename=1234567"...)
	if name, _, err := d.matchCounted(context.Background(), data, 1<<40); name != "R0" || err != nil {
		t.Fatalf("(%q, %v), want R0", name, err)
	}
	d = compileTest(t, "ldb", "R;Engine:81-255,Target:0;0&1;414243;0/(?:\\xc7\\x45|\\x81\\x85)ab/\n")
	expectOutcome(t, d, "ABC \x81\x85ab", anyType, "R", "")
	expectOutcome(t, d, "ABC \x81\x86ab \xc7\x45a", anyType, "", "")
}
