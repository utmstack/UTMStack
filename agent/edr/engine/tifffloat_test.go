package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestTIFFFloatFuzzyRuleEndToEnd(t *testing.T) {
	root := filepath.Join("..", "signatures", "testdata", "tiff_float")
	var rows []struct {
		Name string
		Hash *string
	}
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]*string)
	for _, row := range rows {
		hashes[row.Name] = row.Hash
	}
	for _, name := range []string{"float-3-l-strips-1-p1-s0.tiff", "float-4-b-tiles-5-p3-s0.tiff", "float-3-b-strips-8-p2-s2.tiff", "float-4-l-tiles-32773-p1-s0.tiff", "format-absent-p2.tiff", "padding-p1-kept-60.tiff", "padding-p3-kept-192.tiff", "deflate-old-tag.tiff"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			value := hashes[name]
			if value == nil {
				t.Fatal("missing positive golden")
			}
			for _, typed := range []bool{false, true} {
				files := map[string]string{"probe.ldb": "Image.TIFF;Engine:150-255,Target:0;0;fuzzy_img#" + *value + "\n"}
				if typed {
					files["probe.ftm"] = "0:0:49492a00:TIFF:CL_TYPE_ANY:CL_TYPE_TIFF\n0:0:4d4d002a:TIFF:CL_TYPE_ANY:CL_TYPE_TIFF\n"
				}
				state := pipelineState(t, files)
				expectScan(t, state, data, "Image.TIFF", "")
			}
			wrong := pipelineState(t, map[string]string{"probe.ldb": "Image.TIFF;Engine:150-255,Target:0;0;fuzzy_img#0000000000000000\n"})
			expectScan(t, wrong, data, "", "")
		})
	}
	for _, name := range []string{"deflate-bad-checksum.tiff", "gray-alpha-16-0.tiff", "float-3-l-strips-50000-p1-s0.tiff", "padding-p3-kept-60.tiff"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			state := pipelineState(t, map[string]string{"probe.ldb": "Image.TIFF;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000\n"})
			expectScan(t, state, data, "", signatures.UnknownFuzzyImage)
		})
	}
}
