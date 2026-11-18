package signatures

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// These distinct names expose stage precedence, rather than merely proving
// that some positive rule matched. The inert CBC wire builder is independent
// of both the production parser and VM.
func TestStandaloneHashLogicalAndCallbackStageOrder(t *testing.T) {
	input := []byte("ABCDEF harmless stage-order control\n")
	f := callbackWithLabel([]byte("Strong.Callback"), true)
	f.logical = "Callback.Trigger;Target:0;0;414243444546"
	cbc := f.encode()
	logical := []byte("Strong.Logical;Target:0;0;414243444546\n")
	hash := []byte(fmt.Sprintf("%x:%d:Strong.Hash\n", md5.Sum(input), len(input)))
	for _, tc := range []struct {
		name  string
		files map[string][]byte
		want  string
	}{
		{"callback-before-logical", map[string][]byte{"a.cbc": cbc, "z.ldb": logical}, "Strong.Callback"},
		{"logical-before-callback", map[string][]byte{"a.ldb": logical, "z.cbc": cbc}, "Strong.Logical"},
		{"hash-before-logical", map[string][]byte{"a.ldb": logical, "z.hdb": hash}, "Strong.Hash"},
		{"hash-before-callback", map[string][]byte{"a.cbc": cbc, "z.hdb": hash}, "Strong.Hash"},
		{"standalone-before-logical", map[string][]byte{"a.ldb": logical, "z.ndb": []byte("Strong.Body:0:*:414243444546\n")}, "Strong.Body"},
		{"bm-before-ac", map[string][]byte{"a.ndb": []byte("Strong.AC:0:*:414243??4546\nStrong.BM:0:*:414243444546\n")}, "Strong.BM"},
		{"bm-before-zero-gap-ac", map[string][]byte{"a.ndb": []byte("Strong.AC:0:*:414243{0}444546\nStrong.BM:0:*:414243444546\n")}, "Strong.BM"},
		{"bm-before-legacy-ac", map[string][]byte{"a.db": []byte("Strong.Legacy=414243??4546\n"), "z.ndb": []byte("Strong.BM:0:*:414243444546\n")}, "Strong.BM"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Compile(tc.files)
			if err != nil {
				t.Fatal(err)
			}
			assertMatch(t, d, input, tc.want)
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				dir := t.TempDir()
				args := []string{"--no-summary", "--bytecode-unsigned=yes", "--bytecode=yes"}
				names := make([]string, 0, len(tc.files))
				for name := range tc.files {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					path := filepath.Join(dir, name)
					if err := os.WriteFile(path, tc.files[name], 0600); err != nil {
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
				if code, ok := err.(*exec.ExitError); !ok || code.ExitCode() != 1 ||
					!strings.Contains(string(out), tc.want+" FOUND") && !strings.Contains(string(out), tc.want+".UNOFFICIAL FOUND") {
					t.Fatalf("native stage control: %v: %s; want %q", err, out, tc.want)
				}
				t.Logf("native: %s", strings.TrimSpace(string(out)))
			}
		})
	}
}

func TestStandaloneMatchFamilyPreservesRawSyntaxAcrossTargets(t *testing.T) {
	for _, tc := range []struct {
		target int
		offset string
		body   string
		ac     bool
	}{
		{0, "*", "414243444546", false},
		{1, "*", "414243444546", true},
		{0, "0,0", "414243444546", false},
		{1, "0,0", "414243444546", true},
		{1, "0", "414243444546", false},
		{6, "0", "414243444546", true},
		{0, "*", "414243{0}444546", true},
		{0, "*", "414243(44|64)4546", true},
		{0, "*", "414243??4546", true},
		{1, "VI", "414243444546", true},
	} {
		name := fmt.Sprintf("target%d/%s/%s", tc.target, tc.offset, tc.body)
		t.Run(name, func(t *testing.T) {
			d := compileTest(t, "ndb", fmt.Sprintf("Family:%d:%s:%s\n", tc.target, tc.offset, tc.body))
			if got := d.standaloneAC(&d.rules[0]); got != tc.ac {
				t.Fatalf("AC family=%v, want %v", got, tc.ac)
			}
		})
	}
	// Shared patterns retain target-independent syntax flags. A PE any-offset
	// rule must not change the generic rule's family through the shared ID.
	d := compileTest(t, "ndb", "Typed:1:*:414243444546\nGeneric:0:*:414243444546\n")
	if d.rules[0].patterns[0] != d.rules[1].patterns[0] || !d.standaloneAC(&d.rules[0]) || d.standaloneAC(&d.rules[1]) {
		t.Fatal("shared target-independent pattern lost native matcher routing")
	}
}
