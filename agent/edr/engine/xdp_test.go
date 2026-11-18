package engine

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func xdpFixture(chunks ...[]byte) []byte {
	var out bytes.Buffer
	out.WriteString("<xdp>")
	for _, chunk := range chunks {
		out.WriteString("<chunk>\n")
		out.WriteString(base64.StdEncoding.EncodeToString(chunk))
		out.WriteString("\n</chunk>")
	}
	out.WriteString("</xdp>")
	return out.Bytes()
}

func TestXDPScansPDFChunkAndStopsAtNonPDF(t *testing.T) {
	pdf := []byte("%PDF-1.4\n1 0 obj<</Type/Catalog>>endobj\n%%EOF\n")
	hash := md5.Sum(pdf)
	rule := fmt.Sprintf("%x:%d:XDP.PDF\n", hash, len(pdf))
	ftm := "0:0:3c786470:XDP:CL_TYPE_ANY:CL_TYPE_XDP\n"
	newState := func() *scanState {
		return pipelineState(t, map[string]string{"xdp.ftm": ftm, "pdf.hdb": rule})
	}
	positive := xdpFixture(pdf)
	expectDetected(t, newState(), positive, "XDP.PDF")
	malformedThenPDF := []byte("<xdp><chunk>not-base64!</chunk><chunk>" +
		base64.StdEncoding.EncodeToString(pdf) + "</chunk></xdp>")
	expectScan(t, newState(), malformedThenPDF, "", "")
	limited := newState()
	limited.remaining = int64(len(positive) + len(pdf) - 1)
	expectIncomplete(t, limited, positive, ErrLimit)
	// The reference stops on the first successfully decoded non-PDF chunk.
	expectScan(t, newState(), xdpFixture([]byte("ordinary bytes"), pdf), "", "")
	// Its 1,028-byte magic window checks offsets 0 through 1,027.
	var offsetRules string
	var offsetFixtures [2][]byte
	for _, offset := range []int{1027, 1028} {
		decoded := append(bytes.Repeat([]byte{'x'}, offset), pdf...)
		sum := md5.Sum(decoded)
		offsetRules += fmt.Sprintf("%x:%d:XDP.Offset\n", sum, len(decoded))
		if offset == 1027 {
			offsetFixtures[0] = xdpFixture(decoded)
		} else {
			offsetFixtures[1] = xdpFixture(decoded)
		}
		state := pipelineState(t, map[string]string{"xdp.ftm": ftm,
			"offset.hdb": fmt.Sprintf("%x:%d:XDP.Offset\n", sum, len(decoded))})
		if offset == 1027 {
			expectDetected(t, state, xdpFixture(decoded), "XDP.Offset")
		} else {
			expectScan(t, state, xdpFixture(decoded), "", "")
		}
	}

	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, content := range map[string][]byte{
			"xdp.ftm": []byte(ftm), "pdf.hdb": []byte(rule),
			"offset.hdb":        []byte(offsetRules),
			"positive.xdp":      positive,
			"after-invalid.xdp": malformedThenPDF,
			"early-nonpdf.xdp":  xdpFixture([]byte("ordinary bytes"), pdf),
			"offset-1027.xdp":   offsetFixtures[0],
			"offset-1028.xdp":   offsetFixtures[1],
		} {
			if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"),
			"--database="+dir, "--no-summary", filepath.Join(dir, "positive.xdp"),
			filepath.Join(dir, "after-invalid.xdp"), filepath.Join(dir, "early-nonpdf.xdp"),
			filepath.Join(dir, "offset-1027.xdp"), filepath.Join(dir, "offset-1028.xdp"))
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.ExitError); !ok ||
			!strings.Contains(string(out), "positive.xdp: XDP.PDF.UNOFFICIAL FOUND") ||
			!strings.Contains(string(out), "after-invalid.xdp: OK") ||
			!strings.Contains(string(out), "early-nonpdf.xdp: OK") ||
			!strings.Contains(string(out), "offset-1027.xdp: XDP.Offset.UNOFFICIAL FOUND") ||
			!strings.Contains(string(out), "offset-1028.xdp: OK") {
			t.Fatalf("XDP reference mismatch: err=%v output=%s", err, out)
		}
		t.Log(strings.TrimSpace(string(out)))
	}
}
