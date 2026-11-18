// feedpromote installs a signed native feed release into a local HTTPS mirror.
// It runs on the publisher host; it never holds a signing key.
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
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/feed"
)

const (
	manifestLimit   = 2 << 20
	artifactLimit   = 256 << 20
	generationLimit = 512 << 20
)

type verifiedRelease struct {
	manifest feed.Manifest
	payload  []byte
	data     []byte
	files    map[string]string // digest to local source path
}

func main() {
	var release, mirror, keyPath string
	flag.StringVar(&release, "release", "", "new signed feedpack output directory")
	flag.StringVar(&mirror, "mirror", "", "existing local HTTPS mirror directory")
	flag.StringVar(&keyPath, "public-key", "", "file containing the pinned base64 Ed25519 public key")
	flag.Parse()
	if release == "" || mirror == "" || keyPath == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "feedpromote requires -release, -mirror and -public-key")
		os.Exit(2)
	}
	keyData, err := readRegular(keyPath, 4096)
	if err != nil {
		fmt.Fprintln(os.Stderr, "feedpromote: cannot read public key:", err)
		os.Exit(1)
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(keyData)))
	if err != nil || len(key) != ed25519.PublicKeySize {
		fmt.Fprintln(os.Stderr, "feedpromote: invalid Ed25519 public key")
		os.Exit(1)
	}
	if err := promote(release, mirror, ed25519.PublicKey(key), time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "feedpromote:", err)
		os.Exit(1)
	}
	fmt.Println("Signed native feed promoted to", mirror)
}

func regularDir(name string) error {
	st, err := os.Lstat(name)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return errors.New("path must be a directory, not a link or file")
	}
	return nil
}

func readRegular(name string, limit int64) ([]byte, error) {
	st, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > limit {
		return nil, errors.New("path must be a bounded regular file")
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(b)) > limit {
		return nil, errors.New("file changed or exceeds size limit")
	}
	return b, nil
}

func decodeStrict(b []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}

func signedManifest(data []byte, key ed25519.PublicKey) (feed.Manifest, []byte, error) {
	var env feed.ManifestEnvelope
	var m feed.Manifest
	if err := decodeStrict(data, &env); err != nil {
		return m, nil, errors.New("invalid manifest envelope")
	}
	payload, err := base64.StdEncoding.Strict().DecodeString(env.Payload)
	if err != nil || len(payload) > manifestLimit || len(payload) == 0 {
		return m, nil, errors.New("invalid manifest payload")
	}
	sig, err := base64.StdEncoding.Strict().DecodeString(env.Signature)
	if err != nil || !ed25519.Verify(key, payload, sig) {
		return m, nil, errors.New("manifest signature verification failed")
	}
	if err := decodeStrict(payload, &m); err != nil {
		return m, nil, errors.New("invalid signed manifest payload")
	}
	if m.Sequence == 0 || strings.TrimSpace(m.Version) == "" || len(m.Version) > 128 || len(m.Files) == 0 || len(m.Files) > 128 {
		return m, nil, errors.New("invalid signed manifest metadata")
	}
	names := make(map[string]bool, len(m.Files))
	for _, item := range m.Files {
		name := strings.ToLower(item.Name)
		if !validArtifact(item) || names[name] {
			return m, nil, errors.New("invalid signed artifact description")
		}
		names[name] = true
	}
	return m, payload, nil
}

func validArtifact(item feed.ManifestFile) bool {
	if len(item.Name) == 0 || len(item.Name) > 128 || item.Name == "." || item.Name == ".." {
		return false
	}
	for _, r := range item.Name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	if len(item.SHA256) != 64 || item.URL != "artifacts/"+item.SHA256 {
		return false
	}
	for _, c := range item.SHA256 {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	format := strings.TrimSuffix(strings.TrimSuffix(item.Format, ".tar.gz"), ".gz")
	switch format {
	case "", "signature":
		return item.Format == "" || item.Format == "signature"
	case "md5-list":
		return strings.HasSuffix(strings.ToLower(item.Name), ".hdb")
	case "sha1-list", "sha256-list":
		return strings.HasSuffix(strings.ToLower(item.Name), ".hsb")
	default:
		return false
	}
}

func digestFile(name string) (int64, string, error) {
	st, err := os.Lstat(name)
	if err != nil || !st.Mode().IsRegular() || st.Size() > artifactLimit {
		return 0, "", errors.New("artifact is not a bounded regular file")
	}
	f, err := os.Open(name)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, artifactLimit+1))
	if err != nil || n > artifactLimit || n != st.Size() {
		return 0, "", errors.New("artifact changed or exceeds size limit")
	}
	return n, hex.EncodeToString(h.Sum(nil)), nil
}

func verifyRelease(dir string, key ed25519.PublicKey, now time.Time) (verifiedRelease, error) {
	var out verifiedRelease
	if err := regularDir(dir); err != nil {
		return out, err
	}
	if err := regularDir(filepath.Join(dir, "artifacts")); err != nil {
		return out, err
	}
	data, err := readRegular(filepath.Join(dir, "manifest.json"), manifestLimit)
	if err != nil {
		return out, err
	}
	m, payload, err := signedManifest(data, key)
	if err != nil {
		return out, err
	}
	if !now.Before(m.ExpiresAt) {
		return out, errors.New("release manifest has expired")
	}
	out = verifiedRelease{manifest: m, payload: payload, data: data, files: make(map[string]string)}
	var total int64
	for _, item := range m.Files {
		name := filepath.Join(dir, "artifacts", item.SHA256)
		n, hash, err := digestFile(name)
		if err != nil || hash != item.SHA256 {
			return verifiedRelease{}, fmt.Errorf("release artifact %s failed SHA-256 verification", item.SHA256)
		}
		convertedSize := n
		if item.Format != "" && item.Format != "signature" {
			data, err := readRegular(name, artifactLimit)
			if err != nil {
				return verifiedRelease{}, fmt.Errorf("reading hash-list artifact %s: %w", item.Name, err)
			}
			converted, err := feed.NormalizeArtifact(item, data)
			if err != nil {
				return verifiedRelease{}, fmt.Errorf("invalid hash-list artifact %s: %w", item.Name, err)
			}
			convertedSize = int64(len(converted))
		}
		total += max(n, convertedSize)
		if total > generationLimit {
			return verifiedRelease{}, errors.New("release exceeds generation size limit")
		}
		out.files[item.SHA256] = name
	}
	return out, nil
}

func syncDir(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func copyArtifact(source, dir, digest string) error {
	target := filepath.Join(dir, digest)
	if _, err := os.Lstat(target); err == nil {
		_, hash, err := digestFile(target)
		if err != nil || hash != digest {
			return errors.New("existing mirror artifact does not match signed digest")
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.CreateTemp(dir, ".artifact-*")
	if err != nil {
		return err
	}
	defer os.Remove(out.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), io.LimitReader(in, artifactLimit+1))
	if err != nil || n > artifactLimit || hex.EncodeToString(h.Sum(nil)) != digest {
		_ = out.Close()
		return errors.New("release artifact changed during copy")
	}
	if err := out.Chmod(0644); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Link(out.Name(), target); err != nil {
		if !os.IsExist(err) {
			return err
		}
		_, hash, err := digestFile(target)
		if err != nil || hash != digest {
			return errors.New("concurrent mirror artifact does not match signed digest")
		}
	}
	return nil
}

func promote(release, mirror string, key ed25519.PublicKey, now time.Time) error {
	if len(key) != ed25519.PublicKeySize {
		return errors.New("pinned public key is required")
	}
	r, err := verifyRelease(release, key, now)
	if err != nil {
		return err
	}
	if err := regularDir(mirror); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(mirror, ".feedpromote.lock"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("mirror promotion lock unavailable: %w", err)
	}
	_ = lock.Close()
	defer os.Remove(filepath.Join(mirror, ".feedpromote.lock"))
	artifactsDir := filepath.Join(mirror, "artifacts")
	if err := os.Mkdir(artifactsDir, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	if err := regularDir(artifactsDir); err != nil {
		return err
	}
	currentPath := filepath.Join(mirror, "manifest.json")
	current, err := readRegular(currentPath, manifestLimit)
	if err == nil {
		m, payload, err := signedManifest(current, key)
		if err != nil {
			return fmt.Errorf("current mirror manifest is invalid: %w", err)
		}
		if r.manifest.Sequence < m.Sequence || r.manifest.Sequence == m.Sequence && !bytes.Equal(r.payload, payload) {
			return errors.New("feed rollback or sequence reuse rejected")
		}
		if r.manifest.Sequence == m.Sequence && bytes.Equal(r.payload, payload) {
			for digest, source := range r.files {
				if err := copyArtifact(source, artifactsDir, digest); err != nil {
					return err
				}
			}
			return syncDir(artifactsDir)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	for digest, source := range r.files {
		if err := copyArtifact(source, artifactsDir, digest); err != nil {
			return err
		}
	}
	if err := syncDir(artifactsDir); err != nil {
		return err
	}
	temp, err := os.CreateTemp(mirror, ".manifest-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err := temp.Write(r.data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(0644); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(temp.Name(), currentPath); err != nil {
		return err
	}
	return syncDir(mirror)
}
