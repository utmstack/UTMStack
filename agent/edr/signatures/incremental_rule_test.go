package signatures

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNDBDirectoryRefreshStreamsUnchangedSources(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.ndb"), []byte("Old.Rule:0:*:414243444546\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.ndb"), []byte("Kept.Rule:0:*:4d4e4f505152\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prior, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.ndb"), []byte("New.Rule:0:*:4748494a4b4c\n"), 0o600); err != nil {
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
	if next.Version() != full.Version() || !reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatal("streamed rule refresh differs from full compiler")
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
	assertMatch(t, next, []byte("ABCDEF"), "")
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
}

func TestNDBRefreshSingleSource(t *testing.T) {
	files := map[string][]byte{"feed.ndb": []byte("Old.Rule:0:*:414243444546\n")}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["feed.ndb"] = []byte("New.Rule:0:*:4748494a4b4c\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
}

func TestTarget13RefreshUpdatesLegacyMetadataPass(t *testing.T) {
	files := map[string][]byte{"feed.ndb": []byte("Metadata.Marker:13:*:414243444546\n")}
	prior, err := Compile(files)
	if err != nil || !prior.HasLegacyMetadataRules() {
		t.Fatalf("compile target 13: present=%v err=%v", err == nil && prior.HasLegacyMetadataRules(), err)
	}
	files["feed.ndb"] = []byte("Ordinary.Marker:0:*:414243444546\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled || next.HasLegacyMetadataRules() || !prior.HasLegacyMetadataRules() {
		t.Fatalf("remove target 13: handled=%v next=%v prior=%v err=%v", handled, next != nil && next.HasLegacyMetadataRules(), prior.HasLegacyMetadataRules(), err)
	}
	files["feed.ndb"] = []byte("Metadata.New:13:*:4748494a4b4c\n")
	readded, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: next})
	if err != nil || !handled || !readded.HasLegacyMetadataRules() {
		t.Fatalf("restore target 13: handled=%v present=%v err=%v", handled, readded != nil && readded.HasLegacyMetadataRules(), err)
	}
}

func TestNDBRefreshMiddleSourceExpansionAndContraction(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("First.Kept:0:*:414141414141\n"),
		"b.ndb": []byte("Middle.Old1:0:*:424242424242\nMiddle.Old2:0:*:434343434343\n"),
		"c.ndb": []byte("Last.Kept:0:*:444444444444\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		lines, marker, signature string
	}{
		{"Middle.New1:0:*:454545454545\nMiddle.New2:0:*:464646464646\nMiddle.New3:0:*:474747474747\n", "GGGGGG", "Middle.New3"},
		{"Middle.Final:0:*:484848484848\n", "HHHHHH", "Middle.Final"},
	} {
		files["b.ndb"] = []byte(tc.lines)
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
		if err != nil || !handled {
			t.Fatalf("middle source refresh: handled=%v err=%v", handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if next.Version() != full.Version() || !reflect.DeepEqual(next.Report(), full.Report()) {
			t.Fatalf("middle source report differs from full compile: next=%+v full=%+v", next.Report(), full.Report())
		}
		for _, item := range []struct{ data, name string }{
			{"AAAAAA", "First.Kept"}, {"DDDDDD", "Last.Kept"},
			{tc.marker, tc.signature}, {"BBBBBB", ""}, {"CCCCCC", ""},
		} {
			assertMatch(t, next, []byte(item.data), item.name)
			assertMatch(t, full, []byte(item.data), item.name)
		}
		assertMatch(t, prior, []byte("AAAAAA"), "First.Kept")
		assertMatch(t, prior, []byte("DDDDDD"), "Last.Kept")
		prior = next
	}
}

func TestNDBRefreshUsesFullCompilerNearSignatureLimit(t *testing.T) {
	limits := Limits{MaxSignatures: 2}.defaults()
	files := map[string][]byte{
		"a.ign": []byte("Ignored.Rule\n"),
		"b.ndb": []byte("First.Old:0:*:414243444546\n"),
		"c.ndb": []byte("Later.Kept:0:*:4d4e4f505152\n"),
	}
	prior, err := CompileWithLimits(files, limits)
	if err != nil {
		t.Fatal(err)
	}
	files["b.ndb"] = []byte("First.New:0:*:4748494a4b4c\nIgnored.Rule:0:*:535455565758\n")
	if _, handled, err := tryRuleSourceRefresh(files, limits, compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("near-limit rule refresh bypassed full compiler: handled=%v err=%v", handled, err)
	}
	next, err := compileWithLimits(files, limits, compileOptions{prior: prior})
	if err != nil {
		t.Fatal(err)
	}
	full, err := CompileWithLimits(files, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatal("near-limit rule fallback differs from full compiler")
	}
	assertMatch(t, next, []byte("GHIJKL"), "First.New")
}

func TestNDBRefreshMatchesFullCompilerAndKeepsPriorImmutable(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("Old.One:0:*:414243444546\n"),
		"b.hdb": hashLine([]byte("hash malware fixture"), "Hash.Kept"),
		"c.ndb": []byte("Other.Kept:0:*:4d4e4f505152\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("New.One:0:*:4748494a4b4c\nNew.Two:0:*:535455565758\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || !reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatalf("incremental report=%+v full=%+v", next.Report(), full.Report())
	}
	for _, tc := range []struct {
		data        string
		before, now string
	}{
		{"ABCDEF", "Old.One", ""},
		{"GHIJKL", "", "New.One"},
		{"STUVWX", "", "New.Two"},
		{"MNOPQR", "Other.Kept", "Other.Kept"},
		{"hash malware fixture", "Hash.Kept", "Hash.Kept"},
	} {
		assertMatch(t, prior, []byte(tc.data), tc.before)
		assertMatch(t, next, []byte(tc.data), tc.now)
		assertMatch(t, full, []byte(tc.data), tc.now)
	}
	if next.sourceStats["c.ndb"].ruleStart != full.sourceStats["c.ndb"].ruleStart {
		t.Fatal("later source rule range did not shift with replacement")
	}
}

func TestNDBRefreshRepeatedUpdateDoesNotRetainStaleAnchorVerdicts(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("A.One:0:*:414243444546\n"),
		"b.ndb": []byte("B.Kept:0:*:4d4e4f505152\n"),
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		line string
		data string
		want string
	}{
		{"A.Two:0:*:4748494a4b4c\n", "GHIJKL", "A.Two"},
		{"A.Three:0:*:535455565758\n", "STUVWX", "A.Three"},
	} {
		files["a.ndb"] = []byte(tc.line)
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: d})
		if err != nil || !handled {
			t.Fatalf("repeat refresh: handled=%v err=%v", handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if next.Version() != full.Version() || !reflect.DeepEqual(next.Report(), full.Report()) {
			t.Fatal("repeated refresh differs from full compiler")
		}
		assertMatch(t, next, []byte(tc.data), tc.want)
		assertMatch(t, next, []byte("MNOPQR"), "B.Kept")
		assertMatch(t, next, []byte("ABCDEF"), "")
		d = next
	}
	assertMatch(t, d, []byte("GHIJKL"), "")
}

func TestNDBRefreshRejectsInvalidSourceWithoutChangingPrior(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("Old.One:0:*:414243444546\n"),
		"b.ndb": []byte("Other.Kept:0:*:4d4e4f505152\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("Broken:0:*:zz\n")
	if next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); !handled || err == nil || next != nil {
		t.Fatalf("invalid rule accepted: handled=%v err=%v", handled, err)
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.One")
}

func TestNDBRefreshUpdatesIgnoreAndWorkAccounting(t *testing.T) {
	files := map[string][]byte{
		"a.ign": []byte("Ignored.Rule\n"),
		"b.ndb": []byte("Old.Rule:0:*:414243444546\n"),
		"c.ndb": []byte("Other.Rule:0:*:4d4e4f505152\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["b.ndb"] = []byte("Ignored.Rule:0:*:4748494a4b4c\nNew.Rule:0:*:535455565758\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) || next.Report().Skipped[SkipIgnoreList] != 1 {
		t.Fatalf("ignore accounting: incremental=%+v full=%+v", next.Report(), full.Report())
	}
	for _, data := range [][]byte{[]byte("GHIJKL"), []byte("STUVWX"), []byte("MNOPQR"), []byte("unrelated data")} {
		a, aSteps, aErr := next.matchCounted(context.Background(), data, 10_000_000)
		b, bSteps, bErr := full.matchCounted(context.Background(), data, 10_000_000)
		if a != b || aErr != bErr || aSteps != bSteps {
			t.Fatalf("%q: incremental=(%q,%d,%v) full=(%q,%d,%v)", data, a, aSteps, aErr, b, bSteps, bErr)
		}
	}
}

func TestNDBRefreshCrossSourcePatternSharingAndWorkBudget(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("Old.Rule:0:*:414243444546\n"),
		"b.ndb": []byte("Shared.Kept:0:*:414243444546\n"),
		"c.ndb": []byte("Other.Kept:0:*:4d4e4f505152\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("New.Shared:0:*:414243444546\nNew.Unique:0:*:4748494a4b4c\n")
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{[]byte("ABCDEF"), []byte("GHIJKL"), []byte("MNOPQR"),
		[]byte("ABCDEF GHIJKL MNOPQR"), []byte("unrelated harmless content"),
		[]byte("%PDF-1.7\nABCDEF GHIJKL"), []byte("MZ\x00\x00ABCDEF GHIJKL"),
		[]byte("PK\x03\x04ABCDEF GHIJKL")} {
		for _, budget := range []int64{1, 8, 32, 10_000_000} {
			got, gotSteps, gotErr := next.matchCounted(context.Background(), data, budget)
			want, wantSteps, wantErr := full.matchCounted(context.Background(), data, budget)
			if got != want || gotSteps != wantSteps || !reflect.DeepEqual(gotErr, wantErr) {
				t.Fatalf("%q budget=%d incremental=(%q,%d,%v) full=(%q,%d,%v)",
					data, budget, got, gotSteps, gotErr, want, wantSteps, wantErr)
			}
		}
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
}

func TestNDBRefreshComposesWithHashSourceRefresh(t *testing.T) {
	oldHash, newHash := []byte("old harmless hash target"), []byte("new harmless hash target")
	files := map[string][]byte{
		"a.ndb": []byte("Old.Rule:0:*:414243444546\n"),
		"b.hdb": hashLine(oldHash, "Old.Hash"),
	}
	first, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("New.Rule:0:*:4748494a4b4c\n")
	second, err := compileWithLimits(files, DefaultLimits(), compileOptions{prior: first})
	if err != nil {
		t.Fatal(err)
	}
	files["b.hdb"] = hashLine(newHash, "New.Hash")
	third, err := compileWithLimits(files, DefaultLimits(), compileOptions{prior: second})
	if err != nil {
		t.Fatal(err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if third.Version() != full.Version() || third.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(third.Report(), full.Report()) {
		t.Fatal("composed incremental update differs from full compile")
	}
	assertMatch(t, first, []byte("ABCDEF"), "Old.Rule")
	assertMatch(t, first, oldHash, "Old.Hash")
	assertMatch(t, second, []byte("GHIJKL"), "New.Rule")
	assertMatch(t, second, oldHash, "Old.Hash")
	assertMatch(t, third, []byte("ABCDEF"), "")
	assertMatch(t, third, []byte("GHIJKL"), "New.Rule")
	assertMatch(t, third, oldHash, "")
	assertMatch(t, third, newHash, "New.Hash")
}
