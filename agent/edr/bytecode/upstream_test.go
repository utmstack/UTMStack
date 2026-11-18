package bytecode

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Optional, offline compatibility checks against benign fixtures from the
// pinned upstream revision in REFERENCE.md. They are deliberately not embedded
// in endpoint feeds or downloaded by tests. Set CLAMBC_REFERENCE_DIR to a
// directory containing the named files; hash mismatches are fatal, not skips.
// Expected results come from upstream unit_tests/check_bytecode.c. This checks
// those published expectations, not execution of an upstream C engine here.
func TestPinnedUpstreamPrograms(t *testing.T) {
	dir := os.Getenv("CLAMBC_REFERENCE_DIR")
	if dir == "" {
		t.Skip("set CLAMBC_REFERENCE_DIR for pinned upstream fixture checks")
	}
	read := func(name, digest string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != digest {
			t.Fatalf("%s sha256 %s, want %s", name, got, digest)
		}
		return b
	}
	input := read("apitestfile", "64ff572e97c83b2554bf8846bc8adee6ad88cb54230e5144a05135360448fbf2")
	for _, tc := range []struct {
		name, digest string
		want         uint64
		file         bool
		limit        bool
		options      bool
	}{
		{"retmagic_7.cbc", "92f1f98bb5d5b54e835aa8b152874acf6db99bfa1d0d317f527f54eb37b94ca3", 0x1234f00d, false, false, false},
		{"arith_7.cbc", "bb235a76a83414c995cf5aa11f8b992bd50f3f62e9f9983cffa7120dc6634e93", 0xd55555dd, false, false, false},
		{"apicalls_7.cbc", "9a44837b91a3e92cb0e236de7c05f041078728f4b8ba1512556046825a645586", 0xf00d, false, false, false},
		{"apicalls2_7.cbc", "233063381bb1df61bd87ee16c7108f707d84766dd5570b2d8b847558c274b158", 0xf00d, false, false, false},
		{"api_files_7.cbc", "0d8e7c977dcbb94eaf659f00cd2eeb81b3d904ea3dbf464ceadb7fdfbdd39113", 0xf00d, true, false, false},
		{"lsig_7.cbc", "0ecee467e479232fe61bd2afb1f972cf5871f6108c85fe558c20a27917979255", 0, false, false, false},
		{"debug_7.cbc", "bd25284abddab04f468e2d0dbe82006fa98118b6e203cc71e4ef21eed7a6e705", 0xf00d, false, false, true},
		{"inf_7.cbc", "6bcd15d1fbba319e1f7d407e58bdc5950fc5e0d8883083fb8efc6b55dfab1ced", 0, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(read(tc.name, tc.digest))
			if err != nil {
				t.Fatal(err)
			}
			var data []byte
			if tc.file {
				data = input
			}
			var result Result
			if tc.name == "lsig_7.cbc" || tc.options {
				matches := EmptyMatchContext()
				if tc.options {
					matches.Options = &EngineOptions{Scan: 0x1406037, Database: 0x200a}
				}
				result, err = p.RunWithMatchContext(context.Background(), 0, nil, data, matches, 1_000_000)
			} else {
				result, err = p.RunWithInput(context.Background(), 0, nil, data, 1_000_000)
			}
			if tc.limit {
				if !errors.Is(err, ErrLimit) {
					t.Fatalf("unbounded loop: %v", err)
				}
				return
			}
			if err != nil || result.Value != tc.want {
				t.Fatalf("got %#x, %v; want %#x", result.Value, err, tc.want)
			}
			if result.VirusName != "" {
				t.Fatalf("unexpected detection label %q", result.VirusName)
			}
		})
	}
}

func TestConstantStorageAndAbortTerminator(t *testing.T) {
	// readOperand zeroes the 64-bit constant slot before writing its declared
	// byte width. The VM's read width belongs to the operation, not the literal.
	for _, tc := range []struct {
		stored, read int
		value, want  uint64
	}{
		{8, 32, 0x1ff, 0xff}, {32, 64, 0x123456789, 0x23456789},
		{64, 32, 0x123456789, 0x23456789}, {32, 1, 2, 0}, {64, 1, 3, 1},
	} {
		p, err := Parse(compatProgram(compatFunction{result: tc.read, blocks: [][]string{{compatRet(tc.read, compatConstant(tc.value, tc.stored))}}}))
		if err != nil {
			t.Fatal(err)
		}
		got, err := p.Run(context.Background(), 0, nil, 100)
		if err != nil || got != tc.want {
			t.Fatal(tc, got, err)
		}
	}
	p, err := Parse(compatProgram(compatFunction{result: 32, blocks: [][]string{{"T" + compatFixed(45, 2)}}}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Run(context.Background(), 0, nil, 100); !errors.Is(err, ErrRuntime) {
		t.Fatal(err)
	}
}
