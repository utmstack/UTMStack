package signatures

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

const hookTrigger = "Hook.Test;Engine:51-255,Target:1;0;4142"

// detectingHook sets the name Hook.Label when it runs.
func detectingHook(kind uint64) callbackFixture {
	f := callbackWithLabel([]byte("Hook.Label"), true)
	f.kind, f.logical = kind, hookTrigger
	return f
}

// returningHook returns value and detects nothing.
func returningHook(kind uint64, value uint64) callbackFixture {
	f := callbackWithLabel([]byte("x"), false)
	f.kind, f.logical = kind, hookTrigger
	f.blocks = [][]string{{callbackReturn(32, value)}}
	return f
}

// breakIfEquals is an all-PE hook that returns the break value 0xcea5e when
// the 32 bits at offset of a special global equal want.
func breakIfEquals(global, offset, want uint64) callbackFixture {
	return callbackFixture{
		kind: 259, logical: hookTrigger, result: 32, globalMax: global,
		globals: []string{callbackGlobal(67, offset, global)}, locals: []int{32, 1, 32},
		blocks: [][]string{{
			callbackInst(32, 0, 39, callbackGlobalRef(1)),
			callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(want, 32)),
			callbackInst(32, 2, 31, callbackNum(1), callbackConstant(0xcea5e, 32), callbackConstant(0, 32)),
			"T" + callbackFixed(19, 2) + callbackNum(32) + callbackNum(2),
		}},
	}
}

func compileHooks(t *testing.T, programs ...callbackFixture) *Database {
	t.Helper()
	files := map[string][]byte{}
	for i, f := range programs {
		files[string(rune('a'+i))+".cbc"] = f.encode()
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// packedPE has the empty-section layout with a stream the NRV2B decoder
// reads to its end in the second section.
func packedPE() []byte {
	b := hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})
	copy(b[0x400:], nrv2bLiterals([]byte("a packed program, decoded to its end marker, and more than enough text to cover the entry point reads")))
	return b
}

// withTrigger places the trigger bytes "AB" in the header padding.
func withTrigger(b []byte) []byte {
	copy(b[0x300:], "AB")
	return b
}

func TestPEHookDispatch(t *testing.T) {
	plain := func() []byte { return withTrigger(hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})) }
	plus := plain()
	binary.LittleEndian.PutUint16(plus[152:], 0x20b)
	objects := map[string][]byte{
		"parsed":     plain(),
		"PE32+":      plus,
		"lone":       withTrigger(hookPE(0x800, 0x100, [4]uint32{0x400, 0x1000, 0x200, 0x1000})),
		"misaligned": withTrigger(hookPE(0x800, 0x1100, [4]uint32{0x400, 0x1100, 0x400, 0x400})),
		"empty":      withTrigger(hookPE(0x800, 0x2000, [4]uint32{0x1000, 0x1000, 0, 0}, [4]uint32{0x400, 0x2000, 0x400, 0x400})),
		"packed":     withTrigger(packedPE()),
		"short":      withTrigger(hookPE(0x800, 0x1000+0x400-124, [4]uint32{0x400, 0x1000, 0x400, 0x400})),
		"no trigger": hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400}),
		"aspack":     withTrigger(hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})),
	}
	copy(objects["aspack"][0x400:], []byte{0x60, 0xe8, 0x03, 0, 0, 0, 0xe9, 0xeb}) // ASPack's entry code
	type want struct {
		name       string
		incomplete string // "" clean or detected; otherwise the unknown reason, or "fail"
	}
	check := func(label string, d *Database, cases map[string]want) {
		t.Helper()
		for object, w := range cases {
			name, err := d.Match(nil2ctx(), objects[object])
			var u *UnknownError
			switch {
			case w.incomplete == "":
				if name != w.name || err != nil {
					t.Errorf("%s on %s: (%q, %v), want (%q, nil)", label, object, name, err, w.name)
				}
			case !errors.Is(err, ErrRuleIncomplete) || name != "":
				t.Errorf("%s on %s: (%q, %v), want incomplete", label, object, name, err)
			case w.incomplete != "fail" && (!errors.As(err, &u) || u.Reason != w.incomplete):
				t.Errorf("%s on %s: %v, want reason %s", label, object, err, w.incomplete)
			}
		}
	}
	check("all-PE hook", compileHooks(t, detectingHook(259)), map[string]want{
		"parsed": {name: "Hook.Label"}, "PE32+": {}, "lone": {}, "misaligned": {name: "Hook.Label"}, "no trigger": {},
		"empty": {name: "Hook.Label"}, "short": {name: "Hook.Label"},
	})
	check("unpacker hook", compileHooks(t, detectingHook(257)), map[string]want{
		"parsed": {name: "Hook.Label"}, "PE32+": {}, "lone": {}, "misaligned": {name: "Hook.Label"}, "no trigger": {},
		"empty": {name: "Hook.Label"}, "packed": {incomplete: UnknownPEUnpacker}, "short": {},
	})
	// The all-PE point comes first: its detection wins over an unknown
	// unpacker point.
	check("both hooks", compileHooks(t, detectingHook(257), detectingHook(259)), map[string]want{
		"packed": {name: "Hook.Label"},
	})
	// The break value from an all-PE hook ends the parser before the
	// unpacker hooks.
	check("break", compileHooks(t, returningHook(259, 0xcea5e), detectingHook(257)), map[string]want{
		"parsed": {}, "empty": {},
	})
	check("no break", compileHooks(t, returningHook(259, 7), detectingHook(257)), map[string]want{
		"parsed": {name: "Hook.Label"},
	})
	// A hook that fails at run time leaves the object incomplete.
	faulting := returningHook(259, 0)
	faulting.blocks = [][]string{{callbackInst(32, 0, 39, callbackGlobalRef(0)), callbackReturn(32, 0)}}
	check("failing hook", compileHooks(t, faulting), map[string]want{
		"parsed": {incomplete: "fail"}, "PE32+": {},
	})
	// Hooks see zero match counts and the full header record.
	check("zero counts", compileHooks(t, breakIfEquals(0x8000, 0, 0), detectingHook(257)), map[string]want{"parsed": {}})
	check("record magic", compileHooks(t, breakIfEquals(0x8003, 264, 0x10b), detectingHook(257)), map[string]want{"parsed": {}})
	check("record lfanew", compileHooks(t, breakIfEquals(0x8003, 632, 128), detectingHook(257)), map[string]want{"parsed": {}})
	// Without hook programs, a program the parser would hand to a built-in
	// unpacker this engine lacks is incomplete for that reason; the others
	// are decided.
	plainDB, err := Compile(map[string][]byte{"a.ndb": []byte("Unrelated.Rule:0:*:5a5a5a5a5a5a5a5a\n")})
	if err != nil {
		t.Fatal(err)
	}
	check("no hooks", plainDB, map[string]want{
		"aspack": {incomplete: UnknownPEUnpacker}, "parsed": {}, "empty": {}, "PE32+": {}, "short": {},
	})
	// A non-Windows object never enters the hook phase.
	d := compileHooks(t, detectingHook(259))
	if name, err := d.Match(nil2ctx(), []byte("plain text with AB in it")); name != "" || err != nil {
		t.Fatalf("text: (%q, %v)", name, err)
	}
}

func TestPEHookProgramsLoadAsHooks(t *testing.T) {
	d := compileHooks(t, detectingHook(257), detectingHook(259))
	if n := d.Report().Unknown[UnknownBytecode]; n != 0 {
		t.Fatalf("%d hook programs loaded as unknown", n)
	}
	// A hook declaring write loads too; only a run that writes is unknown
	// (TestProgramsThatWriteAreUnknownOnlyWhenTheyWrite).
	writer := detectingHook(257)
	writer.types = append(writer.types, callbackType(1, 3, 32, 65, 32))
	writer.apiMax, writer.apis = 5, append(writer.apis, callbackAPI(3, 71, "write"))
	if n := compileHooks(t, writer).Report().Unknown[UnknownBytecode]; n != 0 {
		t.Fatalf("a writing hook: %d unknown", n)
	}
}

// A program may run for millions of steps, as the reference's official
// programs do on large executables (BC.Win.Trojan.Xpaj-7 executes 15.2
// million instructions on a 20 MB DLL); the per-run allowance is the
// interpreter's 50 million, within the object's budget.
func TestBytecodeRunsBeyondOneMillionSteps(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.Loop"), true)
	detect := f.blocks[0]
	f.locals = []int{32, 1}
	f.blocks = [][]string{
		{callbackInst(32, 0, 1, callbackConstant(5_500_000, 32), callbackConstant(0, 32)), "T" + callbackFixed(18, 2) + callbackNum(1)},
		{
			callbackInst(32, 0, 2, callbackNum(0), callbackConstant(1, 32)),
			callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(0, 32)),
			"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(2) + callbackNum(1),
		},
		detect,
	}
	if name, err := compileCallback(t, f).Match(nil2ctx(), []byte("AB")); name != "UTMStack.Loop" || err != nil {
		t.Fatalf("(%q, %v), want the detection", name, err)
	}
}

// Programs that fail or run to their allowance must not use up the object's
// budget and stop the signatures after them: a failed run is charged the
// steps it used, and all runs together draw from half of the object's
// budget. The plain signature after them still detects.
func TestBytecodeCannotStarveLaterSignatures(t *testing.T) {
	failing := callbackWithLabel([]byte("x"), false)
	failing.blocks = [][]string{{callbackInst(32, 0, 39, callbackGlobalRef(0)), callbackReturn(32, 0)}}
	looping := callbackWithLabel([]byte("x"), false)
	looping.blocks = [][]string{{"T" + callbackFixed(18, 2) + callbackNum(1)}, {"T" + callbackFixed(18, 2) + callbackNum(1)}}
	for name, tc := range map[string]struct {
		program callbackFixture
		limits  Limits
	}{
		"failing": {failing, Limits{}},
		"looping": {looping, Limits{MaxMatchSteps: 4_000_000}},
	} {
		files := map[string][]byte{"z.ndb": []byte("Real.Malware.Detect:0:*:4142\n")}
		for i := 0; i < 12; i++ {
			files[fmt.Sprintf("p%02d.cbc", i)] = tc.program.encode()
		}
		d, err := CompileWithLimits(files, tc.limits)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := d.Match(nil2ctx(), []byte("xxABxx")); got != "Real.Malware.Detect" || err != nil {
			t.Fatalf("%s programs: (%q, %v), want the plain detection", name, got, err)
		}
	}
}

// writingProgram declares write and, when call is set, writes four bytes
// before returning 0.
func writingProgram(kind uint64, logical string, call bool) callbackFixture {
	f := callbackWithLabel([]byte("data"), false)
	f.kind, f.logical = kind, logical
	f.apiMax, f.apis = 3, []string{callbackAPI(3, 70, "write")}
	body := []string{callbackReturn(32, 0)}
	if call {
		body = append([]string{callbackCall(0, 3, callbackGlobalRef(2), callbackConstant(4, 32))}, body...)
	}
	f.blocks = [][]string{body}
	return f
}

// A program that declares write runs; an object on which it writes output
// is incomplete (the reference would scan that output), and every other
// object is decided. The official LizardNest unpacker hook, whose trigger is
// the usual 32-bit program entry, was unknown on every such program before.
func TestProgramsThatWriteAreUnknownOnlyWhenTheyWrite(t *testing.T) {
	for _, kind := range []uint64{256, 257} {
		logical := hookTrigger
		if kind == 256 {
			logical = "Write.Test;Engine:51-255,Target:0;0;4142"
		}
		silent := compileHooks(t, writingProgram(kind, logical, false))
		if n := silent.Report().Unknown[UnknownBytecode]; n != 0 {
			t.Fatalf("kind %d: loaded as unknown (%d)", kind, n)
		}
		object := withTrigger(hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400}))
		if name, err := silent.Match(nil2ctx(), object); name != "" || err != nil {
			t.Fatalf("kind %d without a write: (%q, %v), want clean", kind, name, err)
		}
		writer := compileHooks(t, writingProgram(kind, logical, true))
		name, err := writer.Match(nil2ctx(), object)
		var u *UnknownError
		if name != "" || !errors.As(err, &u) || u.Reason != UnknownBytecode {
			t.Fatalf("kind %d with a write: (%q, %v), want unknown %q", kind, name, err, UnknownBytecode)
		}
		// Where the trigger does not hold the program never runs.
		if name, err := writer.Match(nil2ctx(), hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})); name != "" || err != nil {
			t.Fatalf("kind %d without its trigger: (%q, %v), want clean", kind, name, err)
		}
	}
}
