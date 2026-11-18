package signatures

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func hashLine(data []byte, name string) []byte {
	sum := md5.Sum(data)
	return []byte(fmt.Sprintf("%x:*:%s:73\n", sum, name))
}

func TestHashOnlyRefreshMatchesFullCompilerAndKeepsOldSnapshot(t *testing.T) {
	oldPayload, newPayload := []byte("old malware payload"), []byte("new malware payload")
	oldSHA := sha256.Sum256(oldPayload)
	files := map[string][]byte{
		"a.ndb": []byte("Pattern.Test:0:*:414243444546\n"),
		"b.hdb": hashLine(oldPayload, "Old.Hash"),
		"c.hsb": []byte(fmt.Sprintf("%x:%d:Old.SHA\n", oldSHA, len(oldPayload))),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["b.hdb"] = hashLine(newPayload, "New.Hash")
	files["c.hsb"] = []byte(fmt.Sprintf("%x:*:New.SHA:73\n", sha256.Sum256(newPayload)))
	next, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh handled=%v err=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version() != full.Version() || next.Count() != full.Count() ||
		next.DetectionCount() != full.DetectionCount() || !reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatalf("incremental database differs: next=%+v full=%+v", next.Report(), full.Report())
	}
	if len(next.patterns) == 0 || &next.patterns[0] != &prior.patterns[0] {
		t.Fatal("unchanged pattern matcher was rebuilt")
	}
	for _, tc := range []struct {
		data     []byte
		old, new string
	}{
		{oldPayload, "Old.Hash", ""},
		{newPayload, "", "New.Hash"},
		{[]byte("ABCDEF"), "Pattern.Test", "Pattern.Test"},
	} {
		gotOld, err := prior.Match(context.Background(), tc.data)
		if err != nil || gotOld != tc.old {
			t.Fatalf("old %q: %q, %v", tc.data, gotOld, err)
		}
		gotNew, err := next.Match(context.Background(), tc.data)
		if err != nil || gotNew != tc.new {
			t.Fatalf("new %q: %q, %v", tc.data, gotNew, err)
		}
		gotFull, err := full.Match(context.Background(), tc.data)
		if err != nil || gotFull != gotNew {
			t.Fatalf("full %q: %q, %v, expected %q", tc.data, gotFull, err, gotNew)
		}
	}
}

func TestHashRefreshRejectsOtherChangesAndInvalidInputs(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("Pattern.Test:0:*:414243444546\n"),
		"b.hdb": hashLine([]byte("malware payload"), "Old.Hash"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.ndb"] = []byte("Pattern.New:0:*:414243444546\n")
	if _, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("changed pattern took hash path: handled=%v err=%v", handled, err)
	}
	files["a.ndb"] = []byte("Pattern.Test:0:*:414243444546\n")
	files["b.hdb"] = []byte("broken hash rule\n")
	if _, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); !handled || err == nil {
		t.Fatalf("invalid hash accepted: handled=%v err=%v", handled, err)
	}
	files["b.hdb"] = hashLine([]byte("other payload"), "New.Hash")
	files["c.hdb"] = hashLine([]byte("third payload"), "Third.Hash")
	if _, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); handled || err != nil {
		t.Fatalf("added file took hash path: handled=%v err=%v", handled, err)
	}
}

func TestHashRefreshPreservesSourceOrderForDuplicateRules(t *testing.T) {
	payload := []byte("duplicate malware")
	files := map[string][]byte{
		"a.hdb": hashLine(payload, "A.First"),
		"b.hdb": hashLine(payload, "B.Second"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	files["a.hdb"] = hashLine([]byte("different malware"), "A.Other")
	next, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	assertMatch(t, prior, payload, "A.First")
	assertMatch(t, next, payload, "B.Second")
}

func TestPreparedHashRefreshMatchesFullCompile(t *testing.T) {
	oldSum := sha256.Sum256([]byte("old threatwind indicator"))
	newSum := sha256.Sum256([]byte("new threatwind indicator"))
	oldWire, oldPrepared, err := PrepareHashList([]byte(fmt.Sprintf("%x\n", oldSum)), "SHA256", 4096)
	if err != nil {
		t.Fatal(err)
	}
	base := map[string][]byte{
		"base.ndb":  []byte("Pattern.Test:0:*:414243444546\n"),
		"intel.hsb": oldWire,
	}
	prior, localErr, fatalErr := CompileWithPreparedLocalFiles(base, nil, map[string]*PreparedHashList{"intel.hsb": oldPrepared}, Limits{})
	if localErr != nil || fatalErr != nil {
		t.Fatalf("prior: local=%v fatal=%v", localErr, fatalErr)
	}
	newWire, newPrepared, err := PrepareHashList([]byte(fmt.Sprintf("%x\n", newSum)), "SHA256", 4096)
	if err != nil {
		t.Fatal(err)
	}
	base["intel.hsb"] = newWire
	next, localErr, fatalErr := CompileWithPreparedLocalFilesAndPrior(base, nil, map[string]*PreparedHashList{"intel.hsb": newPrepared}, Limits{}, prior)
	if localErr != nil || fatalErr != nil {
		t.Fatalf("next: local=%v fatal=%v", localErr, fatalErr)
	}
	full, localErr, fatalErr := CompileWithPreparedLocalFiles(base, nil, map[string]*PreparedHashList{"intel.hsb": newPrepared}, Limits{})
	if localErr != nil || fatalErr != nil {
		t.Fatalf("full: local=%v fatal=%v", localErr, fatalErr)
	}
	if &next.patterns[0] != &prior.patterns[0] || next.Version() != full.Version() ||
		!reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatal("prepared refresh differed from full compile or rebuilt unchanged matcher")
	}
	assertMatch(t, prior, []byte("old threatwind indicator"), "UTMStack.Intel.SHA256."+fmt.Sprintf("%x", oldSum))
	assertMatch(t, next, []byte("new threatwind indicator"), "UTMStack.Intel.SHA256."+fmt.Sprintf("%x", newSum))
}

func TestHashRefreshUpdatesIgnoredRuleAccounting(t *testing.T) {
	priorFiles := map[string][]byte{
		"a.ndb": []byte("Pattern.Test:0:*:414243444546\n"),
		"b.ign": []byte("Ignored.Hash\n"),
		"c.hdb": hashLine([]byte("old malware payload"), "Old.Hash"),
	}
	prior, err := Compile(priorFiles)
	if err != nil {
		t.Fatal(err)
	}
	priorFiles["c.hdb"] = append(hashLine([]byte("ignored payload"), "Ignored.Hash"), hashLine([]byte("new malware payload"), "New.Hash")...)
	next, handled, err := tryHashSourceRefresh(priorFiles, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	full, err := Compile(priorFiles)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.Report(), full.Report()) || next.Report().Skipped[SkipIgnoreList] != 1 {
		t.Fatalf("skip accounting: next=%+v full=%+v", next.Report(), full.Report())
	}
}

func TestShardedHashLookupPrefersExactSizeAcrossSources(t *testing.T) {
	object := []byte("exact-size precedence fixture")
	sum := md5.Sum(object)
	files := map[string][]byte{
		"a.hdb": []byte(fmt.Sprintf("%x:*:Any.First:73\n", sum)),
		"b.hdb": []byte(fmt.Sprintf("%x:%d:Exact.Second\n", sum, len(object))),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, prior, object, "Exact.Second")
	if prior.DetectionCount() != 2 {
		t.Fatalf("distinct exact and wildcard count: %d", prior.DetectionCount())
	}
	files["a.hdb"] = []byte(fmt.Sprintf("%x:*:Any.Updated:73\n", sum))
	next, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	assertMatch(t, next, object, "Exact.Second")
	if next.DetectionCount() != 2 {
		t.Fatalf("refreshed distinct count: %d", next.DetectionCount())
	}
}

func TestDirectoryHashRefreshStreamsUnchangedSources(t *testing.T) {
	dir := t.TempDir()
	pattern := filepath.Join(dir, "a.ndb")
	hash := filepath.Join(dir, "b.hdb")
	if err := os.WriteFile(pattern, []byte("Pattern.Test:0:*:414243444546\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hash, hashLine([]byte("old payload"), "Old.Hash"), 0o600); err != nil {
		t.Fatal(err)
	}
	prior, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hash, hashLine([]byte("new payload"), "New.Hash"), 0o600); err != nil {
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
	if &next.patterns[0] != &prior.patterns[0] || next.Version() != full.Version() ||
		!reflect.DeepEqual(next.Report(), full.Report()) {
		t.Fatal("directory refresh did not share matcher or match full compile")
	}
	assertMatch(t, prior, []byte("old payload"), "Old.Hash")
	assertMatch(t, next, []byte("new payload"), "New.Hash")
}

func TestCompilerTracksRuleRangesByTopLevelSource(t *testing.T) {
	files := map[string][]byte{
		"a.ndb": []byte("A.One:0:*:414243444546\nA.Two:0:*:4748494a4b4c\n"),
		"b.hdb": hashLine([]byte("hash-only source"), "B.Hash"),
		"c.ndb": []byte("C.One:0:*:4d4e4f505152\n"),
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		start, end int
	}{
		{"a.ndb", 0, 2},
		{"b.hdb", 2, 2},
		{"c.ndb", 2, 3},
	} {
		stats := d.sourceStats[tc.name]
		if stats == nil || !stats.ruleSeen || stats.ruleStart != tc.start || stats.ruleEnd != tc.end {
			t.Fatalf("%s rule range = %+v, want [%d,%d)", tc.name, stats, tc.start, tc.end)
		}
	}
	files["b.hdb"] = hashLine([]byte("updated hash-only source"), "B.New")
	next, handled, err := tryHashSourceRefresh(files, DefaultLimits(), compileOptions{prior: d})
	if err != nil || !handled {
		t.Fatalf("refresh: handled=%v err=%v", handled, err)
	}
	if next.sourceStats["a.ndb"] != d.sourceStats["a.ndb"] || next.sourceStats["c.ndb"] != d.sourceStats["c.ndb"] {
		t.Fatal("unchanged source rule provenance was rebuilt")
	}
	if got, want := next.sourceStats["b.hdb"], d.sourceStats["b.hdb"]; !got.ruleSeen || got.ruleStart != want.ruleStart || got.ruleEnd != want.ruleEnd {
		t.Fatalf("changed hash source lost its rule position: got=%+v want=%+v", got, want)
	}
}

func TestCompilerTracksNestedContainerRuleRange(t *testing.T) {
	members := []member{
		{"first.ndb", []byte("Daily.First:0:*:414243444546\n")},
		{"middle.hdb", hashLine([]byte("contained hash"), "Daily.Hash")},
		{"last.ndb", []byte("Daily.Last:0:*:4748494a4b4c\n")},
	}
	daily := cvdFixtureOrdered(t, members)
	d, err := Compile(map[string][]byte{
		"daily.cvd": daily,
		"local.ndb": []byte("Local.Rule:0:*:4d4e4f505152\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		start, end int
	}{
		{"daily.cvd", 0, 2},
		{"local.ndb", 2, 3},
	} {
		stats := d.sourceStats[tc.name]
		if stats == nil || !stats.ruleSeen || stats.ruleStart != tc.start || stats.ruleEnd != tc.end {
			t.Fatalf("%s rule range = %+v, want [%d,%d)", tc.name, stats, tc.start, tc.end)
		}
	}
	if len(d.containerMembers["daily.cvd"]) != len(members) {
		t.Fatal("container member list lost archive order")
	}
	for i, m := range members {
		got := d.containerMembers["daily.cvd"][i]
		if got.name != m.name || got.digest != sha256.Sum256(m.data) || got.size != int64(len(m.data)) {
			t.Fatalf("member %d provenance = %+v", i, got)
		}
		stats := d.memberStats["daily.cvd/"+m.name]
		if stats == nil || !stats.ruleSeen || stats.loaded != 1 || stats.detections != 1 ||
			stats.ruleStart != []int{0, 1, 1}[i] || stats.ruleEnd != []int{1, 1, 2}[i] {
			t.Fatalf("member %d compile range = %+v", i, stats)
		}
	}
}

func TestPatternKeysRetainCrossSourceDeduplicationIdentity(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.ndb": []byte("A.Shared:0:*:414243444546\n"),
		"b.ndb": []byte("B.Shared:0:*:414243444546\nB.Unique:0:*:4748494a4b4c\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.rules) != 3 || len(d.patterns) < 2 || len(d.patternKeys) != len(d.patterns) {
		t.Fatalf("rules=%d patterns=%d keys=%d", len(d.rules), len(d.patterns), len(d.patternKeys))
	}
	if d.rules[0].patterns[0] != d.rules[1].patterns[0] || d.rules[2].patterns[0] == d.rules[0].patterns[0] {
		t.Fatal("unexpected shared pattern identity")
	}
	for i, body := range []string{"414243444546", "4748494a4b4c"} {
		id := d.rules[0].patterns[0]
		if i == 1 {
			id = d.rules[2].patterns[0]
		}
		key := sha256.Sum256([]byte("*\x00" + body))
		if d.patternKeys[id] != [16]byte(key[:16]) {
			t.Fatalf("pattern %d key differs from compiler deduplication key", id)
		}
	}
}
