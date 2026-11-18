package signatures

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

var frozenWindowsImageOracle struct {
	sync.Once
	hashes map[string]string
	err    error
}

// These values come from unmodified native C binary-input runs, independently
// of all Go image calculations. The original Mac/Linux goldens stay intact.
func nativeWindowsImageHash(t *testing.T, data []byte) uint64 {
	t.Helper()
	frozenWindowsImageOracle.Do(func() {
		raw, err := os.ReadFile(filepath.Join("testdata", "resize_math", "windows-native-hashes.json"))
		if err != nil {
			frozenWindowsImageOracle.err = err
			return
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != "3caea8c6b31544770996eafa1a42b25a9f0ee9fe425099947170d27e6508fc2e" {
			frozenWindowsImageOracle.err = fmt.Errorf("frozen native Windows fixture changed")
			return
		}
		var manifest struct {
			Inputs int
			Hashes map[string]string
		}
		if err := json.Unmarshal(raw, &manifest); err != nil {
			frozenWindowsImageOracle.err = err
			return
		}
		if manifest.Inputs != 3798 || len(manifest.Hashes) != 3798 {
			frozenWindowsImageOracle.err = fmt.Errorf("native Windows oracle count changed")
			return
		}
		frozenWindowsImageOracle.hashes = manifest.Hashes
	})
	if err := frozenWindowsImageOracle.err; err != nil {
		t.Fatal(err)
	}
	key := fmt.Sprintf("%x", sha256.Sum256(data))
	value, exists := frozenWindowsImageOracle.hashes[key]
	if !exists {
		t.Fatalf("missing independent native Windows hash for %s", key)
	}
	hash, err := strconv.ParseUint(value, 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
