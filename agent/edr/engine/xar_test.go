package engine

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

func xarZlib(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	if _, err := w.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func xarFixture(t *testing.T, toc string, heap []byte) []byte {
	t.Helper()
	compressed := xarZlib(t, []byte(toc))
	out := make([]byte, 28, 28+len(compressed)+len(heap))
	copy(out, "xar!")
	binary.BigEndian.PutUint16(out[4:], 28)
	binary.BigEndian.PutUint16(out[6:], 1)
	binary.BigEndian.PutUint64(out[8:], uint64(len(compressed)))
	binary.BigEndian.PutUint64(out[16:], uint64(len(toc)))
	out = append(out, compressed...)
	return append(out, heap...)
}

func xarMemberTOC(style string, offset, length, size int) string {
	return fmt.Sprintf(`<file><name>sample</name><data><offset>%d</offset><length>%d</length><size>%d</size><encoding style="%s"/></data></file>`, offset, length, size, style)
}

func TestXARScansHeapExtentsAndKeepsLaterDetection(t *testing.T) {
	payload := []byte(payloadMarker)
	sum := md5.Sum(payload)
	hashRule := fmt.Sprintf("%x:%d:Payload.Hash\n", sum, len(payload))
	xzData := func() []byte {
		var out bytes.Buffer
		w, err := (xz.WriterConfig{DictCap: 1 << 16}).NewWriter(&out)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return out.Bytes()
	}()
	lzmaData := func() []byte {
		var out bytes.Buffer
		w, err := (lzma.WriterConfig{DictCap: 1 << 16, SizeInHeader: true, Size: int64(len(payload))}).NewWriter(&out)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return out.Bytes()
	}()
	variants := map[string][]byte{
		"application/octet-stream": payload,
		"application/x-gzip":       xarZlib(t, payload),
		"application/x-lzma":       lzmaData,
		"application/x-xz":         xzData,
	}
	archives := make(map[string][]byte)
	for style, packed := range variants {
		toc := `<xar><toc>` + xarMemberTOC(style, 0, len(packed), len(payload)) + `</toc></xar>`
		archive := xarFixture(t, toc, packed)
		archives[style] = archive
		expectDetected(t, pipelineState(t, map[string]string{"marker.hdb": hashRule}), archive, "Payload.Hash")
	}
	// A damaged first extent cannot conceal a later intact member.
	bad := []byte{0x78, 0x9c, 0xff}
	toc := `<xar><toc>` + xarMemberTOC("application/x-gzip", 0, len(bad), 100) +
		xarMemberTOC("application/octet-stream", len(bad), len(payload), len(payload)) + `</toc></xar>`
	expectDetected(t, pipelineState(t, map[string]string{"marker.hdb": hashRule}),
		xarFixture(t, toc, append(bad, payload...)), "Payload.Hash")

	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		t.Run("pinned-clamav", func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "marker.hdb"), []byte(hashRule), 0o600); err != nil {
				t.Fatal(err)
			}
			var paths []string
			for style, archive := range archives {
				name := strings.TrimPrefix(style, "application/")
				path := filepath.Join(dir, name+".xar")
				if err := os.WriteFile(path, archive, 0o600); err != nil {
					t.Fatal(err)
				}
				paths = append(paths, path)
			}
			args := append([]string{"--database=" + dir, "--no-summary"}, paths...)
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), args...)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok {
				t.Fatalf("reference exit: %v output=%s", err, out)
			}
			for _, path := range paths {
				if !strings.Contains(string(out), path+": Payload.Hash.UNOFFICIAL FOUND") {
					t.Fatalf("reference missed %s: %s", path, out)
				}
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}

func TestXARBoundsAndUnknownExtents(t *testing.T) {
	missing := `<xar><toc>` + xarMemberTOC("application/octet-stream", 1000, 20, 20) + `</toc></xar>`
	expectIncomplete(t, payloadState(t), xarFixture(t, missing, nil), ErrMalformed)
	broken := xarFixture(t, `<xar><toc><file><data><offset>0</offset></data></file></toc></xar>`, []byte("harmless"))
	expectIncomplete(t, payloadState(t), broken, ErrMalformed)
	valid := xarFixture(t, `<xar><toc>`+xarMemberTOC("application/x-gzip", 0, 100, 100)+`</toc></xar>`, []byte{0x78, 0x9c, 0xff})
	expectIncomplete(t, payloadState(t), valid, ErrMalformed)
	if _, _, err := parseXARTOC([]byte(`<xar><toc><data><offset>18446744073709551616</offset><length>1</length><size>1</size></data></toc></xar>`), 1); !errors.Is(err, ErrMalformed) {
		t.Fatalf("overflowing extent: %v", err)
	}
	outside := `<xar><toc></toc>` + xarMemberTOC("application/octet-stream", 0, len(payloadMarker), len(payloadMarker)) + `</xar>`
	extents, _, err := parseXARTOC([]byte(outside), 2)
	if err != nil || len(extents) != 0 {
		t.Fatalf("extent outside toc: extents=%d err=%v", len(extents), err)
	}
}

func TestXARCreatedBySystemTool(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS xar fixture generation")
	}
	xar, err := exec.LookPath("xar")
	if err != nil {
		t.Skip("xar command unavailable")
	}
	dir := t.TempDir()
	payload := []byte(payloadMarker)
	if err := os.WriteFile(filepath.Join(dir, "sample.txt"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(payload)
	hashRule := fmt.Sprintf("%x:%d:Payload.Hash\n", sum, len(payload))
	if err := os.WriteFile(filepath.Join(dir, "marker.hdb"), []byte(hashRule), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, codec := range []string{"none", "gzip", "bzip2", "lzma"} {
		t.Run(codec, func(t *testing.T) {
			path := filepath.Join(dir, codec+".xar")
			cmd := exec.Command(xar, "-c", "-f", path, "--compression="+codec, "sample.txt")
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				if codec == "lzma" && strings.Contains(string(out), "lzma support not compiled in") {
					t.Skip("system xar lacks LZMA encoder; generated LZMA fixture covers this path")
				}
				t.Fatalf("xar create: %v: %s", err, out)
			}
			archive, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			expectDetected(t, pipelineState(t, map[string]string{"marker.hdb": hashRule}), archive, "Payload.Hash")
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+filepath.Join(dir, "marker.hdb"), "--no-summary", path)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
					"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "Payload.Hash.UNOFFICIAL FOUND") {
					t.Fatalf("reference did not detect member: err=%v output=%s", err, out)
				}
			}
		})
	}
}

func TestXARSubdocumentIsSeparateScannedObject(t *testing.T) {
	subdoc := []byte(`<evidence>subdocument-only</evidence>`)
	toc := `<xar><subdoc>` + string(subdoc) + `</subdoc><toc/></xar>`
	archive := xarFixture(t, toc, nil)
	sum := md5.Sum(subdoc)
	hashRule := fmt.Sprintf("%x:%d:Subdoc.Hash\n", sum, len(subdoc))
	expectDetected(t, pipelineState(t, map[string]string{"subdoc.hdb": hashRule}), archive, "Subdoc.Hash")
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdoc.xar")
		if err := os.WriteFile(path, archive, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "subdoc.hdb"), []byte(hashRule), 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+filepath.Join(dir, "subdoc.hdb"), "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "Subdoc.Hash.UNOFFICIAL FOUND") {
			t.Fatalf("reference did not scan subdoc: err=%v output=%s", err, out)
		}
	}
}
