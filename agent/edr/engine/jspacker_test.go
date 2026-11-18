package engine

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDeanEdwardsPackerClamAVOracle(t *testing.T) {
	fixture, err := os.ReadFile("testdata/packer_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Name, Source, Output string }
	if err := json.Unmarshal(fixture, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 152 {
		t.Fatalf("got %d oracle cases", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			views := normalizeHTML([]byte(tc.Source))
			got, unsupported := normalizeJS(views.scripts)
			if unsupported || string(got) != tc.Output {
				t.Fatalf("got %q, unsupported=%t; want %q", got, unsupported, tc.Output)
			}
		})
	}
}

// Harmless expressions retained from ClamAV 1.5.4 sigtool --html-normalise.
func TestDeanEdwardsPackerReferenceExamples(t *testing.T) {
	for _, tc := range []struct {
		name, source, want string
	}{
		{"anonymous-d", `<html><script>eval(function(p,a,c,k,e,d){return p}('0("1");',2,2,'alert|hi'.split('|'),0,{}))</script></html>`, `<script>eval(function(n000,n001,n002,n003,n004,n005){return n000}(alert("hi");))</script>`},
		{"anonymous-r", `<html><script>eval(function(p,a,c,k,e,r){return p}('0("1");',2,2,'alert|hi'.split('|'),0,{}))</script></html>`, `<script>eval(function(n000,n001,n002,n003,n004,n005){return n000}(alert("hi");))</script>`},
		{"named", `<html><script>function pack(p,a,c,k,e,r){return p} pack('0("1");',2,2,'alert|hi'.split('|'),0,{})</script></html>`, `<script>function n000(n001,n002,n003,n004,n005,n006){return n001}pack(alert("hi");)</script>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			views := normalizeHTML([]byte(tc.source))
			got, unsupported := normalizeJS(views.scripts)
			if unsupported || string(got) != tc.want {
				t.Fatalf("got %q, unsupported=%t; want %q", got, unsupported, tc.want)
			}
		})
	}
}

func TestDeanEdwardsDecodedSignature(t *testing.T) {
	const source = `<html><script>eval(function(p,a,c,k,e,d){return p}('0("1");',2,2,'alert|hi'.split('|'),0,{}))</script></html>`
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Decoded.Packer:3:*:" + hx(`alert("hi")`) + "\n",
	}
	expectScan(t, embeddedState(t, rules), []byte(source), "Decoded.Packer", "")
	control := `<html><script>eval(function(p,a,c,k,e,d){return p}('0("1");',2,2,'warn|hi'.split('|'),0,{}))</script></html>`
	expectScan(t, embeddedState(t, rules), []byte(control), "", "")
}

func TestDeanEdwardsRepeatedCallsAreComplete(t *testing.T) {
	source := []byte(`<html><script>function pack(p,a,c,k,e,r){return p} pack('0("1");',2,2,'alert|hi'.split('|'),0,{}) pack('0("1");',2,2,'alert|hi'.split('|'),0,{})</script></html>`)
	rules := map[string]string{
		"a.ftm": embeddedTypes + "0:0:3c68746d6c:HTML:CL_TYPE_ANY:CL_TYPE_HTML\n",
		"a.ndb": "Other:3:*:" + hx("unrelated") + "\n",
	}
	expectScan(t, embeddedState(t, rules), source, "", "")
}
