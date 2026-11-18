// feedpack validates and signs a native signature generation for an HTTPS mirror.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/feed"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// artifactLimit and generationLimit equal the endpoint feed limits, so every
// published generation can be downloaded. They admit the official containers
// (about 85, 22 and 0.3 MiB) with headroom.
const artifactLimit = 256 << 20
const generationLimit = 512 << 20

type packSpec struct {
	Version string     `json:"version"`
	Files   []packFile `json:"files"`
}
type packFile struct {
	Name                  string `json:"name"`
	Path                  string `json:"path"`
	Format                string `json:"format,omitempty"`
	Official              bool   `json:"official,omitempty"`
	DetachedSignature     string `json:"detached_signature,omitempty"`
	DedupeAgainstOfficial bool   `json:"dedupe_against_official,omitempty"`
}
type packOptions struct {
	Spec     string
	Key      string
	Output   string
	Sequence uint64
	Expires  time.Time
	// AllowUnauthenticatedContainers permits locally built containers that
	// carry no valid upstream publisher signature.
	AllowUnauthenticatedContainers bool
	RequireDetachedSignatures      bool
}

type packArtifact struct {
	input      packFile
	data       []byte
	normalized []byte
}

type dedupeStats struct {
	Official int
	Repeated int
	Retained int
	Omitted  int
}

func main() {
	var o packOptions
	var expiry string
	flag.StringVar(&o.Spec, "spec", "", "JSON input spec (version and files with name/path/format)")
	flag.StringVar(&o.Key, "key", "", "file containing base64 Ed25519 seed or private key")
	flag.StringVar(&o.Output, "out", "", "new output directory (must not already exist)")
	flag.Uint64Var(&o.Sequence, "sequence", 0, "positive monotonically increasing release sequence")
	flag.StringVar(&expiry, "expires", "", "manifest expiration in RFC3339 format")
	flag.BoolVar(&o.AllowUnauthenticatedContainers, "allow-unauthenticated-containers", false, "accept signature containers without a valid upstream publisher signature (locally built containers only)")
	flag.BoolVar(&o.RequireDetachedSignatures, "require-detached-signatures", false, "require pinned-root PKCS#7 .sign verification for each official CVD")
	flag.Parse()
	var err error
	o.Expires, err = time.Parse(time.RFC3339, expiry)
	if err == nil {
		err = publish(o)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "feedpack:", err)
		os.Exit(1)
	}
	fmt.Println("Validated signed generation written to", o.Output)
}

func readRegular(name string, max int64) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > max {
		return nil, errors.New("input must be a bounded regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, errors.New("input exceeds size limit")
	}
	return b, nil
}

func loadKey(name string) (ed25519.PrivateKey, error) {
	b, err := readRegular(name, 4096)
	if err != nil {
		return nil, errors.New("unable to read signing key file")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(b)))
	if err != nil {
		return nil, errors.New("signing key must use base64 encoding")
	}
	switch len(decoded) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(decoded), nil
	case ed25519.PrivateKeySize:
		key := ed25519.NewKeyFromSeed(decoded[:ed25519.SeedSize])
		if !bytes.Equal(key, decoded) {
			return nil, errors.New("Ed25519 private key is inconsistent")
		}
		return key, nil
	default:
		return nil, errors.New("signing key must be a 32-byte seed or 64-byte private key")
	}
}

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

func publish(o packOptions) error {
	if o.Spec == "" || o.Key == "" || o.Output == "" || o.Sequence == 0 || !o.Expires.After(time.Now()) {
		return errors.New("spec, key, new output directory, positive sequence and future expiry are required")
	}
	key, err := loadKey(o.Key)
	if err != nil {
		return err
	}
	b, err := readRegular(o.Spec, 2<<20)
	if err != nil {
		return err
	}
	var spec packSpec
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(&spec); err != nil {
		return fmt.Errorf("invalid spec: %w", err)
	}
	if err = d.Decode(new(any)); err != io.EOF {
		return errors.New("trailing spec JSON")
	}
	if strings.TrimSpace(spec.Version) == "" || len(spec.Version) > 128 || len(spec.Files) == 0 || len(spec.Files) > 128 {
		return errors.New("spec requires a version and between 1 and 128 files")
	}
	manifest := feed.Manifest{Version: spec.Version, Sequence: o.Sequence, ExpiresAt: o.Expires.UTC()}
	compiledFiles := make(map[string][]byte, len(spec.Files))
	artifacts := make(map[string][]byte, len(spec.Files))
	names := make(map[string]bool, len(spec.Files))
	loaded := make([]packArtifact, 0, len(spec.Files))
	officialFiles := make(map[string][]byte)
	needsDedupe := false
	for _, input := range spec.Files {
		nameKey := strings.ToLower(input.Name)
		if !validName.MatchString(input.Name) || names[nameKey] || input.Path == "" {
			return errors.New("artifact output names must be unique portable filenames and paths are required")
		}
		if input.Official && input.DedupeAgainstOfficial || input.Official && input.Format != "" && input.Format != "signature" || input.DedupeAgainstOfficial && !strings.Contains(input.Format, "-list") {
			return fmt.Errorf("invalid provenance or deduplication setting for %s", input.Name)
		}
		names[nameKey] = true
		inputPath := input.Path
		if !filepath.IsAbs(inputPath) {
			inputPath = filepath.Join(filepath.Dir(o.Spec), inputPath)
		}
		data, err := readRegular(inputPath, artifactLimit)
		if err != nil {
			return fmt.Errorf("reading %s: %w", input.Name, err)
		}
		isCVD := strings.EqualFold(filepath.Ext(input.Name), ".cvd") && (input.Format == "" || input.Format == "signature")
		if input.DetachedSignature != "" {
			if !input.Official || !isCVD {
				return fmt.Errorf("detached signature for %s requires an official CVD source", input.Name)
			}
			signPath := input.DetachedSignature
			if !filepath.IsAbs(signPath) {
				signPath = filepath.Join(filepath.Dir(o.Spec), signPath)
			}
			sign, err := readRegular(signPath, 1<<20)
			if err != nil {
				return fmt.Errorf("reading detached signature for %s: %w", input.Name, err)
			}
			if err := signatures.VerifyOfficialDetachedCVD(data, sign); err != nil {
				return fmt.Errorf("verifying detached signature for %s: %w", input.Name, err)
			}
		} else if o.RequireDetachedSignatures && input.Official && isCVD {
			return fmt.Errorf("official CVD %s has no detached signature", input.Name)
		}
		item := feed.ManifestFile{Name: input.Name, Format: input.Format}
		normalized, err := feed.NormalizeArtifact(item, data)
		if err != nil {
			return fmt.Errorf("normalizing %s: %w", input.Name, err)
		}
		loaded = append(loaded, packArtifact{input: input, data: data, normalized: normalized})
		if input.Official || isSelectionSource(input.Name) {
			officialFiles[input.Name] = normalized
		}
		needsDedupe = needsDedupe || input.DedupeAgainstOfficial
	}
	var officialDB *signatures.Database
	if needsDedupe {
		officialCount := 0
		for _, entry := range loaded {
			if entry.input.Official {
				officialCount++
			}
			if ext := strings.ToLower(filepath.Ext(entry.input.Name)); !entry.input.Official && (ext == ".cvd" || ext == ".cld") {
				return fmt.Errorf("deduplication cannot account for selection files in unmarked container %s", entry.input.Name)
			}
		}
		if officialCount == 0 {
			return errors.New("deduplication requires explicitly marked official signature sources")
		}
		var err error
		officialDB, err = signatures.Compile(officialFiles)
		if err != nil {
			return fmt.Errorf("official signature compatibility validation: %w", err)
		}
	}
	seen := make(map[string]bool)
	var stats dedupeStats
	total := 0
	for _, entry := range loaded {
		input, data, normalized := entry.input, entry.data, entry.normalized
		format := input.Format
		if input.DedupeAgainstOfficial {
			var err error
			normalized, err = deduplicateHashArtifact(normalized, officialDB, seen, &stats)
			if err != nil {
				return fmt.Errorf("deduplicating %s: %w", input.Name, err)
			}
			if len(normalized) == 0 {
				stats.Omitted++
				continue
			}
			data, format = normalized, "signature"
		}
		total += max(len(data), len(normalized))
		if total > generationLimit {
			return errors.New("generation exceeds 512 MiB limit; split upstream input")
		}
		hash := sha256.Sum256(data)
		digest := hex.EncodeToString(hash[:])
		item := feed.ManifestFile{Name: input.Name, URL: "artifacts/" + digest, SHA256: digest, Format: format}
		compiledFiles[input.Name] = normalized
		artifacts[digest] = data
		manifest.Files = append(manifest.Files, item)
	}
	officialDB = nil
	db, err := signatures.Compile(compiledFiles)
	if err != nil {
		return fmt.Errorf("signature compatibility validation: %w", err)
	}
	if db.DetectionCount() == 0 {
		return errors.New("generation contains no detection signatures")
	}
	// Signing makes the endpoint trust the payload; upstream provenance must be
	// established first, not replaced by this signature.
	for _, c := range db.Report().Containers {
		if !c.Authenticated && !o.AllowUnauthenticatedContainers {
			return fmt.Errorf("container %s has no valid upstream publisher signature", c.File)
		}
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	envelope := feed.ManifestEnvelope{Payload: base64.StdEncoding.EncodeToString(payload), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))}
	signed, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	// Only a fresh, exclusively created directory is modified. Existing published
	// content cannot be overwritten, and the manifest is written last.
	if err = os.Mkdir(o.Output, 0755); err != nil {
		return fmt.Errorf("output must be a new directory: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(o.Output)
		}
	}()
	if err = os.Mkdir(filepath.Join(o.Output, "artifacts"), 0755); err != nil {
		return err
	}
	for digest, data := range artifacts {
		if err = writeDurable(filepath.Join(o.Output, "artifacts", digest), data); err != nil {
			return err
		}
	}
	if err = writeDurable(filepath.Join(o.Output, ".manifest-staging"), signed); err != nil {
		return err
	}
	if err = os.Rename(filepath.Join(o.Output, ".manifest-staging"), filepath.Join(o.Output, "manifest.json")); err != nil {
		return err
	}
	complete = true
	if needsDedupe {
		fmt.Printf("Hash deduplication: %d official overlaps removed, %d repeated indicators removed, %d retained, %d empty artifacts omitted\n", stats.Official, stats.Repeated, stats.Retained, stats.Omitted)
	}
	return nil
}

// Ignore lists and dynamic configuration anywhere in a generation select the
// official rules too. Include them in the coverage compilation, or a rule
// suppressed in the final generation could incorrectly remove an indicator.
func isSelectionSource(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".ign", ".ign2", ".cfg":
		return true
	}
	return false
}

// deduplicateHashArtifact receives only fully validated HDB/HSB rows from the
// explicit hash-list adapters. The input is compacted in place to avoid a
// second large buffer. A size-specific official rule cannot replace an
// unknown-size ThreatWinds indicator, so only wildcard official rules count.
func deduplicateHashArtifact(normalized []byte, official *signatures.Database, seen map[string]bool, stats *dedupeStats) ([]byte, error) {
	out := normalized[:0]
	for len(normalized) > 0 {
		end := bytes.IndexByte(normalized, '\n')
		if end < 0 {
			return nil, errors.New("unterminated normalized hash row")
		}
		row := normalized[:end+1]
		colon := bytes.IndexByte(row, ':')
		if colon < 0 {
			return nil, errors.New("invalid normalized hash row")
		}
		digest, err := hex.DecodeString(string(row[:colon]))
		if err != nil {
			return nil, errors.New("invalid normalized digest")
		}
		key := string(row[:colon])
		switch {
		case official.HasWholeHashWildcard(digest):
			stats.Official++
		case seen[key]:
			stats.Repeated++
		default:
			out = append(out, row...)
			seen[key] = true
			stats.Retained++
		}
		normalized = normalized[end+1:]
	}
	return out, nil
}

func writeDurable(name string, b []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
