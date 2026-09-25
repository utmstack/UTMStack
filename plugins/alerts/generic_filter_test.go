package main

// Generic input (dataType generic) filter regression.
//
// The engine reads every pipeline definition through go-sdk plugins.GetCfg
// (EventProcessor pkg/parsing/parsing.go), which turns the YAML into JSON and
// decodes it with protojson DiscardUnknown (go-sdk v1.1.36 plugins/config.go).
// A key the SDK does not define, such as field_name on a grok pattern, is
// dropped without any error. The grok step then writes nothing and the json
// step fails on every event. This test loads filters/generic/generic.yml
// through that same SDK loader, in a child process because GetCfg keeps
// process-wide state, and evaluates the json step's where clause with the SDK
// CEL cache the engine uses for step conditions. Every event is fabricated.
// Raw extraction is covered by the EventProcessor playground runs recorded in
// filters/audits/generic.md.
import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	genericFilterPath   = "../../filters/generic/generic.yml"
	genericLoaderChild  = "UTM_GENERIC_LOADER_CHILD"
	genericLoaderOutput = "UTM_GENERIC_LOADER_OUTPUT"
)

func TestGenericFilter(t *testing.T) {
	if os.Getenv(genericLoaderChild) == "1" {
		genericLoaderChildRun(t)
		return
	}

	t.Run("strict decode finds no unknown key", func(t *testing.T) {
		b, err := utils.ReadPbYaml(genericFilterPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := protojson.Unmarshal(b, new(plugins.Config)); err != nil {
			t.Fatalf("the engine loader drops this key without an error: %v", err)
		}
	})

	pipeline := genericLoadWithSDK(t)
	if len(pipeline) != 1 || len(pipeline[0].DataTypes) != 1 || pipeline[0].DataTypes[0] != "generic" {
		t.Fatalf("want one pipeline for dataType generic, got %v", pipeline)
	}
	steps := pipeline[0].Steps
	if len(steps) != 2 || steps[0].Grok == nil || steps[1].Json == nil {
		t.Fatalf("want a grok step then a json step, got %v", steps)
	}

	t.Run("SDK loader keeps the grok target", func(t *testing.T) {
		grok := steps[0].Grok
		if grok.Source != "raw" || len(grok.Patterns) != 1 {
			t.Fatalf("grok step: %v", grok)
		}
		if p := grok.Patterns[0]; p.FieldName != "log.message" || p.Pattern != "(.*)" {
			t.Fatalf("grok pattern after the SDK loader: fieldName %q, pattern %q; want log.message and (.*)",
				p.FieldName, p.Pattern)
		}
	})

	t.Run("json step runs only on a JSON object", func(t *testing.T) {
		step := steps[1].Json
		if step.Source != "log.message" {
			t.Fatalf("json source %q, want log.message", step.Source)
		}
		if step.Where == "" {
			t.Fatal("json step has no where clause, so every text line is stored with a parse error")
		}
		cache := plugins.NewCELCache("generic-filter-test")
		// Each draft is the event as the json step sees it: grok has copied the
		// trimmed first line of raw into log.message. An empty message means grok
		// wrote nothing, as it does for blank raw text.
		for _, tc := range []struct {
			name, raw, message string
			want               bool
		}{
			{"JSON object", `{"user":"bob","src_ip":"192.0.2.20"}`, `{"user":"bob","src_ip":"192.0.2.20"}`, true},
			{"JSON object after spaces", `   {"user":"erin"}`, `{"user":"erin"}`, true},
			{"text starting with a brace", `{not json} login from 192.0.2.26`, `{not json} login from 192.0.2.26`, true},
			{"syslog text", `<34>Oct 11 22:14:15 host1 sshd[123]: Accepted password for alice`,
				`<34>Oct 11 22:14:15 host1 sshd[123]: Accepted password for alice`, false},
			{"syslog header before JSON", `<14>Sep 24 12:00:00 host3 app: {"user":"gina"}`,
				`<14>Sep 24 12:00:00 host3 app: {"user":"gina"}`, false},
			{"key=value text", `user=hank src=192.0.2.23 action=login`, `user=hank src=192.0.2.23 action=login`, false},
			{"JSON array", `[{"a":1},{"b":2}]`, `[{"a":1},{"b":2}]`, false},
			{"JSON number", `42`, `42`, false},
			{"missing log.message", `   `, ``, false},
		} {
			draft := map[string]any{"id": "fabricated-" + tc.name, "dataType": "generic",
				"dataSource": "fixture-generic", "raw": tc.raw}
			if tc.message != "" {
				draft["log"] = map[string]any{"message": tc.message}
			}
			b, err := json.Marshal(draft)
			if err != nil {
				t.Fatal(err)
			}
			text := string(b)
			got, err := cache.Evaluate(&text, step.Where)
			if err != nil {
				t.Errorf("%s: where %q: %v", tc.name, step.Where, err)
				continue
			}
			if got != tc.want {
				t.Errorf("%s: where %q gave %v, want %v", tc.name, step.Where, got, tc.want)
			}
		}
	})
}

// genericLoadWithSDK stages the filter where the config plugin writes it,
// <WORK_DIR>/pipeline/filters/<id>.yaml, and loads it with plugins.GetCfg in a
// child process whose WORK_DIR points at a private temporary folder.
func genericLoadWithSDK(t *testing.T) []*plugins.Pipeline {
	t.Helper()
	b, err := os.ReadFile(genericFilterPath)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	dir := filepath.Join(work, "pipeline", "filters")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "generic.yaml"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(work, "loaded.json")
	command := exec.Command(os.Args[0], "-test.run=^TestGenericFilter$")
	command.Env = append(os.Environ(), genericLoaderChild+"=1", "WORK_DIR="+work, genericLoaderOutput+"="+out)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("SDK loader child: %v\n%s", err, output)
	}
	loaded, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(loaded, cfg); err != nil {
		t.Fatal(err)
	}
	return cfg.Pipeline
}

// genericLoaderChildRun runs only in the child. WORK_DIR was set before the
// process started, so the SDK derived plugins.WorkDir from it as the engine does.
func genericLoaderChildRun(t *testing.T) {
	work, out := os.Getenv("WORK_DIR"), os.Getenv(genericLoaderOutput)
	if work == "" || out == "" || plugins.WorkDir != work {
		t.Fatalf("the child needs WORK_DIR and %s; SDK WorkDir is %q", genericLoaderOutput, plugins.WorkDir)
	}
	cfg := plugins.GetCfg("EventProcessor")
	b, err := protojson.Marshal(&plugins.Config{Pipeline: cfg.Pipeline})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, b, 0o600); err != nil {
		t.Fatal(err)
	}
}
