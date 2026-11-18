package signatures

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMultipleChangedCVDsRefreshInSourceOrder(t *testing.T) {
	oldCBC := callbackWithLabel([]byte("Old.CBC.Label"), true)
	bytecodeMembers := []member{{"hook.cbc", oldCBC.encode()}}
	dailyMembers := []member{
		{"rule.ndb", []byte("Old.Rule:0:*:4f4c444d41524b\n")},
		{"hash.hdb", hashLine([]byte("old harmless hash target"), "Old.Hash")},
	}
	files := map[string][]byte{
		"bytecode.cvd": cvdFixtureOrdered(t, bytecodeMembers),
		"daily.cvd":    cvdFixtureOrdered(t, dailyMembers),
		"local.ndb":    []byte("Kept.Local:0:*:4d4e4f505152\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	newCBC := callbackWithLabel([]byte("New.CBC.Label"), true)
	bytecodeMembers[0].data = newCBC.encode()
	dailyMembers[0].data = []byte("New.Rule:0:*:4e45574d41524b\n")
	dailyMembers[1].data = hashLine([]byte("new harmless hash target"), "New.Hash")
	files["bytecode.cvd"] = cvdFixtureOrdered(t, bytecodeMembers)
	files["daily.cvd"] = cvdFixtureOrdered(t, dailyMembers)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("multiple CVD refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || next.bytecodeBytes != full.bytecodeBytes ||
		!reflect.DeepEqual(next.Report(), full.Report()) || !reflect.DeepEqual(next.sourceStats, full.sourceStats) ||
		!reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatalf("two-CVD incremental/full mismatch: next=%+v full=%+v", next.Report(), full.Report())
	}
	expectOutcome(t, prior, "xxAByy", anyType, "Old.CBC.Label", "")
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
	assertMatch(t, prior, []byte("OLDMARK"), "Old.Rule")
	assertMatch(t, next, []byte("OLDMARK"), "")
	assertMatch(t, next, []byte("NEWMARK"), "New.Rule")
	assertMatch(t, next, []byte("old harmless hash target"), "")
	assertMatch(t, next, []byte("new harmless hash target"), "New.Hash")
	assertMatch(t, next, []byte("MNOPQR"), "Kept.Local")
}

func TestMultipleChangedCVDDirectoryRefreshKeepsLastGood(t *testing.T) {
	dir := t.TempDir()
	oldCBC := callbackWithLabel([]byte("Old.CBC.Label"), true)
	bytecodeMembers := []member{{"hook.cbc", oldCBC.encode()}}
	dailyMembers := []member{{"rule.ndb", []byte("Old.Rule:0:*:4f4c444d41524b\n")}}
	bytecodePath, dailyPath := filepath.Join(dir, "bytecode.cvd"), filepath.Join(dir, "daily.cvd")
	if err := os.WriteFile(bytecodePath, cvdFixtureOrdered(t, bytecodeMembers), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dailyPath, cvdFixtureOrdered(t, dailyMembers), 0o600); err != nil {
		t.Fatal(err)
	}
	prior, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	newCBC := callbackWithLabel([]byte("New.CBC.Label"), true)
	bytecodeMembers[0].data = newCBC.encode()
	dailyMembers[0].data = []byte("New.Rule:0:*:4e45574d41524b\n")
	if err := os.WriteFile(bytecodePath, cvdFixtureOrdered(t, bytecodeMembers), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dailyPath, cvdFixtureOrdered(t, dailyMembers), 0o600); err != nil {
		t.Fatal(err)
	}
	next, err := LoadWithPrior(dir, Limits{}, prior)
	if err != nil {
		t.Fatal(err)
	}
	full, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) || next.Version() != full.Version() {
		t.Fatalf("two-CVD directory/full mismatch: next=%+v full=%+v", next.Report(), full.Report())
	}
	expectOutcome(t, prior, "xxAByy", anyType, "Old.CBC.Label", "")
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
	if err := os.WriteFile(dailyPath, []byte("broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWithPrior(dir, Limits{}, next); err == nil {
		t.Fatal("invalid second CVD displaced the last-good generation")
	}
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
	dailyMembers = append(dailyMembers, member{"extra.ndb", []byte("Added.Rule:0:*:41444445444d41524b\n")})
	if err := os.WriteFile(dailyPath, cvdFixtureOrdered(t, dailyMembers), 0o600); err != nil {
		t.Fatal(err)
	}
	fallback, err := LoadWithPrior(dir, Limits{}, next)
	if err != nil {
		t.Fatal(err)
	}
	full, err = Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if fallback.Version() != full.Version() || !reflect.DeepEqual(fallback.Report(), full.Report()) {
		t.Fatal("changed CVD member set did not fall back to the full compiler")
	}
	assertMatch(t, fallback, []byte("ADDEDMARK"), "Added.Rule")
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
}

func TestOfficialMultipleChangedCVDsUseIncrementalPath(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	dailyNext := os.Getenv("EDR_OFFICIAL_DAILY_NEXT")
	bytecodeNext := os.Getenv("EDR_OFFICIAL_BYTECODE_NEXT")
	if dir == "" || dailyNext == "" || bytecodeNext == "" {
		t.Skip("set official signature directory and both repacked CVD paths")
	}
	files := make(map[string][]byte, 3)
	for _, name := range []string{"main.cvd", "daily.cvd", "bytecode.cvd"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		files[name] = data
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"daily.cvd": dailyNext, "bytecode.cvd": bytecodeNext} {
		files[name], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("official two-CVD delta did not take incremental path: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || next.bytecodeBytes != full.bytecodeBytes ||
		!reflect.DeepEqual(next.Report(), full.Report()) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatal("official two-CVD incremental generation differs from full compile")
	}
}
