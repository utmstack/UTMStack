package release

import (
	"archive/zip"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func archiveFixture(t *testing.T, change func(map[string][]byte)) (string, ed25519.PublicKey, Identity, string) {
	t.Helper()
	envelope, key, want, _, source := signedFixture(t)
	entries := map[string][]byte{"release.json": envelope}
	for _, name := range []string{want.ServiceName, want.ProviderName} {
		b, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		entries[name] = b
	}
	if change != nil {
		change(entries)
	}
	archive := filepath.Join(t.TempDir(), "release.zip")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, data := range entries {
		entry, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return archive, key, want, t.TempDir()
}

func TestExtractSignedPair(t *testing.T) {
	archive, key, want, parent := archiveFixture(t, nil)
	stage, manifest, err := Extract(archive, parent, key, want)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(stage)
	if err := VerifyFiles(stage, manifest); err != nil {
		t.Fatal(err)
	}
	if err := CheckPrivateDirectory(stage); err != nil {
		t.Fatalf("staging directory privacy: %v", err)
	}
}

func TestExtractRejectsUntrustedArchiveWithoutRetainingFiles(t *testing.T) {
	for _, change := range []struct {
		name string
		edit func(map[string][]byte)
	}{
		{"changed service", func(files map[string][]byte) { files["utmstack_edr_windows_amd64.exe"] = []byte("tampered service") }},
		{"missing provider", func(files map[string][]byte) { delete(files, "utmstack_amsi.dll") }},
		{"extra artifact", func(files map[string][]byte) { files["extra.txt"] = []byte("extra") }},
		{"unsafe path", func(files map[string][]byte) { files["../escape"] = []byte("escape") }},
		{"changed manifest", func(files map[string][]byte) {
			var e Envelope
			if err := json.Unmarshal(files["release.json"], &e); err != nil {
				panic(err)
			}
			e.Payload[0] ^= 1
			files["release.json"], _ = json.Marshal(e)
		}},
	} {
		t.Run(change.name, func(t *testing.T) {
			archive, key, want, parent := archiveFixture(t, change.edit)
			if stage, _, err := Extract(archive, parent, key, want); err == nil {
				os.RemoveAll(stage)
				t.Fatal("untrusted archive accepted")
			}
			entries, err := os.ReadDir(parent)
			if err != nil || len(entries) != 0 {
				t.Fatalf("rejected archive retained staging files: %v %v", entries, err)
			}
		})
	}
}
