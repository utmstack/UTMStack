package release

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func signedFixture(t *testing.T) ([]byte, ed25519.PublicKey, Identity, Manifest, string) {
	t.Helper()
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, ed25519.SeedSize))
	public := private.Public().(ed25519.PublicKey)
	want := Identity{PackageVersion: "12.0.56-native", Implementation: "native-go-v231", OS: "windows", Architecture: "amd64", ServiceName: "utmstack_edr_windows_amd64.exe", ProviderName: "utmstack_amsi.dll"}
	base := t.TempDir()
	manifest := Manifest{Schema: 1, PackageVersion: want.PackageVersion, Implementation: want.Implementation, OS: want.OS, Architecture: want.Architecture}
	for _, name := range []string{want.ServiceName, want.ProviderName} {
		content := []byte("inert test artifact: " + name)
		if err := os.WriteFile(filepath.Join(base, name), content, 0600); err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(content)
		manifest.Artifacts = append(manifest.Artifacts, Artifact{Name: name, Bytes: int64(len(content)), SHA256: hex.EncodeToString(h[:])})
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := json.Marshal(Envelope{Payload: payload, Signature: ed25519.Sign(private, payload)})
	if err != nil {
		t.Fatal(err)
	}
	return envelope, public, want, manifest, base
}

func TestSignedPairVerification(t *testing.T) {
	envelope, key, want, manifest, base := signedFixture(t)
	got, err := Verify(envelope, key, want)
	if err != nil || len(got.Artifacts) != 2 {
		t.Fatalf("signed pair rejected: %+v %v", got, err)
	}
	if err := VerifyFiles(base, manifest); err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct {
		name string
		edit func(*Identity)
	}{
		{"package replay", func(i *Identity) { i.PackageVersion = "12.0.58-native" }},
		{"architecture replay", func(i *Identity) { i.Architecture = "arm64" }},
		{"service substitution", func(i *Identity) { i.ServiceName = "other.exe" }},
		{"missing provider expectation", func(i *Identity) { i.ProviderName = "" }},
	} {
		t.Run(change.name, func(t *testing.T) {
			identity := want
			change.edit(&identity)
			if _, err := Verify(envelope, key, identity); err == nil {
				t.Fatal("wrong release identity accepted")
			}
		})
	}
}

func TestSignedPairRejectsMutationAndUnsignedShapes(t *testing.T) {
	envelope, key, want, _, _ := signedFixture(t)
	var original Envelope
	if err := json.Unmarshal(envelope, &original); err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct {
		name string
		edit func(*Envelope)
	}{
		{"modified payload", func(e *Envelope) { e.Payload[0] ^= 1 }},
		{"modified signature", func(e *Envelope) { e.Signature[0] ^= 1 }},
		{"missing signature", func(e *Envelope) { e.Signature = nil }},
	} {
		t.Run(change.name, func(t *testing.T) {
			copy := Envelope{Payload: bytes.Clone(original.Payload), Signature: bytes.Clone(original.Signature)}
			change.edit(&copy)
			data, _ := json.Marshal(copy)
			if _, err := Verify(data, key, want); err == nil {
				t.Fatal("altered envelope accepted")
			}
		})
	}
	if _, err := Verify(envelope, nil, want); err == nil {
		t.Fatal("missing trust anchor accepted")
	}
	if _, err := Verify(append(envelope, []byte(` {}`)...), key, want); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}

func TestSignedPairRejectsAmbiguousPayload(t *testing.T) {
	_, key, want, manifest, _ := signedFixture(t)
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, ed25519.SeedSize))
	canonical, _ := json.Marshal(manifest)
	for _, payload := range [][]byte{
		append([]byte{' '}, canonical...),
		append([]byte(`{"schema":1,`), canonical[1:]...),
	} {
		envelope, _ := json.Marshal(Envelope{Payload: payload, Signature: ed25519.Sign(private, payload)})
		if _, err := Verify(envelope, key, want); err == nil {
			t.Fatal("signed noncanonical or duplicate-field JSON accepted")
		}
	}
}

func TestSignedPairRejectsSignedUnsafeArtifactLists(t *testing.T) {
	_, key, want, baseManifest, _ := signedFixture(t)
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, ed25519.SeedSize))
	for _, change := range []struct {
		name string
		edit func(*Manifest)
	}{
		{"missing DLL", func(m *Manifest) { m.Artifacts = m.Artifacts[:1] }},
		{"duplicate service", func(m *Manifest) { m.Artifacts[1] = m.Artifacts[0] }},
		{"path traversal", func(m *Manifest) { m.Artifacts[1].Name = "../provider.dll" }},
		{"extra file", func(m *Manifest) {
			m.Artifacts = append(m.Artifacts, Artifact{Name: "extra", Bytes: 1, SHA256: strings.Repeat("0", 64)})
		}},
		{"overlarge DLL", func(m *Manifest) { m.Artifacts[1].Bytes = maxProviderBytes + 1 }},
		{"invalid digest", func(m *Manifest) { m.Artifacts[1].SHA256 = strings.Repeat("g", 64) }},
	} {
		t.Run(change.name, func(t *testing.T) {
			m := baseManifest
			m.Artifacts = append([]Artifact(nil), m.Artifacts...)
			change.edit(&m)
			payload, _ := json.Marshal(m)
			envelope, _ := json.Marshal(Envelope{Payload: payload, Signature: ed25519.Sign(private, payload)})
			if _, err := Verify(envelope, key, want); err == nil {
				t.Fatal("signed unsafe artifact list accepted")
			}
		})
	}
}

func TestSignedPairFileDigestAndSymlink(t *testing.T) {
	_, _, want, manifest, base := signedFixture(t)
	provider := filepath.Join(base, want.ProviderName)
	if err := os.WriteFile(provider, []byte("changed same-size content"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFiles(base, manifest); err == nil {
		t.Fatal("mutated provider accepted")
	}
	if err := os.Remove(provider); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(want.ServiceName, provider); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFiles(base, manifest); err == nil {
		t.Fatal("symlink provider accepted")
	}
}
