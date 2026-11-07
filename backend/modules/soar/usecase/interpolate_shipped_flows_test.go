package usecase

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SOAR-F5: every shipped flow uses $(adversary.ip)-style placeholders, but
// NewRootContext (interpolate.go) wraps the alert under a top-level "alert"
// key, so the real gjson path is "alert.adversary.ip" — a bare "adversary.ip"
// never exists in the bag and Interpolate leaves the placeholder untouched,
// verbatim, in the dispatched command.
func TestAlertPlaceholderNeedsAlertPrefix(t *testing.T) {
	alertJSON := []byte(`{"adversary":{"ip":"10.0.0.9"},"target":{"process":"evil.exe"}}`)
	bag := NewRootContext(alertJSON)

	got, err := Interpolate(nil, nil, bag, "kill $(target.process) from $(adversary.ip)", "")
	if err != nil {
		t.Fatalf("Interpolate: %v", err)
	}
	if got != "kill $(target.process) from $(adversary.ip)" {
		t.Fatalf("bare placeholders resolved unexpectedly: %q (shipped flows must use the alert.-prefixed form)", got)
	}

	got, err = Interpolate(nil, nil, bag, "kill $(alert.target.process) from $(alert.adversary.ip)", "")
	if err != nil {
		t.Fatalf("Interpolate: %v", err)
	}
	want := "kill 'evil.exe' from '10.0.0.9'" // escapeInterpolatedValue quotes by default (SOAR-F1)
	if got != want {
		t.Fatalf("Interpolate(alert.-prefixed) = %q, want %q", got, want)
	}
}

var placeholderPattern = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_.]*)\)`)

// Regression guard for SOAR-F5's actual fix: none of the 94 shipped flow
// YAMLs may regress back to a bare, never-resolving placeholder.
func TestShippedFlowsOnlyUseAlertPrefixedPlaceholders(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "definitions", "soar")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("shipped flows directory not found at %s: %v", dir, err)
	}

	checked := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		checked++
		for _, m := range placeholderPattern.FindAllStringSubmatch(string(data), -1) {
			if !strings.HasPrefix(m[1], "alert.") {
				t.Errorf("%s: placeholder $(%s) is missing the alert. prefix and will never resolve", e.Name(), m[1])
			}
		}
	}
	if checked == 0 {
		t.Skip("no shipped flow YAML files found")
	}
	if checked != 94 {
		t.Logf("checked %d shipped flows (expected 94 at the time this test was written — not a failure, just worth noting if the count drifted)", checked)
	}
}
