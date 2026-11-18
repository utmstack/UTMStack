package engine

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestMSXMLPropertiesReference(t *testing.T) {
	for _, tc := range []struct{ typ, root, jsonName string }{
		{"CL_TYPE_XML_WORD", "wordDocument", "WordDocument"},
		{"CL_TYPE_XML_XL", "Workbook", "Workbook"},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			input := []byte(`<?xml version="1.0"?><w:` + tc.root + ` xmlns:w="urn:schemas-microsoft-com:office:word" version="1">` +
				`<DocumentProperties><Author>Alice</Author><Revision>12</Revision><AllowPNG>true</AllowPNG><Author>Bob</Author></DocumentProperties>` +
				`<Fonts><Author>Ignored</Author></Fonts><binData>SEVMTE8=</binData><binData>V09STEQ=</binData></w:` + tc.root + `>`)
			got, err := msxmlProperties(context.Background(), input, 1<<20, tc.typ)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]any{
				tc.jsonName: map[string]any{"Attributes": map[string]any{"w": "urn:schemas-microsoft-com:office:word", "version": "1"}},
				"DocumentProperties": map[string]any{
					"Author":   map[string]any{"Value": []any{"Alice", "Bob"}},
					"Revision": map[string]any{"Value": []any{int32(12)}},
					"AllowPNG": map[string]any{"Value": []any{true}},
				},
				"BinaryData": map[string]any{"Count": int32(2)},
			}
			for key, value := range want {
				if !reflect.DeepEqual(got[key], value) {
					t.Fatalf("%s = %#v; want %#v", key, got[key], value)
				}
			}
			var typ signatures.FileType
			if tc.typ == "CL_TYPE_XML_WORD" {
				typ = signatures.TypeXMLWord
			} else {
				typ = signatures.TypeXMLXL
			}
			preclass, err := rootPreclassMetadata(context.Background(), input, typ)
			if err != nil {
				t.Fatal(err)
			}
			for key, value := range want {
				if !reflect.DeepEqual(preclass[key], value) {
					t.Fatalf("PRECLASS %s = %#v; want %#v", key, preclass[key], value)
				}
			}
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				dir := t.TempDir()
				ftm := "0:0:3c3f786d6c:MS XML:CL_TYPE_ANY:" + tc.typ + "\n"
				if err := os.WriteFile(filepath.Join(dir, "type.ftm"), []byte(ftm), 0o600); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "metadata.xml")
				if err := os.WriteFile(path, input, 0o600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--debug", "--gen-json", "--database="+dir, path)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("reference metadata: %v: %s", err, out)
				}
				start := strings.LastIndex(string(out), "LibClamAV debug: {\n")
				if start < 0 {
					t.Fatalf("missing reference JSON: %s", out)
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
				for key := range want {
					encoded, _ := json.Marshal(got[key])
					var normalized any
					_ = json.Unmarshal(encoded, &normalized)
					if !reflect.DeepEqual(normalized, upstream[key]) {
						t.Fatalf("%s differs: Go=%v reference=%v", key, normalized, upstream[key])
					}
				}
			}
		})
	}
}

func TestMSXMLNumericEntityRepairBounds(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{`&#82V`, `&#82;V`}, {`&#x52V`, `&#x52;V`},
		{`&#82;V`, `&#82;V`}, {`&#82`, `&#82`}, {`&word;`, `&word;`},
	} {
		got, err := repairMSXMLNumericEntities(context.Background(), []byte(tc.input), 32)
		if err != nil || string(got) != tc.want {
			t.Fatalf("repair %q: %q %v; want %q", tc.input, got, err, tc.want)
		}
	}
	if _, err := repairMSXMLNumericEntities(context.Background(), []byte(`&#82V`), 5); !errors.Is(err, ErrLimit) {
		t.Fatalf("expanded limit: %v", err)
	}
	if _, err := repairMSXMLNumericEntities(context.Background(), []byte(`&#82Vtrailing`), 13); !errors.Is(err, ErrLimit) {
		t.Fatalf("expanded limit after repair: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repairMSXMLNumericEntities(ctx, []byte(`&#82V`), 32); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled repair: %v", err)
	}
}

func msxmlFixture(root, binaryTag string, chunks ...[]byte) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?>`)
	b.WriteString("<" + root + ">")
	for _, chunk := range chunks {
		b.WriteString("<" + binaryTag + ">\n")
		b.WriteString(base64.StdEncoding.EncodeToString(chunk))
		b.WriteString("\n</" + binaryTag + ">")
	}
	b.WriteString("</" + root + ">")
	return []byte(b.String())
}

func TestMSXMLSharedReaderRecursionBoundary(t *testing.T) {
	marker := []byte("UTMSTACK-MSOFFICE-DEPTH-HARMLESS")
	sum := md5.Sum(marker)
	files := map[string]string{
		"type.ftm":   "0:0:3c3f786d6c:MS XML:CL_TYPE_ANY:CL_TYPE_XML_WORD\n",
		"marker.hdb": fmt.Sprintf("%x:%d:MSXML.Depth\n", sum, len(marker)),
	}
	child := "<binData>\n" + base64.StdEncoding.EncodeToString(marker) + "\n</binData>"
	fixture := func(wrappers int) []byte {
		return []byte(`<?xml version="1.0"?><wordDocument>` +
			strings.Repeat("<N>", wrappers) + child + strings.Repeat("</N>", wrappers) + "</wordDocument>")
	}
	// The document root is level zero. The 18th wrapper puts binData at
	// level 19; the 19th puts it at the reference's skipped level 20.
	expectDetected(t, pipelineState(t, files), fixture(18), "MSXML.Depth")
	expectIncomplete(t, pipelineState(t, files), fixture(19), ErrUnsupported)
	for _, tc := range []struct {
		wrappers int
		want     bool
	}{
		{17, true}, {18, false},
	} {
		data := []byte(`<?xml version="1.0"?><wordDocument>` + strings.Repeat("<N>", tc.wrappers) +
			"<DocumentProperties><Author>Alice</Author></DocumentProperties>" +
			strings.Repeat("</N>", tc.wrappers) + "</wordDocument>")
		properties, err := msxmlProperties(context.Background(), data, 1<<20, "CL_TYPE_XML_WORD")
		if err != nil {
			t.Fatal(err)
		}
		document, _ := properties["DocumentProperties"].(map[string]any)
		_, found := document["Author"]
		if found != tc.want {
			t.Fatalf("wrappers=%d Author present=%v, want %v", tc.wrappers, found, tc.want)
		}
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, contents := range map[string][]byte{
			"type.ftm": []byte(files["type.ftm"]), "marker.hdb": []byte(files["marker.hdb"]),
			"depth19.xml": fixture(18), "depth20.xml": fixture(19),
		} {
			if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary",
			filepath.Join(dir, "depth19.xml"), filepath.Join(dir, "depth20.xml"))
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if _, ok := err.(*exec.ExitError); !ok ||
			!strings.Contains(string(out), "depth19.xml: MSXML.Depth.UNOFFICIAL FOUND") ||
			!strings.Contains(string(out), "depth20.xml: OK") {
			t.Fatalf("reference recursion boundary: err=%v output=%s", err, out)
		}
	}
}

func TestMSXMLDepthSkipHonorsCancellation(t *testing.T) {
	dec := xml.NewDecoder(strings.NewReader("<N>" + strings.Repeat("<N>", 4096) +
		strings.Repeat("</N>", 4096) + "</N>"))
	if _, err := dec.Token(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := skipMSXMLSubtree(ctx, dec); !errors.Is(err, context.Canceled) {
		t.Fatalf("depth skip after cancellation: %v", err)
	}
}

func TestMSXMLScansBinaryDataChildren(t *testing.T) {
	payload := []byte("EXTRACTED-HARMLESS-MSOFFICE-PAYLOAD")
	sum := md5.Sum(payload)
	hashRule := fmt.Sprintf("%x:%d:MSXML.Child\n", sum, len(payload))
	for _, tc := range []struct {
		name, typeName, root, tag string
	}{
		{"word", "CL_TYPE_XML_WORD", "wordDocument", "binData"},
		{"workbook", "CL_TYPE_XML_XL", "Workbook", "binData"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ftm := "0:0:3c3f786d6c:MS XML:CL_TYPE_ANY:" + tc.typeName + "\n"
			state := func() *scanState {
				return pipelineState(t, map[string]string{"type.ftm": ftm, "marker.hdb": hashRule})
			}
			positive := msxmlFixture(tc.root, tc.tag, payload)
			expectDetected(t, state(), positive, "MSXML.Child")
			withEntity := []byte(strings.Replace(string(positive), "\nRVh", "\n&#82Vh", 1))
			expectDetected(t, state(), withEntity, "MSXML.Child")
			withHexEntity := []byte(strings.Replace(string(positive), "\nRVh", "\n&#x52Vh", 1))
			expectDetected(t, state(), withHexEntity, "MSXML.Child")
			expectDetected(t, state(), positive[:len(positive)-len(tc.root)-3], "MSXML.Child")
			expectIncomplete(t, state(), msxmlFixture(tc.root, tc.tag, []byte("harmless child")), ErrUnsupported)
			multiple := msxmlFixture(tc.root, tc.tag, []byte("harmless first member"), payload)
			expectDetected(t, state(), multiple, "MSXML.Child")
			malformedFirst := []byte(strings.Replace(string(multiple), base64.StdEncoding.EncodeToString([]byte("harmless first member")), "not-base64!", 1))
			expectDetected(t, state(), malformedFirst, "MSXML.Child")
			unwrapped := []byte(strings.ReplaceAll(string(positive), ">\n", ">"))
			unwrapped = []byte(strings.ReplaceAll(string(unwrapped), "\n</", "</"))
			expectDetected(t, state(), unwrapped, "MSXML.Child")
			ignored := map[string][]byte{}
			for _, section := range []string{"Styles", "Fonts"} {
				ignored[strings.ToLower(section)+".xml"] = []byte(`<?xml version="1.0"?><` + tc.root + `><` + section + `><` + tc.tag + `>` +
					base64.StdEncoding.EncodeToString(payload) + `</` + tc.tag + `></` + section + `></` + tc.root + `>`)
				expectIncomplete(t, state(), ignored[strings.ToLower(section)+".xml"], ErrUnsupported)
			}
			limited := state()
			limited.remaining = int64(len(positive) + len(payload) - 1)
			expectIncomplete(t, limited, positive, ErrLimit)
			expectIncomplete(t, state(), []byte(`<?xml version="1.0"?><`+tc.root+`><`+tc.tag+`>broken`), ErrMalformed)

			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				dir := t.TempDir()
				for name, content := range map[string][]byte{
					"type.ftm": []byte(ftm), "marker.hdb": []byte(hashRule),
					"positive.xml": positive, "multiple.xml": multiple, "unwrapped.xml": unwrapped,
					"numeric-entity.xml": withEntity, "hex-entity.xml": withHexEntity,
					"malformed-first.xml": malformedFirst,
					"styles.xml":          ignored["styles.xml"], "fonts.xml": ignored["fonts.xml"],
				} {
					if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary",
					filepath.Join(dir, "positive.xml"), filepath.Join(dir, "multiple.xml"), filepath.Join(dir, "unwrapped.xml"),
					filepath.Join(dir, "malformed-first.xml"), filepath.Join(dir, "styles.xml"), filepath.Join(dir, "fonts.xml"),
					filepath.Join(dir, "numeric-entity.xml"), filepath.Join(dir, "hex-entity.xml"))
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
					"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				if _, ok := err.(*exec.ExitError); !ok ||
					!strings.Contains(string(out), "positive.xml: MSXML.Child.UNOFFICIAL FOUND") ||
					!strings.Contains(string(out), "numeric-entity.xml: MSXML.Child.UNOFFICIAL FOUND") ||
					!strings.Contains(string(out), "hex-entity.xml: MSXML.Child.UNOFFICIAL FOUND") ||
					!strings.Contains(string(out), "multiple.xml: MSXML.Child.UNOFFICIAL FOUND") ||
					!strings.Contains(string(out), "unwrapped.xml: OK") ||
					!strings.Contains(string(out), "malformed-first.xml: MSXML.Child.UNOFFICIAL FOUND") ||
					!strings.Contains(string(out), "styles.xml: OK") || !strings.Contains(string(out), "fonts.xml: OK") {
					t.Fatalf("reference outcome: err=%v output=%s", err, out)
				}
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
