package signatures

import (
	"strings"
	"testing"
)

func auditFixture(t *testing.T, files map[string][]byte, opts AuditOptions) *AuditReport {
	t.Helper()
	if opts.ReferenceLevel == 0 {
		opts.ReferenceLevel = FunctionalityLevel
	}
	r, err := Audit(files, opts)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func reasonCount(r *AuditReport, format, contains string) int {
	n := 0
	for _, reason := range r.Reasons {
		if reason.Format == format && strings.Contains(reason.Reason, contains) {
			n += reason.Count
		}
	}
	return n
}

func TestAuditCountsEveryRuleWithoutStopping(t *testing.T) {
	var ndb strings.Builder
	for i := 0; i < 150; i++ { // more than the compiler's 100-diagnostic cap
		ndb.WriteString("Bad.Macro:0:$0$:41424344\n")
	}
	ndb.WriteString("Good:0:*:41424344\n# comment\n\n")
	r := auditFixture(t, map[string][]byte{"a.ndb": []byte(ndb.String())}, AuditOptions{})
	if r.Totals.Rules != 151 || r.Totals.Accepted != 1 || r.Totals.Rejected != 150 {
		t.Fatalf("totals = %+v", r.Totals)
	}
	if got := reasonCount(r, ".ndb", `invalid signature offset`); got != 150 {
		t.Fatalf("offset reason count = %d, reasons %+v", got, r.Reasons)
	}
	if r.Reasons[0].Features["offset:macro"] != 150 {
		t.Fatalf("features = %+v", r.Reasons[0].Features)
	}
	if len(r.Reasons[0].Examples) != 3 || r.Reasons[0].Examples[0].Line != 1 {
		t.Fatalf("examples = %+v", r.Reasons[0].Examples)
	}
	if r.Complete() {
		t.Fatal("an audit with rejected rules must not be complete")
	}
	if _, err := Compile(map[string][]byte{"a.ndb": []byte(ndb.String())}); err == nil {
		t.Fatal("the audit must not change all-or-nothing compilation")
	}
}

func TestAuditFunctionalityWindows(t *testing.T) {
	ndb := strings.Join([]string{
		"Retired:0:*:41424344:1:80",      // neither engine selects it
		"Future:0:*:41424344:235",        // neither engine selects it
		"NewSyntax:0:*:41424344:150",     // both select it
		"NewMacro:0:$0$:41424344:150",    // both select it; a real syntax gap
		"OldOnly:0:*:41424344:51:100",    // retired before this engine's level
		"OldMacro:0:$0$:41424344:51:100", // idem, with a syntax gap
		"Plain:0:*:41424344",
	}, "\n")
	r := auditFixture(t, map[string][]byte{"w.ndb": []byte(ndb)}, AuditOptions{})
	want := AuditCounts{Rules: 7, Accepted: 2, Rejected: 1, ReferenceSkipped: 4}
	if r.Totals != want || r.SkippedByReason[SkipLevel] != 4 || r.Files[0].SkippedByReason[SkipLevel] != 4 {
		t.Fatalf("totals = %+v skipped %v, want %+v", r.Totals, r.SkippedByReason, want)
	}
	// Against an older reference, rules selected by only one engine differ.
	r = auditFixture(t, map[string][]byte{"w.ndb": []byte(ndb)}, AuditOptions{ReferenceLevel: 100})
	want = AuditCounts{Rules: 7, Accepted: 2, Rejected: 2, ReferenceSkipped: 3, AcceptedOutsideReference: 1}
	if r.Totals != want {
		t.Fatalf("totals = %+v, want %+v", r.Totals, want)
	}
	if reasonCount(r, ".ndb", "syntax otherwise supported") != 1 {
		t.Fatalf("level-only rule not isolated: %+v", r.Reasons)
	}
	if reasonCount(r, ".ndb", `invalid signature offset`) != 1 {
		t.Fatalf("level-gated syntax gap not reported by its syntax: %+v", r.Reasons)
	}
}

// The audit applies the compiler's own selection: ignore lists from another
// container, dynamic configuration, bytecode header windows and PUA files.
func TestAuditUsesCompilerSelection(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.Old"), true)
	f.maximum = 100
	daily := cvdFixtureOrdered(t, []member{
		{"daily.cfg", []byte("# comment\nDOCUMENT:0x5:11:13\nPE:0x0:235:255\n")},
		{"daily.ign2", []byte("Main.Ignored\n")},
		{"daily.ndu", []byte("PUA.Rule:0:*:41424344\n")},
		{"daily.ndb", []byte("Daily.Rule:0:*:41424344\n")},
	})
	main := cvdFixtureOrdered(t, []member{{"main.ndb", []byte("Main.Ignored:0:*:41424344\nMain.Kept:0:*:45464748\n")}})
	bc := cvdFixtureOrdered(t, []member{{"1.cbc", f.encode()}})
	r := auditFixture(t, map[string][]byte{"daily.cvd": daily, "main.cvd": main, "bytecode.cvd": bc}, AuditOptions{})
	if !r.Complete() || r.Totals.Accepted != 3 {
		t.Fatalf("totals = %+v reasons %+v errors %v", r.Totals, r.Reasons, r.Errors)
	}
	want := map[string]int{SkipIgnoreList: 1, SkipLevel: 3, SkipPUA: 1}
	for reason, n := range want {
		if r.SkippedByReason[reason] != n {
			t.Fatalf("skipped by reason = %v, want %v", r.SkippedByReason, want)
		}
	}
	if r.DynamicConfig != DefaultDynamicConfig() {
		t.Fatalf("dynamic configuration = %+v", r.DynamicConfig)
	}
	// The compiler selects the same rules.
	d, err := Compile(map[string][]byte{"daily.cvd": daily, "main.cvd": main, "bytecode.cvd": bc})
	if err != nil {
		t.Fatal(err)
	}
	skipped := d.Report().Skipped
	if skipped[SkipIgnoreList] != 1 || skipped[SkipLevel] != 1 || skipped[SkipPUA] != 1 || d.Count() != 2 {
		t.Fatalf("compiler skipped %v loaded %d", skipped, d.Count())
	}
}

func TestAuditLogicalEngineWindowAndFeatures(t *testing.T) {
	ldb := strings.Join([]string{
		"Pcre;Engine:81-255,Target:0;0&1;41424344;0/abc[0-9]+/i",
		"Macro;Engine:81-255,Target:0;0&1;41424344;${6-7}0$",
		"Icon;Engine:81-255,Target:1,IconGroup1:x;0;41424344",
		"Old;Engine:51-80,Target:0;0;41424344",
		"Ok;Engine:81-255,Target:0;0;41424344",
	}, "\n")
	r := auditFixture(t, map[string][]byte{"l.ldb": []byte(ldb)}, AuditOptions{})
	// The regex and icon-group rules load as unknown when reached.
	if r.Totals.Accepted != 3 || r.Totals.Rejected != 1 || r.Totals.ReferenceSkipped != 1 {
		t.Fatalf("totals = %+v", r.Totals)
	}
	features := map[string]int{}
	for _, reason := range r.Reasons {
		for f, n := range reason.Features {
			features[f] += n
		}
	}
	if features["subsig:macro"] != 1 {
		t.Fatalf("feature subsig:macro = %d in %+v", features["subsig:macro"], features)
	}
}

func TestAuditUnsupportedFormatsAndPUA(t *testing.T) {
	files := map[string][]byte{
		"x.cdb":  []byte("Name:ZIP:*:evil\\.exe:*:*:*:*:*:*\n"),
		"x.ndu":  []byte("PUA.Test:0:*:41424344\n"),
		"x.ign2": []byte("Some.Signature\n"),
		"x.hdb":  []byte("44d88612fea8a8f36de82e1278abb02f:68:Eicar.Test\n"),
	}
	r := auditFixture(t, files, AuditOptions{})
	if r.ByKind[KindPUA].ReferenceSkipped != 1 || r.ByKind[KindPUA].Rejected != 0 {
		t.Fatalf("PUA without reference PUA: %+v", r.ByKind[KindPUA])
	}
	if r.ByKind[KindDetection].Rejected != 1 || r.ByKind[KindDetection].Accepted != 1 {
		t.Fatalf("detections: %+v", r.ByKind[KindDetection])
	}
	if r.ByKind[KindIgnore].Accepted != 1 || r.ByKind[KindIgnore].Rejected != 0 {
		t.Fatalf("ignore list: %+v", r.ByKind[KindIgnore])
	}
	r = auditFixture(t, files, AuditOptions{ReferencePUA: true})
	if r.ByKind[KindPUA].Rejected != 1 || reasonCount(r, ".ndu", "potentially unwanted application databases") != 1 {
		t.Fatalf("PUA with reference PUA: %+v %+v", r.ByKind[KindPUA], r.Reasons)
	}
}

func TestAuditContainerMembersAndLimits(t *testing.T) {
	cvd := cvdFixture(t, map[string][]byte{
		"daily.ndb":  []byte("Test:0:*:41424344\nBad:0:$0$:41424344\n"),
		"daily.info": []byte("metadata\n"),
		"COPYING":    []byte("license text\n"),
	})
	r := auditFixture(t, map[string][]byte{"daily.cvd": cvd}, AuditOptions{})
	if len(r.Containers) != 1 || !r.Containers[0].IntegrityVerified || r.Containers[0].Authenticated {
		t.Fatalf("containers = %+v", r.Containers)
	}
	if len(r.Files) != 1 || r.Files[0].Name != "daily.cvd/daily.ndb" || r.Files[0].Accepted != 1 || r.Files[0].Rejected != 1 {
		t.Fatalf("files = %+v", r.Files)
	}
	if r.ExpandedBytes <= r.InputBytes {
		t.Fatalf("expanded bytes not measured: %d <= %d", r.ExpandedBytes, r.InputBytes)
	}
	corrupt := append([]byte(nil), cvd...)
	corrupt[len(corrupt)-1] ^= 0xff
	r = auditFixture(t, map[string][]byte{"daily.cvd": corrupt}, AuditOptions{})
	if len(r.Errors) != 1 || r.Complete() {
		t.Fatalf("corrupt container must be an audit error: %+v", r.Errors)
	}
	big := strings.Repeat("Test:0:*:41424344\n", 10)
	r = auditFixture(t, map[string][]byte{"big.ndb": []byte(big)}, AuditOptions{Limits: Limits{MaxSignatures: 5}})
	if len(r.LimitFindings) != 0 {
		t.Fatalf("limit findings use production defaults, got %v", r.LimitFindings)
	}
}

func TestAuditBytecodeHookKinds(t *testing.T) {
	f := callbackWithLabel([]byte("Test.Label"), true)
	f.kind = 257
	r := auditFixture(t, map[string][]byte{"u.cbc": f.encode()}, AuditOptions{})
	// A hook program this engine cannot run loads as unknown when reached.
	if r.Totals.Accepted != 1 || r.Totals.Rejected != 0 {
		t.Fatalf("totals = %+v reasons %+v", r.Totals, r.Reasons)
	}
	for _, reason := range r.Reasons {
		if strings.Contains(strings.ToLower(reason.Reason), "clam") {
			t.Fatalf("reason leaks engine branding: %q", reason.Reason)
		}
	}
}

func TestAuditRequiresReferenceLevel(t *testing.T) {
	if _, err := Audit(map[string][]byte{"a.ndb": []byte("A:0:*:41")}, AuditOptions{}); err == nil {
		t.Fatal("missing reference level must fail")
	}
}

func TestAuditCountsUnknownWhenReachedRulesByReason(t *testing.T) {
	// The first regex is evaluated; the second uses recursion, which this
	// engine leaves unknown when reached.
	ldb := "Rx;Engine:81-255,Target:0;0&1;414243;0/x/\nRx2;Engine:81-255,Target:0;0&1;414243;0/x(?R)/\nIcon;Engine:81-255,Target:1,IconGroup1:x;0;41424344\nOk;Engine:81-255,Target:0;0;41424344\n"
	r := auditFixture(t, map[string][]byte{"l.ldb": []byte(ldb)}, AuditOptions{})
	if r.Totals.Accepted != 4 || r.Totals.Unknown != 1 || r.UnknownByReason[UnknownPCRE] != 1 || r.UnknownByReason[UnknownIconGroup] != 0 {
		t.Fatalf("totals = %+v unknown by reason %v", r.Totals, r.UnknownByReason)
	}
	if !r.Complete() {
		t.Fatal("unknown-when-reached rules are loaded; the audit is complete")
	}
}
