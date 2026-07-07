// edr/netblock/parse_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func gz(s string) *bytes.Reader {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(s))
	w.Close()
	return bytes.NewReader(b.Bytes())
}

func TestParseAccumulative(t *testing.T) {
	body := "1.2.3.4\n# comment\n\n10.0.0.0/8\n2001:db8::1\n"
	inds, err := ParseAccumulative(gz(body), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(inds) != 3 {
		t.Fatalf("want 3, got %d: %+v", len(inds), inds)
	}
	// CIDR auto-detected from the slash.
	var sawCIDR bool
	for _, i := range inds {
		if i.Type == "cidr" && i.Value == "10.0.0.0/8" {
			sawCIDR = true
		}
	}
	if !sawCIDR {
		t.Fatal("10.0.0.0/8 not parsed as cidr")
	}
}

func TestParseDaily(t *testing.T) {
	body := `{"value":"1.2.3.4","type":"ip","op":"add"}
{"value":"9.9.9.9","type":"ip","op":"del"}
{"garbage"
{"value":"10.0.0.0/8","type":"cidr","op":"add"}`
	ops, err := ParseDaily(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 3 { // malformed line skipped, not fatal
		t.Fatalf("want 3 ops, got %d", len(ops))
	}
	if ops[1].Op != "del" {
		t.Fatalf("op[1] = %+v", ops[1])
	}
}
