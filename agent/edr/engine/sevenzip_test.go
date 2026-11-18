package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/sevenzip"
)

// The 7-Zip archives are the sevenzip package's test data; their members
// hold the payload marker.
func sevenZipFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "sevenzip", "testdata", name+".7z"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSevenZipMembersAreScanned(t *testing.T) {
	for _, name := range []string{"lzma", "lzma2", "header_plain", "ppmd", "ppmd_restart", "bcj_lzma", "bcj2", "arm",
		"solid", "empty_first", "unsupported_first"} {
		t.Run(name, func(t *testing.T) {
			b := sevenZipFixture(t, name)
			requireHidden(t, b)
			expectDetected(t, payloadState(t), b, "Payload.Test")
		})
	}
}

// As in the reference: members of coders it does not decode either (BZip2,
// Delta, the branch filters other than x86 and ARM) are skipped, encrypted
// members and headers cannot be read, damaged data or a member CRC mismatch
// (which it skips) is reported here, and an archive whose header does not
// check is one it reads nothing of.
func TestSevenZipReferenceOutcomes(t *testing.T) {
	for _, name := range []string{"bzip2", "delta", "armt", "arm64", "ppc", "sparc", "ia64", "riscv"} {
		expectDetected(t, payloadState(t), sevenZipFixture(t, name), "")
	}
	for _, name := range []string{"enc_data", "enc_header"} {
		expectIncomplete(t, payloadState(t), sevenZipFixture(t, name), ErrUnsupported)
	}
	damaged := sevenZipFixture(t, "lzma")
	damaged[40] ^= 0x55
	expectIncomplete(t, payloadState(t), damaged, ErrMalformed)
	stored := sevenZipFixture(t, "copy")
	stored[bytes.Index(stored, []byte(payloadMarker))+3] ^= 1
	expectIncomplete(t, payloadState(t), stored, ErrMalformed)
	header := sevenZipFixture(t, "lzma")
	header[28] ^= 1
	expectDetected(t, payloadState(t), header, "")
}

func TestSevenZipMemberMetadata(t *testing.T) {
	b := sevenZipFixture(t, "enc_data")
	a, err := sevenzip.Open(b, sevenzip.Limits{MaxFiles: 10, MaxUnpack: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	// The reference reports a 7-Zip member's packed size as 0 and counts
	// positions from 0.
	rule := fmt.Sprintf("Sz.Member:CL_TYPE_7Z:*:^text\\.txt$:0:%d:1:0:*:\n", a.Files[0].Size)
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule}), b, "Sz.Member", "")
	rule = fmt.Sprintf("Sz.Member:CL_TYPE_7Z:*:^text\\.txt$:0:%d:0:0:*:\n", a.Files[0].Size)
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule}), sevenZipFixture(t, "lzma"), "Sz.Member", "")
}

func TestSevenZipLimits(t *testing.T) {
	s := payloadState(t)
	s.maxFile = 100
	expectIncomplete(t, s, sevenZipFixture(t, "lzma"), ErrLimit)
	s = payloadState(t)
	s.maxFiles = 1
	if sig, err := s.visit(sevenZipFixture(t, "solid"), 0); sig != "" || err == nil {
		t.Fatalf("two members past a one-file limit: %q, %v", sig, err)
	}
}
