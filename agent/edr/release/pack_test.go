package release

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

func TestPackExtractsExactSignedPair(t *testing.T) {
	_, key, want, _, source := signedFixture(t)
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, ed25519.SeedSize))
	archive := filepath.Join(t.TempDir(), "release.zip")
	manifest, err := Pack(archive, source, private, want)
	if err != nil {
		t.Fatal(err)
	}
	stage, extracted, err := Extract(archive, t.TempDir(), key, want)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(stage)
	if len(manifest.Artifacts) != 2 || len(extracted.Artifacts) != 2 || VerifyFiles(stage, extracted) != nil {
		t.Fatal("published pair failed authenticated roundtrip")
	}
	if _, err := Pack(archive, source, private, want); !os.IsExist(err) {
		t.Fatalf("publisher overwrote an existing release: %v", err)
	}
}

func TestPackRejectsMissingKeyAndSymlinkSource(t *testing.T) {
	_, _, want, _, source := signedFixture(t)
	archive := filepath.Join(t.TempDir(), "release.zip")
	if _, err := Pack(archive, source, nil, want); err == nil {
		t.Fatal("publisher accepted missing key")
	}
	if _, err := os.Stat(archive); !os.IsNotExist(err) {
		t.Fatal("invalid key created archive")
	}
	provider := filepath.Join(source, want.ProviderName)
	if err := os.Remove(provider); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(want.ServiceName, provider); err != nil {
		t.Fatal(err)
	}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, ed25519.SeedSize))
	if _, err := Pack(archive, source, private, want); err == nil {
		t.Fatal("publisher accepted symlink provider")
	}
}

func TestPackLinuxServiceWithoutProvider(t *testing.T) {
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	want := Identity{PackageVersion: "12.0.56-native", Implementation: "native-go-v231", OS: "linux", Architecture: "amd64", ServiceName: "utmstack_edr_linux_amd64"}
	source, parent := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(source, want.ServiceName), []byte("inert Linux service"), 0600); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(parent, "release.zip")
	if _, err := Pack(archive, source, private, want); err != nil {
		t.Fatal(err)
	}
	stage, manifest, err := Extract(archive, parent, private.Public().(ed25519.PublicKey), want)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(stage)
	if len(manifest.Artifacts) != 1 || VerifyFiles(stage, manifest) != nil {
		t.Fatal("Linux single-artifact release did not roundtrip")
	}
}
