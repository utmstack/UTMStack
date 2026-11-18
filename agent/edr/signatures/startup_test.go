package signatures

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func startupFixture(value uint64, abort bool) []byte {
	body := []string{callbackReturn(32, value)}
	if abort {
		body = append([]string{callbackInst(0, 0, 45)}, body...)
	}
	return (callbackFixture{kind: 1, result: 32, blocks: [][]string{body}}).encode()
}

func startupControlFixture(apiID, cond uint64) []byte {
	const reason = "inert test reason"
	values := make([]uint64, len(reason)+1)
	for i := range reason {
		values[i] = uint64(reason[i])
	}
	name := "disable_bytecode_if"
	if apiID == 74 {
		name = "disable_jit_if"
	}
	return (callbackFixture{
		kind: 1, result: 32,
		types:   []string{callbackType(4, len(values), 8), callbackType(1, 4, 32, 65, 32, 32)},
		globals: []string{callbackGlobal(69, values...), callbackGlobal(65, 0, 1)},
		apiMax:  apiID, apis: []string{callbackAPI(int(apiID), 70, name)},
		locals: []int{32}, blocks: [][]string{{
			callbackCall(0, int(apiID), callbackGlobalRef(2), callbackConstant(uint64(len(reason)), 32), callbackConstant(cond, 32)),
			callbackReturn(32, 0xda7aba5e),
		}},
	}).encode()
}

func TestStartupControlSuppressesOnlyBytecode(t *testing.T) {
	logical := callbackWithLabel([]byte("UTM.Startup.Callback"), true)
	for _, tc := range []struct {
		name     string
		apiID    uint64
		cond     uint64
		disabled bool
	}{
		{"jit-disabled", 74, 1, false},
		{"bytecode-disabled", 73, 1, true},
		{"condition-false", 73, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Compile(map[string][]byte{
				"startup.cbc": startupControlFixture(tc.apiID, tc.cond),
				"logical.cbc": logical.encode(),
				"kept.ndb":    []byte("Kept.Rule:0:*:4d4e4f505152\n"),
			})
			if err != nil || d.bytecodeDisabled != tc.disabled {
				t.Fatalf("startup compile: disabled=%t, err=%v", d.bytecodeDisabled, err)
			}
			result, err := d.startupHooks[0].RunWithInput(context.Background(), 0, nil, nil, 10000)
			if err != nil || result.BytecodeDisableStatus != map[uint64]uint32{73: 2, 74: 1}[tc.apiID]*uint32(tc.cond) {
				t.Fatalf("startup result: %+v, %v", result, err)
			}
			want := "UTM.Startup.Callback"
			if tc.disabled {
				want = ""
			}
			expectOutcome(t, d, "xxAByy", anyType, want, "")
			expectOutcome(t, d, "AB"+strings.Repeat("X", 1024), anyType, want, "")
			assertMatch(t, d, []byte("MNOPQR"), "Kept.Rule")
		})
	}
}

func TestStartupFullDisableSuppressesUnsupportedBytecode(t *testing.T) {
	unsupported := callbackWithLabel([]byte("UTM.Unused"), true)
	unsupported.apiMax = 53
	unsupported.apis = append(unsupported.apis, callbackAPI(53, 70, "atoi"))
	for _, cond := range []uint64{0, 1} {
		d, err := Compile(map[string][]byte{
			"startup.cbc": startupControlFixture(73, cond),
			"unknown.cbc": unsupported.encode(),
			"kept.ndb":    []byte("Kept.Rule:0:*:4d4e4f505152\n"),
		})
		if err != nil {
			t.Fatal(err)
		}
		cause := UnknownBytecode
		if cond != 0 {
			cause = ""
		}
		expectOutcome(t, d, "xxAByy", anyType, "", cause)
		assertMatch(t, d, []byte("MNOPQR"), "Kept.Rule")
	}
}

func TestStartupDisableModeAcrossCVDRefresh(t *testing.T) {
	logical := callbackWithLabel([]byte("UTM.Startup.Refresh"), true)
	members := []member{
		{"startup.cbc", startupControlFixture(73, 1)},
		{"logical.cbc", logical.encode()},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil || !prior.bytecodeDisabled {
		t.Fatalf("disabled prior: %v", err)
	}
	members[0].data = startupControlFixture(73, 0)
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if next.bytecodeDisabled || !prior.bytecodeDisabled {
		t.Fatal("startup mode leaked between generations")
	}
	expectOutcome(t, prior, "xxAByy", anyType, "", "")
	expectOutcome(t, next, "xxAByy", anyType, "UTM.Startup.Refresh", "")
	members[0].data = startupControlFixture(73, 1)
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	back := compareBytecodeRefresh(t, files, next)
	if !back.bytecodeDisabled || next.bytecodeDisabled {
		t.Fatal("startup mode was not reapplied")
	}
	expectOutcome(t, back, "xxAByy", anyType, "", "")
}

// Optional harmless database fixtures for a pinned native ClamAV CLI run.
func TestWriteStartupControlOracle(t *testing.T) {
	root := os.Getenv("EDR_STARTUP_ORACLE_DIR")
	if root == "" {
		t.Skip("set EDR_STARTUP_ORACLE_DIR for the native ClamAV oracle")
	}
	logical := callbackWithLabel([]byte("UTM.Startup.Oracle"), true)
	for _, tc := range []struct {
		name  string
		apiID uint64
		cond  uint64
	}{
		{"disable-all", 73, 1},
		{"disable-jit", 74, 1},
		{"condition-false", 73, 0},
	} {
		dir := filepath.Join(root, tc.name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		for name, data := range map[string][]byte{
			"startup.cbc": startupControlFixture(tc.apiID, tc.cond),
			"logical.cbc": logical.encode(),
			"kept.ndb":    []byte("Kept.Rule:0:*:4d4e4f505152\n"),
		} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.WriteFile(filepath.Join(root, "sample-ab.bin"), append([]byte("AB"), bytes.Repeat([]byte{'X'}, 1024)...), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sample-static.bin"), append([]byte("MNOPQR"), bytes.Repeat([]byte{'X'}, 1024)...), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestStartupBytecodeRunsFirstLoadedProgram(t *testing.T) {
	files := map[string][]byte{
		"first.cbc":  startupFixture(0xda7aba5e, false),
		"second.cbc": startupFixture(0, true),
	}
	d, err := Compile(files)
	if err != nil || d.Count() != 2 || d.DetectionCount() != 0 || len(d.startupHooks) != 2 {
		t.Fatalf("startup compile: %+v, %v", d, err)
	}
	// A mismatched self-test value is a warning in ClamAV's normal mode.
	files["first.cbc"] = startupFixture(7, false)
	if _, err := Compile(files); err != nil {
		t.Fatalf("non-magic startup return rejected: %v", err)
	}
	files["first.cbc"] = startupFixture(0, true)
	if _, err := Compile(files); err == nil || !strings.Contains(err.Error(), "startup bytecode") {
		t.Fatalf("failed first startup accepted: %v", err)
	}
}

func TestStartupBytecodeContainerRefreshRejectsFailedReplacement(t *testing.T) {
	members := []member{
		{"startup.cbc", startupFixture(0xda7aba5e, false)},
		{"kept.ndb", []byte("Kept.Rule:0:*:4d4e4f505152\n")},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	prior, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	members[0].data = startupFixture(0, true)
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	if _, handled, err := tryContainerSourceRefresh(files, DefaultLimits(), compileOptions{prior: prior}); !handled || err == nil || !strings.Contains(err.Error(), "startup bytecode") {
		t.Fatalf("failed startup refresh: handled=%v err=%v", handled, err)
	}
	if err := prior.prepareStartup(); err != nil {
		t.Fatalf("prior generation changed: %v", err)
	}
	members[0].data = startupFixture(7, false)
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next := compareBytecodeRefresh(t, files, prior)
	if len(next.startupHooks) != 1 || next.startupHooks[0] == prior.startupHooks[0] {
		t.Fatal("startup program was not replaced")
	}
}
