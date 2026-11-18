package engine

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// This independently encoded, harmless fixture asks one question of the
// reference interpreter: what does wire API 89 return for a parsed PE icon?
func matchIconOracleProgram(kind uint64, group1, group2 string, want uint64) []byte {
	const label = "Bytecode.Icon.Detected"
	const marker = "UTMSTACK-MATCHICON-HARMLESS-TRIGGER"
	strType := func(s string) string {
		return engineCBCFixed(4, 1) + engineCBCNumber(uint64(len(s)+1)) + engineCBCNumber(8)
	}
	fnType := func(parts ...uint64) string {
		out := engineCBCFixed(1, 1) + engineCBCNumber(uint64(len(parts)))
		for _, p := range parts {
			out += engineCBCNumber(p)
		}
		return out
	}
	var globals string
	globals += engineCBCGlobal(0, 0)
	for i, s := range []string{group1, group2, label} {
		var values []uint64
		for _, b := range append([]byte(s), 0) {
			values = append(values, uint64(b))
		}
		globals += engineCBCGlobal(69+i, values...)
		globals += engineCBCGlobal(65, 0, uint64(1+2*i))
	}
	header := "ClamBC" + engineCBCNumber(7) + engineCBCNumber(0) + "|" + engineCBCNumber(0) +
		engineCBCNumber(0) + engineCBCNumber(kind) + strings.Repeat(engineCBCNumber(0), 3) + "|" + engineCBCNumber(0) +
		engineCBCNumber(10) + engineCBCNumber(1) + engineCBCNumber(0x53e5493e9f3d1c30) + engineCBCFixed(42, 2) + ":4096"
	types := "T" + engineCBCFixed(69, 2) + strType(group1) + strType(group2) + strType(label) +
		fnType(32, 65, 32, 65, 32) + fnType(32, 65, 32)
	api := func(id, typ int, name string) string {
		return engineCBCNumber(uint64(id)) + engineCBCNumber(uint64(typ)) + engineCBCString(name)
	}
	call := func(dest, id int, args ...string) string {
		return engineCBCNumber(32) + engineCBCNumber(uint64(dest)) + engineCBCFixed(33, 2) +
			engineCBCFixed(uint64(len(args)), 1) + engineCBCNumber(uint64(id)) + strings.Join(args, "")
	}
	ret := "T" + engineCBCFixed(19, 2) + engineCBCNumber(32) + engineCBCConstant(0, 32)
	match := call(0, 89, engineCBCConstant(2, 0), engineCBCConstant(uint64(len(group1)), 32),
		engineCBCConstant(4, 0), engineCBCConstant(uint64(len(group2)), 32))
	equals := engineCBCNumber(1) + engineCBCNumber(1) + engineCBCFixed(21, 2) +
		engineCBCNumber(32) + engineCBCNumber(0) + engineCBCConstant(want, 32)
	branch := "T" + engineCBCFixed(17, 2) + engineCBCNumber(1) + engineCBCNumber(1) + engineCBCNumber(2)
	setName := call(2, 5, engineCBCConstant(6, 0), engineCBCConstant(uint64(len(label)), 32))
	declaration := "A" + engineCBCFixed(0, 1) + engineCBCNumber(32) + "L" + engineCBCNumber(3) +
		engineCBCNumber(32) + engineCBCFixed(0, 1) + engineCBCNumber(1) + engineCBCFixed(0, 1) +
		engineCBCNumber(32) + engineCBCFixed(0, 1) + "F" + engineCBCNumber(6) + engineCBCNumber(3)
	lines := []string{
		header,
		"Bytecode.Icon.{Detected};Target:1;0;" + hex.EncodeToString([]byte(marker)),
		types,
		"E" + engineCBCNumber(89) + engineCBCNumber(2) + api(5, 73, "setvirusname") + api(89, 72, "matchicon"),
		"G" + engineCBCNumber(0) + engineCBCNumber(7) + globals,
		declaration,
		"B" + match + equals + branch,
		"B" + setName + ret,
		"B" + ret + "E",
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func TestBytecodeMatchIconAPICallContract(t *testing.T) {
	for _, tc := range []struct {
		want   uint64
		detect bool
	}{{0, true}, {1, false}} {
		p, err := bytecode.Parse(matchIconOracleProgram(256, "KNOWN", "GROUP", tc.want))
		if err != nil {
			t.Fatal(err)
		}
		matches := bytecode.EmptyMatchContext()
		matches.PE = &bytecode.PEInfo{EntryPoint: 0x200}
		got, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ harmless"), matches, 1000)
		if err != nil || (got.VirusName == "Bytecode.Icon.Detected") != tc.detect {
			t.Fatalf("API branch %d: %+v, %v", tc.want, got, err)
		}
	}
	invalid, err := bytecode.Parse(matchIconOracleProgram(256, "KNOWN", "GROUP", 0xffffffff))
	if err != nil {
		t.Fatal(err)
	}
	matches := bytecode.EmptyMatchContext()
	matches.PE = &bytecode.PEInfo{}
	got, err := invalid.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ harmless"), matches, 1000)
	if err != nil || got.VirusName != "Bytecode.Icon.Detected" {
		t.Fatalf("no mapped entry point: %+v, %v", got, err)
	}
	oversize, err := bytecode.Parse(matchIconOracleProgram(256, strings.Repeat("A", 128), "GROUP", 0xffffffff))
	if err != nil {
		t.Fatal(err)
	}
	matches.PE.EntryPoint = 0x200
	got, err = oversize.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ harmless"), matches, 1000)
	if err != nil || got.VirusName != "Bytecode.Icon.Detected" {
		t.Fatalf("overlong group name: %+v, %v", got, err)
	}
}

func TestBytecodeMatchIconKnownFalseInScanner(t *testing.T) {
	const marker = "UTMSTACK-MATCHICON-HARMLESS-TRIGGER"
	for _, tc := range []struct {
		want   uint64
		detect bool
	}{{0, true}, {1, false}} {
		e, _ := engineCBCActivate(t, matchIconOracleProgram(256, "KNOWN", "GROUP", tc.want))
		pe := simplePE(append([]byte{0xc3}, []byte(marker)...)).build()
		clean, label, err := e.Scan(context.Background(), pe)
		if err != nil || clean == tc.detect || tc.detect && label != "Bytecode.Icon.Detected" || !tc.detect && label != "" {
			t.Fatalf("want %d: clean=%t label=%q err=%v", tc.want, clean, label, err)
		}
	}
}

func TestExportHarmlessMatchIconOracle(t *testing.T) {
	dir := os.Getenv("EDR_MATCHICON_ORACLE_DIR")
	if dir == "" {
		t.Skip("EDR_MATCHICON_ORACLE_DIR not set")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, group string
		kind        uint64
		want        uint64
	}{
		{"known-zero", "KNOWN", 256, 0}, {"known-one", "KNOWN", 256, 1}, {"known-minus-one", "KNOWN", 256, 0xffffffff},
		{"missing-zero", "MISSING", 256, 0}, {"missing-one", "MISSING", 256, 1}, {"missing-minus-one", "MISSING", 256, 0xffffffff},
		{"any-zero", "", 256, 0}, {"any-one", "", 256, 1},
		{"oversize-minus-one", strings.Repeat("A", 128), 256, 0xffffffff},
		{"hook-known-zero", "KNOWN", 259, 0}, {"hook-known-one", "KNOWN", 259, 1}, {"hook-known-minus-one", "KNOWN", 259, 0xffffffff},
		{"hook-missing-zero", "MISSING", 259, 0}, {"hook-missing-one", "MISSING", 259, 1}, {"hook-missing-minus-one", "MISSING", 259, 0xffffffff},
		{"hook-any-zero", "", 259, 0}, {"hook-any-one", "", 259, 1},
	} {
		group2 := "GROUP"
		if tc.group == "" {
			group2 = ""
		}
		if err := os.WriteFile(filepath.Join(dir, tc.name+".cbc"), matchIconOracleProgram(tc.kind, tc.group, group2, tc.want), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote harmless matchicon programs to %s", dir)
}
