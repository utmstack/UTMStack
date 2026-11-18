package signatures

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestHashTableFirstRuleWinsAndWildcards(t *testing.T) {
	a := md5.Sum([]byte("a"))
	b := sha1.Sum([]byte("b"))
	c := sha256.Sum256([]byte("c"))
	h := newHashTable()
	for _, r := range []struct {
		digest string
		size   int64
		name   string
	}{
		{string(a[:]), 1, "First.A"},
		{string(a[:]), 1, "Second.A"}, // duplicate: the first loaded name wins
		{string(a[:]), -1, "Any.A"},
		{string(b[:]), 7, "Sha1.B"},
		{string(c[:]), -1, "Sha256.C"},
	} {
		if !h.add(r.digest, r.size, r.name) {
			t.Fatalf("add %s failed", r.name)
		}
	}
	if h.add("short", 1, "Bad") {
		t.Fatal("unsupported digest width accepted")
	}
	h.freeze()
	if h.add(string(a[:]), 2, "Late") {
		t.Fatal("add after freeze accepted")
	}
	if h.len() != 4 {
		t.Fatalf("len = %d, want 4 after dropping the duplicate", h.len())
	}
	cases := []struct {
		digest string
		size   int64
		want   string
	}{
		{string(a[:]), 1, "First.A"},
		{string(a[:]), 99, "Any.A"},
		{string(b[:]), 7, "Sha1.B"},
		{string(b[:]), 8, ""},
		{string(c[:]), 12345, "Sha256.C"},
		{string(c[:1]) + string(c[1:]), 0, "Sha256.C"},
		{"", 0, ""},
	}
	for _, tc := range cases {
		if got := h.lookup(tc.digest, tc.size); got != tc.want {
			t.Errorf("lookup(%x, %d) = %q, want %q", tc.digest, tc.size, got, tc.want)
		}
	}
}

func TestHashTableManyEntriesAndNeighbours(t *testing.T) {
	h := newHashTable()
	for i := 0; i < 5000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		h.add(string(d[:]), int64(i), fmt.Sprintf("N.%d", i))
	}
	h.freeze()
	for i := 0; i < 5000; i++ {
		d := md5.Sum([]byte(fmt.Sprint(i)))
		if got := h.lookup(string(d[:]), int64(i)); got != fmt.Sprintf("N.%d", i) {
			t.Fatalf("entry %d = %q", i, got)
		}
		if got := h.lookup(string(d[:]), int64(i)+1); got != "" {
			t.Fatalf("entry %d matched the wrong size: %q", i, got)
		}
	}
	empty := newHashTable()
	empty.freeze()
	if empty.len() != 0 || empty.lookup(string(make([]byte, 16)), 0) != "" {
		t.Fatal("empty table")
	}
}

func TestHashTableInPlaceFreezePreservesFirstLoadedAcrossMixedWidths(t *testing.T) {
	rng := rand.New(rand.NewSource(104))
	type key struct {
		digest string
		size   int64
	}
	h := newHashTable()
	want := map[key]string{}
	var loaded []key
	for i := 0; i < 6000; i++ {
		width := []int{16, 20, 32}[i%3]
		var k key
		if len(loaded) > 0 && i%7 == 0 {
			k = loaded[rng.Intn(len(loaded))]
		} else {
			data := make([]byte, width)
			if _, err := rng.Read(data); err != nil {
				t.Fatal(err)
			}
			k = key{string(data), int64(rng.Intn(2000) + 1)}
		}
		name := fmt.Sprintf("Rule.%d", i)
		if !h.add(k.digest, k.size, name) {
			t.Fatalf("add %d failed", i)
		}
		if _, exists := want[k]; !exists {
			want[k] = name
		}
		loaded = append(loaded, k)
	}
	h.freeze()
	if h.len() != len(want) {
		t.Fatalf("unique records = %d, want %d", h.len(), len(want))
	}
	for k, name := range want {
		if got := h.lookup(k.digest, k.size); got != name {
			t.Fatalf("lookup %x at %d = %q, want %q", k.digest, k.size, got, name)
		}
	}
}

func TestCompileFreezesHashTables(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"a.hdb": []byte("44d88612fea8a8f36de82e1278abb02f:68:Eicar.Test\n44d88612fea8a8f36de82e1278abb02f:68:Eicar.Duplicate\n"),
		"a.fp":  []byte("0cc175b9c0f1b6a831c399e269772661:1:Allowed.A\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !d.hashes.frozen || !d.allow.frozen || !d.sections.frozen {
		t.Fatal("compiled tables must be frozen before use")
	}
	if d.DetectionCount() != 1 {
		t.Fatalf("detection count = %d", d.DetectionCount())
	}
	eicar := []byte(`X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`)
	// Two equal hashes in one artifact select the second native candidate.
	assertMatch(t, d, eicar, "Eicar.Duplicate")
}

func TestHashTableSizes(t *testing.T) {
	h := newHashTable()
	a, b := md5.Sum([]byte("a")), sha256.Sum256([]byte("b"))
	h.add(string(a[:]), 512, "A")
	h.add(string(a[:]), -1, "AnyA")
	h.add(string(b[:]), 4096, "B")
	h.freeze()
	got := h.sizes()
	if len(got) != 3 || !got[512] || !got[-1] || !got[4096] {
		t.Fatalf("sizes = %v", got)
	}
}

func TestHashTableSizeBoundaryDoesNotAliasWildcard(t *testing.T) {
	digest := md5.Sum([]byte("size boundary"))
	h := newHashTable()
	if !h.add(string(digest[:]), int64(hashWildcard)-1, "Max.Size") {
		t.Fatal("largest accepted size rejected")
	}
	if h.add(string(digest[:]), int64(hashWildcard), "Reserved.Size") || h.add(string(digest[:]), 1<<32, "Wrapped.Size") || h.add(string(digest[:]), -2, "Negative.Size") {
		t.Fatal("out-of-range size accepted")
	}
	h.freeze()
	if got := h.lookup(string(digest[:]), int64(hashWildcard)-1); got != "Max.Size" {
		t.Fatalf("largest size = %q", got)
	}
	for _, size := range []int64{int64(hashWildcard), 1 << 32, -1} {
		if got := h.lookup(string(digest[:]), size); got != "" {
			t.Fatalf("size %d aliased the largest size: %q", size, got)
		}
	}
	wildcard := newHashTable()
	wildcard.add(string(digest[:]), -1, "Any.Size")
	wildcard.freeze()
	sharded := newHashTable()
	sharded.shards = []*hashTable{h, wildcard}
	sharded.freeze()
	for _, table := range []*hashTable{wildcard, sharded} {
		for _, size := range []int64{int64(hashWildcard), 1 << 32, -1} {
			if got := table.lookup(string(digest[:]), size); got != "Any.Size" {
				t.Fatalf("wildcard size %d = %q", size, got)
			}
		}
	}
}

func TestHashTableLengthPrefixedNamesAcrossVarintBoundaries(t *testing.T) {
	h := newHashTable()
	for i, length := range []int{0, 1, 127, 128, 16383, 16384} {
		digest := md5.Sum([]byte{byte(i)})
		name := strings.Repeat("n", length)
		if !h.add(string(digest[:]), int64(i+1), name) {
			t.Fatalf("add length %d", length)
		}
		if !h.add(string(digest[:]), int64(i+1), "Discarded.Duplicate") {
			t.Fatalf("duplicate add length %d", length)
		}
	}
	h.freeze()
	if h.len() != 6 {
		t.Fatalf("unique rules = %d", h.len())
	}
	for i, length := range []int{0, 1, 127, 128, 16383, 16384} {
		digest := md5.Sum([]byte{byte(i)})
		if got := h.lookup(string(digest[:]), int64(i+1)); got != strings.Repeat("n", length) {
			t.Fatalf("name length %d returned %d bytes", length, len(got))
		}
	}
}

func TestNestedHashShardsPreserveExactAndLoadOrder(t *testing.T) {
	digest := md5.Sum([]byte("nested hash fixture"))
	firstMember, secondMember, laterSource := newHashTable(), newHashTable(), newHashTable()
	firstMember.add(string(digest[:]), -1, "First.Wildcard")
	secondMember.add(string(digest[:]), -1, "Second.Wildcard")
	secondMember.add(string(digest[:]), 19, "Second.Exact")
	laterSource.add(string(digest[:]), 19, "Later.Exact")
	container := newHashTable()
	container.shards = []*hashTable{firstMember, secondMember}
	all := newHashTable()
	all.shards = []*hashTable{container, laterSource}
	all.freeze()
	if got := all.len(); got != 2 {
		t.Fatalf("distinct nested records = %d, want 2", got)
	}
	if got := all.lookup(string(digest[:]), 19); got != "Second.Exact" {
		t.Fatalf("exact match = %q", got)
	}
	if got := all.lookup(string(digest[:]), 20); got != "First.Wildcard" {
		t.Fatalf("wildcard match = %q", got)
	}
	if !all.sizes()[-1] || !all.sizes()[19] {
		t.Fatalf("nested sizes = %v", all.sizes())
	}
}
