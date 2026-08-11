package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// catalog.json is hand-written and served to clients as
// mcp://utmstack/docs/catalog. It is a design document, not a mirror: by its
// own definition, status="implemented" means wired in tools_*.go today and
// status="planned" means scaffolded but not registered. A planned entry is
// therefore fine; what is not fine is the two disagreeing about what is live.
//
// Nothing else notices: the catalog is embedded at build time, so it compiles
// whatever it claims. An AI client reads it to decide what to call.
func TestEveryRegisteredToolIsDocumentedAsImplemented(t *testing.T) {
	registered := registeredToolNames(t)
	status := catalogToolStatus(t)

	var undocumented, mislabelled []string
	for _, name := range registered {
		s, ok := status[name]
		switch {
		case !ok:
			undocumented = append(undocumented, name)
		case s != "implemented":
			mislabelled = append(mislabelled, name+" (marked "+s+")")
		}
	}

	if len(undocumented) > 0 {
		t.Errorf("registered but absent from catalog.json — clients will never know these exist:\n  %s",
			strings.Join(undocumented, "\n  "))
	}
	if len(mislabelled) > 0 {
		t.Errorf("registered but the catalog says they are not — a client reads this as unavailable:\n  %s",
			strings.Join(mislabelled, "\n  "))
	}
}

// The other direction: a tool the catalog swears is live must answer. A planned
// one is a design note and may legitimately have no code behind it.
func TestNothingClaimedImplementedIsMissing(t *testing.T) {
	registered := registeredToolNames(t)
	inCode := make(map[string]bool, len(registered))
	for _, n := range registered {
		inCode[n] = true
	}

	var missing []string
	for name, s := range catalogToolStatus(t) {
		if s == "implemented" && !inCode[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("catalog claims these are implemented but nothing registers them:\n  %s",
			strings.Join(missing, "\n  "))
	}
}

// registeredToolNames reads the names out of the source rather than standing a
// server up: registration needs every module's dependencies, and this only has
// to know which names exist.
func registeredToolNames(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob("tools_*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no tool files found: %v", err)
	}

	re := regexp.MustCompile(`Name:\s+"([a-z_]+\.[a-z_.]+)"`)
	seen := map[string]bool{}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, m := range re.FindAllStringSubmatch(string(src), -1) {
			seen[m[1]] = true
		}
	}
	return sorted(seen)
}

func catalogToolStatus(t *testing.T) map[string]string {
	t.Helper()

	data, err := os.ReadFile("catalog.json")
	if err != nil {
		t.Fatalf("read catalog.json: %v", err)
	}
	var doc struct {
		Tools []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse catalog.json: %v", err)
	}

	out := make(map[string]string, len(doc.Tools))
	for _, tool := range doc.Tools {
		if tool.Name != "" {
			out[tool.Name] = tool.Status
		}
	}
	return out
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
