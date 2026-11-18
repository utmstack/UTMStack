package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// signature_db_dir is resolved against the process working directory. The
// Windows service runs from %SystemRoot%\System32 while engine-check/scan run
// from the admin's shell, and a bare name in the scan exclusions matches any
// file with that base name, so a relative value must be rejected when set.
func TestApplySetSignatureDBDirRequiresAbsolutePath(t *testing.T) {
	c := config.Default()
	for _, rel := range []string{"sigs", `.\sigs`, "sigs/db", "."} {
		if err := applySet(&c, "signature_db_dir", rel); err == nil {
			t.Fatalf("relative signature_db_dir %q accepted (stored %q)", rel, c.SignatureDBDir)
		}
		if c.SignatureDBDir != "" {
			t.Fatalf("rejected value %q was stored as %q", rel, c.SignatureDBDir)
		}
	}
	dir := t.TempDir()
	if err := applySet(&c, "signature_db_dir", filepath.Join(dir, "sigs", "..", "db")+string(filepath.Separator)); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "db"); c.SignatureDBDir != want {
		t.Fatalf("absolute value not cleaned: %q, want %q", c.SignatureDBDir, want)
	}
	// Empty restores the default directory.
	if err := applySet(&c, "signature_db_dir", ""); err != nil || c.SignatureDBDir != "" {
		t.Fatalf("empty value: err=%v stored=%q", err, c.SignatureDBDir)
	}
}

func TestSetScanResourceLimitsIsValidatedAndReadable(t *testing.T) {
	for _, tc := range []struct{ key, valid, tooLarge string }{
		{"scan_concurrency", "1", "17"},
		{"scan_limits.max_file_size_mb", "5", "201"},
		{"scan_limits.max_scan_size_mb", "12", "401"},
		{"scan_limits.max_recursion", "3", "17"},
		{"scan_limits.max_files", "20", "10001"},
		{"scan_limits.max_scan_time_ms", "750", "120001"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			cfg := config.Default()
			if err := applySet(&cfg, tc.key, tc.valid); err != nil {
				t.Fatal(err)
			}
			if got, ok := getByPath(cfg, tc.key); !ok || got != tc.valid {
				t.Fatalf("get = %q,%v", got, ok)
			}
			before := cfg
			for _, invalid := range []string{"-1", tc.tooLarge, "garbage", "999999999999999999999999999999"} {
				if err := applySet(&cfg, tc.key, invalid); err == nil {
					t.Errorf("accepted %q", invalid)
				}
				if !reflect.DeepEqual(before, cfg) {
					t.Fatalf("rejected value %q mutated configuration", invalid)
				}
			}
			if tc.key != "scan_concurrency" {
				if err := applySet(&cfg, tc.key, "0"); err != nil {
					t.Fatalf("restore auto: %v", err)
				}
			} else if err := applySet(&cfg, tc.key, "0"); err == nil {
				t.Fatal("zero concurrency accepted")
			}
		})
	}
	cfg := config.Default()
	if err := applySet(&cfg, "scan_limits.max_scan_size_mb", "5"); err != nil {
		t.Fatal(err)
	}
	if err := applySet(&cfg, "scan_limits.max_file_size_mb", "10"); err == nil {
		t.Fatal("file budget above total accepted")
	}
}

func TestSetPDFNameHeuristicIsValidatedAndReadable(t *testing.T) {
	cfg := config.Default()
	const key = "scan_heuristics.pdf_obfuscated_name"
	if got, ok := getByPath(cfg, key); !ok || got != "true" {
		t.Fatalf("default = %q,%v", got, ok)
	}
	if err := applySet(&cfg, key, "false"); err != nil {
		t.Fatal(err)
	}
	if got, ok := getByPath(cfg, key); !ok || got != "false" {
		t.Fatalf("disabled = %q,%v", got, ok)
	}
	if err := applySet(&cfg, key, "maybe"); err == nil || cfg.ScanHeuristics.PDFObfuscatedNameOn() {
		t.Fatalf("invalid value changed policy: %v", err)
	}
}

// An edr.json written by the previous (daemon-based) release still carries the
// old engine transport fields. The native engine ignores them, and the
// operator-facing `config show` / `config get` output must not surface them.
func TestConfigShowOmitsLegacyEngineFields(t *testing.T) {
	old := config.ConfigFile
	config.ConfigFile = filepath.Join(t.TempDir(), "edr.json")
	defer func() { config.ConfigFile = old }()
	legacy := `{"enabled":true,"fail_mode":"open","clamd_addr":"127.0.0.1:3310","clamd_path":"C:\\clamav\\clamd.exe","scan_concurrency":2}`
	if err := os.WriteFile(config.ConfigFile, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	runConfig([]string{"show"})
	os.Stdout = stdout
	w.Close()
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), `"scan_concurrency": 2`) {
		t.Fatalf("`config show` did not print the loaded configuration:\n%s", out)
	}
	if strings.Contains(strings.ToLower(string(out)), "clam") {
		t.Fatalf("`config show` surfaces legacy engine fields:\n%s", out)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"clamd_addr", "clamd_path"} {
		if v, ok := getByPath(cfg, key); ok && v != "" {
			t.Fatalf("`config get %s` returns %q", key, v)
		}
	}
}

// settableKeys returns every key `config set` accepts, read from the case
// labels of applySet, so a newly settable key is covered automatically.
func settableKeys(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "manage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "applySet" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			clause, ok := n.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, e := range clause.List {
				if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					key, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					keys = append(keys, key)
				}
			}
			return true
		})
	}
	if len(keys) < 20 {
		t.Fatalf("found only %d settable keys in applySet", len(keys))
	}
	return keys
}

// Every settable key must be readable with `config get`, including while it
// holds its default. Fields omitted from the JSON form when empty or unset
// (signature_db_dir, signature_public_key, sensors.*) used to be reported as
// unknown keys.
func TestConfigGetReadsEverySettableKeyAtDefault(t *testing.T) {
	cfg := config.Default()
	for _, key := range settableKeys(t) {
		if _, ok := getByPath(cfg, key); !ok {
			t.Errorf("config get %s: unknown key at its default value", key)
		}
	}
}

func TestConfigGetReportsDefaultsAndSetValues(t *testing.T) {
	cfg := config.Default()
	for key, want := range map[string]string{
		"signature_db_dir":      "",
		"signature_public_key":  "",
		"sensors.file_watcher":  "true", // an unset sensor is on
		"sensors.process_guard": "true",
		"sensors.amsi":          "true",
		"sensors.behavioral":    "true",
		"allowlist.networks":    "",
	} {
		if got, ok := getByPath(cfg, key); !ok || got != want {
			t.Errorf("config get %s at default = %q, %v; want %q", key, got, ok, want)
		}
	}
	dir := filepath.Join(t.TempDir(), "db")
	key := "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="
	for k, v := range map[string]string{"signature_db_dir": dir, "signature_public_key": key, "sensors.amsi": "false"} {
		if err := applySet(&cfg, k, v); err != nil {
			t.Fatal(err)
		}
	}
	for k, want := range map[string]string{"signature_db_dir": dir, "signature_public_key": key, "sensors.amsi": "false", "sensors.behavioral": "true"} {
		if got, ok := getByPath(cfg, k); !ok || got != want {
			t.Errorf("config get %s after set = %q, %v; want %q", k, got, ok, want)
		}
	}
	// Keys that are not configuration fields stay unknown.
	for _, k := range []string{"no_such_key", "sensors.no_such_sensor", "exclusions", "signature_db_dir.x"} {
		if v, ok := getByPath(cfg, k); ok {
			t.Errorf("config get %s = %q, want unknown key", k, v)
		}
	}
}

// The help must describe what a path allowlist entry really exempts: the
// real-time watcher, the launch guard and process telemetry skip it, while an
// on-demand scan and the ransomware guard (which can quarantine a scored
// program's image) still act there.
func TestHelpDescribesPathAllowlistScope(t *testing.T) {
	// Windows pipes can fill before this synchronous help writer returns.
	// A temporary file captures the complete help without a pipe-size limit.
	capture, err := os.CreateTemp(t.TempDir(), "help-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer capture.Close()
	stdout := os.Stdout
	os.Stdout = capture
	defer func() { os.Stdout = stdout }()
	printUsage()
	os.Stdout = stdout
	if _, err := capture.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(capture)
	if err != nil {
		t.Fatal(err)
	}
	help := strings.Join(strings.Fields(string(out)), " ")
	for _, want := range []string{
		"the real-time watcher never scans or quarantines files there",
		"not scanned, suspended or stopped at launch",
		"left out of process telemetry",
		"can stop them and quarantine their image",
		"An on-demand 'scan <path>' ignores the allowlist",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help does not say %q:\n%s", want, out)
		}
	}
	if strings.Contains(help, "files there are never scanned or quarantined") {
		t.Errorf("help claims allowlisted files are never scanned or quarantined:\n%s", out)
	}
}
