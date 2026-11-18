package engine

import (
	"bytes"
	"compress/flate"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const hwpMLFTM = "0:0:efbbbf3c3f786d6c:HWPML:CL_TYPE_ANY:CL_TYPE_XML_HWP:82\n"

func hwpMLFixture(body string) []byte {
	return []byte("\ufeff<?xml version=\"1.0\"?><HWPML><TAIL><BINDATASTORAGE>" + body + "</BINDATASTORAGE></TAIL></HWPML>")
}

func hwpMLDeflate(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w, err := flate.NewWriter(&out, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func hwpMLCases(t *testing.T, marker []byte) map[string][]byte {
	t.Helper()
	b64 := base64.StdEncoding.EncodeToString(marker)
	compressed := base64.StdEncoding.EncodeToString(hwpMLDeflate(t, marker))
	return map[string][]byte{
		"raw":        hwpMLFixture("<BINDATA>" + string(marker) + "</BINDATA>"),
		"script":     hwpMLFixture("<SCRIPTSOURCE>" + string(marker) + "</SCRIPTSOURCE>"),
		"scripthead": hwpMLFixture("<SCRIPTHEADER>" + string(marker) + "</SCRIPTHEADER>"),
		"base64":     hwpMLFixture("<BINDATA Encoding=\"Base64\" Compress=\"false\">" + b64 + "</BINDATA>"),
		"compressed": hwpMLFixture("<BINDATA Encoding=\"Base64\" Compress=\"true\">" + compressed + "</BINDATA>"),
		"body":       hwpMLFixture("<BODY><BINDATA>" + string(marker) + "</BINDATA></BODY>"),
		"cdata":      hwpMLFixture("<BINDATA><![CDATA[" + string(marker) + "]]></BINDATA>"),
		"control":    hwpMLFixture("<BINDATA Encoding=\"Base64\">" + base64.StdEncoding.EncodeToString([]byte("HARMLESS CONTROL")) + "</BINDATA>"),
		"depth19":    hwpMLFixture(strings.Repeat("<N>", 16) + "<BINDATA>" + string(marker) + "</BINDATA>" + strings.Repeat("</N>", 16)),
		"depth20":    hwpMLFixture(strings.Repeat("<N>", 17) + "<BINDATA>" + string(marker) + "</BINDATA>" + strings.Repeat("</N>", 17)),
	}
}

func TestHWPMLScansBinaryAndScriptChildren(t *testing.T) {
	marker := []byte("UTMSTACK-HWPML-HARMLESS-CHILD")
	sum := md5.Sum(marker)
	files := map[string]string{
		"type.ftm":   hwpMLFTM,
		"marker.hdb": fmt.Sprintf("%x:%d:HWPML.Child\n", sum, len(marker)),
	}
	for name, data := range hwpMLCases(t, marker) {
		t.Run(name, func(t *testing.T) {
			if name == "control" || name == "body" || name == "cdata" || name == "depth20" {
				expectIncomplete(t, pipelineState(t, files), data, ErrUnsupported)
			} else {
				expectDetected(t, pipelineState(t, files), data, "HWPML.Child")
			}
		})
	}
	// The reference's built-in signature recognizes the BOM, declaration and
	// HWPML root even if a local hash-only feed contains no .ftm file.
	expectDetected(t, pipelineState(t, map[string]string{"marker.hdb": files["marker.hdb"]}),
		hwpMLCases(t, marker)["raw"], "HWPML.Child")
}

func TestHWPMLPortableClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("EDR_CLAMAV_PORTABLE not set")
	}
	marker := []byte("UTMSTACK-HWPML-HARMLESS-CHILD")
	sum := md5.Sum(marker)
	dir := t.TempDir()
	for name, contents := range map[string]string{
		"type.ftm":   hwpMLFTM,
		"marker.hdb": fmt.Sprintf("%x:%d:HWPML.Child\n", sum, len(marker)),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var paths []string
	for name, data := range hwpMLCases(t, marker) {
		path := filepath.Join(dir, name+".xml")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append([]string{"--database=" + dir, "--no-summary"}, paths...)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference exit: %v output=%s", err, out)
	}
	for _, name := range []string{"raw", "script", "scripthead", "depth19"} {
		if !strings.Contains(string(out), name+".xml: HWPML.Child.UNOFFICIAL FOUND") {
			t.Fatalf("reference missed %s: %s", name, out)
		}
	}
	// In 1.5.4, the decoded temp file is not rewound before scanning, so
	// Base64 scans zero bytes. The Go decoder intentionally scans the child.
	for _, name := range []string{"base64", "compressed", "body", "cdata", "control", "depth20"} {
		if !strings.Contains(string(out), name+".xml: OK") {
			t.Fatalf("reference control %s: %s", name, out)
		}
	}
}
