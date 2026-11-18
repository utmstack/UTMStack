package signatures

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// Independently encode harmless public ClamBC wire programs. These fixtures do
// not use the parser, VM instructions, or callback implementation as builders.
type callbackFixture struct {
	kind, minimum, maximum uint64
	logical                string
	types, globals, apis   []string
	globalMax, apiMax      uint64
	args, locals           []int
	result                 int
	blocks                 [][]string
}

func callbackNum(v uint64) string {
	var b []byte
	for v != 0 {
		b = append(b, 0x60|byte(v&15))
		v >>= 4
	}
	return string(append([]byte{0x60 + byte(len(b))}, b...))
}
func callbackFixed(v uint64, width int) string {
	b := make([]byte, width)
	for i := range b {
		b[i] = 0x60 | byte(v&15)
		v >>= 4
	}
	return string(b)
}
func callbackConstant(v uint64, bits int) string {
	b := []byte(callbackNum(v))
	b[0] &^= 0x20
	return string(b) + callbackFixed(uint64(bits/8), 1)
}
func callbackGlobalRef(id int) string { return callbackConstant(uint64(id), 0) }
func callbackType(kind, count int, elements ...int) string {
	s := callbackFixed(uint64(kind), 1)
	if kind != 5 {
		s += callbackNum(uint64(count))
	}
	for _, id := range elements {
		s += callbackNum(uint64(id))
	}
	return s
}
func callbackGlobal(typ int, values ...uint64) string {
	s := callbackNum(uint64(typ))
	for _, v := range values {
		b := []byte(callbackNum(v))
		b[0] &^= 0x20
		s += string(b)
	}
	return s + "`"
}
func callbackAPI(id, typ int, name string) string {
	s := callbackNum(uint64(id)) + callbackNum(uint64(typ)) + "|" + callbackNum(uint64(len(name)+1))
	for _, b := range append([]byte(name), 0) {
		s += callbackFixed(uint64(b), 2)
	}
	return s
}
func callbackInst(typ, dest, op int, args ...string) string {
	return callbackNum(uint64(typ)) + callbackNum(uint64(dest)) + callbackFixed(uint64(op), 2) + strings.Join(args, "")
}
func callbackCall(dest, api int, args ...string) string {
	fields := []string{callbackFixed(uint64(len(args)), 1), callbackNum(uint64(api))}
	return callbackInst(32, dest, 33, append(fields, args...)...)
}
func callbackReturn(width int, value uint64) string {
	return "T" + callbackFixed(19, 2) + callbackNum(uint64(width)) + callbackConstant(value, width)
}
func (f callbackFixture) encode() []byte {
	header := "ClamBC" + callbackNum(7) + callbackNum(0) + "|" + callbackNum(0) + callbackNum(0) + callbackNum(f.kind) + callbackNum(f.minimum) + callbackNum(f.maximum) + callbackNum(0) + "|" + callbackNum(0) + callbackNum(uint64(5+len(f.types))) + callbackNum(1) + callbackNum(0x53e5493e9f3d1c30) + callbackFixed(42, 2) + ":65536"
	lines := []string{
		header, f.logical, "T" + callbackFixed(69, 2) + strings.Join(f.types, ""),
		"E" + callbackNum(f.apiMax) + callbackNum(uint64(len(f.apis))) + strings.Join(f.apis, ""),
		"G" + callbackNum(f.globalMax) + callbackNum(uint64(len(f.globals)+1)) + callbackNum(0) + "@`" + strings.Join(f.globals, ""),
	}
	declaration := "A" + callbackFixed(uint64(len(f.args)), 1) + callbackNum(uint64(f.result)) + "L" + callbackNum(uint64(len(f.locals)))
	for _, group := range [][]int{f.args, f.locals} {
		for _, typ := range group {
			declaration += callbackNum(uint64(typ&0x7fff)) + callbackFixed(uint64(typ>>15), 1)
		}
	}
	count := 0
	for _, block := range f.blocks {
		count += len(block)
	}
	lines = append(lines, declaration+"F"+callbackNum(uint64(count))+callbackNum(uint64(len(f.blocks))))
	for i, block := range f.blocks {
		line := "B" + strings.Join(block, "")
		if i == len(f.blocks)-1 {
			line += "E"
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func callbackWithLabel(label []byte, set bool) callbackFixture {
	values := make([]uint64, len(label)+1)
	for i, b := range label {
		values[i] = uint64(b)
	}
	body := []string{callbackReturn(32, 0xf00d)}
	if set {
		body = append([]string{callbackCall(0, 5, callbackGlobalRef(2), callbackConstant(uint64(len(label)), 32))}, body...)
	}
	return callbackFixture{
		kind: 256, logical: "Callback.Trigger;Engine:51-255,Target:0;0;4142", result: 32,
		types:   []string{callbackType(4, len(values), 8), callbackType(1, 3, 32, 65, 32)},
		globals: []string{callbackGlobal(69, values...), callbackGlobal(65, 0, 1)},
		apiMax:  5, apis: []string{callbackAPI(5, 70, "setvirusname")}, locals: []int{32}, blocks: [][]string{body},
	}
}
func compileCallback(t *testing.T, f callbackFixture) *Database {
	t.Helper()
	d, err := Compile(map[string][]byte{"test.cbc": f.encode()})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestLogicalBytecodeRunsAfterBodyAndWholeHash(t *testing.T) {
	// The pinned ClamAV oracle gives a static .ndb match precedence over a
	// whole-file .hdb match and a logical ClamBC callback; with the .ndb
	// absent, the .hdb wins over the callback.
	input := []byte("xxABharmless")
	f := callbackWithLabel([]byte("Callback.Hit"), true)
	h := md5.Sum(input)
	files := map[string][]byte{
		"a.cbc": f.encode(),
		"b.hdb": []byte(fmt.Sprintf("%x:%d:Hash.Hit\n", h, len(input))),
		"z.ndb": []byte("Body.Hit:0:*:4142\n"),
	}
	check := func(want string) {
		t.Helper()
		d, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := d.Match(context.Background(), input); got != want || err != nil {
			t.Fatalf("Match = %q, %v; want %q", got, err, want)
		}
	}
	check("Body.Hit")
	delete(files, "z.ndb")
	check("Hash.Hit")
	delete(files, "b.hdb")
	check("Callback.Hit")
}

func TestExtractNewIsIncompleteOnlyOnReachedCall(t *testing.T) {
	f := callbackWithLabel([]byte("unused"), false)
	f.logical = "Extract.Test;Engine:51-255,Target:0;0;4142"
	f.types = append(f.types, callbackType(1, 2, 32, 32))
	f.apiMax, f.apis = 22, []string{callbackAPI(22, 71, "extract_new")}
	f.blocks = [][]string{{callbackReturn(32, 0)}}
	silent := compileCallback(t, f)
	if got, err := silent.Match(context.Background(), []byte("xxAByy")); got != "" || err != nil {
		t.Fatalf("unused extract_new: %q, %v", got, err)
	}
	f.blocks = [][]string{{callbackCall(0, 22, callbackConstant(0xffffffff, 32)), callbackReturn(32, 0)}}
	calling := compileCallback(t, f)
	if got, err := calling.Match(context.Background(), []byte("xxAByy")); got != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("reached extract_new: %q, %v", got, err)
	}
	if got, err := calling.Match(context.Background(), []byte("nothing")); got != "" || err != nil {
		t.Fatalf("unreached trigger: %q, %v", got, err)
	}
}

func TestBytecodeOptionQueryUsesEnclosingScanProfile(t *testing.T) {
	const expected = 0x01406037
	f := callbackWithLabel([]byte("UTM.Options.Match"), true)
	f.types = append(f.types, callbackType(1, 1, 32))
	f.apiMax = 68
	f.apis = append(f.apis, callbackAPI(68, 71, "engine_scan_options"))
	f.locals = []int{32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 68),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(expected, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	d := compileCallback(t, f)
	input := []byte("xxAByy")
	match := func(profile *bytecode.EngineOptions) (string, error) {
		result, err := d.MatchObject(context.Background(), input, MatchContext{Type: TypeAny, BytecodeOptions: profile})
		return result.Name, err
	}
	if name, err := match(nil); name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("missing profile: %q, %v", name, err)
	}
	if name, err := match(&bytecode.EngineOptions{Scan: expected}); name != "UTM.Options.Match" || err != nil {
		t.Fatalf("matching profile: %q, %v", name, err)
	}
	if name, err := match(&bytecode.EngineOptions{Scan: expected &^ 0x20}); name != "" || err != nil {
		t.Fatalf("different profile: %q, %v", name, err)
	}
}

func TestBytecodeJSONQueryUsesScannerMetadata(t *testing.T) {
	f := callbackWithLabel([]byte("FileType"), true)
	f.types = append(f.types, callbackType(1, 4, 32, 65, 32, 32), callbackType(1, 2, 32, 32))
	f.apiMax = 97
	f.apis = append(f.apis, callbackAPI(93, 71, "json_get_object"), callbackAPI(97, 72, "json_get_string_length"))
	f.locals = []int{32, 32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 93, callbackGlobalRef(2), callbackConstant(8, 32), callbackConstant(0, 32)),
		callbackCall(1, 97, callbackNum(0)),
		callbackInst(1, 2, 21, callbackNum(32), callbackNum(1), callbackConstant(16, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(2) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	d := compileCallback(t, f)
	input := []byte("xxAByy")
	match := func(metadata map[string]any) (string, error) {
		res, err := d.MatchObject(context.Background(), input, MatchContext{Type: TypeAny, Metadata: metadata})
		return res.Name, err
	}
	if name, err := match(nil); name != "" || err != nil {
		t.Fatalf("missing metadata: %q, %v", name, err)
	}
	if name, err := match(map[string]any{"FileType": "CL_TYPE_XML_WORD"}); name != "FileType" || err != nil {
		t.Fatalf("matched metadata: %q, %v", name, err)
	}
	if name, err := match(map[string]any{"FileType": "OTHER"}); name != "" || err != nil {
		t.Fatalf("other metadata: %q, %v", name, err)
	}
	if res, err := d.MatchObject(context.Background(), input, MatchContext{Type: TypeAny, Metadata: map[string]any{}, MetadataPartial: true}); res.Name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("absent partial metadata: %+v, %v", res, err)
	}
}

func TestBytecodeJSONActiveMatchesReferenceCollectionOption(t *testing.T) {
	f := callbackWithLabel([]byte("UTM.JSON.Active"), true)
	f.types = append(f.types, callbackType(1, 1, 32))
	f.apiMax = 92
	f.apis = append(f.apis, callbackAPI(92, 71, "json_is_active"))
	f.locals = []int{32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 92),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(1, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	input := []byte(`<?xml version="1.0"?><WordDocument>AB</WordDocument>`)
	d := compileCallback(t, f)
	for _, tc := range []struct {
		metadata map[string]any
		want     string
	}{{nil, ""}, {map[string]any{}, "UTM.JSON.Active"}} {
		got, err := d.MatchObject(context.Background(), input, MatchContext{Type: TypeXMLWord, Metadata: tc.metadata})
		if err != nil || got.Name != tc.want {
			t.Fatalf("metadata=%v: %+v, %v; want %q", tc.metadata, got, err, tc.want)
		}
	}
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "active.cbc"), f.encode(), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "document.xml")
	if err := os.WriteFile(path, input, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		collect bool
		found   bool
	}{{false, false}, {true, true}} {
		args := []string{"--database=" + dir, "--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary"}
		if tc.collect {
			args = append(args, "--gen-json")
		}
		args = append(args, path)
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), args...)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if tc.found {
			if code, ok := err.(*exec.ExitError); !ok || code.ExitCode() != 1 || !strings.Contains(string(out), "UTM.JSON.Active") {
				t.Fatalf("reference collect=%t: %v: %s", tc.collect, err, out)
			}
		} else if err != nil || strings.Contains(string(out), " FOUND") {
			t.Fatalf("reference collect=%t: %v: %s", tc.collect, err, out)
		}
	}
}

func TestBytecodeChildDetectionWinsAfterLaterFault(t *testing.T) {
	child := []byte("HARMLESS-CHILD")
	f := callbackWithLabel(child, false)
	f.types = append(f.types, callbackType(1, 2, 32, 32))
	f.apiMax = 22
	f.apis = []string{callbackAPI(3, 70, "write"), callbackAPI(22, 71, "extract_new")}
	f.blocks = [][]string{{
		callbackCall(0, 3, callbackGlobalRef(2), callbackConstant(uint64(len(child)), 32)),
		callbackCall(0, 22, callbackConstant(7, 32)),
		"T" + callbackFixed(45, 2), // a fault after the child was detected
	}}
	d := compileCallback(t, f)
	var calls int
	res, err := d.MatchObject(context.Background(), []byte("xxAByy"), MatchContext{Type: TypeAny, ScanExtracted: func(data []byte, id int32) (string, error) {
		calls++
		if string(data) != string(child) || id != -1 {
			t.Errorf("child output %q, id %d", data, id)
		}
		return "Child.Callback.Hash", nil
	}})
	if err != nil || res.Name != "Child.Callback.Hash" || calls != 1 {
		t.Fatalf("child detection: %+v, %v, calls %d", res, err, calls)
	}
	if got, err := d.Match(context.Background(), []byte("xxAByy")); got != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("standalone matcher without child scanner: %q, %v", got, err)
	}
}

type callbackCheck struct {
	target uint64
	index  int
	want   uint32
}

func callbackWithChecks(logical string, checks ...callbackCheck) callbackFixture {
	f := callbackWithLabel([]byte("UTMStack.Callback.Match"), true)
	f.logical, f.globalMax = logical, 0x8005
	f.locals = []int{32, 32, 1, 1}
	body := []string{callbackInst(0, 0, 34, callbackConstant(1, 8), callbackNum(3))}
	for _, check := range checks {
		f.globals = append(f.globals, callbackGlobal(67, uint64(check.index*4), check.target))
		body = append(body,
			callbackInst(32, 1, 39, callbackGlobalRef(len(f.globals))),
			callbackInst(1, 2, 21, callbackNum(32), callbackNum(1), callbackConstant(uint64(check.want), 32)),
			callbackInst(1, 3, 11, callbackNum(3), callbackNum(2)))
	}
	body = append(body, "T"+callbackFixed(17, 2)+callbackNum(3)+callbackNum(1)+callbackNum(2))
	f.blocks = [][]string{body, f.blocks[0], {callbackReturn(32, 0)}}
	return f
}

func TestBytecodeCallbackRequiresExplicitDetectionLabel(t *testing.T) {
	for _, set := range []bool{false, true} {
		f := callbackWithLabel([]byte("UTMStack.Full.Callback.Label"), set)
		d := compileCallback(t, f)
		want := ""
		if set {
			want = "UTMStack.Full.Callback.Label"
		}
		got, err := d.Match(context.Background(), []byte("AB"))
		if err != nil || got != want {
			t.Fatalf("set=%t: Match = %q, %v; want %q, nil", set, got, err, want)
		}
		if d.Count() != 1 || d.DetectionCount() != 1 {
			t.Fatalf("logical bytecode counted incorrectly: %d/%d", d.Count(), d.DetectionCount())
		}
	}
}

func TestBytecodeCallbackGetsFullCountsAndOffsets(t *testing.T) {
	for _, tc := range []struct {
		name, logical, data string
		checks              []callbackCheck
	}{
		{"boolean-trigger-full-overlap-count", "Callback;Target:0;0;4141", "AAAA", []callbackCheck{{0x8000, 0, 3}, {0x8005, 0, 0}}},
		{"leading-wildcard-start", "Callback;Target:0;0;??4142", "zzxABxAB", []callbackCheck{{0x8000, 0, 2}, {0x8005, 0, 2}}},
		{"leading-gap-start", "Callback;Target:0;0;{1-2}4142", "xyzAB", []callbackCheck{{0x8000, 0, 2}, {0x8005, 0, 1}}},
		{"absent-anchored", "Callback;Target:0;0&1=0;4142;ffff", "zzAB", []callbackCheck{{0x8000, 0, 1}, {0x8005, 0, 2}, {0x8000, 1, 0}, {0x8005, 1, 0xfffffffe}}},
		{"absent-unanchored", "Callback;Target:0;0&1=0;4142;f?ff", "zzAB", []callbackCheck{{0x8000, 1, 0}, {0x8005, 1, 0xfffffffe}}},
		{"unused-array-slot", "Callback;Target:0;0;4142", "AB", []callbackCheck{{0x8000, 63, 0}, {0x8005, 63, 0xfffffffe}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := compileCallback(t, callbackWithChecks(tc.logical, tc.checks...))
			got, err := d.Match(context.Background(), []byte(tc.data))
			if err != nil || got != "UTMStack.Callback.Match" {
				t.Fatalf("callback metadata mismatch: %q, %v", got, err)
			}
		})
	}
	// The callback's count predicate must also reject a triggered input with
	// a different full count; evaluating the LDB alone is insufficient.
	d := compileCallback(t, callbackWithChecks("Callback;Target:0;0;4141", callbackCheck{0x8000, 0, 3}))
	if got, err := d.Match(context.Background(), []byte("AAA")); err != nil || got != "" {
		t.Fatalf("callback predicate ignored: %q, %v", got, err)
	}
}

func TestBytecodeCallbackRejectsUnsafeDetectionLabels(t *testing.T) {
	for _, label := range [][]byte{
		{}, []byte("   "), {0xff, 'x'}, []byte("bad\nname"), []byte("bad\x7fname"), []byte("bad\u2028name"),
		[]byte("PUA.Test"), []byte("Heuristics.Test"), []byte("BC.Heuristics.Test"), []byte("Weak.Test"),
	} {
		t.Run(fmt.Sprintf("%x", label), func(t *testing.T) {
			d := compileCallback(t, callbackWithLabel(label, true))
			got, err := d.Match(context.Background(), []byte("AB"))
			if err == nil || got != "" {
				t.Fatalf("unsafe label returned a verdict: %q, %v", got, err)
			}
			if strings.HasPrefix(string(label), "PUA.") || strings.HasPrefix(string(label), "Heuristics.") || strings.HasPrefix(string(label), "BC.Heuristics.") || strings.HasPrefix(string(label), "Weak.") {
				if !errors.Is(err, bytecode.ErrUnsupported) {
					t.Fatalf("classification error = %v; want ErrUnsupported", err)
				}
			}
		})
	}
}

func TestBytecodeCallbackRuntimeFailureIsIncomplete(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.Must.Not.Escape"), true)
	f.blocks[0] = append(f.blocks[0][:1], callbackInst(0, 0, 45), callbackReturn(32, 0))
	d := compileCallback(t, f)
	if got, err := d.Match(context.Background(), []byte("AB")); got != "" || !errors.Is(err, bytecode.ErrRuntime) {
		t.Fatalf("partial label after VM failure: %q, %v", got, err)
	}
	if got, err := d.Match(context.Background(), []byte("no trigger")); got != "" || err != nil {
		t.Fatalf("untriggered callback ran: %q, %v", got, err)
	}
	f = callbackWithLabel(nil, false)
	f.blocks = [][]string{{"T" + callbackFixed(18, 2) + callbackNum(1)}, {"T" + callbackFixed(18, 2) + callbackNum(1)}}
	d = compileCallback(t, f)
	if got, err := d.Match(context.Background(), []byte("AB")); got != "" || !errors.Is(err, bytecode.ErrLimit) {
		t.Fatalf("unbounded callback returned clean: %q, %v", got, err)
	}
	f = callbackWithLabel([]byte("UTMStack.Budget"), true)
	d, err := CompileWithLimits(map[string][]byte{"test.cbc": f.encode()}, Limits{MaxMatchSteps: 16})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := d.Match(context.Background(), []byte("AB")); got != "" || err == nil {
		t.Fatalf("shared work budget returned a verdict: %q, %v", got, err)
	}
}

// Programs the reference would load and run but this engine cannot run load
// as unknown when their trigger holds (TestUnsupportedBytecodeKinds...);
// malformed programs still reject the whole generation.
func TestBytecodeCallbackCompileRejectsUnsupportedGeneration(t *testing.T) {
	cases := map[string]func(*callbackFixture){
		"missing-trigger":     func(f *callbackFixture) { f.logical = "" },
		"malformed-trigger":   func(f *callbackFixture) { f.logical = "not an LDB" },
		"invalid-expression":  func(f *callbackFixture) { f.logical = "Bad;Target:0;1;4142" },
		"unselected-target":   func(f *callbackFixture) { f.logical = "Bad;Target:15;0;4142" },
		"ldb-functionality":   func(f *callbackFixture) { f.logical = "Bad;Engine:1-20,Target:0;0;4142" },
		"wrong-api-prototype": func(f *callbackFixture) { f.types[1] = callbackType(1, 3, 32, 66, 32) },
	}
	previous := compileCallback(t, callbackWithLabel([]byte("UTMStack.Previous"), true))
	version := previous.Version()
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := callbackWithLabel([]byte("UTMStack.Rejected"), true)
			mutate(&f)
			d, err := Compile(map[string][]byte{"a-valid.ndb": []byte("Valid:0:*:4142"), "z-unsupported.cbc": f.encode()})
			if err == nil || d != nil {
				t.Fatalf("partially accepted generation: database=%v error=%v", d, err)
			}
			var diagnostic *CompileError
			if !errors.As(err, &diagnostic) || len(diagnostic.Diagnostics) == 0 {
				t.Fatalf("missing compilation diagnostics: %v", err)
			}
			got, err := previous.Match(context.Background(), []byte("AB"))
			if err != nil || got != "UTMStack.Previous" || previous.Version() != version {
				t.Fatalf("rejected generation changed prior database: %q, %v", got, err)
			}
		})
	}
}

func TestBytecodeCallbackConcurrentInputIsolation(t *testing.T) {
	f := callbackFixture{
		kind: 256, logical: "Input.Trigger;Target:0;0;00", result: 32,
		types:  []string{callbackType(4, 64, 8), callbackType(1, 3, 32, 65, 32)},
		apiMax: 5, apis: []string{callbackAPI(2, 70, "read"), callbackAPI(5, 70, "setvirusname")},
		locals: []int{69 | 0x8000, 32}, blocks: [][]string{{
			callbackCall(1, 2, callbackNum(0), callbackConstant(64, 32)),
			callbackCall(1, 5, callbackNum(0), callbackConstant(0, 32)), callbackReturn(32, 0),
		}},
	}
	d := compileCallback(t, f)
	const workers = 24
	start := make(chan struct{})
	failures := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			for run := 0; run < 8; run++ {
				want := fmt.Sprintf("UTMStack.Input.%02d.%d", id, run)
				input := []byte(want + "\x00payload")
				if run%3 == 0 {
					input, want = []byte("untriggered input"), ""
				}
				ctx, cancel := context.WithCancel(context.Background())
				canceled := id%5 == 0 && run%2 == 0
				if canceled {
					cancel()
				}
				got, err := d.Match(ctx, input)
				cancel()
				if canceled {
					if got != "" || !errors.Is(err, context.Canceled) {
						failures <- fmt.Errorf("worker %d cancellation: %q, %v", id, got, err)
						return
					}
				} else if err != nil || got != want {
					failures <- fmt.Errorf("worker %d run %d: %q, %v; want %q", id, run, got, err, want)
					return
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

func TestUnsupportedBytecodeKindsLoadAsUnknownWhenReached(t *testing.T) {
	cases := map[string]func(*callbackFixture){
		"generic-kind":    func(f *callbackFixture) { f.kind = 0 },
		"other-hook":      func(f *callbackFixture) { f.kind = 258 },
		"entry-arguments": func(f *callbackFixture) { f.args = []int{32} },
		"entry-result":    func(f *callbackFixture) { f.result = 64; f.blocks = [][]string{{callbackReturn(64, 0)}} },
		"unused-unknown-api": func(f *callbackFixture) {
			f.apiMax = 53
			f.apis = append(f.apis, callbackAPI(53, 70, "atoi"))
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := callbackWithLabel([]byte("UTMStack.Unknown"), true)
			mutate(&f)
			d := compileCallback(t, f)
			if d.Report().Unknown[UnknownBytecode] != 1 {
				t.Fatalf("not counted: %v", d.Report().Unknown)
			}
			got, err := d.Match(context.Background(), []byte("xxAByy"))
			if got != "" || unknownReason(err) != UnknownBytecode {
				t.Fatalf("trigger holds: %q, %v", got, err)
			}
			if got, err := d.Match(context.Background(), []byte("nothing")); got != "" || err != nil {
				t.Fatalf("trigger false: %q, %v", got, err)
			}
		})
	}
}
