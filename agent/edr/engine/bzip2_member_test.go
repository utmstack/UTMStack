package engine

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func bzip2MemberFixtures(t *testing.T) map[string][]byte {
	t.Helper()
	f, err := os.Open("testdata/bzip2_member_scanner.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	a := tar.NewReader(z)
	files := map[string][]byte{}
	for {
		h, err := a.Next()
		if err == io.EOF {
			break
		}
		if err != nil || h.Size > 1<<20 || !h.FileInfo().Mode().IsRegular() {
			t.Fatal("invalid bzip2 scanner fixture", err)
		}
		files[h.Name], err = io.ReadAll(a)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) != 6 {
		t.Fatal("wrong bzip2 scanner fixture inventory")
	}
	return files
}

func TestBzip2MemberBoundaryAndDamagedPrefix(t *testing.T) {
	files := bzip2MemberFixtures(t)
	first := []byte("first harmless archive member\n")
	second := []byte("second harmless archive member\n")
	combined := append(append([]byte(nil), first...), second...)
	crcBody := make([]byte, 25000)
	for i := range crcBody {
		crcBody[i] = byte(i*29 + 71)
	}
	for _, tc := range []struct {
		name   string
		hashed []byte
		hit    bool
		reason string
	}{
		{"single-first.bz2", first, true, ""},
		{"concatenated-first.bz2", first, true, ""},
		{"concatenated-first.bz2", second, true, ""},
		{"concatenated-first.bz2", combined, false, ""},
		{"trailing-text-first.bz2", first, true, ""},
		{"trailing-text-first.bz2", combined, false, ""},
		{"crc-8192.bz2", crcBody[:8192], false, "Malformed"},
		{"crc-8193.bz2", crcBody[:8192], true, ""},
		{"crc-25000.bz2", crcBody[:24576], true, ""},
		{"crc-25000.bz2", crcBody, false, "Malformed"},
	} {
		t.Run(fmt.Sprintf("%s/%d", tc.name, len(tc.hashed)), func(t *testing.T) {
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			rule := []byte(fmt.Sprintf("%x:%d:Bzip.Member.Control\n", md5.Sum(tc.hashed), len(tc.hashed)))
			if _, err := e.ActivateGeneration(map[string][]byte{"member.hdb": rule}, 1, manifestHash(tc.name)); err != nil {
				t.Fatal(err)
			}
			clean, name, err := e.Scan(context.Background(), files[tc.name])
			if tc.reason != "" {
				if clean || name != "" || IncompleteReason(err) != tc.reason {
					t.Fatalf("damaged member: clean=%v name=%q error=%v", clean, name, err)
				}
			} else if err != nil || clean == tc.hit || (name == "Bzip.Member.Control") != tc.hit {
				t.Fatalf("member: clean=%v name=%q error=%v", clean, name, err)
			}
		})
	}
}

func TestBzip2MemberChargesDiscardedExpansion(t *testing.T) {
	data := bzip2MemberFixtures(t)["crc-25000.bz2"]
	for _, budget := range []int64{0, 1, 8191, 8192, 8193, 24576, 24999, 25000} {
		s := pipelineState(t, map[string]string{"never.ndb": "Bzip.Never:0:*:4142434445463736353433323130\n"})
		s.remaining = budget
		s.maxFile = 1 << 20
		sig, err := s.bzip2Member(data, 0)
		if sig != "" || s.remaining != max(int64(0), budget-25000) {
			t.Fatalf("budget=%d remaining=%d signature=%q", budget, s.remaining, sig)
		}
		want := "ResourceLimit"
		if budget == 25000 {
			want = "Malformed"
		}
		if IncompleteReason(err) != want {
			t.Fatalf("budget=%d error=%v want=%s", budget, err, want)
		}
	}
}

func TestBzip2MemberEmptyStreamsAreBounded(t *testing.T) {
	empty := []byte("BZh9\x17\x72\x45\x38\x50\x90\x00\x00\x00\x00")
	s := pipelineState(t, map[string]string{"never.ndb": "Bzip.Never:0:*:4142434445463736353433323130\n"})
	s.maxFiles = 2
	name, err := s.bzip2Member(bytes.Repeat(empty, 3), 0)
	if name != "" || IncompleteReason(err) != "ResourceLimit" {
		t.Fatalf("empty member bound: name=%q error=%v", name, err)
	}
}
