package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// A parser error must preserve independent hooks and proved root identity.
// A no-hit hook must retain the original incomplete verdict.
func TestPreclassRunsAfterMetadataParserError(t *testing.T) {
	for _, doc := range []struct{ name, typ, body string }{
		{"word-truncated", "CL_TYPE_XML_WORD", `<wordDocument><DocumentProperties><Author>harmless`},
		{"word-entity", "CL_TYPE_XML_WORD", `<wordDocument>&undefined;</wordDocument>`},
		{"workbook-truncated", "CL_TYPE_XML_XL", `<Workbook><DocumentProperties><Author>harmless`},
		{"workbook-entity", "CL_TYPE_XML_XL", `<Workbook>&undefined;</Workbook>`},
		{"hwp-truncated", "CL_TYPE_XML_HWP", `<HWPML><HEAD><DOCSUMMARY><AUTHOR>harmless`},
		{"hwp-entity", "CL_TYPE_XML_HWP", `<HWPML>&undefined;</HWPML>`},
	} {
		data := []byte(`<?xml version="1.0"?><!--` + engineCBCMarker + `-->` + doc.body)
		for _, query := range []struct {
			mode string
			hit  bool
		}{
			{"detect", true}, {"no-label", false}, {"json-active", true},
			{fmt.Sprintf("json-field-int:FileSize:%d", len(data)), true},
			{"json-field-int:FileSize:0", false},
			{fmt.Sprintf("json-field-length:RootFileType:%d", len(doc.typ)), true},
		} {
			t.Run(doc.name+"/"+query.mode, func(t *testing.T) {
				program := engineCBCProgram(260, query.mode, false)
				ftm := []byte("0:0:3c3f786d6c:Office XML error control:CL_TYPE_ANY:" + doc.typ + "\n")
				if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
					dir := t.TempDir()
					for name, contents := range map[string][]byte{"preclass.cbc": program, "type.ftm": ftm, "input.xml": data} {
						if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
							t.Fatal(err)
						}
					}
					cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--bytecode-unsigned", "--gen-json", "--database="+dir, "--no-summary", filepath.Join(dir, "input.xml"))
					cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
					out, err := cmd.CombinedOutput()
					t.Logf("native expected_detection=%v error=%v output=%s", query.hit, err, out)
					checkNativeMetadataVerdict(t, out, err, query.hit, engineCBCLabel)
				}
				cfg := config.Default()
				cfg.CollectMetadata = true
				cfg.SignatureDBDir = t.TempDir()
				e := New(cfg)
				if _, err := e.ActivateGeneration(map[string][]byte{"preclass.cbc": program, "type.ftm": ftm}, 1, manifestHash(doc.name+query.mode)); err != nil {
					t.Fatal(err)
				}
				typ, _ := e.current.Load().db.DetermineType(data)
				if "CL_TYPE_"+typ.String() != doc.typ {
					t.Fatalf("wrong root type %s", typ)
				}
				if _, err := rootPreclassMetadata(context.Background(), data, typ); IncompleteReason(err) != "Malformed" {
					t.Fatalf("fixture must fail metadata parsing: %v", err)
				}
				clean, name, err := e.Scan(context.Background(), data)
				t.Logf("Go clean=%v signature=%q error=%v expected_detection=%v", clean, name, err, query.hit)
				if query.hit {
					if clean || name != engineCBCLabel || err != nil {
						t.Fatalf("independent hook suppressed: clean=%v signature=%q error=%v", clean, name, err)
					}
				} else if clean || name != "" || IncompleteReason(err) != "Malformed" {
					t.Fatalf("parser error lost: clean=%v signature=%q error=%v", clean, name, err)
				}
			})
		}
	}
}

func TestMetadataParserFailurePreservesProvedRootOnly(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><wordDocument><DocumentProperties><Author>harmless`)
	meta, err := initialRootMetadata(context.Background(), data, signatures.TypeXMLWord, "/tmp/root.xml")
	if err != nil {
		t.Fatal(err)
	}
	got, err := finishRootMetadata(context.Background(), data, signatures.TypeXMLWord, meta)
	if IncompleteReason(err) != "Malformed" || got["FileName"] != "root.xml" || got["FileSize"] != uint64(len(data)) || got["RootFileType"] != "CL_TYPE_XML_WORD" {
		t.Fatalf("proved root lost: metadata=%v error=%v", got, err)
	}
	for _, key := range []string{"WordDocument", "DocumentProperties", "BinaryData"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unproved parser subtree published: %s", key)
		}
	}
}
