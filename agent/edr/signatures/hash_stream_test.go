package signatures

import (
	"crypto/md5"
	"fmt"
	"strings"
	"testing"
)

func TestHashSourceScannerBufferIsNotRetained(t *testing.T) {
	var lines strings.Builder
	var first, last [md5.Size]byte
	var ignored string
	for i := 0; i < 1200; i++ {
		digest := md5.Sum([]byte(fmt.Sprintf("payload-%04d", i)))
		if i == 0 {
			first = digest
		}
		if i == 1199 {
			last = digest
		}
		line := fmt.Sprintf("%x:12:Rule.%04d", digest, i)
		if i == 600 {
			ignored = line
		}
		lines.WriteString(line)
		lines.WriteByte('\n')
	}
	ignoreDigest := md5.Sum([]byte(ignored))
	db, err := Compile(map[string][]byte{
		"sample.hdb":  []byte(lines.String()),
		"sample.ign2": []byte(fmt.Sprintf("Rule.0600:%x\n", ignoreDigest)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if db.DetectionCount() != 1199 {
		t.Fatalf("detection count %d, want 1199 after exact ignore", db.DetectionCount())
	}
	if got := db.hashes.lookup(string(first[:]), 12); got != "Rule.0000" {
		t.Fatalf("first hash name overwritten by scanner reuse: %q", got)
	}
	if got := db.hashes.lookup(string(last[:]), 12); got != "Rule.1199" {
		t.Fatalf("last hash name = %q", got)
	}
	ignoredHash := md5.Sum([]byte("payload-0600"))
	if got := db.hashes.lookup(string(ignoredHash[:]), 12); got != "" {
		t.Fatalf("ignored hash survived: %q", got)
	}
}
