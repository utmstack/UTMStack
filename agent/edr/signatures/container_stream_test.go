package signatures

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestHashReserveCountsByDigestWidth(t *testing.T) {
	md5Digest := md5.Sum([]byte("md5"))
	shaDigest := sha256.Sum256([]byte("sha256"))
	var hint hashReserve
	hint.observe(".hdb", []byte(hex.EncodeToString(md5Digest[:])+":7:First.Name"))
	hint.observe(".mdb", []byte("9:"+hex.EncodeToString(md5Digest[:])+":Section.Name"))
	hint.observe(".hsb", []byte(hex.EncodeToString(shaDigest[:])+":*:SHA.Name:123"))
	hint.observe(".hdb", []byte("# ignored"))
	hint.observe(".hdb", []byte("missing-fields"))
	hint.observe(".hdb", []byte(hex.EncodeToString(md5Digest[:])+":7:"+strings.Repeat("x", 257)))
	if hint[0].count != 2 || hint[1].count != 0 || hint[2].count != 1 {
		t.Fatalf("digest counts = %+v", hint)
	}
	if got, want := hint[0].nameBytes, int(hashNameEncodedSize(len("First.Name"))+hashNameEncodedSize(len("Section.Name"))); got != want {
		t.Fatalf("MD5 name reservation = %d, want %d", got, want)
	}
	if got, want := hint[2].nameBytes, int(hashNameEncodedSize(len("SHA.Name"))); got != want {
		t.Fatalf("SHA256 name reservation = %d, want %d", got, want)
	}
	if !hint.bounded(DefaultLimits()) {
		t.Fatal("small valid reservation rejected")
	}
}

func TestReservedHashTablePreservesFirstRule(t *testing.T) {
	digest := md5.Sum([]byte("same digest"))
	hint := hashReserve{{count: 2, nameBytes: int(hashNameEncodedSize(len("First")) + hashNameEncodedSize(len("Second")))}}
	h := newHashTable()
	h.reserve(hint)
	if cap(h.groups[0].records) < 2*20 || cap(h.groups[0].offsets) < 2 || cap(h.groups[0].arena) < hint[0].nameBytes {
		t.Fatal("reservation was not applied")
	}
	if !h.add(string(digest[:]), 5, "First") || !h.add(string(digest[:]), 5, "Second") {
		t.Fatal("failed to add reserved entries")
	}
	h.freeze()
	if h.len() != 1 || h.lookup(string(digest[:]), 5) != "First" {
		t.Fatal("reservation changed duplicate resolution")
	}
}

func TestPatternCandidatesAreBoundedUpperEstimate(t *testing.T) {
	for _, tc := range []struct {
		ext, line string
		want      int
	}{
		{".ldb", "Name;Engine:81-255,Target:0;0;0:aa;1:bb", 2},
		{".ldb", "# comment;ignored;0;0:aa", 0},
		{".ndb", "Name:0:*:aabb", 1},
		{".db", "Name=616263", 1},
		{".ndb", "", 0},
	} {
		if got := patternCandidates(tc.ext, []byte(tc.line)); got != tc.want {
			t.Errorf("%s %q: %d, want %d", tc.ext, tc.line, got, tc.want)
		}
	}
}
