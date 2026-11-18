package engine

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestEncodedHTMLClamAVViews(t *testing.T) {
	fixture, err := os.ReadFile("testdata/html_screnc_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name      string            `json:"name"`
		SourceB64 string            `json:"source_b64"`
		ViewsB64  map[string]string `json:"views_b64"`
	}
	if err := json.Unmarshal(fixture, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 90 {
		t.Fatalf("got %d oracle cases", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			source, err := base64.StdEncoding.DecodeString(tc.SourceB64)
			if err != nil {
				t.Fatal(err)
			}
			views := normalizeHTML(source)
			js, jsUnsupported := normalizeJS(views.scripts)
			if views.unsupported != "" || jsUnsupported {
				if !strings.Contains(tc.Name, "partial-") && !strings.Contains(tc.Name, "long-declared") {
					t.Fatalf("unexpected incomplete view: %q js=%t", views.unsupported, jsUnsupported)
				}
			}
			for key, got := range map[string][]byte{"page": views.noComment, "text": views.noTags, "javascript": js} {
				want, err := base64.StdEncoding.DecodeString(tc.ViewsB64[key])
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != string(want) {
					t.Errorf("%s: got %q (%d bytes), want %q (%d bytes)", key, got[:min(len(got), 120)], len(got), want[:min(len(want), 120)], len(want))
				}
			}
		})
	}
}

func TestEncodedASPDecodedSignature(t *testing.T) {
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Decoded.ASP:3:*:" + hx(`msgbox "x"`) + "\n",
	}
	encoded := []byte("<html><%@ LANGUAGE = \"VBScript.Encode\" %>\n<% #@~^CgAAAA==\\ko$K6,J(rDAMAAA==^#~@ %></html>")
	expectScan(t, embeddedState(t, rules), encoded, "Decoded.ASP", "")
	control := []byte("<html><%@ LANGUAGE = \"VBScript.Encode\" %>\n<% ordinary text %></html>")
	expectScan(t, embeddedState(t, rules), control, "", "")
}

func TestEncodedHTMLRandomClamAVViews(t *testing.T) {
	fixture, err := os.ReadFile("testdata/html_screnc_random_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name      string            `json:"name"`
		SourceB64 string            `json:"source_b64"`
		ViewsB64  map[string]string `json:"views_b64"`
	}
	if err := json.Unmarshal(fixture, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 120 {
		t.Fatalf("got %d random cases", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			source, err := base64.StdEncoding.DecodeString(tc.SourceB64)
			if err != nil {
				t.Fatal(err)
			}
			views := normalizeHTML(source)
			js, jsUnsupported := normalizeJS(views.scripts)
			if views.unsupported != "" || jsUnsupported {
				t.Fatalf("unexpected incomplete view: %q js=%t", views.unsupported, jsUnsupported)
			}
			for key, got := range map[string][]byte{"page": views.noComment, "text": views.noTags, "javascript": js} {
				want, err := base64.StdEncoding.DecodeString(tc.ViewsB64[key])
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != string(want) {
					t.Errorf("%s: got %q (%d), want %q (%d)", key, got[:min(len(got), 120)], len(got), want[:min(len(want), 120)], len(want))
				}
			}
		})
	}
}

func TestTruncatedEncodedHTMLStillScansPartialView(t *testing.T) {
	page := []byte("<html><script language=\"JScript.Encode\">#@~^CwAAAA==C^+.D`EC</script></html>")
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Partial.ScriptEncoder:3:*:" + hx(`alert("h`) + "\n",
	}
	expectScan(t, embeddedState(t, rules), page, "Partial.ScriptEncoder", "")
	rules["a.ndb"] = "Other:3:*:" + hx("unrelated") + "\n"
	expectScan(t, embeddedState(t, rules), page, "", "")
}

func TestLongDeclaredScriptEncoderStillScansPartialView(t *testing.T) {
	page := []byte("<html><script language=\"JScript.Encode\">#@~^6AMAAA==C^+.D`ECkE#aAMAAA==^#~@</script></html>")
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Partial.ScriptEncoder:3:*:" + hx(`alert("h`) + "\n",
	}
	expectScan(t, embeddedState(t, rules), page, "Partial.ScriptEncoder", "")
	rules["a.ndb"] = "Other:3:*:" + hx("unrelated") + "\n"
	expectScan(t, embeddedState(t, rules), page, "", "html-normalization")
}
