// Command rulecheck reports which correlation rules the engine can actually
// load, using the same path the cel plugin uses: YAML -> JSON -> protojson ->
// Rule, then the CEL expression.
//
// It exists because every way a rule can be wrong is silent. A rule that fails
// to parse, fails to compile, or names a field that does not exist is logged
// once at startup and then never fires — there is nothing at runtime that says
// a detection is missing.
//
//	go run ./tools/rulecheck definitions/rules
//
// Exits non-zero when a rule fails to load, so it can gate a build.
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

type finding struct {
	file   string
	stage  string
	detail string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: rulecheck <rules-dir>")
		os.Exit(2)
	}
	root := os.Args[1]

	var files []string
	if err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml") {
			files = append(files, p)
		}
		return nil
	}); err != nil {
		fmt.Fprintln(os.Stderr, "walking", root+":", err)
		os.Exit(2)
	}

	cache := plugins.NewCELCache("rulecheck")

	var failures, warnings []finding
	loaded := 0

	for _, f := range files {
		rel, _ := filepath.Rel(root, f)

		data, err := os.ReadFile(f)
		if err != nil {
			failures = append(failures, finding{rel, "read", err.Error()})
			continue
		}

		var generic any
		if err := yaml.Unmarshal(data, &generic); err != nil {
			failures = append(failures, finding{rel, "yaml", oneLine(err.Error())})
			continue
		}

		jsonBytes, err := json.Marshal(generic)
		if err != nil {
			failures = append(failures, finding{rel, "json", oneLine(err.Error())})
			continue
		}
		jsonStr := string(jsonBytes)

		var rule plugins.Rule
		if err := utils.StringToProtoMessage(&jsonStr, &rule); err != nil {
			failures = append(failures, finding{rel, "proto", oneLine(err.Error())})
			continue
		}

		// The engine parses with DiscardUnknown, so a misspelled key is not an
		// error there — the rule loads and quietly does less than it says. A
		// second, strict pass is the only way to see them.
		if err := (protojson.UnmarshalOptions{}).Unmarshal([]byte(jsonStr), &plugins.Rule{}); err != nil {
			warnings = append(warnings, finding{rel, "unknown-field", oneLine(err.Error())})
		}

		// An empty event: a rule that only fails because a field is absent has
		// still compiled, which is what is being counted.
		if _, err := cache.Eval(rule.Where, &plugins.Event{}); err != nil {
			if strings.Contains(err.Error(), "compile") {
				failures = append(failures, finding{rel, "cel", oneLine(err.Error())})
				continue
			}
		}

		loaded++
	}

	fmt.Printf("rules:  %d\n", len(files))
	fmt.Printf("loaded: %d\n", loaded)
	fmt.Printf("failed: %d\n", len(failures))
	if len(warnings) > 0 {
		fmt.Printf("warned: %d\n", len(warnings))
	}

	report("FAILED", failures)
	report("WARNING", warnings)

	if len(failures) > 0 {
		os.Exit(1)
	}
}

func report(label string, items []finding) {
	if len(items) == 0 {
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].file < items[j].file })
	fmt.Printf("\n%s\n", label)
	for _, it := range items {
		fmt.Printf("  [%s] %s\n      %s\n", it.stage, it.file, it.detail)
	}
}

func oneLine(s string) string {
	if i := strings.Index(s, "\n"); i > 0 {
		s = s[:i]
	}
	if len(s) > 160 {
		s = s[:160] + "..."
	}
	return s
}
