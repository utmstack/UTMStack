package main

// Offline AWS extraction model, not the closed EventProcessor.
// Explicit YAML JSON/key sanitization, grok, rename, add and delete steps are modeled.
// CEL and Event serialization use SDK v1.1.31. History requests are tested separately
// with that SDK. External geolocation is mocked only when a fixture declares it.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type awsFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
	Enrichment map[string]any `json:"enrichment"`
}

func awsPut(m map[string]any, path string, value any, remove bool) {
	p := strings.Split(path, ".")
	for _, k := range p[:len(p)-1] {
		n, ok := m[k].(map[string]any)
		if !ok {
			if remove {
				return
			}
			n = map[string]any{}
			m[k] = n
		}
		m = n
	}
	if remove {
		delete(m, p[len(p)-1])
	} else {
		m[p[len(p)-1]] = value
	}
}
func awsGet(m map[string]any, p string) (any, bool) {
	var v any = m
	for _, k := range strings.Split(p, ".") {
		n, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok = n[k]
		if !ok {
			return nil, false
		}
	}
	return v, true
}
func awsConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/aws/aws.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func awsRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	var pattern strings.Builder
	for i, p := range g.Patterns {
		if p.FieldName != "" {
			fmt.Fprintf(&pattern, "(?P<f%d>%s)", i, p.Pattern)
		} else {
			pattern.WriteString("(?:" + p.Pattern + ")")
		}
	}
	pats := map[string]string{"greedy": ".*", "data": ".*?", "word": "[A-Za-z0-9_-]+", "space": "\\s+"}
	for k, v := range cfg.Patterns {
		pats[k] = v
	}
	tmpl, e := template.New("grok").Option("missingkey=error").Parse(pattern.String())
	if e != nil {
		t.Fatal(e)
	}
	var b bytes.Buffer
	if e = tmpl.Execute(&b, pats); e != nil {
		t.Fatal(e)
	}
	r, e := regexp.Compile(b.String())
	if e != nil {
		t.Fatal(e)
	}
	return r
}

// awsStoredName is the name the parser plugins store for a grok, rename or add target:
// utils.SanitizeField keeps only letters, digits and dots.
func awsStoredName(name string) string {
	utils.SanitizeField(&name)
	return name
}

func awsParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache, enrichment ...map[string]any) string {
	return awsParseMode(t, cfg, raw, dataSource, cache, false, enrichment...)
}

// Both modes model the unresolved nested-key behavior of the closed JSON step.
func awsParseMode(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache, preserveNested bool, enrichment ...map[string]any) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "aws", "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "aws" {
				matched = true
			}
		}
		if !matched {
			continue
		}
		for _, s := range stage.Steps {
			b, e := protojson.Marshal(s)
			if e != nil {
				t.Fatal(e)
			}
			var step map[string]map[string]any
			if e = json.Unmarshal(b, &step); e != nil {
				t.Fatal(e)
			}
			for kind, body := range step {
				if w, ok := body["where"].(string); ok && w != "" {
					snapshot, err := json.Marshal(draft)
					if err != nil {
						t.Fatal(err)
					}
					match, e := cache.Eval(w, string(snapshot))
					if e != nil {
						t.Fatal(e)
					}
					if !match {
						continue
					}
				}
				switch kind {
				case "grok":
					g := s.Grok
					src := g.Source
					if src == "" {
						src = "raw"
					}
					v, ok := awsGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := awsRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							awsPut(draft, awsStoredName(p.FieldName), m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := awsGet(draft, p); ok {
							awsPut(draft, awsStoredName(s.Rename.To), v, false)
							awsPut(draft, p, nil, true)
							break
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					awsPut(draft, awsStoredName(s.Add.Params["key"].GetStringValue()), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						awsPut(draft, p, nil, true)
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := awsGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed. Fixtures may explicitly supply its mocked output.
					for _, fields := range enrichment {
						for path, value := range fields {
							if !strings.HasPrefix(path, "origin.geolocation.") {
								t.Fatal("unexpected enrichment field")
							}
							awsPut(draft, path, value, false)
						}
					}
				case "json":
					source, ok := awsGet(draft, s.Json.Source)
					if !ok {
						continue
					}
					str, ok := source.(string)
					if !ok {
						t.Fatalf("JSON source is not a string")
					}
					var parsed map[string]any
					if e := json.Unmarshal([]byte(str), &parsed); e != nil {
						t.Fatal(e)
					}
					normalized := awsSanitizeJSON(parsed)
					if preserveNested {
						normalized = map[string]any{}
						for key, value := range parsed {
							utils.SanitizeField(&key)
							normalized[key] = value
						}
					}
					for key, value := range normalized {
						awsPut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := awsGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								awsPut(draft, field, utils.CastString(value), false)
							case "int":
								awsPut(draft, field, utils.CastInt64(value), false)
							default:
								t.Fatalf("unsupported cast %s", s.Cast.To)
							}
						}
					}
				case "drop":
					return ""
				default:
					t.Fatalf("unsupported filter step %s", kind)
				}
			}
		}
	}
	b, e := json.Marshal(draft)
	if e != nil {
		t.Fatal(e)
	}
	in := string(b)
	ev := new(plugins.Event)
	if e = utils.StringToProtoMessage(&in, ev); e != nil {
		t.Fatal(e)
	}
	out, e := utils.ProtoMessageToString(ev)
	if e != nil {
		t.Fatal(e)
	}
	return *out
}
func awsRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths := []string{}
	e := filepath.WalkDir("../../rules/cloud/aws", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			paths = append(paths, path)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}

	out := map[string]*plugins.Rule{}
	for _, p := range paths {
		b, e := utils.ReadPbYaml(p)
		if e != nil {
			t.Fatal(e)
		}
		r := new(plugins.Rule)
		if e = protojson.Unmarshal(b, r); e != nil {
			t.Fatal(e)
		}
		r.Normalize()
		out[strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))] = r
	}
	return out
}

func awsSanitizeJSON(input map[string]any) map[string]any {
	var walk func(any) any
	walk = func(input any) any {
		switch v := input.(type) {
		case map[string]any:
			out := map[string]any{}
			for key, value := range v {
				utils.SanitizeField(&key)
				out[key] = walk(value)
			}
			return out
		case []any:
			out := make([]any, len(v))
			for i, x := range v {
				out[i] = walk(x)
			}
			return out
		default:
			return input
		}
	}
	return walk(input).(map[string]any)
}

func awsFixtures(t *testing.T) []awsFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/aws_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []awsFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func TestAWSRawContracts(t *testing.T) {
	cfg, rules, cache := awsConfig(t), awsRules(t), plugins.NewCELCache("aws-raw")
	positive := map[string]int{}
	negative := map[string]int{}
	if len(rules) != 82 {
		t.Fatalf("rules: %d", len(rules))
	}
	for _, f := range awsFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			out := awsParse(t, cfg, f.Raw, f.DataSource, cache, f.Enrichment)
			for field, want := range f.Expected {
				got := gjson.Get(out, field)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s got %v want %v", field, got.Value(), want)
				}
			}
			for _, field := range f.Absent {
				if gjson.Get(out, field).Exists() {
					t.Errorf("unexpected %s", field)
				}
			}
			if gjson.Get(out, "raw").String() != f.Raw {
				t.Error("raw altered")
			}
			expected := map[string]bool{}
			for _, n := range f.Matches {
				expected[n] = true
			}
			for name, r := range rules {
				yes, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Fatalf("%s: %v", name, e)
				}
				if yes != expected[name] {
					t.Errorf("%s matched %v want %v", name, yes, expected[name])
				}
				if yes {
					positive[name]++
				} else {
					negative[name]++
				}
				if yes {
					for _, search := range r.Correlation {
						for _, term := range search.With {
							value := term.Value.GetStringValue()
							if strings.HasPrefix(value, "{{.") {
								field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
								if !gjson.Get(out, field).Exists() {
									t.Errorf("%s unresolved %s", name, field)
								}
							}
						}
					}
				}
				if yes {
					ev := new(plugins.Event)
					if e := utils.StringToProtoMessage(&out, ev); e != nil {
						t.Fatal(e)
					}
					if r.Adversary != "origin" {
						t.Errorf("unexpected actor direction %s", r.Adversary)
					}
					alert := &plugins.Alert{Adversary: ev.Origin, Target: ev.Target, Events: []*plugins.Event{ev}}
					wire, e := utils.ProtoMessageToString(alert)
					if e != nil {
						t.Fatal(e)
					}
					if gjson.Get(out, "target.ip").String() != gjson.Get(*wire, "target.ip").String() {
						t.Error("endpoint identity lost")
					}
					if gjson.Get(out, "origin.ip").String() != gjson.Get(*wire, "adversary.ip").String() {
						t.Error("attacker identity lost")
					}
				}
			}
		})
	}
	for name := range rules {
		if positive[name] == 0 || negative[name] == 0 {
			t.Errorf("%s missing positive/negative coverage: %d/%d", name, positive[name], negative[name])
		}
	}
}
func TestAWSOfficialExamples(t *testing.T) {
	input := os.Getenv("AWS_OFFICIAL_EXAMPLES")
	if input == "" {
		t.Skip("official documentation snapshots supplied separately")
	}
	b, e := os.ReadFile(input)
	if e != nil {
		t.Fatal(e)
	}
	var docs []struct {
		Reference string         `json:"reference"`
		Event     map[string]any `json:"event"`
		Expected  map[string]any `json:"expected"`
		Absent    []string       `json:"absent"`
		Matches   []string       `json:"matches"`
	}
	if e = json.Unmarshal(b, &docs); e != nil {
		t.Fatal(e)
	}
	cfg, rules, cache := awsConfig(t), awsRules(t), plugins.NewCELCache("aws-official")
	results := []map[string]any{}
	for _, d := range docs {
		raw, e := json.Marshal(d.Event)
		if e != nil {
			t.Fatal(e)
		}
		parsed := awsParse(t, cfg, string(raw), "collector-test", cache)
		if len(d.Expected) == 0 || d.Matches == nil {
			t.Fatal("official sample lacks explicit expectations")
		}
		for path, want := range d.Expected {
			if got := gjson.Get(parsed, path); !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
				t.Errorf("%s %s: %s got %v want %v", d.Reference, d.Event["eventName"], path, got.Value(), want)
			}
		}
		for _, path := range d.Absent {
			if gjson.Get(parsed, path).Exists() {
				t.Errorf("unexpected %s", path)
			}
		}
		expectedMatches := map[string]bool{}
		for _, name := range d.Matches {
			expectedMatches[name] = true
		}
		matches := []string{}
		errors := map[string]string{}
		for name, r := range rules {
			yes, err := cache.Eval(r.Where, parsed)
			if err != nil {
				errors[name] = err.Error()
				t.Errorf("official predicate %s: %v", name, err)
			} else if yes {
				matches = append(matches, name)
			}
		}
		for name, r := range rules {
			yes, e := cache.Eval(r.Where, parsed)
			if e != nil || yes != expectedMatches[name] {
				t.Errorf("official %s %s matched %v want %v: %v", d.Event["eventName"], name, yes, expectedMatches[name], e)
			}
		}
		var value map[string]any
		if e = json.Unmarshal([]byte(parsed), &value); e != nil {
			t.Fatal(e)
		}
		results = append(results, map[string]any{"reference": d.Reference, "parsed": value, "matches": matches, "errors": errors})
	}
	if path := os.Getenv("AWS_OFFICIAL_OUTPUT"); path != "" {
		b, e := json.MarshalIndent(results, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("modeled %d official examples; no live AWS telemetry or alerts", len(docs))
}

// Stored Azure Event Grid records retain punctuation in nested claim keys. That is
// not proof of the AWS parser's behavior, so header consumers accept both layouts.
func TestAWSNestedKeyCompatibility(t *testing.T) {
	cfg, rules, cache := awsConfig(t), awsRules(t), plugins.NewCELCache("aws-nested-keys")
	raw := `{"eventVersion":"1.09","eventTime":"2026-09-17T12:00:00Z","eventType":"AwsApiCall","eventName":"PutBucketAcl","eventSource":"s3.amazonaws.com","requestParameters":{"x-amz-acl":"public-read","x-amz-server-side-encryption":"AES256"},"responseElements":{"x-amz-expiration":"expiry-test","x-amz-server-side-encryption":"AES256"},"additionalEventData":{"x-amz-id-2":"request-test"}}`
	for _, preserved := range []bool{false, true} {
		t.Run(fmt.Sprint(preserved), func(t *testing.T) {
			out := awsParseMode(t, cfg, raw, "collector-test", cache, preserved)
			expected := map[string]string{"log.requestParametersXAmzAcl": "public-read", "log.requestParametersXAmzServerSideEncryption": "AES256", "log.responseElementsXAmzExpiration": "expiry-test", "log.responseElementsXAmzServerSideEncryption": "AES256", "log.additionalEventDataXamzId2": "request-test"}
			for field, want := range expected {
				if got := gjson.Get(out, field).String(); got != want {
					t.Errorf("%s=%q want %q", field, got, want)
				}
			}
			yes, e := cache.Eval(rules["s3_bucket_public_exposure"].Where, out)
			if e != nil || !yes {
				t.Fatalf("public ACL lost: %v %v", yes, e)
			}
			key := "log.requestParameters.xamzacl"
			if preserved {
				key = "log.requestParameters.x-amz-acl"
			}
			if gjson.Get(out, key).String() != "public-read" {
				t.Error("original header lost")
			}
			var private map[string]any
			if err := json.Unmarshal([]byte(raw), &private); err != nil {
				t.Fatal(err)
			}
			private["requestParameters"].(map[string]any)["x-amz-acl"] = "private"
			private["requestParametersXAmzAcl"] = "public-read"
			encoded, _ := json.Marshal(private)
			negative := awsParseMode(t, cfg, string(encoded), "collector-test", cache, preserved)
			if yes, e := cache.Eval(rules["s3_bucket_public_exposure"].Where, negative); e != nil || yes {
				t.Fatalf("flat input alias fabricates ACL exposure: %v %v", yes, e)
			}

		})
	}
}
