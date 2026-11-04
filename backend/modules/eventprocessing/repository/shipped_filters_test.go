package repository

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const definitionsFiltersDir = "../../../../definitions/filters"

// The values the engine understands for actionResult. The threat-intelligence
// analysis skips indicator lookups only for denied and failed (and blocked,
// which it reads as denied), so a filter that wrote "failure" would have its
// failed sign-ins looked up and alerted on as malicious.
var actionResults = map[string]bool{"success": true, "failed": true, "denied": true}

func TestEveryActionResultAFilterWritesIsOneTheEngineRecognizes(t *testing.T) {
	checked := 0
	err := filepath.WalkDir(definitionsFiltersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || (filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var doc any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Errorf("%s: %v", path, err)
			return nil
		}
		walkAddSteps(doc, func(key, value string) {
			checked++
			if key == "actionResult" && !actionResults[value] {
				t.Errorf("%s writes actionResult %q: use success, failed or denied", path, value)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatalf("no add steps found under %s: is the path right?", definitionsFiltersDir)
	}
}

// walkAddSteps calls fn for every {key, value} pair under a params map, which is
// how an add step names the field it writes.
func walkAddSteps(n any, fn func(key, value string)) {
	switch v := n.(type) {
	case map[string]any:
		if p, ok := v["params"].(map[string]any); ok {
			k, kok := p["key"].(string)
			val, vok := p["value"].(string)
			if kok && vok {
				fn(strings.TrimSpace(k), strings.TrimSpace(val))
			}
		}
		for _, c := range v {
			walkAddSteps(c, fn)
		}
	case []any:
		for _, c := range v {
			walkAddSteps(c, fn)
		}
	}
}
