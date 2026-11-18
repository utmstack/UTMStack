package release

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxArchiveBytes = maxServiceBytes + maxProviderBytes + maxEnvelopeBytes + 1<<20

// Extract authenticates a bounded release archive before returning its private
// staging directory. Only the exact signed artifact names are extracted. The
// caller owns removal of the returned directory after promotion or rejection.
func Extract(archivePath, parent string, key ed25519.PublicKey, want Identity) (string, Manifest, error) {
	var empty Manifest
	info, err := os.Lstat(archivePath)
	if err != nil {
		return "", empty, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxArchiveBytes {
		return "", empty, errors.New("EDR release archive is not a bounded regular file")
	}
	f, err := os.Open(archivePath)
	if err != nil {
		return "", empty, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || opened.Size() != info.Size() {
		return "", empty, errors.New("EDR release archive changed before verification")
	}
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		return "", empty, err
	}
	if len(zr.File) < 2 || len(zr.File) > 3 {
		return "", empty, errors.New("EDR release archive has missing or extra entries")
	}
	entries := make(map[string]*zip.File, len(zr.File))
	for _, member := range zr.File {
		if !safeName(member.Name) || member.Flags&1 != 0 || member.FileInfo().Mode()&os.ModeType != 0 || entries[member.Name] != nil {
			return "", empty, errors.New("EDR release archive contains an unsafe entry")
		}
		entries[member.Name] = member
	}
	meta := entries["release.json"]
	if meta == nil || meta.UncompressedSize64 > maxEnvelopeBytes {
		return "", empty, errors.New("EDR release archive has no bounded signed manifest")
	}
	reader, err := meta.Open()
	if err != nil {
		return "", empty, err
	}
	envelope, readErr := io.ReadAll(io.LimitReader(reader, maxEnvelopeBytes+1))
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil {
		return "", empty, errors.Join(readErr, closeErr)
	}
	manifest, err := Verify(envelope, key, want)
	if err != nil {
		return "", empty, err
	}
	if len(entries) != len(manifest.Artifacts)+1 {
		return "", empty, errors.New("EDR release archive entries do not match signed manifest")
	}
	stage, err := MakePrivateTempDir(parent, ".edr-release-")
	if err != nil {
		return "", empty, err
	}
	complete := false
	defer func() {
		if !complete {
			os.RemoveAll(stage)
		}
	}()
	for _, a := range manifest.Artifacts {
		member := entries[a.Name]
		if member == nil || member.UncompressedSize64 != uint64(a.Bytes) {
			return "", empty, fmt.Errorf("EDR release artifact %q is missing or has wrong size", a.Name)
		}
		input, err := member.Open()
		if err != nil {
			return "", empty, err
		}
		output, err := os.OpenFile(filepath.Join(stage, a.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			input.Close()
			return "", empty, err
		}
		h := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(output, h), io.LimitReader(input, a.Bytes+1))
		closeOut, closeIn := output.Close(), input.Close()
		if err := errors.Join(copyErr, closeOut, closeIn); err != nil {
			return "", empty, err
		}
		if n != a.Bytes || hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
			return "", empty, fmt.Errorf("EDR release artifact %q digest mismatch", a.Name)
		}
	}
	if err := os.WriteFile(filepath.Join(stage, "release.json"), envelope, 0600); err != nil {
		return "", empty, err
	}
	complete = true
	return stage, manifest, nil
}
