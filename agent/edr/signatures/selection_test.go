package signatures

import (
	"errors"
	"strings"
	"testing"
)

// compileSkips compiles one file next to a plain loadable rule and returns
// the skip counts, failing when compilation fails.
func compileSkips(t *testing.T, ext, text string) map[string]int {
	t.Helper()
	d, err := Compile(map[string][]byte{"anchor.ndb": []byte("Anchor:0:*:41424344"), "test." + ext: []byte(text)})
	if err != nil {
		t.Fatalf("compile %s %q: %v", ext, text, err)
	}
	return d.Report().Skipped
}

func mustReject(t *testing.T, ext, text string) {
	t.Helper()
	if _, err := Compile(map[string][]byte{"anchor.ndb": []byte("Anchor:0:*:41424344"), "test." + ext: []byte(text)}); err == nil {
		t.Errorf("%s %q compiled; the reference rejects the database", ext, text)
	}
}

func TestCHelpers(t *testing.T) {
	for s, want := range map[string]int32{"": 0, "12": 12, "  7x": 7, "+5": 5, "-5": -5, "x1": 0, "99999999999": 1<<31 - 1, "-99999999999": -1 << 31} {
		if got := cAtoi(s); got != want {
			t.Errorf("cAtoi(%q) = %d, want %d", s, got, want)
		}
	}
	for s, want := range map[string]bool{"": true, "0": true, "216": true, "-1": false, " 1": false, "1a": false} {
		if isNumber(s) != want {
			t.Errorf("isNumber(%q) = %v", s, !want)
		}
	}
	got := cStrtokenize("a,b,,c", ',', 3)
	if strings.Join(got, "|") != "a|b|" {
		t.Errorf("cStrtokenize keeps empty fields and drops the rest: %q", got)
	}
	for _, tt := range []struct {
		line  string
		field int
		want  string
		ok    bool
	}{{"PE:0x1:10:20", 2, "10", true}, {"PE:0x1::217", 2, "217", true}, {"PE:0x1", 2, "", false}, {"PE:0x1:", 2, "", false}, {":PE", 0, "", false}} {
		if got, ok := cStrtok(tt.line, tt.field, ':'); got != tt.want || ok != tt.ok {
			t.Errorf("cStrtok(%q,%d) = %q,%v", tt.line, tt.field, got, ok)
		}
	}
}

func TestExtendedLevelWindows(t *testing.T) {
	for _, line := range []string{
		"N:0:*:41414141:235",    // minimum above the level
		"N:0:*:41414141:0:26",   // maximum below the level
		"N:0:*:41414141:51:",    // an empty maximum is 0
		"N:0:*:41414141:999:ab", // the minimum decides before the maximum is read
		"N:4:*:zz:235",          // the window is checked before the target and body
		"N:0:*:41414141:0235",   // leading zeros are digits
	} {
		if got := compileSkips(t, "ndb", line); got[SkipLevel] != 1 {
			t.Errorf("%q: skipped %v, want one level skip", line, got)
		}
	}
	for _, line := range []string{"N:0:*:41414141:234", "N:0:*:41414141:234:234", "N:0:*:41414141::234", "N:*:*:41414141", "N::*:41414141", "N:0:*:41414141:0:000234"} {
		if got := compileSkips(t, "ndb", line); len(got) != 0 {
			t.Errorf("%q must load, skipped %v", line, got)
		}
	}
	for _, line := range []string{"N:15:*:41414141", "N:999999999999:*:zz"} {
		if got := compileSkips(t, "ndb", line); got[SkipTarget] != 1 {
			t.Errorf("%q: skipped %v, want a target skip", line, got)
		}
	}
	for _, line := range []string{"N:0:*:41414141:abc", "N:0:*:41414141:51:x", "N:x:*:41414141", "N:-1:*:41414141", "N:0:*:41414141: 51"} {
		mustReject(t, "ndb", line)
	}
}

func TestHashLevelWindows(t *testing.T) {
	const md5 = "d41d8cd98f00b204e9800998ecf8427e"
	for _, line := range []string{
		md5 + ":5:N:73:",   // empty maximum: 0
		md5 + ":5:N:235",   // minimum above the level
		md5 + ":0:N:235",   // the window is checked before the size
		md5 + ":*:N:-1",    // a negative minimum compares as a huge unsigned value
		"zz:5:N:73:100",    // and before the digest
		md5 + ":5:N:x:100", // a non-numeric minimum reads as 0; the maximum excludes
		md5 + ":5:N: +235", // atoi skips blanks and reads a sign
		md5 + ":5:N:99999999999",
	} {
		if got := compileSkips(t, "hdb", line); got[SkipLevel] != 1 {
			t.Errorf("%q: skipped %v, want one level skip", line, got)
		}
	}
	for _, line := range []string{md5 + ":5:N:73:-1", md5 + ":5:N:x", md5 + ":*:N:73", md5 + ":+5:N", md5 + ": 5:N"} {
		if got := compileSkips(t, "hdb", line); len(got) != 0 {
			t.Errorf("%q must load, skipped %v", line, got)
		}
	}
	for _, line := range []string{md5 + ":*:N", md5 + ":*:N:72", md5 + ":0:N", md5 + ":4294967295:N", md5 + ":-5:N", md5 + ":5x:N", md5 + ":5", md5 + ":5:N:73:216:x"} {
		mustReject(t, "hdb", line)
	}
}

func TestLogicalSelectionOrder(t *testing.T) {
	skips := map[string]string{
		"N;Engine:51-,Target:0;0;4141":                   SkipLevel,
		"N;Engine:235-255,Target:0;0;zz":                 SkipLevel, // subsignatures are parsed only after the window
		"N;Engine:51-255,Foo:1,Target:0;0;4141":          SkipUnknownAttribute,
		"N;Foo:1,Engine:x;0;4141":                        SkipUnknownAttribute, // the first unknown attribute returns at once
		"N;Engine:51-255,Container:CL_TYPE_NOPE;0;4141":  SkipUnknownAttribute,
		"N;Engine:51-255,Intermediates:CL_TYPE_ZIP>X;0;": SkipUnknownAttribute,
		"N;Engine:51-255,Target:15;0;4141":               SkipTarget,
		"N;Engine:51-255,Target:0,Target:99;0;4141":      SkipTarget, // the last duplicate wins
		// A known attribute this engine cannot evaluate does not stop the
		// pass: a later unknown one still skips the rule.
		"N;Engine:51-255,Target:0,Container:CL_TYPE_ZIP,Foo:1;0;4141": SkipUnknownAttribute,
	}
	for line, reason := range skips {
		if got := compileSkips(t, "ldb", line); got[reason] != 1 {
			t.Errorf("%q: skipped %v, want %s", line, got, reason)
		}
	}
	// A PCRE subsignature may contain the field separator.
	if got := compileSkips(t, "ldb", "N;Engine:235-255,Target:0;0&1;4141;0/a;b/"); got[SkipLevel] != 1 {
		t.Errorf("PCRE tokenization: %v", got)
	}
	for _, line := range []string{
		"N;Target:0,Engine:51-255;0;4141",          // Engine not first
		"N;Engine:51-255;0;4141",                   // no target
		"N;Engine:51,Target:0;0;4141",              // Engine needs a hyphen
		"N;Engine:51-x,Target:0;0;4141",            // digits only
		"N;Engine:235-255,Target:0;1;4141",         // expression names a missing subsignature
		"N;Engine:235-255,Target:0;0&;4141",        // expression syntax before the window
		"N;Engine:235-255,Target:0;0;4141;4242",    // unused subsignature
		"N;Engine:235-255,Target:0,Bad;0;4141",     // attribute without a value
		"N;Target:0,FileSize:1-100;0;4141",         // new attributes need Engine >= 51
		"N;Engine:50-255,Target:0,FileSize:1-9;0;", // idem
		"N;Engine:51-255,Target:0,FileSize:5;0;41", // ranges need a hyphen
		"N;Engine:51-255,Target:0,IconGroup1:x;0;4141",
		"N;Engine:51-255,Target:0,EntryPoint:1-2;0;4141",
		"N;Engine:51-255,Target:0",
		// Attributes are read in order: a malformed one before an unknown
		// one is an error, not a skip.
		"N;Engine:51-255,Target:x,Foo:1;0;4141",
		"N;Engine:51-255,Bad,Foo:1,Target:0;0;4141",
		"N;Engine:51-255,Engine:51-255,Target:0;0;4141",
	} {
		mustReject(t, "ldb", line)
	}
	// A repeated attribute is read last-wins by the reference; this engine
	// still rejects the generation once such a rule is selected.
	mustReject(t, "ldb", "N;Engine:51-255,Target:1,Target:1;0;4141")
	for _, line := range []string{"N;Engine:51-234,Target:0;0;4141", "N;Engine:234-234,Target:1,FileSize:-100;0;4141", "N;Target:0;0;4141",
		"N;Engine:51-255,Target:;0;4141", // an empty target reads as 0
		"N;Engine:0051-99999999999,Target:0;0;4141"} {
		if got := compileSkips(t, "ldb", line); len(got) != 0 {
			t.Errorf("%q must load, skipped %v", line, got)
		}
	}
}

func TestReferenceExpressionCheck(t *testing.T) {
	for expr, want := range map[string]int{
		"0": 0, "0&1|2": 2, "(0|1&2)": 2, "(0&1)>5": 1, "0 & 1 > 200": 1, "0:0&1": 1, "(8==9)": 8,
		"0=2,&1&2": 2, "(0|1)>2,2": 1, "((0))": 0, "0&(1|2)>1": 2, "3abc": 3, "+1": 1,
	} {
		got, err := lsigHighestIndex(expr)
		if err != nil || got != want {
			t.Errorf("lsigHighestIndex(%q) = %d, %v; want %d", expr, got, err, want)
		}
	}
	for _, expr := range []string{"", "0&", "&0", "(0", "0)", "(0)>1", "a", "0>", "(0|1)>", "0&()", "-1", "(0|1)2"} {
		if _, err := lsigHighestIndex(expr); err == nil {
			t.Errorf("lsigHighestIndex(%q) accepted", expr)
		}
	}
}

func TestPhishingLevelSuffix(t *testing.T) {
	for line, stripped := range map[string]string{
		"H:amazon.com:17":        "H:amazon.com:17", // no hyphen: part of the pattern
		"H:amazon.com:17-":       "H:amazon.com",
		"H:amazon.com:-216":      "H:amazon.com", // inclusive maximum
		"H:amazon.com:216-216":   "H:amazon.com",
		"H:amazon.com:a-b":       "H:amazon.com:a-b",
		"H:amazon.com:-":         "H:amazon.com",
		"M:a.example:b.example":  "M:a.example:b.example",
		"H:amazon.com:1-2-3":     "H:amazon.com:1-2-3",
		"H:amazon.com:0017-0216": "H:amazon.com",
	} {
		got, err := phishingLevel(line, 216)
		if err != nil || got != stripped {
			t.Errorf("phishingLevel(%q) = %q, %v; want %q", line, got, err, stripped)
		}
	}
	for _, line := range []string{"H:amazon.com:217-", "H:amazon.com:-215", "H:amazon.com:20-10"} {
		if _, err := phishingLevel(line, 216); !isSkip(err, SkipLevel) {
			t.Errorf("phishingLevel(%q) = %v, want a level skip", line, err)
		}
	}
	if got := compileSkips(t, "pdb", "H:amazon.com:235-\nH:bank.example:-215"); got[SkipLevel] != 2 {
		t.Errorf("phishing skips = %v", got)
	}
	d := compileTest(t, "pdb", "H:amazon.com:-234")
	htmlMatch(t, d, `<a href="https://evil.test">amazon.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
}

func TestBytecodeHeaderWindowSkipsBeforeParse(t *testing.T) {
	for name, mutate := range map[string]func(*callbackFixture){
		"high": func(f *callbackFixture) { f.minimum = FunctionalityLevel + 1 },
		"low":  func(f *callbackFixture) { f.maximum = FunctionalityLevel - 1 },
	} {
		f := callbackWithLabel([]byte("UTMStack.Skipped"), true)
		mutate(&f)
		f.blocks = nil // the body would not parse; the reference never reads it
		if got := compileSkips(t, "cbc", string(f.encode())); got[SkipLevel] != 1 {
			t.Errorf("%s: skipped %v", name, got)
		}
	}
	f := callbackWithLabel([]byte("UTMStack.Skipped"), true)
	b := f.encode()
	b[len("ClamBC")+1] = 0x65 // format 5
	if got := compileSkips(t, "cbc", string(b)); got[SkipLevel] != 1 {
		t.Errorf("format level: skipped %v", got)
	}
	mustReject(t, "cbc", "NotBytecode")
	// The trigger's own window disagreeing with the header is malformed.
	f = callbackWithLabel([]byte("UTMStack.Skipped"), true)
	f.logical = "Bad;Engine:1-20,Target:0;0;4142"
	mustReject(t, "cbc", string(f.encode()))
	f.logical = "Bad;Engine:51-255,Target:0,Odd:1;0;4142"
	mustReject(t, "cbc", string(f.encode()))
}

func TestPUAFilesAreSkippedAndCounted(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("Ok:0:*:41424344"),
		"a.ndu": []byte("PUA.One:0:*:41424344\n# comment\n\nPUA.Two:4:*:zz"),
		"a.ldu": []byte("PUA.Three;Engine:51-255,Target:0;0;4142"),
		"a.hdu": []byte("d41d8cd98f00b204e9800998ecf8427e:0:PUA.Four"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Report().Skipped[SkipPUA]; got != 4 {
		t.Fatalf("PUA skips = %d, want 4", got)
	}
	if d.Count() != 1 {
		t.Fatalf("loaded = %d", d.Count())
	}
}

func TestSkippedReportIsACopy(t *testing.T) {
	d, err := Compile(map[string][]byte{"a.ndb": []byte("Ok:0:*:41424344\nOld:0:*:41424344:1:2")})
	if err != nil {
		t.Fatal(err)
	}
	r := d.Report()
	r.Skipped[SkipLevel] = 99
	if d.Report().Skipped[SkipLevel] != 1 {
		t.Fatal("report exposes mutable skip counts")
	}
}

func TestLegacyDoubleEqualsIsSkipped(t *testing.T) {
	if got := compileSkips(t, "db", "Old==4142\nNew=41424344"); got[SkipLegacyFormat] != 1 {
		t.Fatalf("skips = %v", got)
	}
	mustReject(t, "db", "NoEquals")
}

func isSkip(err error, reason string) bool {
	var s *skipError
	return errors.As(err, &s) && s.reason == reason
}
