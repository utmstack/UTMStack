package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestTIFFPlanesFuzzyRuleEndToEnd(t *testing.T) {
	root := filepath.Join("..", "signatures", "testdata", "tiff_planes")
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
	for _, name := range []string{"planes-0-p1-r0-s0.tiff", "planes-1-p0-r2-s1.tiff", "planes-2-p1-r0-s0.tiff", "planes-3-p1-r0-s1.tiff", "planes-6-p0-r2-s0.tiff", "planes-7-p1-r0-s0.tiff", "planes-8-p1-r0-s0.tiff", "planes-9-p0-r2-s0.tiff", "planes-11-p0-r0-s1-rgba.tiff", "planes-12-p1-r0-s0.tiff", "planes-15-p1-r0-s1.tiff", "planes-17-p0-r2-s0.tiff"} {
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
	for _, name := range []string{"planes-4-p0-r0-s0.tiff", "planes-6-p1-r0-s0.tiff", "planes-9-p1-r0-s0.tiff", "planes-17-p1-r2-s1.tiff"} {
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
