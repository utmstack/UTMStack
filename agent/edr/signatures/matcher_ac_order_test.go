package signatures

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

func nativeFixedACControl(t *testing.T, input []byte, files map[string][]byte, want string) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	dir := t.TempDir()
	args := []string{"--no-summary"}
	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, files[name], 0600); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--database="+path)
	}
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), filepath.Join(ref, "bin", "clamscan"), append(args, path)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if code, ok := err.(*exec.ExitError); !ok || code.ExitCode() != 1 || !strings.Contains(string(out), want+".UNOFFICIAL FOUND") {
		t.Fatalf("native fixed AC: %v: %s; want %q", err, out, want)
	}
	t.Logf("native fixed AC: %s", strings.TrimSpace(string(out)))
}

func TestFixedACNativeCandidateSelection(t *testing.T) {
	body, err := os.ReadFile("testdata/ac_fixed_candidate_selection.json")
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
	if err = json.Unmarshal(body, &fixture); err != nil || len(fixture.Cases) != 220 {
		t.Fatalf("native fixed AC inventory: %v; cases=%d", err, len(fixture.Cases))
	}
	for index, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			input, err := hex.DecodeString(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			var rules strings.Builder
			for i, pattern := range row.Patterns {
				fmt.Fprintf(&rules, "Candidate.R%03d:0:*:%s\n", i, pattern)
			}
			files := map[string][]byte{"rules.ndb": []byte(rules.String())}
			d, err := Compile(files)
			if err != nil {
				t.Fatal(err)
			}
			if d.fixedAC[0] == nil {
				t.Fatal("fixed AC selection not enabled for native control")
			}
			want := fmt.Sprintf("Candidate.R%03d", row.Expected)
			assertMatch(t, d, input, want)
			if index < 10 || index >= 100 && index < 110 {
				nativeFixedACControl(t, input, files, want)
			}
		})
	}
}

func TestFixedACNativeGroupAndAnchorOrder(t *testing.T) {
	input := []byte("ABCDEF harmless fixed AC group control\n")
	for _, tc := range []struct {
		name  string
		files map[string][]byte
		want  string
	}{
		{"offset-independent-groups", map[string][]byte{
			"rules.ndb": []byte("First:0:*:41424344??46\nUnreached:0:999:414243??4546\nGrouped:0:*:414243??4546\n"),
		}, "Grouped"},
		{"logical-registration-counts", map[string][]byte{
			"a.ndb": []byte("First:0:*:41424344??46\n"),
			"b.ldb": []byte("Never.Alert;Target:0;0>2;414243??4546\n"),
			"z.ndb": []byte("Grouped:0:*:414243??4546\n"),
		}, "Grouped"},
		{"logical-first-group-breaks-tie", map[string][]byte{
			"a.ldb": []byte("Never.A;Target:0;0>2;414243??4546\n"),
			"b.ndb": []byte("First:0:*:41424344??46\nGrouped:0:*:414243??4546\n"),
			"z.ldb": []byte("Never.B;Target:0;0>2;41424344??46\n"),
		}, "Grouped"},
		{"deeper-key-before-failure-link", map[string][]byte{
			"rules.ndb": []byte("Short:0:*:??4344??46\nDeep:0:*:??424344??46\n"),
		}, "Deep"},
		{"anchor-before-body-end", map[string][]byte{
			"rules.ndb": []byte("Later:0:*:424344??46\nEarlier:0:*:414243??4546206861726d6c657373\n"),
		}, "Earlier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Compile(tc.files)
			if err != nil {
				t.Fatal(err)
			}
			if d.fixedAC[0] == nil {
				t.Fatal("fixed AC selection not enabled")
			}
			assertMatch(t, d, input, tc.want)
			nativeFixedACControl(t, input, tc.files, tc.want)
		})
	}
}

func TestFixedACComplexRegistrationsUseNativePartStream(t *testing.T) {
	for _, other := range []struct{ name, text string }{
		{"other.ndb", "Other:0:*:78797a*717273\n"},
		{"other.ldb", "Other;Target:0;0;78797a(70|71)7273\n"},
		{"other.ldb", "Other;Target:0;0;78797a::i\n"},
	} {
		files := map[string][]byte{
			"rules.ndb": []byte("First:0:*:424344??46\nEarlier:0:*:414243??4546\n"),
			other.name:  []byte(other.text),
		}
		d, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if d.fixedAC[0] != nil {
			t.Fatal("partially modeled root published a fixed AC index")
		}
		if d.nativeAC[0] == nil {
			t.Fatal("general native part stream not enabled")
		}
		input := []byte("ABCDEF harmless unmodeled registration control")
		assertMatch(t, d, input, "Earlier")
		nativeFixedACControl(t, input, files, "Earlier")
	}
}

func TestFixedACTypedRootIsIndependentOfGenericFallback(t *testing.T) {
	input := testPE()
	copy(input[512:], []byte("ABCDEF harmless fixed AC typed-root control"))
	files := map[string][]byte{"rules.ndb": []byte("Typed.Later:1:*:424344??46\nTyped.Earlier:1:*:414243??4546\nGeneric.Unmodeled:0:*:78797a*717273\n")}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if d.fixedAC[0] != nil || d.fixedAC[1] == nil {
		t.Fatal("generic fallback changed the typed root's AC eligibility")
	}
	assertMatch(t, d, input, "Typed.Earlier")
	nativeFixedACControl(t, input, files, "Typed.Earlier")
}

func TestFixedACRefreshKeepsConcurrentOldSnapshots(t *testing.T) {
	input := []byte("ABCDEF harmless fixed AC generation control")
	files := map[string][]byte{
		"a.ndb": []byte("First:0:*:41424344??46\nGrouped:0:*:414243??4546\n"),
		"z.ndb": []byte("Other:0:*:78797a??717273\n"),
	}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, prior, input, "First")
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for range 100 {
				name, err := prior.Match(context.Background(), input)
				if err != nil || name != "First" {
					t.Errorf("old fixed AC generation changed: %q %v", name, err)
					return
				}
			}
		})
	}
	defer readers.Wait()
	current := prior
	for i := range 20 {
		want := "First"
		files["z.ndb"] = []byte("Other:0:*:78797a??717273\n")
		if i%2 == 0 {
			files["z.ndb"] = []byte("Extra:0:*:414243??4546\n")
			want = "Grouped"
		}
		next, handled, err := tryRuleSourceRefresh(files, DefaultLimits(), compileOptions{prior: current})
		if err != nil || !handled {
			t.Fatalf("fixed AC refresh: handled=%v: %v", handled, err)
		}
		full, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		assertMatch(t, next, input, want)
		assertMatch(t, full, input, want)
		assertMatch(t, prior, input, "First")
		current = next
	}
}
