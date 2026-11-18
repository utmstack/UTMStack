// Package release verifies the exact files in a native EDR release before
// either the service executable or its AMSI provider can be installed.
package release

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxEnvelopeBytes = 16 << 10
	maxServiceBytes  = 256 << 20
	maxProviderBytes = 32 << 20
)

// Envelope signs the literal JSON payload bytes. Re-serializing JSON before
// verification would allow publishers and endpoints to disagree on the bytes.
type Envelope struct {
	Payload   []byte `json:"payload"`
	Signature []byte `json:"signature"`
}

type Artifact struct {
	Name   string `json:"name"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	Schema         int        `json:"schema"`
	PackageVersion string     `json:"package_version"`
	Implementation string     `json:"implementation"`
	OS             string     `json:"os"`
	Architecture   string     `json:"architecture"`
	Artifacts      []Artifact `json:"artifacts"`
}

// Identity is pinned by the agent build, not obtained from the downloaded
// executable or from mutable server metadata.
type Identity struct {
	PackageVersion, Implementation, OS, Architecture string
	ServiceName, ProviderName                        string
}

func strictJSON(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

// Verify authenticates release metadata and requires exactly one service
// executable and, on Windows, exactly one AMSI provider DLL.
func Verify(data []byte, publicKey ed25519.PublicKey, want Identity) (Manifest, error) {
	var empty Manifest
	if len(data) == 0 || len(data) > maxEnvelopeBytes {
		return empty, errors.New("EDR release envelope exceeds bounds")
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return empty, errors.New("EDR publisher key is not configured")
	}
	var envelope Envelope
	if err := strictJSON(data, &envelope); err != nil {
		return empty, fmt.Errorf("EDR release envelope: %w", err)
	}
	if len(envelope.Payload) == 0 || len(envelope.Payload) > maxEnvelopeBytes || len(envelope.Signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, envelope.Payload, envelope.Signature) {
		return empty, errors.New("EDR release signature is invalid")
	}
	var manifest Manifest
	if err := strictJSON(envelope.Payload, &manifest); err != nil {
		return empty, fmt.Errorf("EDR signed manifest: %w", err)
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, envelope.Payload) {
		return empty, errors.New("EDR signed manifest is not canonical")
	}
	if manifest.Schema != 1 || manifest.PackageVersion == "" || manifest.PackageVersion != want.PackageVersion || manifest.Implementation == "" || manifest.Implementation != want.Implementation || manifest.OS != want.OS || manifest.Architecture != want.Architecture {
		return empty, errors.New("EDR signed release identity mismatch")
	}
	if want.OS != "windows" && want.ProviderName != "" || want.OS == "windows" && want.ProviderName == "" || !safeName(want.ServiceName) || want.ProviderName != "" && !safeName(want.ProviderName) || want.ProviderName == want.ServiceName {
		return empty, errors.New("EDR expected artifact names are invalid")
	}
	expected := map[string]int64{want.ServiceName: maxServiceBytes}
	if want.ProviderName != "" {
		expected[want.ProviderName] = maxProviderBytes
	}
	if len(manifest.Artifacts) != len(expected) {
		return empty, errors.New("EDR release has missing or extra artifacts")
	}
	seen := make(map[string]bool, len(expected))
	for _, a := range manifest.Artifacts {
		limit, ok := expected[a.Name]
		if !ok || seen[a.Name] || a.Bytes < 1 || a.Bytes > limit || len(a.SHA256) != 64 || strings.ToLower(a.SHA256) != a.SHA256 {
			return empty, errors.New("EDR release artifact is invalid or duplicated")
		}
		if _, err := hex.DecodeString(a.SHA256); err != nil {
			return empty, errors.New("EDR release artifact hash is invalid")
		}
		seen[a.Name] = true
	}
	return manifest, nil
}

func safeName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name && !strings.ContainsAny(name, "/\\:\x00") && len(name) <= 128
}

// VerifyFiles reads the exact staged files through open descriptors. The
// lstat/fstat identity check rejects symlinks and replacement between checks.
func VerifyFiles(dir string, manifest Manifest) error {
	for _, a := range manifest.Artifacts {
		if !safeName(a.Name) {
			return errors.New("unsafe EDR artifact name")
		}
		path := filepath.Join(dir, a.Name)
		before, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !before.Mode().IsRegular() || before.Size() != a.Bytes {
			return fmt.Errorf("EDR artifact %q is not the expected regular file", a.Name)
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		opened, err := file.Stat()
		if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) || opened.Size() != a.Bytes {
			file.Close()
			return fmt.Errorf("EDR artifact %q changed before verification", a.Name)
		}
		h := sha256.New()
		n, err := io.Copy(h, io.LimitReader(file, a.Bytes+1))
		file.Close()
		if err != nil || n != a.Bytes || hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
			return fmt.Errorf("EDR artifact %q digest or size mismatch", a.Name)
		}
	}
	return nil
}
