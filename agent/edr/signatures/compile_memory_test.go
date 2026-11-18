package signatures

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

func TestCommonRuleStorageAndGenerationOwnership(t *testing.T) {
	if got := unsafe.Sizeof(rule{}); got != 64 {
		t.Fatalf("common rule grew to %d bytes", got)
	}
	d, err := Compile(map[string][]byte{"test.ndb": []byte("Simple:0:*:41424344\n")})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.rules) != 1 || d.rules[0].ruleCold != &emptyRuleCold {
		t.Fatal("ordinary rule retained optional storage")
	}
	copyOf := d.rules[0]
	copyOf.ownCold().unknown = UnknownBytecode
	if d.rules[0].unknown != "" || emptyRuleCold.unknown != "" {
		t.Fatal("writing a copied rule changed the published zero value")
	}
	copyOf.cloneCold().unknown = UnknownIconGroup
	if copyOf.unknown != UnknownIconGroup || d.rules[0].unknown != "" {
		t.Fatal("cold record copy did not isolate generations")
	}
}

func TestUntrustedTruncatedContainerDoesNotPreallocateClaimedMember(t *testing.T) {
	var payload bytes.Buffer
	tw := tar.NewWriter(&payload)
	if err := tw.WriteHeader(&tar.Header{Name: "large.ndb", Size: 384 << 20, Mode: 0600}); err != nil {
		t.Fatal(err)
	}
	// Deliberately omit the member body. Both local CLD and unsigned CVD
	// headers can be constructed by an untrusted administrator input file.
	for _, name := range []string{"truncated.cld", "truncated.cvd"} {
		header := fmt.Sprintf("ClamAV-VDB:now:1:1:90:%x:unverified:builder:1", md5.Sum(payload.Bytes()))
		data := append([]byte(header+strings.Repeat(" ", 512-len(header))), payload.Bytes()...)
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		_, info, err := unpackContainer(name, data, &archiveBudget{limits: Limits{}.defaults()})
		runtime.ReadMemStats(&after)
		if err == nil || info.Authenticated {
			t.Fatalf("%s: info=%+v error=%v", name, info, err)
		}
		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8<<20 {
			t.Fatalf("%s: truncated body allocated %d bytes", name, allocated)
		}
	}
}

func TestSizedSignatureReadRejectsChangedFiles(t *testing.T) {
	for _, tc := range []struct {
		data  string
		size  int64
		valid bool
	}{
		{"", 0, true}, {"abc", 3, true}, {"abc", 2, false},
		{"ab", 3, false}, {"", -1, false},
	} {
		got, err := readSizedFile(strings.NewReader(tc.data), tc.size)
		if (err == nil) != tc.valid {
			t.Fatalf("read(%q,%d): %q, %v", tc.data, tc.size, got, err)
		}
		if tc.valid && string(got) != tc.data {
			t.Fatalf("changed content: %q", got)
		}
	}
}

func TestCompiledLogicalTreesShareOnlyWithinGeneration(t *testing.T) {
	files := map[string][]byte{"test.ldb": []byte("First;Target:0;0&1;4141;4242\nSecond;Target:0;0&1;4343;4444\n")}
	a, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if a.rules[0].expr != a.rules[1].expr {
		t.Fatal("identical logical trees were not shared")
	}
	if a.rules[0].expr == b.rules[0].expr {
		t.Fatal("generation trees unexpectedly shared")
	}
	if a.expressionIDs != nil {
		t.Fatal("compile-only lookup retained")
	}
	// Counts remain per rule and per input despite the shared immutable tree.
	assertMatch(t, a, []byte("AABB"), "First")
	assertMatch(t, a, []byte("CCDD"), "Second")
	assertMatch(t, a, []byte("AADD"), "")
	assertMatch(t, a, []byte("AABB"), "First")
	d := newDatabase(Limits{}.defaults(), 0)
	if _, err := d.sharedExpression("1", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := d.sharedExpression("1", 1); err == nil {
		t.Fatal("cache bypassed subsignature validation")
	}
}

func TestPatternNodePreallocationMatchesEmittedGraph(t *testing.T) {
	for _, body := range []string{"4142", "41??{2-5}42", "41(42|43(44|45))46", "41!(4243|4445)46", "4142::wa", "4142::f"} {
		p, err := compileBody("*", body)
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if p.graph != nil && len(p.graph.nodes) != cap(p.graph.nodes) {
			t.Errorf("%s: nodes len=%d cap=%d", body, len(p.graph.nodes), cap(p.graph.nodes))
		}
	}
}

func TestHashFreezeExactStorageAndDuplicateNames(t *testing.T) {
	h := newHashTable()
	for i := 0; i < 1000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		for j := 0; j < 4; j++ {
			name := fmt.Sprintf("First.%d", i)
			if j != 0 {
				name = strings.Repeat("discarded", j*5)
			}
			if !h.add(string(d[:]), int64(i), name) {
				t.Fatal("add failed")
			}
		}
	}
	h.freeze()
	if h.len() != 1000 {
		t.Fatalf("kept %d", h.len())
	}
	g := &h.groups[0]
	if len(g.arena) != cap(g.arena) || len(g.records) != cap(g.records) || len(g.offsets) != cap(g.offsets) {
		t.Fatal("frozen arrays retain unused backing storage")
	}
	if bytes.Contains(g.arena, []byte("discarded")) {
		t.Fatal("duplicate name retained")
	}
	for i := 0; i < 1000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		if got := h.lookup(string(d[:]), int64(i)); got != fmt.Sprintf("First.%d", i) {
			t.Fatalf("entry %d: %q", i, got)
		}
	}
}

func TestHashFreezeUniqueNamesAvoidsSecondArena(t *testing.T) {
	h := newHashTable()
	for i := 0; i < 1000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		if !h.add(string(d[:]), int64(i), fmt.Sprintf("Distinct.%d", i)) {
			t.Fatal("add failed")
		}
	}
	g := &h.groups[0]
	// This input has distinct names and little unused capacity, so the
	// original arena can be kept while record order changes.
	original := &g.arena[0]
	originalOffsets := &g.offsets[0]
	h.freeze()
	if len(g.spans) != 1000 || g.offsets != nil || &g.arena[0] != original {
		t.Fatal("unique names caused a second arena allocation")
	}
	if &g.spans[0] != originalOffsets {
		t.Fatal("freeze allocated a second per-rule name-offset array")
	}
	for i := 0; i < 1000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		if got := h.lookup(string(d[:]), int64(i)); got != fmt.Sprintf("Distinct.%d", i) {
			t.Fatalf("entry %d: %q", i, got)
		}
	}
}
