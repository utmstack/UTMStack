package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestDryRunScanReportsEachFileWithoutActing(t *testing.T) {
	db := t.TempDir()
	// The feed's own test names carry the upstream vendor's name, which output must not.
	rules := "Clamav.Test.Marker:0:*:" + hex.EncodeToString([]byte("DRYRUN-MARKER")) + "\n"
	pcre := "UTMStack.Test.Regex;Engine:81-255,Target:0;0&1;" + hex.EncodeToString([]byte("NEEDS-REGEX")) + ";0/x+(?R)?y/\n"
	if err := os.WriteFile(filepath.Join(db, "local.ndb"), []byte(rules), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(db, "local.ldb"), []byte(pcre), 0o600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := map[string]string{
		"detected.txt":   "prefix DRYRUN-MARKER suffix",
		"clean.txt":      "nothing to see in this file",
		"incomplete.txt": "this one NEEDS-REGEX to decide",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	code := runDryRun([]string{"--dry-run", "--json", "--db", db, "--workers", "2", dir}, &out)
	if code != 1 {
		t.Fatalf("exit code %d, want 1 (a detection)\n%s", code, out.String())
	}
	lines := map[string]dryRunLine{}
	var summary struct {
		Summary *dryRunSummary `json:"summary"`
	}
	sc := bufio.NewScanner(&out)
	for sc.Scan() {
		text := sc.Text()
		if strings.HasPrefix(text, `{"summary"`) {
			if err := json.Unmarshal([]byte(text), &summary); err != nil {
				t.Fatal(err)
			}
			continue
		}
		var l dryRunLine
		if err := json.Unmarshal([]byte(text), &l); err != nil {
			t.Fatalf("line %q: %v", text, err)
		}
		lines[filepath.Base(l.Path)] = l
	}
	if len(lines) != 3 {
		t.Fatalf("lines = %+v", lines)
	}
	if l := lines["detected.txt"]; l.Clean || l.Signature != "UTMStack.Test.Marker" || l.Size != int64(len(files["detected.txt"])) {
		t.Fatalf("detected = %+v", l)
	}
	if l := lines["clean.txt"]; !l.Clean || l.Signature != "" || l.IncompleteReason != "" || l.Error != "" {
		t.Fatalf("clean = %+v", l)
	}
	if l := lines["incomplete.txt"]; l.Clean || l.Signature != "" || l.IncompleteReason != "Unsupported.pcre" {
		t.Fatalf("incomplete = %+v", l)
	}
	s := summary.Summary
	if s == nil || s.Files != 3 || s.Detected != 1 || s.Clean != 1 || s.Incomplete != 1 || s.IncompleteByReason["Unsupported.pcre"] != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if strings.Contains(strings.ToLower(out.String()), "clam") {
		t.Fatalf("output leaks the engine name:\n%s", out.String())
	}
	for _, p := range []string{config.Default().QuarantineDir, config.DBFile, config.SpoolFile} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("dry run created %s", p)
		}
	}
	for name := range files {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("dry run touched %s: %v", name, err)
		}
	}
}

func TestDryRunScanRejectsUsageWithoutJSON(t *testing.T) {
	var out bytes.Buffer
	if code := runDryRun([]string{"--dry-run", t.TempDir()}, &out); code != 2 {
		t.Fatalf("exit code %d, want 2", code)
	}
}

func TestOnDemandScanLineKeepsVendorNameOut(t *testing.T) {
	if got := verdictLine("malicious", "Clamav.Test.File-6"); got != "UTMStack EDR verdict: malicious UTMStack.Test.File-6" {
		t.Fatalf("line = %q", got)
	}
}
