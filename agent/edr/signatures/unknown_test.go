package signatures

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// unknownReason returns the reason of an unknown-when-reached result, or "".
func unknownReason(err error) string {
	var u *UnknownError
	if errors.As(err, &u) && errors.Is(err, ErrRuleIncomplete) {
		return u.Reason
	}
	return ""
}

func expectOutcome(t *testing.T, d *Database, data string, mc MatchContext, wantName, wantReason string) {
	t.Helper()
	name, err := matchObject(t, d, data, mc)
	if name != wantName || unknownReason(err) != wantReason || wantReason == "" && err != nil {
		t.Fatalf("on %q (%s): %q, %v; want %q with unknown reason %q", data, mc.Type, name, err, wantName, wantReason)
	}
}

var anyType = MatchContext{Type: TypeAny}

// A regex with a construct this engine does not evaluate ((?R) recursion
// here) is unknown exactly where its trigger holds.
func TestRegexSubsignatureIsUnknownOnlyWhenItsTriggerHolds(t *testing.T) {
	d := compileTest(t, "ldb", "Rx.And;Engine:81-255,Target:0;0&1;414243;0/x+(?R)?y/i\n")
	expectOutcome(t, d, "nothing here", anyType, "", "")
	expectOutcome(t, d, "ABC and xxy", anyType, "", UnknownPCRE)
	d = compileTest(t, "ldb", "Rx.Or;Engine:81-255,Target:0;0|1;414243;0/x(?R)/\n")
	expectOutcome(t, d, "ABC", anyType, "Rx.Or", "")
	expectOutcome(t, d, "nothing", anyType, "", "")
	// The trigger may need a higher count than the rule expression does.
	d = compileTest(t, "ldb", "Rx.Count;Engine:81-255,Target:0;0&1;414243;0>2/x(?R)/\n")
	expectOutcome(t, d, "ABC ABC", anyType, "", "")
	expectOutcome(t, d, "ABC ABC ABC", anyType, "", UnknownPCRE)
	// A regex may hold the field separators ';' and ':'.
	d = compileTest(t, "ldb", "Rx.Sep;Engine:81-255,Target:0;0&1;414243;0/a;b:c(?R)/\n")
	expectOutcome(t, d, "ABC", anyType, "", UnknownPCRE)
}

func TestFuzzyImageHashIsUnknownOnlyOnImages(t *testing.T) {
	d := compileTest(t, "ldb", "Img;Engine:150-255,Target:0;0;fuzzy_img#af2ad01ed42993c7\nImg.Zero;Engine:150-255,Target:0;0;fuzzy_img#af2ad01ed42993c8#0\n")
	for _, typ := range []FileType{TypePNG, TypeGIF, TypeJPEG, TypeTIFF, TypeGraphics} {
		expectOutcome(t, d, "\x89PNG image bytes", MatchContext{Type: typ}, "", UnknownFuzzyImage)
	}
	expectOutcome(t, d, "plain text", anyType, "", "")
	expectOutcome(t, d, "MZ executable", MatchContext{Type: TypeMSEXE}, "", "")
}

func TestUnsupportedByteCompareIsUnknownOnlyWhenItsSubsignatureMatched(t *testing.T) {
	d := compileTest(t, "ldb", "Cmp;Engine:100-255,Target:0;1;414243;0(>>26#ib3#>512)\n")
	expectOutcome(t, d, "nothing", anyType, "", "")
	expectOutcome(t, d, "xxABCxx", anyType, "", UnknownByteCompare)
}

func TestDescriptorsAreEvaluatedOnlyAfterTheExpression(t *testing.T) {
	pe := labProgram([]labIcon{{w: 16, bpp: 24, rgba: make([][4]byte, 256)}})
	copy(pe[len(pe)-4:], "ABCD")
	d := iconDatabase(t, "Win.Test.Icon:GROUP:ANY:"+resourceFlatHash+"\n", "Icon;Engine:51-255,Target:1,IconGroup1:GROUP;0;41424344")
	expectOutcome(t, d, string(pe), MatchContext{Type: TypeMSEXE}, "Icon", "")
	other := append([]byte(nil), pe...)
	copy(other[len(other)-4:], "WXYZ")
	expectOutcome(t, d, string(other), MatchContext{Type: TypeMSEXE}, "", "")
	// Without a parsed header the icon group is known false.
	expectOutcome(t, d, "0123MZ.......ABCD", MatchContext{Type: TypeMSEXE}, "", "")
	d = compileTest(t, "ldb", "Handler;Engine:51-255,Target:0,HandlerType:CL_TYPE_PDF;0;41424344\n")
	expectOutcome(t, d, "xxABCDxx", anyType, "", UnknownHandlerType)
	expectOutcome(t, d, "xxABCDxx", MatchContext{Type: TypeAny, Mode: BufferMode}, "Handler", "")
	expectOutcome(t, d, "xxxxxxxx", anyType, "", "")
	d = compileTest(t, "ldb", "ScriptHandler;Engine:51-255,Target:0,HandlerType:CL_TYPE_SCRIPT;0;41424344\n")
	expectOutcome(t, d, "xxABCDxx", MatchContext{Type: TypeAny, Mode: BufferMode}, "ScriptHandler", "")
}

func TestVersionInfoOffsetMatchesNothingWithoutVersionStrings(t *testing.T) {
	d := compileTest(t, "ldb", "Vi;Engine:51-255,Target:1;0;VI:41424344\nViElf;Engine:51-255,Target:6;0;VI:41424344\n")
	// No resources, so no version strings: known false, not unknown.
	expectOutcome(t, d, string(testPE()), MatchContext{Type: TypeMSEXE}, "", "")
	expectOutcome(t, d, "\x7fELF with ABCD", MatchContext{Type: TypeELF}, "", "")
	// An unmapped entry point makes the reference reject the PE header.
	expectOutcome(t, d, string(append(brokenPE(), "ABCD"...)), MatchContext{Type: TypeMSEXE}, "", "")
}

func TestDetectionWinsOverUnknownRules(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ldb": []byte("Rx;Engine:81-255,Target:0;0&1;414243;0/x/\n"),
		"b.ndb": []byte("Hit:0:*:414243\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	expectOutcome(t, d, "ABC", anyType, "Hit", "")
}

func TestUnknownRulesAreCountedByReason(t *testing.T) {
	d, err := Compile(map[string][]byte{"a.idb": []byte("Win.Test.Icon:MS:GROUP:" + iconMetrics + "\n"), "a.ldb": []byte(strings.Join([]string{
		"Rx;Engine:81-255,Target:0;0&1;414243;0/x(?R)/",
		"Rx2;Engine:81-255,Target:0;0&1&2;414243;0/x\\R/;0/y/",
		"Rx3;Engine:81-255,Target:0;0&1;414243;0/x/",
		"Img;Engine:150-255,Target:0;0;fuzzy_img#af2ad01ed42993c7",
		"Cmp;Engine:100-255,Target:0;1;414243;0(>>26#ib3#>512)",
		"Icon;Engine:51-255,Target:1,IconGroup2:GROUP;0;41424344",
		"Handler;Engine:51-255,Target:0,HandlerType:CL_TYPE_ZIP;0;41424344",
		"Vi;Engine:51-255,Target:1;0;VI:41424344",
		"Plain;Engine:51-255,Target:0;0;41424344",
	}, "\n") + "\n")})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{UnknownPCRE: 2, UnknownFuzzyImage: 1, UnknownByteCompare: 1}
	got := d.Report().Unknown
	for reason, n := range want {
		if got[reason] != n {
			t.Errorf("Report().Unknown[%s] = %d, want %d (all: %v)", reason, got[reason], n, got)
		}
	}
	if d.Count() != 9 {
		t.Errorf("Count() = %d, want every rule loaded", d.Count())
	}
}

func TestSyntaxOutsideEveryKnownCategoryStillFails(t *testing.T) {
	for _, line := range []string{
		"Bad;Engine:81-255,Target:0;0&1;414243;/x/",           // a regex without a trigger
		"Bad;Engine:81-255,Target:0;0&1;414243;1/x/",          // a trigger naming itself
		"Bad;Engine:81-255,Target:0;0&1;414243;0/x/G",         // an unknown regex flag
		"Bad;Engine:100-255,Target:0;1;414243;1(>>26#ib2#>5)", // a comparison of itself
		"Bad;Engine:150-255,Target:0;0;fuzzy_img#nothex",
		"Bad;Engine:150-255,Target:0;0;fuzzy_img#af2ad01ed42993c7#1",
		"Bad;Engine:51-255,Target:0;0;4142::x",
	} {
		if _, err := Compile(map[string][]byte{"a.ldb": []byte(line + "\n")}); err == nil {
			t.Errorf("%q compiled", line)
		}
	}
}

func TestUnsupportedBytecodeIsUnknownWhenItsTriggerHolds(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.Unsupported"), true)
	f.apiMax = 53
	f.apis = append(f.apis, callbackAPI(53, 70, "atoi"))
	d := compileCallback(t, f)
	expectOutcome(t, d, "xxAByy", anyType, "", UnknownBytecode)
	expectOutcome(t, d, "nothing", anyType, "", "")
	if d.Report().Unknown[UnknownBytecode] != 1 {
		t.Fatalf("unknown bytecode not counted: %v", d.Report().Unknown)
	}
}

func TestHookBytecodeWithoutTriggerIsUnknownOnItsObjects(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.PDF.Hook"), true)
	f.kind = 258
	f.logical = ""
	// an API this engine lacks keeps the program unknown
	f.apiMax = 53
	f.apis = append(f.apis, callbackAPI(53, 70, "atoi"))
	d, err := Compile(map[string][]byte{"hook.cbc": f.encode(), "a.ndb": []byte("Other:0:*:57575757\n")})
	if err != nil {
		t.Fatal(err)
	}
	expectOutcome(t, d, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "", UnknownBytecode)
	expectOutcome(t, d, "%PDF-1.4 document", MatchContext{Type: TypePDF, Mode: BufferMode}, "", "")
	expectOutcome(t, d, "plain text", anyType, "", "")
}

func TestIconAndCertificateDatabasesLoad(t *testing.T) {
	icon := "Win.Test.Icon:MS:IE:" + iconMetrics
	cert := "Test.Trust;1;4a532974c46ae5048824c6da8cfb8e163705b693;1d80f656104b69a763a5d40db6abe01546b88861;ABCD;0;1;1;0;0;0"
	d, err := Compile(map[string][]byte{"a.idb": []byte(icon + "\n"), "a.crb": []byte(cert + "\n"), "a.ndb": []byte("Hit:0:*:414243\n")})
	if err != nil {
		t.Fatal(err)
	}
	if d.Report().Unknown[UnknownCertificateTrust] != 1 {
		t.Fatalf("certificate trust rules not counted: %v", d.Report().Unknown)
	}
	assertMatch(t, d, []byte("xxABCxx"), "Hit")
}

func TestUnsupportedPhishingAllowRuleMakesOnlyPairsItCouldAllowIncomplete(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"test.pdb": []byte("H:rbcroyalbank.com\nH:example.com\n"),
		"test.wdb": []byte("M:rbc.com%zz:rbcroyalbank.com\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Report().Unknown[UnknownPhishingAllowlist] != 1 {
		t.Fatalf("unsupported allow rule not counted: %v", d.Report().Unknown)
	}
	// The allow rule could apply: real host holds rbc and com.
	got, err := d.MatchHTML(context.Background(), []byte(`<a href="https://secure.rbc.com/">https://www.rbcroyalbank.com/</a>`))
	if got != "" || unknownReason(err) != UnknownPhishingAllowlist {
		t.Fatalf("pair the rule could allow: %q, %v", got, err)
	}
	// It cannot: the displayed host is not rbcroyalbank.
	htmlMatch(t, d, `<a href="https://secure.rbc.com/">https://www.example.com/</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	// It cannot: the real URL lacks the literal "rbc".
	htmlMatch(t, d, `<a href="https://evil.test/">https://www.rbcroyalbank.com/</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
}

func TestRegexWhoseTriggerHoldsWithoutMatchesIsReachedEverywhere(t *testing.T) {
	// The trigger 0=0 holds on an object without the marker, so the
	// reference runs the regex there.
	d := compileTest(t, "ldb", "Rx.Zero;Engine:81-255,Target:0;0|1;414243;0=0/xyz/\n")
	expectOutcome(t, d, "has xyz but no marker", anyType, "Rx.Zero", "")
	expectOutcome(t, d, "neither here", anyType, "", "")
	expectOutcome(t, d, "ABC marker", anyType, "Rx.Zero", "")
	// A regex this engine does not evaluate is unknown there, not false.
	d = compileTest(t, "ldb", "Rx.Zero;Engine:81-255,Target:0;0|1;414243;0=0/xyz(?R)/\n")
	expectOutcome(t, d, "has xyz but no marker", anyType, "", UnknownPCRE)
	expectOutcome(t, d, "ABC marker", anyType, "Rx.Zero", "")
}
