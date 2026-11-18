package engine

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Run with EDR_LHASA_ARCHIVE_ROOT pointing at the archives directory of
// Lhasa commit 75ed83559f23e9538e0045c62f53f77ab03d03d6. The local
// package suite does not need to download the upstream historical collection.
func TestLHABroadHistoricalDifferential(t *testing.T) {
	root := os.Getenv("EDR_LHASA_ARCHIVE_ROOT")
	if root == "" {
		t.Skip("set EDR_LHASA_ARCHIVE_ROOT for the 111-archive differential")
	}
	b, err := os.ReadFile("../../../docs/edr/native-engine/audits/2026-10-28-v141-lh1/broad-lhasa-positive.json")
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		LhasaCommit string `json:"lhasa_commit"`
		Cases       []struct {
			Path       string `json:"path"`
			ArchiveSHA string `json:"archive_sha256"`
			Signature  string `json:"signature"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.LhasaCommit != "75ed83559f23e9538e0045c62f53f77ab03d03d6" || len(receipt.Cases) != 111 {
		t.Fatal("unexpected historical reference manifest")
	}
	rules := map[string]string{"test.hdb": "b234ee4d69f5fce4486a80fdaf4a4263:18092:Historical.Short\na6264a9c4ba8e0128f7d17b3f8c5f5c8:1241658:Historical.Long\n"}
	for _, tc := range receipt.Cases {
		t.Run(strings.ReplaceAll(tc.Path, "/", "_"), func(t *testing.T) {
			archive, err := os.ReadFile(filepath.Join(root, tc.Path))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.ArchiveSHA {
				t.Fatalf("historical archive changed: %s", got)
			}
			want := strings.TrimSuffix(tc.Signature, ".UNOFFICIAL FOUND")
			expectDetected(t, pipelineState(t, rules), archive, want)
		})
	}
}
