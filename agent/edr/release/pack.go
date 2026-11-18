package release

import (
	"archive/zip"
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
)

// Pack produces a signed release archive from already-built and, on Windows,
// already-Authenticode-signed files. The private key is supplied only by the
// publisher. It must never be packaged with the endpoint or repository.
func Pack(archivePath, sourceDir string, privateKey ed25519.PrivateKey, want Identity) (manifest Manifest, err error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return Manifest{}, errors.New("EDR publisher signing key is invalid")
	}
	if want.PackageVersion == "" || want.Implementation == "" || want.OS == "" || want.Architecture == "" || !safeName(want.ServiceName) || want.OS == "windows" && !safeName(want.ProviderName) || want.OS != "windows" && want.ProviderName != "" || want.ProviderName == want.ServiceName {
		return Manifest{}, errors.New("EDR release identity is invalid")
	}
	manifest = Manifest{Schema: 1, PackageVersion: want.PackageVersion, Implementation: want.Implementation, OS: want.OS, Architecture: want.Architecture}
	names := []string{want.ServiceName}
	if want.ProviderName != "" {
		names = append(names, want.ProviderName)
	}
	for _, name := range names {
		path := filepath.Join(sourceDir, name)
		info, e := os.Lstat(path)
		if e != nil {
			return Manifest{}, e
		}
		limit := int64(maxServiceBytes)
		if name == want.ProviderName {
			limit = maxProviderBytes
		}
		if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > limit {
			return Manifest{}, fmt.Errorf("EDR release source %q is not a bounded regular file", name)
		}
		f, e := os.Open(path)
		if e != nil {
			return Manifest{}, e
		}
		opened, e := f.Stat()
		if e != nil || !os.SameFile(info, opened) || opened.Size() != info.Size() {
			f.Close()
			return Manifest{}, fmt.Errorf("EDR release source %q changed before hashing", name)
		}
		h := sha256.New()
		n, readErr := io.Copy(h, io.LimitReader(f, limit+1))
		closeErr := f.Close()
		if e := errors.Join(readErr, closeErr); e != nil {
			return Manifest{}, e
		}
		if n != info.Size() {
			return Manifest{}, fmt.Errorf("EDR release source %q changed during hashing", name)
		}
		manifest.Artifacts = append(manifest.Artifacts, Artifact{Name: name, Bytes: n, SHA256: hex.EncodeToString(h.Sum(nil))})
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, err
	}
	envelope, err := json.Marshal(Envelope{Payload: payload, Signature: ed25519.Sign(privateKey, payload)})
	if err != nil {
		return Manifest{}, err
	}
	if _, err := Verify(envelope, privateKey.Public().(ed25519.PublicKey), want); err != nil {
		return Manifest{}, err
	}
	output, err := os.OpenFile(archivePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return Manifest{}, err
	}
	complete := false
	defer func() {
		if !complete {
			os.Remove(archivePath)
		}
	}()
	zw := zip.NewWriter(output)
	writeEntry := func(name string, content io.Reader) error {
		h := &zip.FileHeader{Name: name, Method: zip.Store}
		h.SetMode(0600)
		entry, e := zw.CreateHeader(h)
		if e != nil {
			return e
		}
		_, e = io.Copy(entry, content)
		return e
	}
	if err := writeEntry("release.json", bytes.NewReader(envelope)); err != nil {
		zw.Close()
		output.Close()
		return Manifest{}, err
	}
	for _, a := range manifest.Artifacts {
		f, e := os.Open(filepath.Join(sourceDir, a.Name))
		if e != nil {
			zw.Close()
			output.Close()
			return Manifest{}, e
		}
		h := sha256.New()
		e = writeEntry(a.Name, io.TeeReader(io.LimitReader(f, a.Bytes+1), h))
		f.Close()
		if e != nil || hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
			zw.Close()
			output.Close()
			return Manifest{}, fmt.Errorf("EDR release source %q changed during packaging: %w", a.Name, e)
		}
	}
	if err := errors.Join(zw.Close(), output.Close()); err != nil {
		return Manifest{}, err
	}
	complete = true
	return manifest, nil
}
