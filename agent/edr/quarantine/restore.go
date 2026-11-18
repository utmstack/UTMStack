package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

var (
	ErrRestoreIntegrity             = errors.New("quarantine integrity check failed")
	ErrQuarantineBusy               = errors.New("quarantine item is busy; retry the operation")
	ErrFreshRestoreApprovalRequired = errors.New("fresh restore requires its exact explicit execution approval")
)

// Restorable input has the same ceiling as files admitted by the scanner.
// A growing or forged stored file cannot fill a disk during recovery.
const maxRestoreBytes = int64(config.MaxFileSizeMB) << 20

type restoreDirectory struct {
	file *os.File
	path string
}

func (d *restoreDirectory) Close() error { return d.file.Close() }

func ordinaryQuarantineID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func restoreLeaf(destination string) (string, error) {
	name := filepath.Base(destination)
	if !filepath.IsAbs(destination) || name == "." || name == ".." || name == string(filepath.Separator) || strings.ContainsAny(name, "/\x00") || (filepath.Separator == '\\' && strings.ContainsAny(name, "\\:")) {
		return "", fmt.Errorf("%w: invalid original destination", ErrRestoreIntegrity)
	}
	return name, nil
}

func restoreDigest(text string) ([sha256.Size]byte, error) {
	var expected [sha256.Size]byte
	decoded, err := hex.DecodeString(text)
	if err != nil || len(decoded) != len(expected) {
		return expected, fmt.Errorf("%w: invalid recorded SHA-256", ErrRestoreIntegrity)
	}
	copy(expected[:], decoded)
	return expected, nil
}

func openRestoreDirectory(path string, create bool) (*restoreDirectory, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if create {
		if err := os.MkdirAll(absolute, 0700); err != nil {
			return nil, err
		}
	}
	// Preserve initially configured directory aliases, then pin the resolved
	// directory without following any newly introduced symlink components.
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	return platformRestoreDirectory(resolved)
}

func restoreInputInfo(f *os.File) (os.FileInfo, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxRestoreBytes {
		return nil, fmt.Errorf("%w: stored entry is not a regular file within the scanner's file limit", ErrRestoreIntegrity)
	}
	return info, nil
}

func (s *Store) copyRestoreSnapshot(out *os.File, input *os.File, initial os.FileInfo, expected [sha256.Size]byte) error {
	copyFn := s.restoreCopy
	if copyFn == nil {
		copyFn = io.Copy
	}
	hash := sha256.New()
	n, err := copyFn(io.MultiWriter(out, hash), io.NewSectionReader(input, 0, initial.Size()))
	if err != nil {
		return err
	}
	if n != initial.Size() || !equalRestoreHash(hash.Sum(nil), expected) {
		return fmt.Errorf("%w: stored contents do not match the recorded SHA-256", ErrRestoreIntegrity)
	}
	after, err := input.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(initial, after) || after.Size() != initial.Size() || !after.ModTime().Equal(initial.ModTime()) {
		return fmt.Errorf("%w: stored file changed during recovery", ErrRestoreIntegrity)
	}
	return out.Sync()
}

func equalRestoreHash(got []byte, expected [sha256.Size]byte) bool {
	if len(got) != len(expected) {
		return false
	}
	var different byte
	for i := range got {
		different |= got[i] ^ expected[i]
	}
	return different == 0
}

func verifiedRestoreHash(input *os.File, info os.FileInfo, expected [sha256.Size]byte) error {
	hash := sha256.New()
	n, err := io.Copy(hash, io.NewSectionReader(input, 0, info.Size()))
	if err != nil {
		return err
	}
	if n != info.Size() || !equalRestoreHash(hash.Sum(nil), expected) {
		return fmt.Errorf("%w: stored contents do not match the recorded SHA-256", ErrRestoreIntegrity)
	}
	return nil
}
