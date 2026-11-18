package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestStandaloneBMNativeCandidateSelection(t *testing.T) {
	data, err := os.ReadFile("testdata/bm_candidate_selection.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name     string   `json:"case"`
			Input    string   `json:"input_hex"`
			Patterns []string `json:"patterns"`
			Expected int      `json:"expected_index"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || len(fixture.Cases) != 100 {
		t.Fatalf("native BM inventory: %v, cases=%d", err, len(fixture.Cases))
	}
	for index, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			input, err := hex.DecodeString(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			for _, pe := range []bool{false, true} {
				var text strings.Builder
				for i, value := range row.Patterns {
					target, off := 0, "*"
					if pe {
						pattern, err := hex.DecodeString(value)
						if err != nil {
							t.Fatal(err)
						}
						target, off = 1, fmt.Sprint(512+bytes.Index(input, pattern))
					}
					fmt.Fprintf(&text, "Candidate.R%03d:%d:%s:%s\n", i, target, off, value)
				}
				object := input
				if pe {
					object = testPE()
					copy(object[512:], input)
				}
				d := compileTest(t, "ndb", text.String())
				want := fmt.Sprintf("Candidate.R%03d", row.Expected)
				assertMatch(t, d, object, want)
				if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" && pe && index < 20 {
					dir := t.TempDir()
					path, rules := filepath.Join(dir, "input.exe"), filepath.Join(dir, "rules.ndb")
					if err := os.WriteFile(path, object, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(rules, []byte(text.String()), 0600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.CommandContext(context.Background(), filepath.Join(ref, "bin", "clamscan"), "--no-summary", "--database="+rules, path)
					cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
					out, err := cmd.CombinedOutput()
					if code, ok := err.(*exec.ExitError); !ok || code.ExitCode() != 1 || !strings.Contains(string(out), want+".UNOFFICIAL FOUND") {
						t.Fatalf("native PE BM control: %v: %s; want %q", err, out, want)
					}
					t.Logf("native PE: %s", strings.TrimSpace(string(out)))
				}
			}
		})
	}
}

func TestStandaloneBMRefreshRebuildsCandidateOrder(t *testing.T) {
	input := []byte("ABCDEF harmless source-order refresh control")
	line := func(i int) []byte { return []byte(fmt.Sprintf("Candidate.R%03d:0:*:414243444546\n", i)) }
	files := map[string][]byte{"a.ndb": line(0), "z.ndb": append(line(1), line(2)...)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, prior, input, "Candidate.R000")
	files["a.ndb"] = append(append(line(0), line(3)...), append(line(4), line(5)...)...)
	next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior})
	if err != nil || !handled {
		t.Fatalf("incremental refresh: handled=%v, error=%v", handled, err)
	}
	full, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	// Six equal six-byte literals exhaust the four distinct triplet buckets;
	// the last source entry retains prefix0 and precedes the first there.
	assertMatch(t, next, input, "Candidate.R002")
	assertMatch(t, full, input, "Candidate.R002")
	assertMatch(t, prior, input, "Candidate.R000")
}

func TestStandaloneBMConcurrentRefreshSnapshots(t *testing.T) {
	input := []byte("ABCDEF harmless immutable candidate-index control")
	line := func(i int) []byte { return []byte(fmt.Sprintf("Candidate.R%03d:0:*:414243444546\n", i)) }
	files := map[string][]byte{"a.ndb": line(0), "z.ndb": append(line(1), line(2)...)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	stop, failures := make(chan struct{}), make(chan error, 4)
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				name, err := prior.Match(context.Background(), input)
				if name != "Candidate.R000" || err != nil {
					failures <- fmt.Errorf("old snapshot changed: %q, %v", name, err)
					return
				}
			}
		}()
	}
	defer func() {
		close(stop)
		readers.Wait()
		close(failures)
		for err := range failures {
			t.Error(err)
		}
	}()
	current := prior
	for round := 0; round < 20; round++ {
		files["a.ndb"] = line(0)
		want := "Candidate.R000"
		if round%2 == 0 {
			files["a.ndb"] = append(append(line(0), line(3)...), append(line(4), line(5)...)...)
			want = "Candidate.R002"
		}
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: current})
		if err != nil || !handled {
			t.Fatalf("round%d refresh: handled=%v, error=%v", round, handled, err)
		}
		assertMatch(t, next, input, want)
		current = next
	}
}
