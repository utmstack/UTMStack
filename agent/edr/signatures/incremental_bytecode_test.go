package signatures

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func compareBytecodeRefresh(t *testing.T, files map[string][]byte, prior *Database) *Database {
	t.Helper()
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("CBC member refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || next.bytecodeBytes != full.bytecodeBytes ||
		next.bytecodeDisabled != full.bytecodeDisabled ||
		next.hookCount != full.hookCount || !reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) ||
		!reflect.DeepEqual(next.unknownHooks, full.unknownHooks) {
		t.Fatalf("incremental CBC runtime/report differs from full compilation: next=%+v full=%+v", next.Report(), full.Report())
	}
	return next
}

func TestCVDBytecodeMemberRefreshReplacesLogicalAndPDFHooks(t *testing.T) {
	logical := callbackWithLabel([]byte("Old.CBC.Label"), true)
	pdf := callbackWithLabel([]byte("Old.PDF.Label"), true)
	pdf.kind, pdf.logical = 258, ""
	members := []member{
		{"logical.cbc", logical.encode()},
		{"pdf.cbc", pdf.encode()},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	logical = callbackWithLabel([]byte("New.CBC.Label"), true)
	pdf = callbackWithLabel([]byte("New.PDF.Label"), true)
	pdf.kind, pdf.logical = 258, ""
	members[0].data, members[1].data = logical.encode(), pdf.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if len(next.pdfHooks) != 1 || len(prior.pdfHooks) != 1 || next.pdfHooks[0] == prior.pdfHooks[0] ||
		len(next.bytecodeSources) != 2 {
		t.Fatal("PDF program or CBC source provenance was not replaced")
	}
	expectOutcome(t, prior, "xxAByy", anyType, "Old.CBC.Label", "")
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
	expectOutcome(t, prior, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "Old.PDF.Label", "")
	expectOutcome(t, next, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "New.PDF.Label", "")
	assertMatch(t, next, []byte("MNOPQR"), "Kept.Rule")
}

func TestCVDBytecodeMemberRefreshRestoresPEHookOrder(t *testing.T) {
	members := []member{
		{"first.cbc", returningHook(259, 0).encode()},
		{"middle.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
		{"second.cbc", returningHook(257, 0).encode()},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	changed := returningHook(259, 1)
	members[0].data = changed.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if next.hookCount != 2 || next.rules[0].hookOrder != 0 || next.rules[2].hookOrder != 1 ||
		prior.rules[0].hookOrder != 0 || prior.rules[2].hookOrder != 1 {
		t.Fatal("PE hook order changed across generation boundaries")
	}
	for _, data := range [][]byte{[]byte("MNOPQR"), []byte("ordinary clean bytes")} {
		got, gotWork, gotErr := next.matchCounted(context.Background(), data, 100_000)
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		want, wantWork, wantErr := full.matchCounted(context.Background(), data, 100_000)
		if got != want || gotWork != wantWork || !reflect.DeepEqual(gotErr, wantErr) {
			t.Fatalf("%q: incremental=(%q,%d,%v) full=(%q,%d,%v)", data, got, gotWork, gotErr, want, wantWork, wantErr)
		}
	}
}

func TestCVDBytecodeMemberRefreshRestoresBareAndTriggeredHookOrder(t *testing.T) {
	bare := callbackWithLabel([]byte("Old.Bare.Hook"), true)
	bare.kind, bare.logical = 261, ""
	triggered := callbackWithLabel([]byte("Triggered.Hook"), true)
	triggered.kind, triggered.logical = 261, "Triggered.Hook;Engine:51-255,Target:6;0;4142"
	members := []member{{"bare.cbc", bare.encode()}, {"triggered.cbc", triggered.encode()}}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	bare = callbackWithLabel([]byte("New.Bare.Hook"), true)
	bare.kind, bare.logical = 261, ""
	members[0].data = bare.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if next.DetectionCount() != 2 || len(next.bareHooks[6]) != 1 || next.bareHooks[6][0].hookOrder != 0 ||
		len(next.rules) != 1 || next.rules[0].hookOrder != 1 || prior.bareHooks[6][0].hookOrder != 0 {
		t.Fatalf("bare/triggered order across refresh: prior=%+v next=%+v", prior.bareHooks[6], next.bareHooks[6])
	}
	elf := elf64(binary.LittleEndian, 1, 0, nil, 1, 64)
	copy(elf[len(elf)-2:], "AB")
	for _, tc := range []struct {
		d    *Database
		want string
	}{{prior, "Old.Bare.Hook"}, {next, "New.Bare.Hook"}} {
		res, err := tc.d.MatchObject(context.Background(), elf, MatchContext{Type: TypeELF})
		if err != nil || res.Name != tc.want {
			t.Fatalf("refresh hook detection: %+v, %v, want %s", res, err, tc.want)
		}
	}
}

func TestCVDBytecodeMemberRefreshDropsUnsupportedHook(t *testing.T) {
	unsupported := callbackWithLabel([]byte("Unused.PDF.Label"), true)
	unsupported.kind, unsupported.logical = 258, ""
	unsupported.apiMax = 53
	unsupported.apis = append(unsupported.apis, callbackAPI(53, 70, "atoi"))
	members := []member{
		{"pdf.cbc", unsupported.encode()},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if prior.unknownHooks[10] != UnknownBytecode {
		t.Fatal("fixture lacks an unsupported PDF hook")
	}
	supported := callbackWithLabel([]byte("New.PDF.Label"), true)
	supported.kind, supported.logical = 258, ""
	members[0].data = supported.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if next.unknownHooks[10] != "" || len(next.pdfHooks) != 1 {
		t.Fatal("old unsupported PDF hook survived replacement")
	}
	expectOutcome(t, prior, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "", UnknownBytecode)
	expectOutcome(t, next, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "New.PDF.Label", "")
	members[0].data = unsupported.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	back := compareBytecodeRefresh(t, files, next)
	if back.unknownHooks[10] != UnknownBytecode || len(back.pdfHooks) != 0 {
		t.Fatal("returning unsupported hook was not restored")
	}
}

func TestCVDMixedBytecodeRuleAndHashRefreshMatchesFullCompile(t *testing.T) {
	logical := callbackWithLabel([]byte("Old.CBC.Label"), true)
	members := []member{
		{"logical.cbc", logical.encode()},
		{"rule.ndb", []byte("Old.Rule:0:*:4f4c444d41524b\n")},
		{"hash.hdb", hashLine([]byte("old harmless hash target"), "Old.Hash")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	logical = callbackWithLabel([]byte("New.CBC.Label"), true)
	members[0].data = logical.encode()
	members[1].data = []byte("New.Rule:0:*:4e45574d41524b\n")
	members[2].data = hashLine([]byte("new harmless hash target"), "New.Hash")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	expectOutcome(t, prior, "xxAByy", anyType, "Old.CBC.Label", "")
	expectOutcome(t, next, "xxAByy", anyType, "New.CBC.Label", "")
	assertMatch(t, next, []byte("OLDMARK"), "")
	assertMatch(t, next, []byte("NEWMARK"), "New.Rule")
	assertMatch(t, next, []byte("old harmless hash target"), "")
	assertMatch(t, next, []byte("new harmless hash target"), "New.Hash")
}

func TestCVDBytecodeMemberRefreshFallsBackWhenSelectionChanges(t *testing.T) {
	f := callbackWithLabel([]byte("Selected.Label"), true)
	f.logical = "Selected.Trigger;Engine:51-255,Target:0;0;4142"
	members := []member{
		{"daily.ign2", []byte("Ignored.Trigger\n")},
		{"logical.cbc", f.encode()},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	f.logical = "Ignored.Trigger;Engine:51-255,Target:0;0;4142"
	members[1].data = f.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("CBC selection change did not fall back: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil || full.Count() != prior.Count()-1 || full.Report().Skipped[SkipIgnoreList] != 1 {
		t.Fatalf("full compile after CBC selection change: count=%d report=%+v err=%v", full.Count(), full.Report(), err)
	}
}

func TestCVDBytecodeMemberRefreshKeepsOtherUnknownHook(t *testing.T) {
	unsupported := callbackWithLabel([]byte("Unused.PDF.Label"), true)
	unsupported.kind, unsupported.logical = 258, ""
	unsupported.apiMax = 53
	unsupported.apis = append(unsupported.apis, callbackAPI(53, 70, "atoi"))
	members := []member{
		{"first.cbc", unsupported.encode()},
		{"second.cbc", unsupported.encode()},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	supported := callbackWithLabel([]byte("Supported.PDF.Label"), true)
	supported.kind, supported.logical = 258, ""
	members[0].data = supported.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	first := compareBytecodeRefresh(t, files, prior)
	if first.unknownHooks[10] != UnknownBytecode || first.Report().Unknown[UnknownBytecode] != 1 || len(first.pdfHooks) != 1 {
		t.Fatalf("unchanged unsupported hook was lost: %+v", first.Report())
	}
	members[1].data = supported.encode()
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	second := compareBytecodeRefresh(t, files, first)
	if second.unknownHooks[10] != "" || second.Report().Unknown[UnknownBytecode] != 0 || len(second.pdfHooks) != 2 ||
		prior.unknownHooks[10] != UnknownBytecode || first.unknownHooks[10] != UnknownBytecode {
		t.Fatal("unsupported hook state did not follow the correct generation")
	}
}

func TestOfficialCVDBytecodeMemberRefreshUsesIncrementalPath(t *testing.T) {
	dir, changedPath := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR"), os.Getenv("EDR_OFFICIAL_BYTECODE_NEXT")
	if dir == "" || changedPath == "" {
		t.Skip("set EDR_OFFICIAL_SIGNATURE_DIR and EDR_OFFICIAL_BYTECODE_NEXT for official CVD delta test")
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
	files["bytecode.cvd"], err = os.ReadFile(changedPath)
	if err != nil {
		t.Fatal(err)
	}
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("official CBC member did not take incremental path: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || next.bytecodeBytes != full.bytecodeBytes ||
		!reflect.DeepEqual(next.Report(), full.Report()) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatal("official incremental CBC member differs from full compile")
	}
}
