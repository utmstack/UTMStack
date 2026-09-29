package main

import (
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// This checks configuration loading against the linked SDK. It does not model
// raw-log extraction, evaluate detections, or claim that a live alert was created.
func TestPaloAltoFilterConfiguration(t *testing.T) {
	data, err := utils.ReadPbYaml("../../filters/paloalto/pa_firewall.yml")
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(data, cfg); err != nil {
		t.Fatalf("filter must decode without dropping unknown SDK properties: %v", err)
	}
	if len(cfg.Pipeline) == 0 {
		t.Fatal("filter has no pipeline")
	}
	eventPaths := map[string]bool{}
	contractPaths(new(plugins.Event).ProtoReflect().Descriptor(), "", eventPaths)
	for stageIndex, stage := range cfg.Pipeline {
		if len(stage.Steps) == 0 {
			t.Errorf("stage %d has no steps", stageIndex)
		}
		for stepIndex, step := range stage.Steps {
			if csv := step.Csv; csv != nil {
				if csv.Source == "" || csv.Separator == "" || len(csv.Headers) == 0 {
					t.Errorf("stage %d step %d: incomplete CSV configuration", stageIndex, stepIndex)
				}
			}
			if grok := step.Grok; grok != nil && len(grok.Patterns) == 0 {
				t.Errorf("stage %d step %d: grok has no patterns", stageIndex, stepIndex)
			}
			if cast := step.Cast; cast != nil {
				for _, field := range cast.Fields {
					if !eventPaths[field] && !strings.HasPrefix(field, "log.") {
						t.Errorf("stage %d step %d: cast targets unknown Event field %q", stageIndex, stepIndex, field)
					}
				}
				// Supported types from the official SDK Filter-Steps-Reference.
				switch cast.To {
				case "int", "float", "string", "bool", "[]string":
				default:
					t.Errorf("stage %d step %d: unsupported documented cast type %q", stageIndex, stepIndex, cast.To)
				}
			}
		}
	}
}
