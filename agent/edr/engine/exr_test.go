package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestEXRFuzzyRuleEndToEnd(t *testing.T) {
	root := filepath.Join("..", "exr154", "testdata")
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
	names := []string{"multipart-gray-rgb.exr", "piz-wide-range.exr", "meta-raw-fallback-11.exr"}
	for codec := 0; codec < 8; codec++ {
		names = append(names, fmt.Sprintf("codec-%d-type-0-layout-0.exr", codec))
	}
	for _, name := range names {
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
				ftm := ""
				wanted := ""
				if typed {
					ftm = "0:0:762f3101:EXR:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"
					wanted = "Image.EXR"
				}
				files := map[string]string{"probe.ldb": "Image.EXR;Engine:150-255,Target:0;0;fuzzy_img#" + *value + "\n"}
				if typed {
					files["probe.ftm"] = ftm
				}
				state := pipelineState(t, files)
				expectScan(t, state, data, wanted, "")
			}
			wrong := pipelineState(t, map[string]string{"probe.ftm": "0:0:762f3101:EXR:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "probe.ldb": "Image.EXR;Engine:150-255,Target:0;0;fuzzy_img#0000000000000000\n"})
			expectScan(t, wrong, data, "", "")
		})
	}
	for _, name := range []string{"meta-type-empty.exr", "multipart-deep-rgb.exr"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			state := pipelineState(t, map[string]string{"probe.ftm": "0:0:762f3101:EXR:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n", "probe.ldb": "Image.EXR;Engine:150-255,Target:0;0;fuzzy_img#b9c2953b642b5ab8\n"})
			expectScan(t, state, data, "", signatures.UnknownFuzzyImage)
		})
	}
}
