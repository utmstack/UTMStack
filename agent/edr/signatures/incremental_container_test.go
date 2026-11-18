package signatures

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCVDNDBMemberRefreshMatchesFullCompile(t *testing.T) {
	containedHash := []byte("contained harmless hash target")
	members := []member{
		{"first.ndb", []byte("Old.Rule:0:*:414243444546\n")},
		{"middle.hdb", hashLine(containedHash, "Kept.Hash")},
		{"last.ndb", []byte("Later.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{
		"daily.cvd": cvdFixtureOrdered(t, members),
		"local.ndb": []byte("Local.Rule:0:*:535455565758\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = []byte("New.One:0:*:4748494a4b4c\nNew.Two:0:*:616263646566\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("CVD refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || !reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatalf("incremental report=%+v full=%+v", next.Report(), full.Report())
	}
	if next.sourceHashes["daily.cvd"] != prior.sourceHashes["daily.cvd"] {
		t.Fatal("unchanged CVD hash member was rebuilt")
	}
	if next.memberStats["daily.cvd/last.ndb"].ruleStart != full.memberStats["daily.cvd/last.ndb"].ruleStart ||
		next.sourceStats["local.ndb"].ruleStart != full.sourceStats["local.ndb"].ruleStart {
		t.Fatal("later member/source rule ranges did not shift")
	}
	for _, data := range [][]byte{[]byte("ABCDEF"), []byte("GHIJKL"), []byte("abcdef"),
		[]byte("MNOPQR"), []byte("STUVWX"), containedHash, []byte("unrelated harmless data")} {
		got, gotSteps, gotErr := next.matchCounted(context.Background(), data, 10_000_000)
		want, wantSteps, wantErr := full.matchCounted(context.Background(), data, 10_000_000)
		if got != want || gotSteps != wantSteps || !reflect.DeepEqual(gotErr, wantErr) {
			t.Fatalf("%q incremental=(%q,%d,%v) full=(%q,%d,%v)", data, got, gotSteps, gotErr, want, wantSteps, wantErr)
		}
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
	assertMatch(t, next, []byte("ABCDEF"), "")
	assertMatch(t, next, []byte("GHIJKL"), "New.One")
	// The prior snapshot must still serve its original CVD rules.
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
}

func TestCVDNDBMemberDirectoryRefreshAndInvalidUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daily.cvd")
	members := []member{
		{"first.ndb", []byte("Old.Rule:0:*:414243444546\n")},
		{"last.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	if err := os.WriteFile(path, cvdFixtureOrdered(t, members), 0o600); err != nil {
		t.Fatal(err)
	}
	prior, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = []byte("New.Rule:0:*:4748494a4b4c\n")
	if err := os.WriteFile(path, cvdFixtureOrdered(t, members), 0o600); err != nil {
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
		t.Fatal("directory CVD refresh differs from full compile")
	}
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
	assertMatch(t, next, []byte("ABCDEF"), "")
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
	members[0].data = []byte("Bad.Rule:0:*:zz\n")
	if err := os.WriteFile(path, cvdFixtureOrdered(t, members), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWithPrior(dir, Limits{}, next); err == nil {
		t.Fatal("invalid CVD member displaced last-good matcher")
	}
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
}

func TestCVDRefreshFallsBackWhenSelectionChangesAndHandlesHashMember(t *testing.T) {
	members := []member{
		{"daily.ign", []byte("Ignored.Rule\n")},
		{"first.ndb", []byte("Old.Rule:0:*:414243444546\n")},
		{"middle.hdb", hashLine([]byte("old hash target"), "Old.Hash")},
		{"other.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = []byte("Different.Rule\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("selection member took incremental path: handled=%v err=%v", handled, err)
	}
	members[0].data = []byte("Ignored.Rule\n")
	members[2].data = hashLine([]byte("new hash target"), "New.Hash")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	hashNext, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("changed hash member did not refresh: handled=%v err=%v", handled, err)
	}
	hashFull, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if hashNext.DetectionCount() != hashFull.DetectionCount() || !reflect.DeepEqual(hashNext.Report(), hashFull.Report()) ||
		!reflect.DeepEqual(hashNext.sourceStats, hashFull.sourceStats) || !reflect.DeepEqual(hashNext.memberStats, hashFull.memberStats) {
		t.Fatal("hash-member refresh differs from full compile")
	}
	assertMatch(t, prior, []byte("old hash target"), "Old.Hash")
	assertMatch(t, hashNext, []byte("old hash target"), "")
	assertMatch(t, hashNext, []byte("new hash target"), "New.Hash")
	members[2].data = hashLine([]byte("old hash target"), "Old.Hash")
	members[1].data = []byte("New.Rule:0:*:4748494a4b4c\n")
	members[3].data = []byte("Other.New:0:*:535455565758\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("two changed NDB members did not refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || !reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatal("two-member refresh differs from full compile")
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
	assertMatch(t, next, []byte("ABCDEF"), "")
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
	assertMatch(t, next, []byte("STUVWX"), "Other.New")
}

func TestCVDNDBMemberRefreshPreservesIgnoreAccounting(t *testing.T) {
	members := []member{
		{"daily.ign", []byte("Ignored.Rule\n")},
		{"first.ndb", []byte("Old.Rule:0:*:414243444546\n")},
		{"other.ndb", []byte("Other.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[1].data = []byte("Ignored.Rule:0:*:4748494a4b4c\nNew.Rule:0:*:535455565758\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("CVD refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) || next.Report().Skipped[SkipIgnoreList] != 1 ||
		!reflect.DeepEqual(next.sourceStats["daily.cvd"].skipped, full.sourceStats["daily.cvd"].skipped) ||
		!reflect.DeepEqual(next.memberStats["daily.cvd/first.ndb"].skipped, full.memberStats["daily.cvd/first.ndb"].skipped) {
		t.Fatalf("incremental selection accounting differs: next=%+v full=%+v", next.Report(), full.Report())
	}
	assertMatch(t, next, []byte("GHIJKL"), "")
	assertMatch(t, next, []byte("STUVWX"), "New.Rule")
}

func TestCVDNDBDeltaVerifierRejectsCorruptOrUnsafeArchive(t *testing.T) {
	good := cvdFixtureOrdered(t, []member{{"daily.ndb", []byte("Old.Rule:0:*:414243444546\n")}})
	files := map[string][]byte{"daily.cvd": good}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte(nil), good...)
	corrupt[len(corrupt)-1] ^= 1
	files["daily.cvd"] = corrupt
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); !handled || err == nil {
		t.Fatalf("corrupt checksum accepted: handled=%v err=%v", handled, err)
	}
	unsafe := cvdFixtureOrdered(t, []member{{"../daily.ndb", []byte("New.Rule:0:*:4748494a4b4c\n")}})
	files["daily.cvd"] = unsafe
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); !handled || err == nil {
		t.Fatalf("unsafe member accepted: handled=%v err=%v", handled, err)
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Old.Rule")
}

func TestCVDNDBMemberRefreshAfterEarlierRuleAndRepeated(t *testing.T) {
	members := []member{
		{"first.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
		{"last.ndb", []byte("Old.Rule:0:*:414243444546\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	first, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		line, marker, name string
	}{
		{"New.One:0:*:4748494a4b4c\nNew.Two:0:*:535455565758\n", "GHIJKL", "New.One"},
		{"Newest.Rule:0:*:616263646566\n", "abcdef", "Newest.Rule"},
	} {
		members[1].data = []byte(tc.line)
		files["daily.cvd"] = cvdFixtureOrdered(t, members)
		next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: first})
		if err != nil || !handled {
			t.Fatalf("repeated CVD refresh: handled=%v err=%v", handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if next.Version() != full.Version() || !reflect.DeepEqual(next.Report(), full.Report()) ||
			next.memberStats["daily.cvd/last.ndb"].ruleStart != full.memberStats["daily.cvd/last.ndb"].ruleStart ||
			next.memberStats["daily.cvd/last.ndb"].ruleEnd != full.memberStats["daily.cvd/last.ndb"].ruleEnd {
			t.Fatal("repeated CVD refresh differs from full compile")
		}
		assertMatch(t, next, []byte("MNOPQR"), "Kept.Rule")
		assertMatch(t, next, []byte(tc.marker), tc.name)
		first = next
	}
	assertMatch(t, first, []byte("ABCDEF"), "")
	assertMatch(t, first, []byte("GHIJKL"), "")
}

func TestTopLevelNDBThenCVDMemberRefreshKeepsRuleRanges(t *testing.T) {
	members := []member{{"first.ndb", []byte("Archive.Old:0:*:4d4e4f505152\n")}}
	files := map[string][]byte{
		"a.ndb":    []byte("Local.Old:0:*:414243444546\n"),
		"main.cvd": cvdFixtureOrdered(t, members),
	}
	first, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("Local.One:0:*:4748494a4b4c\nLocal.Two:0:*:535455565758\n")
	second, err := compileWithLimits(files, DefaultLimits(), compileOptions{prior: first})
	if err != nil {
		t.Fatal(err)
	}
	if second.memberStats["main.cvd/first.ndb"].ruleStart != 2 {
		t.Fatal("top-level NDB refresh left later CVD member range stale")
	}
	members[0].data = []byte("Archive.New:0:*:616263646566\n")
	files["main.cvd"] = cvdFixtureOrdered(t, members)
	third, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: second})
	if err != nil || !handled {
		t.Fatalf("CVD after top-level refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if third.Version() != full.Version() || !reflect.DeepEqual(third.Report(), full.Report()) ||
		third.memberStats["main.cvd/first.ndb"].ruleStart != full.memberStats["main.cvd/first.ndb"].ruleStart {
		t.Fatal("chained top-level/CVD update differs from full compile")
	}
	assertMatch(t, first, []byte("ABCDEF"), "Local.Old")
	assertMatch(t, second, []byte("MNOPQR"), "Archive.Old")
	assertMatch(t, third, []byte("MNOPQR"), "")
	assertMatch(t, third, []byte("abcdef"), "Archive.New")
}

func TestCVDMultipleNDBMembersShiftInterveningAndLaterRanges(t *testing.T) {
	members := []member{
		{"first.ndb", []byte("First.Old:0:*:414243444546\n")},
		{"middle.hdb", hashLine([]byte("unchanged hash target"), "Middle.Hash")},
		{"last.ndb", []byte("Last.Old:0:*:4748494a4b4c\n")},
		{"after.ndb", []byte("After.Kept:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{
		"daily.cvd": cvdFixtureOrdered(t, members),
		"local.ndb": []byte("Local.Kept:0:*:535455565758\n"),
	}
	previous, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for round, replacement := range []struct {
		first, last []byte
		probe, want string
	}{
		{[]byte("First.One:0:*:616263646566\nFirst.Two:0:*:6768696a6b6c\n"),
			[]byte("Last.One:0:*:6d6e6f707172\nLast.Two:0:*:737475767778\nLast.Three:0:*:797a30313233\n"),
			"abcdef", "First.One"},
		{[]byte("First.Final:0:*:313233343536\n"),
			[]byte("Last.Final:0:*:373839303132\n"),
			"123456", "First.Final"},
	} {
		members[0].data, members[2].data = replacement.first, replacement.last
		files["daily.cvd"] = cvdFixtureOrdered(t, members)
		next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: previous})
		if err != nil || !handled {
			t.Fatalf("round %d multi-member refresh: handled=%v err=%v", round, handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if next.Version() != full.Version() || next.DetectionCount() != full.DetectionCount() ||
			!reflect.DeepEqual(next.Report(), full.Report()) ||
			!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
			t.Fatalf("round %d multi-member metadata differs from full compile", round)
		}
		for _, data := range [][]byte{[]byte(replacement.probe), []byte("MNOPQR"), []byte("STUVWX"),
			[]byte("unchanged hash target"), []byte("unrelated harmless data")} {
			got, gotSteps, gotErr := next.matchCounted(context.Background(), data, 10_000_000)
			want, wantSteps, wantErr := full.matchCounted(context.Background(), data, 10_000_000)
			if got != want || gotSteps != wantSteps || !reflect.DeepEqual(gotErr, wantErr) {
				t.Fatalf("round %d %q incremental=(%q,%d,%v) full=(%q,%d,%v)",
					round, data, got, gotSteps, gotErr, want, wantSteps, wantErr)
			}
		}
		assertMatch(t, next, []byte(replacement.probe), replacement.want)
		assertMatch(t, next, []byte("MNOPQR"), "After.Kept")
		assertMatch(t, next, []byte("STUVWX"), "Local.Kept")
		previous = next
	}
}

func TestCVDMultipleNDBMembersPreserveSkippedAccounting(t *testing.T) {
	members := []member{
		{"daily.ign", []byte("Ignored.Rule\n")},
		{"first.ndb", []byte("First.Old:0:*:414243444546\n")},
		{"second.ndb", []byte("Second.Old:0:*:4748494a4b4c\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[1].data = []byte("First.New:0:*:4d4e4f505152\nIgnored.Rule:0:*:535455565758\n")
	members[2].data = []byte("Ignored.Rule:0:*:616263646566\nSecond.New:0:*:6768696a6b6c\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("multi-member skip refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) ||
		!reflect.DeepEqual(next.memberStats, full.memberStats) ||
		next.Report().Skipped[SkipIgnoreList] != 2 {
		t.Fatalf("multi-member skip accounting differs: next=%+v full=%+v", next.Report(), full.Report())
	}
	assertMatch(t, next, []byte("MNOPQR"), "First.New")
	assertMatch(t, next, []byte("ghijkl"), "Second.New")
	assertMatch(t, next, []byte("STUVWX"), "")
}

func TestCVDMemberRefreshUsesFullCompilerNearSignatureLimit(t *testing.T) {
	limits := Limits{MaxSignatures: 2}.defaults()
	members := []member{
		{"daily.ign", []byte("Ignored.Rule\n")},
		{"first.ndb", []byte("First.Old:0:*:414243444546\n")},
	}
	files := map[string][]byte{
		"daily.cvd": cvdFixtureOrdered(t, members),
		"local.ndb": []byte("Local.Kept:0:*:4d4e4f505152\n"),
	}
	prior, err := CompileWithLimits(files, limits)
	if err != nil {
		t.Fatal(err)
	}
	members[1].data = []byte("First.New:0:*:4748494a4b4c\n# comment\nIgnored.Rule:0:*:535455565758\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	// An ignored line after the first new rule is legal before the later
	// local rule is loaded. The incremental baseline already includes local.
	if _, handled, err := tryContainerSourceRefresh(files, limits, compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("near-limit CVD refresh bypassed full compiler: handled=%v err=%v", handled, err)
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
		t.Fatal("near-limit fallback differs from full compiler")
	}
	assertMatch(t, next, []byte("GHIJKL"), "First.New")
}

func TestCVDMixedNDBAndHashMemberRefreshMatchesFullCompile(t *testing.T) {
	shared := []byte("shared hash member fixture")
	oldSHA := []byte("old sha member fixture")
	newSHA := []byte("new sha member fixture")
	oldSHA256, newSHA256 := sha256.Sum256(oldSHA), sha256.Sum256(newSHA)
	members := []member{
		{"first.hdb", hashLine(shared, "First.Hash")},
		{"middle.ndb", []byte("Middle.Old:0:*:414243444546\n")},
		{"second.hsb", []byte(fmt.Sprintf("%x:*:Old.SHA:73\n", oldSHA256))},
		{"third.hdb", hashLine(shared, "Third.Hash")},
		{"after.ndb", []byte("After.Kept:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{
		"daily.cvd": cvdFixtureOrdered(t, members),
		"local.hdb": hashLine(shared, "Local.Hash"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	// Member shards belong to one native CVD candidate group.
	assertMatch(t, prior, shared, "Third.Hash")
	members[0].data = hashLine([]byte("new md5 member fixture"), "First.New")
	members[1].data = []byte("Middle.One:0:*:4748494a4b4c\nMiddle.Two:0:*:535455565758\n")
	members[2].data = []byte(fmt.Sprintf("%x:*:New.SHA:73\n", newSHA256))
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("mixed CVD refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatalf("mixed refresh metadata differs: next=%+v full=%+v", next.Report(), full.Report())
	}
	if next.memberHashes["daily.cvd/third.hdb"] != prior.memberHashes["daily.cvd/third.hdb"] ||
		next.memberHashes["daily.cvd/first.hdb"] == prior.memberHashes["daily.cvd/first.hdb"] {
		t.Fatal("hash member sharing/replacement differs from provenance")
	}
	for _, data := range [][]byte{shared, oldSHA, newSHA, []byte("new md5 member fixture"),
		[]byte("ABCDEF"), []byte("GHIJKL"), []byte("STUVWX"), []byte("MNOPQR"), []byte("unrelated harmless data")} {
		got, gotSteps, gotErr := next.matchCounted(context.Background(), data, 10_000_000)
		want, wantSteps, wantErr := full.matchCounted(context.Background(), data, 10_000_000)
		if got != want || gotSteps != wantSteps || !reflect.DeepEqual(gotErr, wantErr) {
			t.Fatalf("%q incremental=(%q,%d,%v) full=(%q,%d,%v)", data, got, gotSteps, gotErr, want, wantSteps, wantErr)
		}
	}
	assertMatch(t, next, shared, "Third.Hash")
	assertMatch(t, next, oldSHA, "")
	assertMatch(t, next, newSHA, "New.SHA")
	assertMatch(t, next, []byte("GHIJKL"), "Middle.One")
	// Member shards belong to one native CVD candidate group.
	assertMatch(t, prior, shared, "Third.Hash")
	assertMatch(t, prior, oldSHA, "Old.SHA")
	assertMatch(t, prior, []byte("ABCDEF"), "Middle.Old")
	// A second hash-only member rotation must retain the first update and the
	// unchanged member shard order.
	members[2].data = []byte(fmt.Sprintf("%x:*:Newest.SHA:73\n", sha256.Sum256([]byte("newest sha member fixture"))))
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	again, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: next})
	if err != nil || !handled {
		t.Fatalf("repeated hash member refresh: handled=%v err=%v", handled, err)
	}
	againFull, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Report(), againFull.Report()) || again.DetectionCount() != againFull.DetectionCount() ||
		!reflect.DeepEqual(again.memberStats, againFull.memberStats) {
		t.Fatal("repeated hash member refresh differs from full compiler")
	}
	assertMatch(t, again, newSHA, "")
	assertMatch(t, again, []byte("newest sha member fixture"), "Newest.SHA")
	assertMatch(t, again, shared, "Third.Hash")
}

func TestCVDHashMemberCanBecomeDetectingAndRejectInvalidUpdate(t *testing.T) {
	members := []member{
		{"daily.ign", []byte("Ignored.Hash\n")},
		{"first.hdb", hashLine([]byte("old ignored payload"), "Ignored.Hash")},
		{"kept.ndb", []byte("Kept.Rule:0:*:414243444546\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if prior.memberHashes["daily.cvd/first.hdb"] != nil {
		t.Fatal("ignored member unexpectedly retained a hash shard")
	}
	members[1].data = hashLine([]byte("new detecting payload"), "Active.Hash")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("activating hash member: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) || next.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatalf("activated hash member differs from full compile: next=%#v full=%#v", next.Report(), full.Report())
	}
	assertMatch(t, next, []byte("new detecting payload"), "Active.Hash")
	assertMatch(t, prior, []byte("new detecting payload"), "")
	members[1].data = []byte("invalid hash member\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: next}); !handled || err == nil {
		t.Fatalf("invalid hash member accepted: handled=%v err=%v", handled, err)
	}
	assertMatch(t, next, []byte("new detecting payload"), "Active.Hash")
}

func TestCVDMixedDirectoryRefreshSharesUnchangedHashMember(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daily.cvd")
	members := []member{
		{"first.hdb", hashLine([]byte("old hash payload"), "Old.Hash")},
		{"middle.ndb", []byte("Old.Rule:0:*:414243444546\n")},
		{"kept.hdb", hashLine([]byte("kept hash payload"), "Kept.Hash")},
	}
	if err := os.WriteFile(path, cvdFixtureOrdered(t, members), 0o600); err != nil {
		t.Fatal(err)
	}
	prior, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = hashLine([]byte("new hash payload"), "New.Hash")
	members[1].data = []byte("New.Rule:0:*:4748494a4b4c\n")
	if err := os.WriteFile(path, cvdFixtureOrdered(t, members), 0o600); err != nil {
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
	if next.memberHashes["daily.cvd/kept.hdb"] != prior.memberHashes["daily.cvd/kept.hdb"] ||
		next.memberHashes["daily.cvd/first.hdb"] == prior.memberHashes["daily.cvd/first.hdb"] ||
		!reflect.DeepEqual(next.Report(), full.Report()) || next.DetectionCount() != full.DetectionCount() {
		t.Fatal("directory mixed refresh rebuilt unchanged hash member or differs from full compile")
	}
	assertMatch(t, prior, []byte("old hash payload"), "Old.Hash")
	assertMatch(t, next, []byte("old hash payload"), "")
	assertMatch(t, next, []byte("new hash payload"), "New.Hash")
	assertMatch(t, next, []byte("GHIJKL"), "New.Rule")
	assertMatch(t, next, []byte("kept hash payload"), "Kept.Hash")
}

func TestCVDLogicalMemberRefreshMatchesFullCompile(t *testing.T) {
	members := []member{
		{"first.ldb", []byte("Logical.Old;Target:0;0;414243444546\n")},
		{"middle.hdb", hashLine([]byte("kept logical hash payload"), "Kept.Hash")},
		{"last.ndb", []byte("Later.Kept:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{
		"daily.cvd": cvdFixtureOrdered(t, members),
		"local.ndb": []byte("Local.Kept:0:*:535455565758\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = []byte("Logical.One;Target:0;0;4748494a4b4c\nLogical.Two;Target:0;0&1;616263;646566\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("logical member refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) {
		t.Fatalf("logical member refresh differs from full: next=%#v full=%#v", next.Report(), full.Report())
	}
	if next.expressionIDs != nil {
		t.Fatal("logical compilation map retained after publication")
	}
	oldPattern := prior.rules[0].patterns[0]
	if !next.retiredPatterns.has(oldPattern) || !reflect.DeepEqual(next.patterns[oldPattern], bodyPattern{}) {
		t.Fatal("replaced logical pattern body was retained")
	}
	for _, data := range [][]byte{[]byte("ABCDEF"), []byte("GHIJKL"), []byte("abcdef"),
		[]byte("MNOPQR"), []byte("STUVWX"), []byte("kept logical hash payload")} {
		got, gotSteps, gotErr := next.matchCounted(context.Background(), data, 10_000_000)
		want, wantSteps, wantErr := full.matchCounted(context.Background(), data, 10_000_000)
		if got != want || gotSteps != wantSteps || !reflect.DeepEqual(gotErr, wantErr) {
			t.Fatalf("%q incremental=(%q,%d,%v) full=(%q,%d,%v)", data, got, gotSteps, gotErr, want, wantSteps, wantErr)
		}
	}
	assertMatch(t, prior, []byte("ABCDEF"), "Logical.Old")
	assertMatch(t, next, []byte("ABCDEF"), "")
	assertMatch(t, next, []byte("GHIJKL"), "Logical.One")
	assertMatch(t, next, []byte("abcdef"), "Logical.Two")
	assertMatch(t, next, []byte("MNOPQR"), "Later.Kept")
}

func TestCVDLogicalIconGroupRefreshResolvesDeadRule(t *testing.T) {
	members := []member{{"daily.ldb", []byte("Old.Rule;Target:0;0;414243444546\n")}}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = []byte("Dead.Icon;Engine:51-255,Target:1,IconGroup1:Missing;0;4748494a4b4c\nLive.Rule;Target:0;0;535455565758\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("icon-group logical refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) ||
		!reflect.DeepEqual(next.sourceStats, full.sourceStats) || !reflect.DeepEqual(next.memberStats, full.memberStats) ||
		len(next.rules) != len(full.rules) || !next.rules[0].dead || !full.rules[0].dead {
		t.Fatalf("logical icon resolution differs: next=%#v full=%#v", next.Report(), full.Report())
	}
	assertMatch(t, next, []byte("STUVWX"), "Live.Rule")
}

func TestRepeatedCVDLogicalRefreshReusesRetiredPatternSlots(t *testing.T) {
	marker := func(n int) []byte { return []byte(fmt.Sprintf("L%05d", n)) }
	rule := func(n int) []byte { return []byte(fmt.Sprintf("Logical.%d;Target:0;0;%x\n", n, marker(n))) }
	members := []member{
		{"daily.ldb", rule(0)},
		{"kept.hdb", hashLine([]byte("kept repeated logical hash"), "Kept.Hash")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	previous, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	initialSlots := len(previous.patterns)
	for round := 1; round <= 32; round++ {
		members[0].data = rule(round)
		files["daily.cvd"] = cvdFixtureOrdered(t, members)
		next, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: previous})
		if err != nil || !handled {
			t.Fatalf("round %d logical CVD refresh: handled=%v err=%v", round, handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if len(next.patterns) > initialSlots+1 || !reflect.DeepEqual(next.Report(), full.Report()) ||
			next.sourceHashes["daily.cvd"] != previous.sourceHashes["daily.cvd"] {
			t.Fatalf("round %d accumulated %d slots or changed shared hash source", round, len(next.patterns))
		}
		assertMatch(t, previous, marker(round-1), fmt.Sprintf("Logical.%d", round-1))
		assertMatch(t, next, marker(round-1), "")
		assertMatch(t, next, marker(round), fmt.Sprintf("Logical.%d", round))
		assertMatch(t, next, []byte("kept repeated logical hash"), "Kept.Hash")
		previous = next
	}
}
