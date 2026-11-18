package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestTIFFCMYKFuzzyRuleEndToEnd(t *testing.T) {
	root := filepath.Join("..", "signatures", "testdata", "tiff_cmyk")
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
	for _, depth := range []int{8, 16} {
		for _, codec := range []int{1, 5, 8, 32773} {
			name := fmt.Sprintf("cmyk-%d-b-tiles-%d-p1-s0.tiff", depth, codec)
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
						files["probe.ftm"] = "0:0:4d4d002a:TIFF:CL_TYPE_ANY:CL_TYPE_TIFF\n"
					}
					state := pipelineState(t, files)
					expectScan(t, state, data, "Image.TIFF", "")
				}
				wrong := pipelineState(t, map[string]string{"probe.ldb": "Image.TIFF;Engine:150-255,Target:0;0;fuzzy_img#0000000000000000\n"})
				expectScan(t, wrong, data, "", "")
			})
		}
	}
	for _, name := range []string{"palette-8.tiff", "planar-separated.tiff", "pixel-offset-into-shadow.tiff"} {
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
