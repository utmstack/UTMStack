package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/feed"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func packFixture(t *testing.T) (packOptions, ed25519.PublicKey, packSpec) {
	t.Helper()
	dir := t.TempDir()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	o := packOptions{Spec: filepath.Join(dir, "spec.json"), Key: filepath.Join(dir, "key.secret"), Output: filepath.Join(dir, "release"), Sequence: 7, Expires: time.Now().Add(time.Hour)}
	if err = os.WriteFile(o.Key, []byte(base64.StdEncoding.EncodeToString(private.Seed())), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "test.ndb"), []byte("Test.Feed:0:*:41424344\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "hashes.list"), []byte(strings.Repeat("a", 64)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	spec := packSpec{Version: "unit-test-release", Files: []packFile{{Name: "test.ndb", Path: "test.ndb", Format: "signature"}, {Name: "intel.hsb", Path: "hashes.list", Format: "sha256-list"}}}
	writeSpec(t, o.Spec, spec)
	return o, public, spec
}

func writeSpec(t *testing.T, name string, spec packSpec) {
	t.Helper()
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, b, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPublishSignsExactPayloadAndPreservesOriginalArtifacts(t *testing.T) {
	o, public, _ := packFixture(t)
	if err := publish(o); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(o.Output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope feed.ManifestEnvelope
	if err = json.Unmarshal(b, &envelope); err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(public, payload, sig) {
		t.Fatal("invalid manifest signature")
	}
	var manifest feed.Manifest
	if err = json.Unmarshal(payload, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Sequence != 7 || len(manifest.Files) != 2 {
		t.Fatal("incorrect manifest")
	}
	for _, item := range manifest.Files {
		artifact, err := os.ReadFile(filepath.Join(o.Output, filepath.FromSlash(item.URL)))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(artifact)
		if hex.EncodeToString(digest[:]) != item.SHA256 {
			t.Fatal("checksum does not cover original artifact")
		}
		if _, err := feed.NormalizeArtifact(item, artifact); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(o.Output)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "artifacts" || entries[1].Name() != "manifest.json" {
		t.Fatal("unexpected published file (possibly key or staging data)")
	}
	if err := publish(o); err == nil {
		t.Fatal("existing release overwritten")
	}
}

func TestPublishDetachedSignaturePolicyFailsClosed(t *testing.T) {
	o, _, spec := packFixture(t)
	dir := filepath.Dir(o.Spec)
	if err := os.WriteFile(filepath.Join(dir, "official.cvd"), append([]byte("ClamAV-VDB:"), make([]byte, 512)...), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files = []packFile{{Name: "official.cvd", Path: "official.cvd", Official: true}}
	writeSpec(t, o.Spec, spec)
	o.RequireDetachedSignatures = true
	if err := publish(o); err == nil || !strings.Contains(err.Error(), "no detached signature") {
		t.Fatalf("required detached signature omitted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "official.sign"), []byte("#clamsign-1.0\n90::pkcs7-pem:AAAA\n"), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files[0].DetachedSignature = "official.sign"
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err == nil || !strings.Contains(err.Error(), "verifying detached signature") {
		t.Fatalf("invalid detached signature accepted: %v", err)
	}
	if _, err := os.Stat(o.Output); !os.IsNotExist(err) {
		t.Fatalf("failed verification created release directory: %v", err)
	}
}

func TestPublishDeduplicatesThreatWindsHashesWithoutLosingSizeCoverage(t *testing.T) {
	o, _, spec := packFixture(t)
	dir := filepath.Dir(o.Spec)
	a := md5.Sum([]byte("official wildcard"))
	b := md5.Sum([]byte("official exact"))
	c := md5.Sum([]byte("threatwinds only"))
	ah, bh, ch := hex.EncodeToString(a[:]), hex.EncodeToString(b[:]), hex.EncodeToString(c[:])
	official := ah + ":*:Official.Wildcard:73\n" + bh + ":14:Official.Exact\n"
	if err := os.WriteFile(filepath.Join(dir, "official.hdb"), []byte(official), 0600); err != nil {
		t.Fatal(err)
	}
	list := strings.Join([]string{ah, bh, ch, ch, ""}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "md5.list"), []byte(list), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files = []packFile{
		{Name: "official.hdb", Path: "official.hdb", Format: "signature", Official: true},
		{Name: "intel.hdb", Path: "md5.list", Format: "md5-list", DedupeAgainstOfficial: true},
	}
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err != nil {
		t.Fatal(err)
	}
	bundle, err := os.ReadFile(filepath.Join(o.Output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope feed.ManifestEnvelope
	if err := json.Unmarshal(bundle, &envelope); err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var manifest feed.Manifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 2 || manifest.Files[1].Format != "signature" {
		t.Fatalf("deduplicated artifact must be signed as native signature: %+v", manifest.Files)
	}
	data, err := os.ReadFile(filepath.Join(o.Output, filepath.FromSlash(manifest.Files[1].URL)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), ah) || strings.Count(string(data), bh) != 2 || strings.Count(string(data), ch) != 2 || strings.Count(string(data), "\n") != 2 {
		t.Fatalf("wrong deduplicated rows: %q", data)
	}
	db, err := signatures.Compile(map[string][]byte{"official.hdb": []byte(official), "intel.hdb": data})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ object, want string }{
		{"official wildcard", "Official.Wildcard"},
		{"official exact", "Official.Exact"},
		{"threatwinds only", "UTMStack.Intel.MD5." + ch},
	} {
		got, err := db.Match(context.Background(), []byte(tc.object))
		if err != nil || got != tc.want {
			t.Fatalf("match %q: %q, %v; want %q", tc.object, got, err, tc.want)
		}
	}
}

func TestPublishDeduplicationRequiresOfficialSourcesAndOmitsEmptyArtifact(t *testing.T) {
	o, _, spec := packFixture(t)
	spec.Files[1].DedupeAgainstOfficial = true
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err == nil || !strings.Contains(err.Error(), "official") {
		t.Fatalf("missing official source accepted: %v", err)
	}
	// An all-covered ThreatWinds source is omitted, not published as a
	// zero-rule artifact that the endpoint would reject.
	dir := filepath.Dir(o.Spec)
	digest := strings.Repeat("a", 64)
	if err := os.WriteFile(filepath.Join(dir, "official.hsb"), []byte(digest+":*:Official.Hash:73\n"), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files = []packFile{
		{Name: "official.hsb", Path: "official.hsb", Official: true},
		{Name: "intel.hsb", Path: "hashes.list", Format: "sha256-list", DedupeAgainstOfficial: true},
	}
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err != nil {
		t.Fatal(err)
	}
	bundle, err := os.ReadFile(filepath.Join(o.Output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope feed.ManifestEnvelope
	if err := json.Unmarshal(bundle, &envelope); err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var manifest feed.Manifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Name != "official.hsb" {
		t.Fatalf("expected only official rule: %+v", manifest.Files)
	}
}

func TestPublishDeduplicationHonorsGenerationIgnoreLists(t *testing.T) {
	o, _, spec := packFixture(t)
	dir := filepath.Dir(o.Spec)
	digest := strings.Repeat("a", 64)
	if err := os.WriteFile(filepath.Join(dir, "official.hsb"), []byte(digest+":*:Official.Ignored:73\n"+strings.Repeat("b", 64)+":*:Official.Kept:73\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore.ign2"), []byte("Official.Ignored\n"), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files = []packFile{
		{Name: "official.hsb", Path: "official.hsb", Official: true},
		{Name: "ignore.ign2", Path: "ignore.ign2"},
		{Name: "intel.hsb", Path: "hashes.list", Format: "sha256-list", DedupeAgainstOfficial: true},
	}
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err != nil {
		t.Fatal(err)
	}
	bundle, err := os.ReadFile(filepath.Join(o.Output, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope feed.ManifestEnvelope
	if err := json.Unmarshal(bundle, &envelope); err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var manifest feed.Manifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 3 {
		t.Fatalf("ignored official rule improperly removed indicator: %+v", manifest.Files)
	}
	data, err := os.ReadFile(filepath.Join(o.Output, filepath.FromSlash(manifest.Files[2].URL)))
	if err != nil || !strings.HasPrefix(string(data), digest+":*:") {
		t.Fatalf("ignored official rule should leave ThreatWinds indicator: %q, %v", data, err)
	}
}

func TestPublishDeduplicationRejectsUnmarkedSelectionContainer(t *testing.T) {
	o, _, spec := packFixture(t)
	dir := filepath.Dir(o.Spec)
	if err := os.WriteFile(filepath.Join(dir, "extra.cvd"), unsignedContainer(t, "extra.ign2", []byte("Official.Hash\n")), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files[0].Official = true
	spec.Files[1].DedupeAgainstOfficial = true
	spec.Files = append(spec.Files, packFile{Name: "extra.cvd", Path: "extra.cvd"})
	writeSpec(t, o.Spec, spec)
	if err := publish(o); err == nil || !strings.Contains(err.Error(), "unmarked container") {
		t.Fatalf("unmarked container accepted: %v", err)
	}
}

func TestPublisherRejectsUnsupportedOrInvalidInputWithoutPublishing(t *testing.T) {
	cases := []struct {
		name   string
		change func(*packOptions, *packSpec)
	}{
		{"zero sequence", func(o *packOptions, _ *packSpec) { o.Sequence = 0 }},
		{"expired", func(o *packOptions, _ *packSpec) { o.Expires = time.Now().Add(-time.Minute) }},
		{"missing key", func(o *packOptions, _ *packSpec) { o.Key += "-missing" }},
		{"unsupported rules", func(_ *packOptions, s *packSpec) { s.Files[0].Name = "test.cbc" }},
		{"duplicate filename", func(_ *packOptions, s *packSpec) { s.Files = append(s.Files, s.Files[0]) }},
		{"portable case collision", func(_ *packOptions, s *packSpec) {
			s.Files = append(s.Files, packFile{Name: "TEST.NDB", Path: "test.ndb"})
		}},
		{"traversal", func(_ *packOptions, s *packSpec) { s.Files[0].Name = "../test.ndb" }},
		{"malformed hash list", func(_ *packOptions, s *packSpec) { s.Files[1].Path = "test.ndb" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o, _, s := packFixture(t)
			tc.change(&o, &s)
			writeSpec(t, o.Spec, s)
			if err := publish(o); err == nil {
				t.Fatal("invalid publication accepted")
			}
			if _, err := os.Stat(o.Output); !os.IsNotExist(err) {
				t.Fatal("failed validation left a published directory")
			}
		})
	}
}

func unsignedContainer(t *testing.T, member string, content []byte) []byte {
	t.Helper()
	var payload bytes.Buffer
	z := gzip.NewWriter(&payload)
	tw := tar.NewWriter(z)
	if err := tw.WriteHeader(&tar.Header{Name: member, Size: int64(len(content)), Mode: 0600}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	header := fmt.Sprintf("ClamAV-VDB:now:5:1:90:%x:unsigned:builder:1", md5.Sum(payload.Bytes()))
	return append([]byte(header+strings.Repeat(" ", 512-len(header))), payload.Bytes()...)
}

// Re-signing must not launder provenance: a container is published only when
// its upstream publisher signature verifies, unless explicitly overridden.
func TestPublishRequiresAuthenticatedContainers(t *testing.T) {
	o, _, spec := packFixture(t)
	dir := filepath.Dir(o.Spec)
	if err := os.WriteFile(filepath.Join(dir, "daily.cvd"), unsignedContainer(t, "daily.ndb", []byte("Test.Container:0:*:45464748\n")), 0600); err != nil {
		t.Fatal(err)
	}
	spec.Files = append(spec.Files, packFile{Name: "daily.cvd", Path: "daily.cvd", Format: "signature"})
	writeSpec(t, o.Spec, spec)
	err := publish(o)
	if err == nil || !strings.Contains(err.Error(), "publisher signature") {
		t.Fatalf("unsigned container published: %v", err)
	}
	if _, statErr := os.Stat(o.Output); !os.IsNotExist(statErr) {
		t.Fatal("rejected publication left output behind")
	}
	o.AllowUnauthenticatedContainers = true
	if err = publish(o); err != nil {
		t.Fatalf("explicit override: %v", err)
	}
}
