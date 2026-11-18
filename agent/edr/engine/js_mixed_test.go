package engine

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMixedDecodersNativeReference(t *testing.T) {
	raw, err := os.ReadFile("testdata/mixed_decoder_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Name, Source, Output string }
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 363 {
		t.Fatalf("native control inventory %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			v := normalizeHTML([]byte(tc.Source))
			got, unknown := normalizeJS(v.scripts)
			if unknown || string(got) != tc.Output {
				t.Fatalf("Go=%q unknown=%t native=%q", got, unknown, tc.Output)
			}
		})
	}
}
