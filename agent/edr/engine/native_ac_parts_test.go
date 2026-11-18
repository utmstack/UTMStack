package engine

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestNativeACMultipartHistoryAndCounts(t *testing.T) {
	checkNativeACFixture(t, "native_multipart_history.json", 16)
}

func TestNativeACMultipartMatrix(t *testing.T) {
	checkNativeACFixture(t, "native_multipart_matrix.json", 336)
}

func checkNativeACFixture(t *testing.T, filename string, count int) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../signatures/testdata", filename))
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Cases []struct {
			Name     string            `json:"name"`
			Input    string            `json:"input_hex"`
			Files    map[string]string `json:"files"`
			Expected string            `json:"expected"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(data, &wire); err != nil || len(wire.Cases) != count {
		t.Fatalf("multipart inventory: %v; cases=%d", err, len(wire.Cases))
	}
	for _, row := range wire.Cases {
		t.Run(row.Name, func(t *testing.T) {
			input, err := hex.DecodeString(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			t.Cleanup(func() { _ = e.Stop() })
			files := map[string][]byte{}
			for name, text := range row.Files {
				files[name] = []byte(text)
			}
			if _, err = e.ActivateGeneration(files, 1, strings.Repeat("1", 64)); err != nil {
				t.Fatal(err)
			}
			clean, name, _, err := e.ScanSnapshot(context.Background(), input)
			if err != nil || clean != (row.Expected == "") || name != row.Expected {
				t.Fatalf("multipart scan: clean=%v name=%q err=%v; native=%q", clean, name, err, row.Expected)
			}
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				dir := t.TempDir()
				var names []string
				for name := range files {
					names = append(names, name)
				}
				sort.Strings(names)
				args := []string{"--no-summary"}
				for _, name := range names {
					p := filepath.Join(dir, name)
					if err = os.WriteFile(p, files[name], 0600); err != nil {
						t.Fatal(err)
					}
					args = append(args, "--database="+p)
				}
				path := filepath.Join(dir, "input")
				if err = os.WriteFile(path, input, 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.CommandContext(context.Background(), filepath.Join(ref, "bin", "clamscan"), append(args, path)...)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, runErr := cmd.CombinedOutput()
				code := 0
				if runErr != nil {
					exit, ok := runErr.(*exec.ExitError)
					if !ok {
						t.Fatal(runErr)
					}
					code = exit.ExitCode()
				}
				want := 0
				if row.Expected != "" {
					want = 1
				}
				if code != want || row.Expected != "" && !strings.Contains(string(out), row.Expected+".UNOFFICIAL FOUND") {
					t.Fatalf("native multipart: exit=%d: %s", code, out)
				}
				t.Logf("native multipart: %s", strings.TrimSpace(string(out)))
			}
		})
	}
}

// The mixed native control exits on SIGSEGV, so it cannot supply a detection
// oracle. The two independently positive isolated rules still require a Go
// detection, and the Go scan must terminate without an error or process fault.
func TestNativeACReferenceCrashDoesNotCrashGo(t *testing.T) {
	data, err := os.ReadFile("../signatures/testdata/native_multipart_reference_crash.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Input     string            `json:"input_hex"`
		Files     map[string]string `json:"files"`
		Signal    int               `json:"native_signal"`
		Qualified bool              `json:"parity_qualified"`
	}
	if err = json.Unmarshal(data, &fixture); err != nil || fixture.Signal != 11 || fixture.Qualified {
		t.Fatal("reference crash inventory", err)
	}
	input, err := hex.DecodeString(fixture.Input)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	t.Cleanup(func() { _ = e.Stop() })
	files := map[string][]byte{}
	for n, v := range fixture.Files {
		files[n] = []byte(v)
	}
	if _, err = e.ActivateGeneration(files, 1, strings.Repeat("1", 64)); err != nil {
		t.Fatal(err)
	}
	clean, name, _, err := e.ScanSnapshot(context.Background(), input)
	if err != nil || clean || name == "" {
		t.Fatalf("Go crash control: clean=%v name=%q err=%v", clean, name, err)
	}
}
