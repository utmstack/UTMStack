package engine

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNamedPackerCallsNativeReference(t *testing.T) {
	raw, err := os.ReadFile("testdata/named_packer_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Name, Source, Output string }
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 204 {
		t.Fatalf("native control inventory %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			views := normalizeHTML([]byte(tc.Source))
			got, unknown := normalizeJS(views.scripts)
			if unknown || string(got) != tc.Output {
				t.Fatalf("Go=%q unknown=%t native=%q", got, unknown, tc.Output)
			}
		})
	}
}

func TestNamedPackerDecodedOnlySignature(t *testing.T) {
	declaration := `function pack(p,a,c,k,e,r){return p} `
	first := `pack('HARMLESS',2,2,''.split('|'),0,{});`
	for _, tail := range []string{"CHILD;", "SAFE;"} {
		source := []byte(`<html><script>` + declaration + first + `pack('` + tail + `',2,2,''.split('|'),0,{});</script></html>`)
		if strings.Contains(strings.ToLower(string(source)), "harmlesschild") {
			t.Fatal("marker must exist only after decoding")
		}
		rules := map[string]string{"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n", "a.ndb": "Decoded.Named.Multi:3:*:" + hx("harmlesschild") + "\n"}
		want := ""
		if tail == "CHILD;" {
			want = "Decoded.Named.Multi"
		}
		expectScan(t, embeddedState(t, rules), source, want, "")
	}
}

func TestNamedPackerCallLimitRemainsIncomplete(t *testing.T) {
	source := `<html><script>function pack(p,a,c,k,e,r){return p} ` + strings.Repeat(`pack('0;',2,2,'safe'.split('|'),0,{});`, 129) + `</script></html>`
	rules := map[string]string{"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n", "a.ndb": "Other:3:*:" + hx("unrelated") + "\n"}
	expectScan(t, embeddedState(t, rules), []byte(source), "", "html-normalization")
}
