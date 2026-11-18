// utmstack-v12/edr/internal/mirror/writer.go
// Package mirror maintains the static mirror tree the platform serves to
// agents at https://<server>:9001/private/edr/ — artifacts are written
// atomically and never removed on failure (stale, never empty).
package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// WriteArtifact atomically writes data to root/rel and a sibling
// root/rel.sha256 holding the bare lowercase sha256 hex of data (the exact
// format the agent's feed client verifies). The artifact lands before the
// checksum, so a torn read can only produce a mismatch (agent skips and
// retries), never a false verification.
func WriteArtifact(root, rel string, data []byte) error {
	dst := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := atomicWrite(dst, data); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	return atomicWrite(dst+".sha256", []byte(hex.EncodeToString(sum[:])))
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
