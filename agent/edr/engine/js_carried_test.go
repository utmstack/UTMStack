package engine

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCarriedDecodersNativeReference(t *testing.T) {
	raw, e := os.ReadFile("testdata/carried_decoder_oracle.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Name           string
		Source, Output []byte
	}
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2846 {
		t.Fatalf("native inventory %d", len(rows))
	}
	for _, r := range rows {
		t.Run(r.Name, func(t *testing.T) {
			v := normalizeHTML(r.Source)
			got, unknown := normalizeJS(v.scripts)
			if unknown || string(got) != string(r.Output) {
				t.Fatalf("Go=%q unknown=%t native=%q", got, unknown, r.Output)
			}
		})
	}
}
