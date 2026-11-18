package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestTIFFJPEGFuzzyRuleEndToEnd(t *testing.T) {
	root := filepath.Join("..", "signatures", "testdata", "tiff_jpeg")
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
	for _, name := range []string{"jpeg-1-b-strips-t3-q31-s0.tiff", "jpeg-2-l-tiles-t0-q90-s0.tiff", "jpeg-5-b-tiles-t1-q90-s1.tiff", "pillow-RGB-p1-sampling-2.tiff", "cmyk-as-rgba-tiles-s0.tiff", "ycbcr-as-rgb-strips-s0.tiff", "reinterpret-32-p3.tiff", "body-prefix-ignored.tiff", "body-length-642.tiff", "table-tail-ignored.tiff"} {
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
	for _, name := range []string{"table-mode-1-removed.tiff", "table-short-1.tiff", "jpeg-6-l-tiles-t0-q31-s0.tiff", "float8-format.tiff"} {
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
