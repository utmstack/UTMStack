package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestScriptEncoderClamAVOracle(t *testing.T) {
	fixture, err := os.ReadFile("testdata/screnc_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name                   string `json:"name"`
		InputB64               string `json:"input_b64"`
		ReferenceScrencHTMLB64 string `json:"reference_screnc_html_b64"`
	}
	if err := json.Unmarshal(fixture, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			input, err := base64.StdEncoding.DecodeString(tc.InputB64)
			if err != nil {
				t.Fatal(err)
			}
			want, err := base64.StdEncoding.DecodeString(tc.ReferenceScrencHTMLB64)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := decodeScrEnc(context.Background(), input)
			partial := tc.Name == "truncated" || tc.Name == "partial"
			if partial && !errors.Is(err, ErrUnsupported) || !partial && err != nil {
				t.Fatal(err)
			}
			got := append(append([]byte("<script>"), decoded...), "</script>"...)
			if !bytes.Equal(got, want) {
				t.Fatalf("decoded wrapper differs: got %d bytes, reference %d", len(got), len(want))
			}
		})
	}
}

func TestScriptEncoderReferenceVectors(t *testing.T) {
	for _, tt := range []struct{ encoded, decoded string }{
		{"#@~^CwAAAA==C^+.D`ECkE#aAMAAA==^#~@", `alert('Hi')`},
		{"#@~^CgAAAA==\\ko$K6,J(rDAMAAA==^#~@", `MsgBox "X"`},
	} {
		got, err := decodeScrEnc(context.Background(), []byte(tt.encoded))
		if err != nil || string(got) != tt.decoded {
			t.Errorf("%q: got %q, %v; want %q", tt.encoded, got, err, tt.decoded)
		}
	}
}

func TestScriptEncoderChildIsScanned(t *testing.T) {
	const encoded = "#@~^CwAAAA==C^+.D`ECkE#aAMAAA==^#~@"
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c736372697074:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Decoded.ScriptEncoder:3:*:" + hx(`alert("hi")`) + "\n",
	}
	expectScan(t, embeddedState(t, rules), []byte(encoded), "Decoded.ScriptEncoder", "")
	bad := encoded[:len(encoded)-1]
	expectScan(t, embeddedState(t, rules), []byte(bad), "Decoded.ScriptEncoder", "")
	rules["a.ndb"] = "Other:3:*:" + hx("unrelated") + "\n"
	expectScan(t, embeddedState(t, rules), []byte(encoded), "", "")
	sig, err := embeddedState(t, rules).visit([]byte(bad), 0)
	if sig != "" || !errors.Is(err, ErrUnsupported) || !strings.Contains(err.Error(), "format-screnc") {
		t.Fatalf("truncated envelope without detection: %q, %v", sig, err)
	}
}

func TestScriptEncoderIgnoresChecksumAsReferenceDoes(t *testing.T) {
	const encoded = "#@~^CwAAAA==C^+.D`ECkE#aAMAAA==^#~@"
	badChecksum := []byte(encoded)
	badChecksum[len(badChecksum)-12] = 'Z'
	got, err := decodeScrEnc(context.Background(), badChecksum)
	if err != nil || string(got) != `alert('Hi')` {
		t.Fatalf("checksum-only change: %q, %v", got, err)
	}
}

func TestEncodedHTMLScriptDecodedObjectIsScanned(t *testing.T) {
	const encoded = "#@~^CwAAAA==C^+.D`ECkE#aAMAAA==^#~@"
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c736372697074:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n" +
			"0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Decoded.EmbeddedScriptEncoder:3:*:" + hx(`alert("hi")`) + "\n",
	}
	page := []byte(`<html><script language="JScript.Encode">` + encoded + `</script></html>`)
	expectScan(t, embeddedState(t, rules), page, "Decoded.EmbeddedScriptEncoder", "")
	rules["a.ndb"] = "Other:3:*:" + hx(`alert("other")`) + "\n"
	expectScan(t, embeddedState(t, rules), page, "", "")
}

func TestScriptEncoderCancellation(t *testing.T) {
	fixture, err := os.ReadFile("testdata/screnc_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string `json:"name"`
		InputB64 string `json:"input_b64"`
	}
	if err := json.Unmarshal(fixture, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		if tc.Name != "chunk" {
			continue
		}
		input, err := base64.StdEncoding.DecodeString(tc.InputB64)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := decodeScrEnc(ctx, input); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled decode: %v", err)
		}
		return
	}
	t.Fatal("missing chunk oracle")
}
