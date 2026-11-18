package feed

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func hashListArchive(t *testing.T, data []byte, members int) []byte {
	t.Helper()
	var out bytes.Buffer
	z := gzip.NewWriter(&out)
	w := tar.NewWriter(z)
	for i := 0; i < members; i++ {
		if err := w.WriteHeader(&tar.Header{Name: "hashes.list", Mode: 0600, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestPreparedHashListArtifactFormatsMatchLegacyBytes(t *testing.T) {
	for _, tc := range []struct {
		format, ext string
		length      int
	}{
		{"md5-list", ".hdb", 32}, {"sha1-list", ".hsb", 40}, {"sha256-list", ".hsb", 64},
	} {
		input := []byte("# indicators\n" + strings.Repeat("A", tc.length) + "\r\n")
		var compressed bytes.Buffer
		z := gzip.NewWriter(&compressed)
		if _, err := z.Write(input); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		for suffix, raw := range map[string][]byte{
			"": input, ".gz": compressed.Bytes(), ".tar.gz": hashListArchive(t, input, 1),
		} {
			t.Run(tc.format+suffix, func(t *testing.T) {
				item := ManifestFile{Name: "intel" + tc.ext, Format: tc.format + suffix}
				legacy, err := NormalizeArtifact(item, raw)
				if err != nil {
					t.Fatal(err)
				}
				wire, prepared, err := normalizeArtifact(item, raw, true)
				if err != nil || prepared == nil || !bytes.Equal(wire, legacy) {
					t.Fatalf("native artifact differs: err=%v wire=%q legacy=%q", err, wire, legacy)
				}
				db, localErr, fatalErr := signatures.CompileWithPreparedLocalFiles(map[string][]byte{item.Name: wire}, nil,
					map[string]*signatures.PreparedHashList{item.Name: prepared}, signatures.Limits{})
				if db == nil || localErr != nil || fatalErr != nil || db.DetectionCount() != 1 {
					t.Fatalf("native artifact rejected: local=%v fatal=%v", localErr, fatalErr)
				}
			})
		}
	}
}

func TestHashListsNormalizeOnlyValidWholeDigests(t *testing.T) {
	for _, tc := range []struct {
		format, ext string
		length      int
	}{{"md5-list", ".hdb", 32}, {"sha1-list", ".hsb", 40}, {"sha256-list", ".hsb", 64}} {
		t.Run(tc.format, func(t *testing.T) {
			input := []byte("# upstream snapshot\n" + strings.Repeat("A", tc.length) + "\r\n")
			out, err := NormalizeArtifact(ManifestFile{Name: "intel" + tc.ext, Format: tc.format}, input)
			if err != nil || !strings.Contains(string(out), strings.Repeat("a", tc.length)+":*:UTMStack.Intel.") || !strings.HasSuffix(string(out), ":73\n") {
				t.Fatalf("%q %v", out, err)
			}
			archive := hashListArchive(t, input, 1)
			archived, err := NormalizeArtifact(ManifestFile{Name: "intel" + tc.ext, Format: tc.format + ".tar.gz"}, archive)
			if err != nil || !bytes.Equal(out, archived) {
				t.Fatalf("tar decode: %q %v", archived, err)
			}
			if _, err := NormalizeArtifact(ManifestFile{Name: "intel" + tc.ext, Format: tc.format}, append(input, []byte("garbage\n")...)); err == nil {
				t.Fatal("malformed line silently ignored")
			}
		})
	}
}

func TestHashListArchiveRejectsAmbiguousOrEmptyInput(t *testing.T) {
	item := ManifestFile{Name: "intel.hsb", Format: "sha256-list.tar.gz"}
	for _, n := range []int{0, 2} {
		if _, err := NormalizeArtifact(item, hashListArchive(t, []byte(strings.Repeat("a", 64)+"\n"), n)); err == nil {
			t.Fatalf("accepted %d archive members", n)
		}
	}
	if _, err := NormalizeArtifact(item, []byte("not gzip")); err == nil {
		t.Fatal("malformed gzip accepted")
	}
	if _, err := NormalizeArtifact(ManifestFile{Name: "intel.hsb", Format: "sha256-list"}, []byte("# empty\n")); err == nil {
		t.Fatal("empty list accepted")
	}
	if _, err := NormalizeArtifact(ManifestFile{Name: "intel.hdb", Format: "sha256-list"}, []byte(strings.Repeat("a", 64))); err == nil {
		t.Fatal("wrong extension accepted")
	}
}

// The reference engine never matches any signature against an object of five
// bytes or fewer, so an indicator for the empty object can never fire there.
// The adapter drops it instead of emitting a rule that flags every empty file.
func TestHashListsDropEmptyObjectDigests(t *testing.T) {
	for _, tc := range []struct {
		format, name, empty string
	}{
		{"md5-list", "intel.hdb", "d41d8cd98f00b204e9800998ecf8427e"},
		{"sha1-list", "intel.hsb", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"sha256-list", "intel.hsb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			other := strings.Repeat("a", len(tc.empty))
			input := []byte(strings.ToUpper(tc.empty) + "\n" + other + "\n" + tc.empty + "\n")
			out, err := NormalizeArtifact(ManifestFile{Name: tc.name, Format: tc.format}, input)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(out), tc.empty) || strings.Count(string(out), "\n") != 1 || !strings.HasPrefix(string(out), other+":*:") {
				t.Fatalf("empty-object digest kept or other indicator lost: %q", out)
			}
			if _, err := NormalizeArtifact(ManifestFile{Name: tc.name, Format: tc.format}, []byte(tc.empty+"\n")); err == nil {
				t.Fatal("list with no usable indicator accepted")
			}
		})
	}
}
