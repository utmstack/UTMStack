package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/feed"
)

func releaseFixture(t *testing.T, key ed25519.PrivateKey, sequence uint64, body string, expires time.Time) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "artifacts"), 0700); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(body))
	digest := hex.EncodeToString(h[:])
	if err := os.WriteFile(filepath.Join(dir, "artifacts", digest), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	m := feed.Manifest{Version: "fixture", Sequence: sequence, ExpiresAt: expires,
		Files: []feed.ManifestFile{{Name: "test.ndb", URL: "artifacts/" + digest, SHA256: digest}}}
	payload, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	env := feed.ManifestEnvelope{Payload: base64.StdEncoding.EncodeToString(payload), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func testKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestPromoteArtifactsBeforeAtomicManifest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("publisher promotion requires the documented POSIX directory-sync contract")
	}
	pub, priv := testKey(t)
	now := time.Now()
	one := releaseFixture(t, priv, 1, "first", now.Add(time.Hour))
	two := releaseFixture(t, priv, 2, "second", now.Add(time.Hour))
	mirror := t.TempDir()
	if err := promote(one, mirror, pub, now); err != nil {
		t.Fatal(err)
	}
	firstManifest, err := os.ReadFile(filepath.Join(mirror, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := promote(two, mirror, pub, now); err != nil {
		t.Fatal(err)
	}
	secondManifest, err := os.ReadFile(filepath.Join(mirror, "manifest.json"))
	if err != nil || bytes.Equal(firstManifest, secondManifest) {
		t.Fatalf("manifest was not switched: %v", err)
	}
	for _, body := range []string{"first", "second"} {
		h := sha256.Sum256([]byte(body))
		b, err := os.ReadFile(filepath.Join(mirror, "artifacts", hex.EncodeToString(h[:])))
		if err != nil || string(b) != body {
			t.Fatalf("promoted artifact missing: %v", err)
		}
	}
	if err := promote(one, mirror, pub, now); err == nil || !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("rollback accepted: %v", err)
	}
	if err := promote(two, mirror, pub, now); err != nil {
		t.Fatalf("idempotent promotion failed: %v", err)
	}
	if err := os.Remove(filepath.Join(mirror, "artifacts", digestText("second"))); err != nil {
		t.Fatal(err)
	}
	if err := promote(two, mirror, pub, now); err != nil {
		t.Fatalf("same-sequence artifact repair failed: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(mirror, "artifacts", digestText("second"))); err != nil || string(b) != "second" {
		t.Fatalf("artifact not repaired: %v", err)
	}
}

func digestText(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func rewriteManifest(t *testing.T, dir string, key ed25519.PrivateKey, change func(*feed.Manifest)) {
	t.Helper()
	path := filepath.Join(dir, "manifest.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var env feed.ManifestEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var m feed.Manifest
	if err := json.Unmarshal(payload, &m); err != nil {
		t.Fatal(err)
	}
	change(&m)
	payload, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	env.Payload = base64.StdEncoding.EncodeToString(payload)
	env.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))
	b, err = json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPromoteRejectsInvalidReleaseWithoutChangingActiveManifest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("publisher promotion requires the documented POSIX directory-sync contract")
	}
	pub, priv := testKey(t)
	now := time.Now()
	baseline := releaseFixture(t, priv, 3, "baseline", now.Add(time.Hour))
	mirror := t.TempDir()
	if err := promote(baseline, mirror, pub, now); err != nil {
		t.Fatal(err)
	}
	active, err := os.ReadFile(filepath.Join(mirror, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		seq        uint64
		modify     func(string) error
		want       string
	}{
		{"sequence reused", "other", 3, nil, "sequence reuse"},
		{"expired", "expired", 4, nil, "expired"},
		{"tampered artifact", "tamper", 4, func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "artifacts", digestText("tamper")), []byte("changed"), 0600)
		}, "SHA-256"},
		{"tampered signature", "signature", 4, func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"payload":"eA==","signature":"AA=="}`), 0600)
		}, "signature"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expiry := now.Add(time.Hour)
			if tc.name == "expired" {
				expiry = now.Add(-time.Second)
			}
			release := releaseFixture(t, priv, tc.seq, tc.body, expiry)
			if tc.modify != nil {
				if err := tc.modify(release); err != nil {
					t.Fatal(err)
				}
			}
			if err := promote(release, mirror, pub, now); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("promotion error = %v; want %q", err, tc.want)
			}
			got, err := os.ReadFile(filepath.Join(mirror, "manifest.json"))
			if err != nil || !bytes.Equal(got, active) {
				t.Fatalf("active manifest changed on rejection: %v", err)
			}
		})
	}
}

func TestPromoteRejectsLockAndExistingWrongArtifact(t *testing.T) {
	pub, priv := testKey(t)
	now := time.Now()
	release := releaseFixture(t, priv, 1, "known", now.Add(time.Hour))
	mirror := t.TempDir()
	if err := os.WriteFile(filepath.Join(mirror, ".feedpromote.lock"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := promote(release, mirror, pub, now); err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("concurrent promotion allowed: %v", err)
	}
	if err := os.Remove(filepath.Join(mirror, ".feedpromote.lock")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(mirror, "artifacts"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirror, "artifacts", digestText("known")), []byte("wrong"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := promote(release, mirror, pub, now); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("wrong artifact accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(mirror, "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("manifest published after failed artifact check: %v", err)
	}
}

func TestSignedArtifactDescriptionRejectsUnsafePathsAndFormats(t *testing.T) {
	item := feed.ManifestFile{Name: "intel.hdb", URL: "artifacts/" + strings.Repeat("a", 64), SHA256: strings.Repeat("a", 64), Format: "md5-list.tar.gz"}
	if !validArtifact(item) {
		t.Fatal("valid hash-list artifact rejected")
	}
	for _, changed := range []feed.ManifestFile{
		{Name: "../intel.hdb", URL: item.URL, SHA256: item.SHA256, Format: item.Format},
		{Name: item.Name, URL: "../artifacts/" + item.SHA256, SHA256: item.SHA256, Format: item.Format},
		{Name: item.Name, URL: item.URL, SHA256: strings.ToUpper(item.SHA256), Format: item.Format},
		{Name: "intel.hsb", URL: item.URL, SHA256: item.SHA256, Format: item.Format},
		{Name: item.Name, URL: item.URL, SHA256: item.SHA256, Format: "invented"},
	} {
		if validArtifact(changed) {
			t.Fatalf("unsafe artifact accepted: %+v", changed)
		}
	}
}

func TestPromoteRejectsSignedButInvalidHashList(t *testing.T) {
	pub, priv := testKey(t)
	now := time.Now()
	release := releaseFixture(t, priv, 1, "not a digest", now.Add(time.Hour))
	rewriteManifest(t, release, priv, func(m *feed.Manifest) {
		m.Files[0].Name = "intel.hdb"
		m.Files[0].Format = "md5-list"
	})
	mirror := t.TempDir()
	if err := promote(release, mirror, pub, now); err == nil || !strings.Contains(err.Error(), "invalid hash-list") {
		t.Fatalf("invalid signed list accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(mirror, "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("invalid list was published: %v", err)
	}
}
