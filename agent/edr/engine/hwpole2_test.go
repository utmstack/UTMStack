package engine

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func hwpOLE2Fixture(ole []byte, declared uint32) []byte {
	wrapped := make([]byte, 4, len(ole)+4)
	binary.LittleEndian.PutUint32(wrapped, declared)
	return append(wrapped, ole...)
}

func TestHWPOLE2NestedCompressedRepresentation(t *testing.T) {
	const marker = "HWP-ORACLE-CHILD-ONLY-MARKER"
	packed := gzipFixture(t, []byte("innocuous prefix "+marker+" suffix"))
	if bytes.Contains(packed, []byte(marker)) {
		t.Fatal("marker visible without nested decoding")
	}
	rules := map[string]string{
		"type.ftm": "0:4:1f8b08:HWP wrapper:CL_TYPE_ANY:CL_TYPE_HWPOLE2\n",
		"hash.hdb": fmt.Sprintf("%x:%d:HWP.Child\n", md5.Sum(packed), len(packed)),
	}
	for _, declared := range []uint32{uint32(len(packed)), 1} {
		wrapped := hwpOLE2Fixture(packed, declared)
		if typ, _ := pipelineState(t, rules).db.DetermineType(wrapped); typ != signatures.TypeHWPOLE2 {
			t.Fatalf("declared size %d: type %v", declared, typ)
		}
		expectScan(t, pipelineState(t, rules), wrapped, "HWP.Child", "")
	}
	clean := hwpOLE2Fixture(gzipFixture(t, []byte("innocuous text")), 14)
	expectScan(t, pipelineState(t, rules), clean, "", "")
}

func TestHWPOLE2ScansNestedDocumentDespiteSizeMismatch(t *testing.T) {
	inner := oleFixture(t, 3, "Contents", []byte("harmless OLE stream"), true, false, false)
	sum := md5.Sum(inner)
	rules := map[string]string{
		"type.ftm": "0:4:d0cf11e0a1b11ae1:HWP OLE:CL_TYPE_ANY:CL_TYPE_HWPOLE2\n",
		"hash.hdb": fmt.Sprintf("%x:%d:HWP.Inner\n", sum, len(inner)),
	}
	state := pipelineState(t, rules)
	for _, declared := range []uint32{uint32(len(inner)), 1} {
		wrapped := hwpOLE2Fixture(inner, declared)
		if out := os.Getenv("EDR_HWP_ORACLE_OUT"); out != "" {
			name := "ole-correct-size.hwpole2"
			if declared == 1 {
				name = "ole-mismatched-size.hwpole2"
			}
			if err := os.WriteFile(filepath.Join(out, name), wrapped, 0600); err != nil {
				t.Fatal(err)
			}
		}
		if typ, _ := state.db.DetermineType(wrapped); typ != signatures.TypeHWPOLE2 {
			t.Fatalf("declared size %d: type %v", declared, typ)
		}
		expectScan(t, pipelineState(t, rules), wrapped, "HWP.Inner", "")
	}

	other := oleFixture(t, 3, "Contents", []byte("a different harmless OLE stream"), true, false, false)
	clean := hwpOLE2Fixture(other, uint32(len(other)))
	if out := os.Getenv("EDR_HWP_ORACLE_OUT"); out != "" {
		if err := os.WriteFile(filepath.Join(out, "ole-benign.hwpole2"), clean, 0600); err != nil {
			t.Fatal(err)
		}
	}
	expectScan(t, pipelineState(t, rules), clean, "", "")
}

func TestHWPOLE2PinnedClamAVOracleFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine",
		"audits", "2026-10-28-v150-hwpole2")
	ftm, err := os.ReadFile(filepath.Join(root, "rules", "type.ftm"))
	if err != nil {
		t.Fatal(err)
	}
	hdb, err := os.ReadFile(filepath.Join(root, "rules", "hash.hdb"))
	if err != nil {
		t.Fatal(err)
	}
	oleHDB, err := os.ReadFile(filepath.Join(root, "rules", "ole.hash.hdb"))
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]string{
		"type.ftm":     string(ftm),
		"hash.hdb":     string(hdb),
		"ole.hash.hdb": string(oleHDB),
	}
	for name, want := range map[string]string{
		"benign.hwpole2":              "",
		"correct-size.hwpole2":        "HWP.Child",
		"mismatched-size.hwpole2":     "HWP.Child",
		"ole-benign.hwpole2":          "",
		"ole-correct-size.hwpole2":    "HWP.OLE.Child",
		"ole-mismatched-size.hwpole2": "HWP.OLE.Child",
	} {
		data, err := os.ReadFile(filepath.Join(root, "fixtures", name))
		if err != nil {
			t.Fatal(err)
		}
		expectScan(t, pipelineState(t, rules), data, want, "")
	}
}
