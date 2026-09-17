package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func contractPaths(d protoreflect.MessageDescriptor, prefix string, out map[string]bool) {
	for i := 0; i < d.Fields().Len(); i++ {
		f := d.Fields().Get(i)
		p := prefix + f.JSONName()
		out[p] = true
		if f.Message() != nil && !f.IsMap() && !strings.HasPrefix(string(f.Message().FullName()), "google.protobuf.") {
			contractPaths(f.Message(), p+".", out)
		}
	}
}
func TestFilterAndRuleContracts(t *testing.T) {
	eventPaths := map[string]bool{}
	alertPaths := map[string]bool{}
	contractPaths(new(plugins.Event).ProtoReflect().Descriptor(), "", eventPaths)
	contractPaths(new(plugins.Alert).ProtoReflect().Descriptor(), "", alertPaths)
	arrayIndex := regexp.MustCompile(`\.[0-9]+(\.|$)`)
	eventPath := func(p string) bool {
		p = strings.TrimSuffix(p, ".keyword")
		p = arrayIndex.ReplaceAllString(p, "$1")
		return eventPaths[p] || strings.HasPrefix(p, "log.") || strings.HasPrefix(p, "compliance.")
	}
	alertPath := func(p string) bool {
		p = strings.TrimSuffix(p, ".keyword")
		if strings.HasPrefix(p, "lastEvent.") {
			return eventPath(strings.TrimPrefix(p, "lastEvent."))
		}
		p = arrayIndex.ReplaceAllString(p, "$1")
		return alertPaths[p]
	}
	cache := plugins.NewCELCache("filter-rule-contract-test")
	sample := `{"log":{"messageId":0,"severity":0},"origin":{},"target":{},"action":"","actionResult":"","protocol":"","severity":"","connectionStatus":"","raw":"","dataType":"","dataSource":"","deviceTime":"","tenantId":"","tenantName":"","statusCode":0}`
	var expressions func(*testing.T, any)
	expressions = func(t *testing.T, v any) {
		switch n := v.(type) {
		case map[string]any:
			for k, x := range n {
				if k == "where" {
					if w, ok := x.(string); ok && w != "" {
						_, err := cache.Eval(w, sample)
						// Direct selectors can fail on this empty sample after a successful compile.
						// Their presence is not a syntax error or evidence that parsed logs fail.
						if err != nil && !strings.Contains(err.Error(), "failed to evaluate program") {
							t.Errorf("CEL compilation: %v", err)
						}
					}
				} else {
					expressions(t, x)
				}
			}
		case []any:
			for _, x := range n {
				expressions(t, x)
			}
		}
	}
	for _, dir := range []string{"filters", "rules"} {
		err := filepath.WalkDir(filepath.Join("../..", dir), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || (filepath.Ext(path) != ".yml" && filepath.Ext(path) != ".yaml") {
				return nil
			}
			t.Run(path, func(t *testing.T) {
				b, err := utils.ReadPbYaml(path)
				if err != nil {
					t.Fatal(err)
				}
				var doc any
				if err = json.Unmarshal(b, &doc); err != nil {
					t.Fatal(err)
				}
				expressions(t, doc)
				if dir == "filters" {
					cfg := new(plugins.Config)
					if err = protojson.Unmarshal(b, cfg); err != nil {
						t.Fatal(err)
					}
					for _, stage := range cfg.Pipeline {
						for _, step := range stage.Steps {
							fields := []string{}
							if s := step.Rename; s != nil {
								fields = append(fields, s.To)
							}
							if s := step.Grok; s != nil {
								for _, p := range s.Patterns {
									if p.FieldName != "" { // Empty grok names are non-capturing separators.
										fields = append(fields, p.FieldName)
									}
								}
							}
							if s := step.Csv; s != nil {
								fields = append(fields, s.Headers...)
							}
							if s := step.Add; s != nil {
								fields = append(fields, s.Params["key"].GetStringValue())
							}
							if s := step.Cast; s != nil {
								fields = append(fields, s.Fields...)
							}
							for _, p := range fields {
								if !eventPath(p) {
									t.Errorf("unknown event write/cast field %q", p)
								}
							}
						}
					}
				} else {
					rule := new(plugins.Rule)
					if err = protojson.Unmarshal(b, rule); err != nil {
						t.Fatal(err)
					}
					rule.Normalize()
					if len(rule.GroupBy) > 0 && len(rule.DeduplicateBy) > 0 {
						t.Error("groupBy and deduplicateBy are mutually exclusive")
					}
					if rule.Adversary != "" && rule.Adversary != "origin" && rule.Adversary != "target" {
						t.Errorf("unknown adversary %s", rule.Adversary)
					}
					for _, p := range append(rule.GroupBy, rule.DeduplicateBy...) {
						if !alertPath(p) {
							t.Errorf("unknown alert grouping field %q", p)
						}
					}
					var checkSearch func([]*plugins.SearchRequest)
					checkSearch = func(searches []*plugins.SearchRequest) {
						for _, s := range searches {
							if s.Within != "" {
								if _, err := time.ParseDuration(s.Within); err != nil {
									t.Error(err)
								}
							}
							for _, x := range s.With {
								valid := eventPath(x.Field)
								if strings.Contains(s.IndexPattern, "-alert-") {
									valid = alertPath(x.Field)
								}
								if !valid {
									t.Errorf("unknown search field %q", x.Field)
								}
								if x.Value == nil {
									t.Errorf("missing search value for %s", x.Field)
									continue
								}
								v := x.Value.GetStringValue()
								if strings.HasPrefix(v, "{{.") && strings.HasSuffix(v, "}}") {
									p := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
									if !eventPath(p) {
										t.Errorf("unknown event placeholder %q", p)
									}
								}
								switch x.Operator {
								case "filter_term", "filter_match", "must_not_term", "must_not_match":
								default:
									t.Error(fmt.Sprintf("unknown search operator %s", x.Operator))
								}
							}
							checkSearch(s.Or)
						}
					}
					checkSearch(rule.Correlation)
				}
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
