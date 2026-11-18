package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunExitCodesAndGate(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "db")
	if err := os.Mkdir(db, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, db, "a.ndb", "Good:0:*:41424344\nMacro:0:$0$:41424344\nMacro2:0:$0$:45464748\n")
	writeFile(t, db, "b.hdb", "44d88612fea8a8f36de82e1278abb02f:68:Eicar.Test\n")
	report := filepath.Join(dir, "report.json")
	var out bytes.Buffer
	code, err := run(options{Inputs: []string{db}, ReferenceLevel: 216, JSON: report, Examples: 1, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 2 {
		t.Fatalf("incomplete feed: code %d err %v\n%s", code, err, out.String())
	}
	if !strings.Contains(out.String(), `invalid signature offset`) || !strings.Contains(out.String(), "TOTAL") {
		t.Fatalf("summary missing reason or totals:\n%s", out.String())
	}
	var r signatures.AuditReport
	b, err := os.ReadFile(report)
	if err != nil || json.Unmarshal(b, &r) != nil {
		t.Fatalf("json report: %v", err)
	}
	if r.Totals.Accepted != 2 || r.Totals.Rejected != 2 {
		t.Fatalf("json totals = %+v", r.Totals)
	}

	reason := r.Reasons[0].Reason
	gate := func(max int) string {
		g, _ := json.Marshal(Gate{Approved: []Approval{{Format: ".ndb", Reason: reason, Max: max, Note: "macro offsets are planned"}}})
		return writeFile(t, dir, "gate.json", string(g))
	}
	out.Reset()
	code, err = run(options{Inputs: []string{db}, ReferenceLevel: 216, Gate: gate(2), Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 0 || !strings.Contains(out.String(), "release gate: PASS") {
		t.Fatalf("approved gap: code %d err %v\n%s", code, err, out.String())
	}
	out.Reset()
	code, err = run(options{Inputs: []string{db}, ReferenceLevel: 216, Gate: gate(1), Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 2 || !strings.Contains(out.String(), "release gate: FAIL") {
		t.Fatalf("gap above its approval: code %d err %v\n%s", code, err, out.String())
	}
	prefix, _ := json.Marshal(Gate{Approved: []Approval{{Format: "*", Reason: "invalid signature offset*", Max: -1}}})
	out.Reset()
	code, err = run(options{Inputs: []string{db}, ReferenceLevel: 216, Gate: writeFile(t, dir, "prefix.json", string(prefix)), Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 0 {
		t.Fatalf("prefix approval: code %d err %v\n%s", code, err, out.String())
	}
}

func TestRunReportsSkippedRulesByReason(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.ndb", "Good:0:*:41424344\nIgnored:0:*:41424344\nOld:0:*:41424344:1:80\n")
	writeFile(t, dir, "a.ign2", "Ignored\n")
	writeFile(t, dir, "a.ndu", "PUA.Rule:0:*:41424344\n")
	var out bytes.Buffer
	code, err := run(options{Inputs: []string{dir}, ReferenceLevel: 216, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 0 {
		t.Fatalf("code %d err %v\n%s", code, err, out.String())
	}
	for _, want := range []string{"reference-skipped rules by reason:", "1  ignore-list", "1  level", "1  pua"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("summary lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "dynamic configuration differs") {
		t.Fatalf("default configuration reported as changed:\n%s", out.String())
	}
}

func TestDefaultReferenceLevelIncludesPinnedClamAVWindow(t *testing.T) {
	if defaultReferenceLevel != 234 {
		t.Fatalf("default reference level = %d, want pinned 1.5.4 level 234", defaultReferenceLevel)
	}
	dir := t.TempDir()
	writeFile(t, dir, "window.ndb", "Older:0:*:41424344:216\nCurrent:0:*:45464748:234\nFuture:0:*:494a4b4c:235\n")
	var out bytes.Buffer
	code, err := run(options{Inputs: []string{dir}, ReferenceLevel: defaultReferenceLevel, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 0 {
		t.Fatalf("pinned-window audit: code %d err %v\n%s", code, err, out.String())
	}
	if !strings.Contains(out.String(), "engine level 234, reference level 234") || strings.Contains(out.String(), "syntax otherwise supported") {
		t.Fatalf("level-234 rule was not selected by both engines:\n%s", out.String())
	}
}

func TestRunCompleteFeedAndInputErrors(t *testing.T) {
	dir := t.TempDir()
	good := writeFile(t, dir, "ok.ndb", "Good:0:*:41424344\n")
	var out bytes.Buffer
	if code, err := run(options{Inputs: []string{good}, ReferenceLevel: 216, Top: 5, MaxBytes: 1 << 20}, &out); err != nil || code != 0 {
		t.Fatalf("complete feed: code %d err %v", code, err)
	}
	out.Reset()
	if code, err := run(options{Inputs: []string{good}, ReferenceLevel: 216, StrictRuntime: true, Top: 5, MaxBytes: 1 << 20}, &out); err != nil || code != 0 {
		t.Fatalf("strict complete feed: code %d err %v", code, err)
	}
	if _, err := run(options{ReferenceLevel: 216, MaxBytes: 1 << 20}, &out); err == nil {
		t.Fatal("no inputs must fail")
	}
	if _, err := run(options{Inputs: []string{good}, ReferenceLevel: 216, MaxBytes: 4}, &out); err == nil {
		t.Fatal("input above the byte bound must fail")
	}
	bad := writeFile(t, dir, "bad.json", `{"approved":[{"format":".ndb","reason":"x","max":1,"unknown":true}]}`)
	if _, err := run(options{Inputs: []string{good}, ReferenceLevel: 216, Gate: bad, MaxBytes: 1 << 20}, &out); err == nil {
		t.Fatal("gate files with unknown fields must be rejected")
	}
}

func TestRunReportsUnknownWhenReachedRulesSeparately(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.ldb", "Rx;Engine:81-255,Target:0;0&1;414243;0/x(?R)/\nOk;Engine:81-255,Target:0;0;41424344\n")
	var out bytes.Buffer
	code, err := run(options{Inputs: []string{dir}, ReferenceLevel: 216, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 0 {
		t.Fatalf("code %d err %v\n%s", code, err, out.String())
	}
	if !strings.Contains(out.String(), "unknown when reached") || !strings.Contains(out.String(), "pcre") {
		t.Fatalf("summary does not report unknown-when-reached rules:\n%s", out.String())
	}
	out.Reset()
	code, err = run(options{Inputs: []string{dir}, ReferenceLevel: 216, StrictRuntime: true, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 2 || !strings.Contains(out.String(), "strict runtime gate: FAIL") {
		t.Fatalf("strict runtime gate: code %d err %v\n%s", code, err, out.String())
	}
	gatePath := writeFile(t, t.TempDir(), "gate.json", `{"approved":[]}`)
	out.Reset()
	code, err = run(options{Inputs: []string{dir}, ReferenceLevel: 216, StrictRuntime: true, Gate: gatePath, Top: 5, MaxBytes: 1 << 20}, &out)
	if err != nil || code != 2 || !strings.Contains(out.String(), "0 unapproved gap reasons, 1 unresolved runtime rules") {
		t.Fatalf("strict runtime gate with approved compilation: code %d err %v\n%s", code, err, out.String())
	}
}
