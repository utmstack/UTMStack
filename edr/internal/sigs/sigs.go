// edr/internal/sigs/sigs.go
// Package sigs maintains the mirrored ClamAV signature databases via the
// official cvdupdate tool. cvdupdate is rate-limit-compliant against the CDN
// and maintains the exact CVD + cdiff layout freshclam PrivateMirror expects;
// content integrity is ClamAV's own signing, verified by the agents.
package sigs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

type Syncer struct {
	DBDir string
	Run   func(name string, args ...string) (string, error)
	Rec   *mirror.Recorder
}

func New(dbDir string, rec *mirror.Recorder) *Syncer {
	return &Syncer{DBDir: dbDir, Rec: rec, Run: runCommand}
}

func runCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// EnsureConfig points cvdupdate's database directory at the mirror tree.
func (s *Syncer) EnsureConfig() error {
	if err := os.MkdirAll(s.DBDir, 0o755); err != nil {
		return err
	}
	out, err := s.Run("cvd", "config", "set", "--dbdir", s.DBDir)
	if err != nil {
		return fmt.Errorf("cvd config: %v: %s", err, out)
	}
	return nil
}

// SyncOnce runs one cvdupdate cycle (full CVDs on first run, cdiffs after)
// and records the resulting database versions in status.json.
func (s *Syncer) SyncOnce() error {
	out, err := s.Run("cvd", "update")
	if err != nil {
		s.Rec.SigFailure(fmt.Sprintf("%v: %s", err, truncate(out, 500)))
		return fmt.Errorf("cvd update: %v", err)
	}
	s.Rec.SigSuccess(DatabaseVersions(s.DBDir))
	return nil
}

// DatabaseVersions maps each *.cvd in dir to its version (from the CVD header).
func DatabaseVersions(dir string) map[string]string {
	dbs := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dbs
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cvd") {
			continue
		}
		if v, err := ReadCVDVersion(filepath.Join(dir, e.Name())); err == nil {
			dbs[e.Name()] = v
		}
	}
	return dbs
}

// ReadCVDVersion extracts the version from a CVD's 512-byte text header:
// "ClamAV-VDB:<build time>:<version>:...".
func ReadCVDVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	head := make([]byte, 512)
	n, err := f.Read(head)
	if err != nil {
		return "", err
	}
	fields := strings.Split(string(head[:n]), ":")
	if len(fields) < 3 || fields[0] != "ClamAV-VDB" {
		return "", fmt.Errorf("%s: not a CVD header", path)
	}
	return fields[2], nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
