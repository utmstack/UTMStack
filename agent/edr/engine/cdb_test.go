package engine

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// encryptedZip stores one member and marks it encrypted in both of its
// headers, as an encrypting archiver would.
func encryptedZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	f.Write(content)
	w.Close()
	out := b.Bytes()
	for _, sig := range []struct {
		magic string
		flags int
	}{{"PK\x03\x04", 6}, {"PK\x01\x02", 8}} {
		i := bytes.Index(out, []byte(sig.magic))
		flags := binary.LittleEndian.Uint16(out[i+sig.flags:])
		binary.LittleEndian.PutUint16(out[i+sig.flags:], flags|1)
	}
	return out
}

func TestContainerMetadataRulesRunBeforeExtraction(t *testing.T) {
	content := []byte("encrypted-looking body of the member")
	rule := fmt.Sprintf("Zip.Invoice:CL_TYPE_ZIP:*:invoice\\.exe$:%d:%d:1:1:*:\nRar.Member:CL_TYPE_RAR:*:^x$:*:%d:0:1:*:\n", len(content), len(content), 9)
	s := pipelineState(t, map[string]string{"a.cdb": rule})
	expectScan(t, s, encryptedZip(t, "docs/invoice.exe", content), "Zip.Invoice", "")
	// Not in an Office document, which is typed as such and not as a zip.
	var doc bytes.Buffer
	w := zip.NewWriter(&doc)
	f, _ := w.Create("word/document.xml")
	f.Write([]byte("<w:document/>"))
	f, _ = w.CreateHeader(&zip.FileHeader{Name: "docs/invoice.exe", Method: zip.Store})
	f.Write(content)
	w.Close()
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule}), doc.Bytes(), "", "")
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule}), rarStoredFixture([]byte("123456789"), false), "Rar.Member", "")
}

func TestContainerPOSIXRegexUnknownPropagatesThroughZIP(t *testing.T) {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.CreateHeader(&zip.FileHeader{Name: "report", Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("harmless child")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	rule := "Word:CL_TYPE_ZIP:*:\\<report\\>:*:*:*:*:*:\n"
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule}), b.Bytes(), "", "container-name-regex")
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule, "child.ndb": "Child:0:*:6861726d6c657373206368696c64\n"}), b.Bytes(), "Child", "")
}
