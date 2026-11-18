package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWebPFuzzyRuleEndToEnd(t *testing.T) {
	ftm := "0:0:52494646:RIFF:CL_TYPE_ANY:CL_TYPE_RIFF\n"
	graphicsFTM := ftm + "0:8:57454250:WebP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
	for _, tc := range []struct {
		name, fixture, rule, typeRules, wantSig, wantUnknown string
	}{
		{"custom-lossless", "webp-lossless.webp", "95cc150b9649973f", graphicsFTM, "Image.WebP", ""},
		{"custom-alpha", "webp-alpha.webp", "95cc150b9649973f", graphicsFTM, "Image.WebP", ""},
		{"custom-extended", "webp-extended-alpha.webp", "95cc150b9649973f", graphicsFTM, "Image.WebP", ""},
		{"custom-lossy", "webp-lossy.webp", "954c150b9e49973f", graphicsFTM, "Image.WebP", ""},
		{"custom-extended-lossy", "webp-extended-lossy.webp", "954c150b9e49973f", graphicsFTM, "Image.WebP", ""},
		{"custom-alpha-lossy", "webp-alpha-lossy.webp", "95cc150b9649973f", graphicsFTM, "Image.WebP", ""},
		{"official-RIFF", "webp-lossless.webp", "95cc150b9649973f", ftm, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "signatures", "testdata", tc.fixture))
			if err != nil {
				t.Fatal(err)
			}
			s := pipelineState(t, map[string]string{
				"probe.ftm": tc.typeRules,
				"probe.ldb": "Image.WebP;Engine:150-255,Target:0;0;fuzzy_img#" + tc.rule + "\n",
			})
			expectScan(t, s, data, tc.wantSig, tc.wantUnknown)
		})
	}
}
