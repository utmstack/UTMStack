package engine

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func hwpMLPropertiesFixture() []byte {
	return []byte("\ufeff" + `<?xml version="1.0"?><HWPML Version="1"><HEAD>` +
		`<DOCSUMMARY><TITLE>Hello</TITLE><AUTHOR>Alice</AUTHOR><DATE>2026</DATE></DOCSUMMARY>` +
		`<DOCSETTING><BEGINNUMBER Page="2"/><CARETPOS List="3"/></DOCSETTING>` +
		`<BODY><TITLE>Ignored</TITLE></BODY></HEAD><TAIL><BINDATASTORAGE>` +
		`<BINDATA Encoding="Base64" Compress="false">SEVMTE8=</BINDATA>` +
		`<BINDATA ID="B">SAFE</BINDATA></BINDATASTORAGE>` +
		`<SCRIPTCODE Name="x"><SCRIPTHEADER>code</SCRIPTHEADER>` +
		`<SCRIPTSOURCE><![CDATA[ignored]]>source</SCRIPTSOURCE></SCRIPTCODE>` +
		`</TAIL></HWPML>`)
}

func TestHWPMLPropertiesReference(t *testing.T) {
	input := hwpMLPropertiesFixture()
	properties, err := hwpMLProperties(context.Background(), input, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"Attributes": map[string]any{"Version": "1"},
		"DocumentProperties": map[string]any{
			"Title":  map[string]any{"Value": []any{"Hello"}},
			"Author": map[string]any{"Value": []any{"Alice"}},
			"Date":   map[string]any{"Value": []any{int32(2026)}},
		},
		"DocumentSettings": map[string]any{
			"BeginNumber": map[string]any{"Attributes": map[string]any{"Page": "2"}},
			"CaretPos":    map[string]any{"Attributes": map[string]any{"List": "3"}},
		},
		"BinaryDataStorage": map[string]any{
			"BinaryData": map[string]any{"Attributes": map[string]any{
				"Encoding": "Base64", "Compress": "false", "ID": "B",
			}},
		},
		"ScriptCodeStorage": map[string]any{
			"Attributes":   map[string]any{"Name": "x"},
			"ScriptHeader": map[string]any{"Value": []any{"code"}},
			"ScriptSource": map[string]any{"Value": []any{"source"}},
		},
	}
	if !reflect.DeepEqual(properties["HWPML"], want) {
		t.Fatalf("HWPML properties differ: got=%#v want=%#v", properties["HWPML"], want)
	}
	metadata, err := rootPreclassMetadata(context.Background(), input, signatures.TypeXMLHWP)
	if err != nil || !reflect.DeepEqual(metadata["HWPML"], want) {
		t.Fatalf("PRECLASS HWPML properties: got=%#v err=%v", metadata["HWPML"], err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "type.ftm"), []byte(hwpMLFTM), 0o600); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "metadata.xml")
		if err := os.WriteFile(path, input, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--debug", "--gen-json", "--database="+dir, path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("reference metadata: %v: %s", err, out)
		}
		start := strings.LastIndex(string(out), "LibClamAV debug: {\n")
		if start < 0 {
			t.Fatalf("reference JSON missing: %s", out)
		}
		var jsonLines []string
		for _, line := range strings.Split(string(out)[start:], "\n") {
			if !strings.HasPrefix(line, "LibClamAV debug: ") {
				break
			}
			jsonLines = append(jsonLines, strings.TrimPrefix(line, "LibClamAV debug: "))
			if strings.TrimSpace(line) == "LibClamAV debug: }" {
				break
			}
		}
		var upstream map[string]any
		if err := json.Unmarshal([]byte(strings.Join(jsonLines, "\n")), &upstream); err != nil {
			t.Fatalf("reference JSON parse: %v: %s", err, jsonLines)
		}
		encoded, _ := json.Marshal(properties["HWPML"])
		var normalized any
		_ = json.Unmarshal(encoded, &normalized)
		if !reflect.DeepEqual(normalized, upstream["HWPML"]) {
			t.Fatalf("HWPML metadata differs: Go=%v reference=%v", normalized, upstream["HWPML"])
		}
	}
}

func TestHWPMLPropertiesDepthBoundary(t *testing.T) {
	for _, tc := range []struct {
		wrappers int
		want     bool
	}{
		{17, true}, {18, false},
	} {
		input := []byte("\ufeff" + `<?xml version="1.0"?><HWPML>` + strings.Repeat("<N>", tc.wrappers) +
			"<DOCSUMMARY><TITLE>Hello</TITLE></DOCSUMMARY>" + strings.Repeat("</N>", tc.wrappers) + "</HWPML>")
		properties, err := hwpMLProperties(context.Background(), input, 1<<20)
		if err != nil {
			t.Fatal(err)
		}
		hwp, _ := properties["HWPML"].(map[string]any)
		doc, _ := hwp["DocumentProperties"].(map[string]any)
		_, found := doc["Title"]
		if found != tc.want {
			t.Fatalf("wrappers=%d Title present=%v want=%v", tc.wrappers, found, tc.want)
		}
	}
}

func TestHWPMLPreclassBytecodeReadsPropertyObject(t *testing.T) {
	program := engineCBCProgram(260, "json-hwpml-object", false)
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	if _, err := e.ActivateGeneration(map[string][]byte{
		"type.ftm": []byte(hwpMLFTM), "preclass.cbc": program,
	}, 1, manifestHash("hwpml-preclass-json")); err != nil {
		t.Fatal(err)
	}
	input := []byte(strings.Replace(string(hwpMLPropertiesFixture()), "Hello", "Hello "+engineCBCMarker, 1))
	clean, name, err := e.Scan(context.Background(), input)
	if err != nil || clean || name != engineCBCLabel {
		t.Fatalf("HWPML PRECLASS property query: clean=%v name=%q err=%v", clean, name, err)
	}
	negative := []byte(strings.ReplaceAll(strings.ReplaceAll(string(input), "<HWPML ", "<OTHER "), "</HWPML>", "</OTHER>"))
	if clean, name, err := e.Scan(context.Background(), negative); name != "" || clean || err == nil {
		t.Fatalf("missing HWPML object produced a verdict: clean=%v name=%q err=%v", clean, name, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, contents := range map[string][]byte{
			"type.ftm": []byte(hwpMLFTM), "preclass.cbc": program, "input.xml": input,
			"negative.xml": negative,
		} {
			if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--gen-json", "--bytecode-unsigned", "--database="+dir, "--no-summary", filepath.Join(dir, "input.xml"), filepath.Join(dir, "negative.xml"))
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "input.xml: "+engineCBCLabel+" FOUND") ||
			!strings.Contains(string(out), "negative.xml: OK") {
			t.Fatalf("native HWPML PRECLASS property query: err=%v output=%s", err, out)
		}
	}
}
