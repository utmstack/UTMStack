package bytecode

import (
	"bytes"
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

func apiTestBytes(t *testing.T, mem *memory, data []byte) uint64 {
	t.Helper()
	ptr, err := mem.alloc(len(data), false)
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.write(ptr, data); err != nil {
		t.Fatal(err)
	}
	return ptr
}

func apiTestCall(t *testing.T, state *apiState, mem *memory, id int, want uint64, args ...uint64) {
	t.Helper()
	got, err := state.invoke(context.Background(), id, args, mem, func(uint64) error { return nil })
	if err != nil || got != want {
		t.Fatalf("API %d(%v) = %#x, %v; want %#x", id, args, got, err, want)
	}
}

func TestAPIDeclarations(t *testing.T) {
	want := map[int]apiDefinition{
		1: {"test1", 32, []int{32, 32}}, 2: {"read", 32, []int{65, 32}},
		3: {"write", 32, []int{65, 32}}, 4: {"seek", 32, []int{32, 32}},
		5: {"setvirusname", 32, []int{65, 32}}, 6: {"debug_print_str", 32, []int{65, 32}},
		9: {"trace_directory", 32, []int{65, 32}}, 10: {"trace_scope", 32, []int{65, 32}},
		11: {"trace_source", 32, []int{65, 32}}, 12: {"trace_op", 32, []int{65, 32}},
		13: {"trace_value", 32, []int{65, 32}}, 14: {"trace_ptr", 32, []int{65, 32}},
		18: {"malloc", 65, []int{32}}, 22: {"extract_new", 32, []int{32}}, 23: {"read_number", 32, []int{32}}, 51: {"memstr", 32, []int{65, 32, 65, 32}},
		24: {"hashset_new", 32, nil}, 25: {"hashset_add", 32, []int{32, 32}},
		26: {"hashset_remove", 32, []int{32, 32}}, 27: {"hashset_contains", 32, []int{32, 32}},
		28: {"hashset_done", 32, []int{32}}, 29: {"hashset_empty", 32, []int{32}},
		30: {"buffer_pipe_new", 32, []int{32}}, 31: {"buffer_pipe_new_fromfile", 32, []int{32}},
		32: {"buffer_pipe_read_avail", 32, []int{32}}, 33: {"buffer_pipe_read_get", 65, []int{32, 32}},
		34: {"buffer_pipe_read_stopped", 32, []int{32, 32}}, 35: {"buffer_pipe_write_avail", 32, []int{32}},
		36: {"buffer_pipe_write_get", 65, []int{32, 32}}, 37: {"buffer_pipe_write_stopped", 32, []int{32, 32}},
		38: {"buffer_pipe_done", 32, []int{32}},
		39: {"inflate_init", 32, []int{32, 32, 32}}, 40: {"inflate_process", 32, []int{32}},
		41: {"inflate_done", 32, []int{32}},
		43: {"jsnorm_init", 32, []int{32}}, 44: {"jsnorm_process", 32, []int{32}},
		45: {"jsnorm_done", 32, []int{32}},
		57: {"map_new", 32, []int{32, 32}}, 58: {"map_addkey", 32, []int{65, 32, 32}},
		59: {"map_setvalue", 32, []int{65, 32, 32}}, 60: {"map_remove", 32, []int{65, 32, 32}},
		61: {"map_find", 32, []int{65, 32, 32}}, 62: {"map_getvaluesize", 32, []int{32}},
		63: {"map_getvalue", 65, []int{32, 32}}, 64: {"map_done", 32, []int{32}},
		52: {"hex2ui", 32, []int{32, 32}}, 56: {"entropy_buffer", 32, []int{65, 32}},
		15: {"pe_rawaddr", 32, []int{32}}, 20: {"get_pe_section", 32, []int{91, 32}},
		21: {"fill_buffer", 32, []int{65, 32, 32, 32, 32}},
		72: {"get_environment", 32, []int{86, 32}}, 8: {"disasm_x86", 32, []int{95, 32}},
		73: {"disable_bytecode_if", 32, []int{65, 32, 32}},
		74: {"disable_jit_if", 32, []int{65, 32, 32}},
		75: {"version_compare", 32, []int{65, 32, 65, 32}},
		76: {"check_platform", 32, []int{32, 32, 32}},
		7:  {"debug_print_uint", 32, []int{32}}, 16: {"file_find", 32, []int{65, 32}},
		17: {"file_byteat", 32, []int{32}}, 54: {"debug_print_str_start", 32, []int{65, 32}},
		19: {"test2", 32, []int{32}},
		55: {"debug_print_str_nonl", 32, []int{65, 32}}, 65: {"file_find_limit", 32, []int{65, 32, 32}},
		42: {"bytecode_rt_error", 32, []int{32}}, 46: {"ilog2", 32, []int{32, 32}},
		47: {"ipow", 32, []int{32, 32, 32}}, 48: {"iexp", 32, []int{32, 32, 32}},
		49: {"isin", 32, []int{32, 32, 32}}, 50: {"icos", 32, []int{32, 32, 32}},
		66: {"engine_functionality_level", 32, nil}, 67: {"engine_dconf_level", 32, nil},
		68: {"engine_scan_options", 32, nil}, 69: {"engine_db_options", 32, nil},
		70: {"extract_set_container", 32, []int{32}},
		71: {"input_switch", 32, []int{32}},
		77: {"pdf_get_obj_num", 32, nil}, 78: {"pdf_get_flags", 32, nil}, 79: {"pdf_set_flags", 32, []int{32}},
		80: {"pdf_lookupobj", 32, []int{32}}, 81: {"pdf_getobjsize", 32, []int{32}}, 82: {"pdf_getobj", 65, []int{32, 32}},
		83: {"pdf_getobjid", 32, []int{32}}, 84: {"pdf_getobjflags", 32, []int{32}}, 85: {"pdf_setobjflags", 32, []int{32, 32}},
		86: {"pdf_get_offset", 32, []int{32}}, 87: {"pdf_get_phase", 32, nil}, 88: {"pdf_get_dumpedobjid", 32, nil},
		89: {"matchicon", 32, []int{65, 32, 65, 32}},
		90: {"running_on_jit", 32, nil},
		91: {"get_file_reliability", 32, nil},
		92: {"json_is_active", 32, nil}, 93: {"json_get_object", 32, []int{65, 32, 32}},
		94: {"json_get_type", 32, []int{32}}, 95: {"json_get_array_length", 32, []int{32}},
		96: {"json_get_array_idx", 32, []int{32, 32}}, 97: {"json_get_string_length", 32, []int{32}},
		98: {"json_get_string", 32, []int{65, 32, 32}}, 99: {"json_get_boolean", 32, []int{32}},
		100: {"json_get_int", 32, []int{32}},
		101: {"engine_scan_options_ex", 32, []int{65, 32}},
		102: {"lzma_init", 32, []int{32, 32}},
		103: {"lzma_process", 32, []int{32}},
		104: {"lzma_done", 32, []int{32}},
		105: {"bzip2_init", 32, []int{32, 32}},
		106: {"bzip2_process", 32, []int{32}},
		107: {"bzip2_done", 32, []int{32}},
	}
	if !reflect.DeepEqual(builtinAPIs, want) {
		t.Fatal("API wire IDs, names, or prototypes changed")
	}
	// PRE has no objects yet; later phases need the PDF parser's object model.
	apiTestCall(t, &apiState{pdfPhase: 3}, nil, 87, 3)
	apiTestCall(t, &apiState{}, nil, 87, 0)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 77, 0)
	apiTestCall(t, &apiState{pdfPhase: 4, pdfFlagsKnown: true, pdfFlags: 5}, nil, 78, 5)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 80, apiFailure, 123)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 81, 0, 0)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 83, apiFailure, 0)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 84, apiFailure, 0)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 88, apiFailure)
	apiTestCall(t, &apiState{pdfPhase: 4}, nil, 86, apiFailure, 0)
	apiTestCall(t, &apiState{}, nil, 77, apiFailure)
	apiTestCall(t, &apiState{}, nil, 78, apiFailure)
	apiTestCall(t, &apiState{}, nil, 80, apiFailure, 0)
	apiTestCall(t, &apiState{}, nil, 81, 0, 0)
	apiTestCall(t, &apiState{}, nil, 86, apiFailure, 0)
	apiTestCall(t, &apiState{pdfPhase: 2}, nil, 88, 0)
	for _, id := range []int{77, 78, 79, 80, 82, 83, 84, 85, 86} {
		args := make([]uint64, len(builtinAPIs[id].args))
		if _, err := (&apiState{pdfPhase: 2}).invoke(context.Background(), id, args, nil, func(uint64) error { return nil }); !errors.Is(err, ErrPDFObjects) {
			t.Fatalf("API %d: %v; want ErrPDFObjects", id, err)
		}
	}
	for _, state := range []*apiState{{}, {scanning: true, output: []byte("unscanned child")}} {
		if _, err := state.invoke(context.Background(), 22, []uint64{0}, nil, func(uint64) error { return nil }); !errors.Is(err, ErrExtraction) {
			t.Fatalf("extract_new without a child scanner: %v; want ErrExtraction", err)
		}
	}
	state := &apiState{}
	apiTestCall(t, state, nil, 1, 0x12345678, 0xf00dbeef, 0xbeeff00d)
	apiTestCall(t, state, nil, 1, 0x55, 0xbeeff00d, 0xf00dbeef)
	apiTestCall(t, state, nil, 19, 0xd00f, 0xf00d)
	apiTestCall(t, state, nil, 19, 0x5555, 0xf00e)
	apiTestCall(t, state, nil, 66, EngineLevel)
	apiTestCall(t, state, nil, 67, EngineLevel)
	if EngineLevel != 234 {
		t.Fatalf("EngineLevel = %d; the pinned ClamAV 1.5.4 release reports 234", EngineLevel)
	}
}

func TestAPIMatchIconReferenceGate(t *testing.T) {
	var mem memory
	first := apiTestBytes(t, &mem, []byte("KNOWN"))
	second := apiTestBytes(t, &mem, []byte("GROUP"))
	state := &apiState{pe: &PEInfo{EntryPoint: 0x200}}
	apiTestCall(t, state, &mem, 89, 0, first, 5, second, 5)
	state.pe.EntryPoint = 0
	apiTestCall(t, state, &mem, 89, apiFailure, first, 5, second, 5)
	state.pe.EntryPoint = 0x200
	long := apiTestBytes(t, &mem, bytes.Repeat([]byte{'A'}, 128))
	apiTestCall(t, state, &mem, 89, apiFailure, long, 128, second, 5)
	for _, args := range [][]uint64{{first, 5, 0, 5}, {0, 5, second, 5}} {
		if _, err := state.invoke(context.Background(), 89, args, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
			t.Fatalf("invalid pointer %v: %v; want ErrRuntime", args, err)
		}
	}
}

// engineLevelProgram calls one engine level API and sets label only when the
// returned level equals want. It always returns the level it received.
func engineLevelProgram(id int, name string, want uint64, label string) memoryFixture {
	components := make([]uint64, 16)
	for i := 0; i < len(label); i++ {
		components[i] = uint64(label[i])
	}
	return memoryFixture{
		types:   append(memoryArrayTypes(16, 8), memoryType(1, 1, 32), memoryType(1, 3, 32, 65, 32)),
		globals: []string{memoryGlobal(69, components...), memoryGlobal(65, 0, 1)},
		apiMax:  id, apis: []string{memoryAPI(5, 72, "setvirusname"), memoryAPI(id, 71, name)},
		functions: []compatFunction{{locals: []int{32, 1, 32}, result: 32, blocks: [][]string{
			{
				memoryAPICall(0, id),
				compatInst(1, 1, 21, compatNum(32), compatNum(0), compatConstant(want, 32)),
				"T" + compatFixed(17, 2) + compatNum(1) + compatNum(1) + compatNum(2),
			},
			{memoryAPICall(2, 5, memoryGlobalRef(2), compatConstant(uint64(len(label)), 32)), compatRet(32, compatNum(0))},
			{compatRet(32, compatNum(0))},
		}}},
	}
}

func TestAPIEngineLevelQueries(t *testing.T) {
	// bytecode_api.c returns cl_retflevel() (CL_FLEVEL) and CL_FLEVEL_DCONF,
	// Both report the pinned reference level. Neither API reads its context.
	const label = "UTMStack.Level"
	for _, api := range []struct {
		id   int
		name string
	}{{66, "engine_functionality_level"}, {67, "engine_dconf_level"}} {
		t.Run(api.name, func(t *testing.T) {
			for _, tc := range []struct {
				want    uint64
				labeled bool
			}{{EngineLevel, true}, {EngineLevel - 1, false}, {EngineLevel + 1, false}} {
				p := memoryParse(t, engineLevelProgram(api.id, api.name, tc.want, label))
				got, err := p.RunWithInput(context.Background(), 0, nil, nil, 1000)
				if err != nil || got.Value != EngineLevel || got.VirusNameSet != tc.labeled {
					t.Fatalf("level == %d: %+v, %v", tc.want, got, err)
				}
				if tc.labeled && got.VirusName != label {
					t.Fatalf("label %q; want %q", got.VirusName, label)
				}
				p.Kind = 256
				p.LogicalSignature = "UTMStack.Level;Engine:56-255,Target:1;0;4142"
				if err := p.ValidateLogical(EngineLevel); err != nil {
					t.Fatalf("logical profile: %v", err)
				}
			}
		})
	}
}

func TestAPIEngineLevelDeclarationsStayExact(t *testing.T) {
	// All four engine queries use the same no-argument 32-bit prototype.
	body := []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}
	for _, tc := range []struct {
		name, declaration, prototype string
		want                         error
	}{
		{"functionality", memoryAPI(66, 69, "engine_functionality_level"), memoryType(1, 1, 32), nil},
		{"dconf", memoryAPI(67, 69, "engine_dconf_level"), memoryType(1, 1, 32), nil},
		{"scan-options", memoryAPI(68, 69, "engine_scan_options"), memoryType(1, 1, 32), nil},
		{"db-options", memoryAPI(69, 69, "engine_db_options"), memoryType(1, 1, 32), nil},
		{"swapped-name", memoryAPI(66, 69, "engine_dconf_level"), memoryType(1, 1, 32), ErrMalformed},
		{"argument", memoryAPI(66, 69, "engine_functionality_level"), memoryType(1, 2, 32, 32), ErrMalformed},
		{"wide-result", memoryAPI(67, 69, "engine_dconf_level"), memoryType(1, 1, 64), ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := memoryFixture{types: []string{tc.prototype}, apis: []string{tc.declaration}, apiMax: 4095, functions: body}
			if _, err := Parse(m.encode()); !errors.Is(err, tc.want) {
				t.Fatalf("Parse error = %v; want %v", err, tc.want)
			}
		})
	}
	state := &apiState{}
	for _, id := range []int{66, 67, 68, 69} {
		if _, err := state.invoke(context.Background(), id, []uint64{0}, nil, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
			t.Fatalf("API %d with an argument: %v", id, err)
		}
	}
	for _, id := range []int{68, 69} {
		if _, err := state.invoke(context.Background(), id, nil, nil, func(uint64) error { return nil }); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("API %d: %v", id, err)
		}
	}
}

func TestAPIEngineOptionsUseScannerProfile(t *testing.T) {
	profile := &EngineOptions{Scan: 0x2015, Database: 0x200a}
	for _, tc := range []struct {
		id   int
		name string
		want uint64
	}{{68, "engine_scan_options", 0x2015}, {69, "engine_db_options", 0x200a}} {
		t.Run(tc.name, func(t *testing.T) {
			p := memoryParse(t, engineLevelProgram(tc.id, tc.name, tc.want, "UTM.Options"))
			if _, err := p.RunWithInput(context.Background(), 0, nil, nil, 1000); !errors.Is(err, ErrUnsupported) {
				t.Fatalf("missing scanner profile: %v", err)
			}
			m := EmptyMatchContext()
			m.Options = profile
			got, err := p.RunWithMatchContext(context.Background(), 0, nil, nil, m, 1000)
			if err != nil || got.Value != tc.want || got.VirusName != "UTM.Options" {
				t.Fatalf("profile query: %+v, %v", got, err)
			}
		})
	}
}

func TestAPIScanOptionsExPinnedBehavior(t *testing.T) {
	// The pinned 1.5.4 native interpreter returns zero for these documented
	// names even with the corresponding scan options enabled. Keep the wire
	// API usable without manufacturing a detection that it would not make.
	state := &apiState{options: &EngineOptions{Scan: 0x10000001, Database: 0x200a}}
	var mem memory
	for _, name := range []string{"general collect metadata", "parse archive", "general heuristics", "unknown option"} {
		ptr := apiTestBytes(t, &mem, []byte(name))
		apiTestCall(t, state, &mem, 101, 0, ptr, uint64(len(name)))
	}
	ptr := apiTestBytes(t, &mem, []byte("parse archive"))
	apiTestCall(t, state, &mem, 101, 0, ptr, 0)
	if _, err := state.invoke(context.Background(), 101, []uint64{ptr, 100}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("out-of-bounds option name: %v", err)
	}
	if _, err := state.invoke(context.Background(), 101, []uint64{ptr, 13}, &mem, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) {
		t.Fatalf("exhausted work budget: %v", err)
	}
	for _, tc := range []struct {
		name, declaration, prototype string
		wantErr                      error
	}{
		{"valid", memoryAPI(101, 69, "engine_scan_options_ex"), memoryType(1, 3, 32, 65, 32), nil},
		{"wrong-name", memoryAPI(101, 69, "engine_scan_options"), memoryType(1, 3, 32, 65, 32), ErrMalformed},
		{"wrong-prototype", memoryAPI(101, 69, "engine_scan_options_ex"), memoryType(1, 2, 32, 65), ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := memoryFixture{types: []string{tc.prototype}, apis: []string{tc.declaration}, apiMax: 101,
				functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}}
			if _, err := Parse(fixture.encode()); !errors.Is(err, tc.wantErr) {
				t.Fatalf("Parse: %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAPIEngineLevelQueriesAreBoundedAndStateless(t *testing.T) {
	// Upstream ignores the context for both queries, so a call must leave the
	// input cursor, output, label and diagnostics untouched. Each call still
	// costs one unit of work and honors cancellation like every other API.
	for _, id := range []int{66, 67} {
		state := &apiState{input: []byte("abc"), offset: 2, output: []byte("out"), virusName: "Kept", virusNameSet: true, debug: []string{"d"}, debugBytes: 2}
		want := *state
		charged := uint64(0)
		got, err := state.invoke(context.Background(), id, nil, nil, func(n uint64) error { charged += n; return nil })
		if err != nil || got != EngineLevel || charged != 1 || !reflect.DeepEqual(*state, want) {
			t.Fatalf("API %d = %d, %v; charged %d; state %+v", id, got, err, charged, *state)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := state.invoke(ctx, id, nil, nil, func(uint64) error { return nil }); !errors.Is(err, context.Canceled) {
			t.Fatalf("API %d with a canceled context: %v", id, err)
		}
		if _, err := state.invoke(context.Background(), id, nil, nil, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) {
			t.Fatalf("API %d with an exhausted budget: %v", id, err)
		}
	}
}

func TestAPIReadShortEOFAndSeek(t *testing.T) {
	state := &apiState{input: []byte("abcabc")}
	mem := &memory{}
	ptr, err := mem.alloc(8, false)
	if err != nil {
		t.Fatal(err)
	}
	apiTestCall(t, state, mem, 2, 2, ptr, 2)
	if got, err := mem.read(ptr, 2); err != nil || string(got) != "ab" || state.offset != 2 {
		t.Fatalf("first read: %q, %v, offset %d", got, err, state.offset)
	}
	apiTestCall(t, state, mem, 2, 4, ptr, 8)
	if got, err := mem.read(ptr, 4); err != nil || string(got) != "cabc" || state.offset != 6 {
		t.Fatalf("short read: %q, %v, offset %d", got, err, state.offset)
	}
	if _, err := mem.read(ptr+4, 1); !errors.Is(err, ErrRuntime) {
		t.Fatalf("short read initialized unread bytes: %v", err)
	}
	apiTestCall(t, state, mem, 2, 0, ptr, 1)
	apiTestCall(t, state, mem, 2, 0, ptr, 0)
	apiTestCall(t, state, mem, 4, 5, apiFailure, 2)
	apiTestCall(t, state, mem, 17, 'c', 5)
	apiTestCall(t, state, mem, 17, apiFailure, 6)
	if state.offset != 5 {
		t.Fatal("byteat changed cursor")
	}
	apiTestCall(t, state, mem, 4, apiFailure, 1, 2)
	apiTestCall(t, state, mem, 4, apiFailure, 0, 3)
	if state.offset != 5 {
		t.Fatal("failed seek changed cursor")
	}
	apiTestCall(t, state, mem, 4, 6, 1, 1)
	apiTestCall(t, state, mem, 4, 0, 0, 0)
	apiTestCall(t, state, mem, 4, apiFailure, 0x80000000, 1)
	apiTestCall(t, state, mem, 4, apiFailure, 0x7fffffff, 2)
	if state.offset != 0 {
		t.Fatal("overflow or negative target changed cursor")
	}
}

func TestAPIInputValidationBeforeMutation(t *testing.T) {
	state := &apiState{input: []byte("abc")}
	mem := &memory{}
	ptr := apiTestBytes(t, mem, []byte("...."))
	ro, err := mem.alloc(4, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		ptr  uint64
		n    uint64
	}{
		{"null", 0, 0}, {"readonly", ro, 1}, {"full span", ptr, 5},
		{"negative signed length", ptr, apiFailure}, {"invalid region", 0xffff00000000, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := state.invoke(context.Background(), 2, []uint64{tc.ptr, tc.n}, mem, func(uint64) error { return nil })
			if !errors.Is(err, ErrRuntime) || state.offset != 0 {
				t.Fatalf("got %v, offset %d", err, state.offset)
			}
		})
	}
	got, err := mem.read(ptr, 4)
	if err != nil || string(got) != "...." {
		t.Fatalf("failed read modified memory: %q, %v", got, err)
	}
}

func TestAPIFileFindWindowAndLimit(t *testing.T) {
	mem := &memory{}
	ptr := apiTestBytes(t, mem, []byte("abc"))
	state := &apiState{input: []byte("abcabc"), offset: 1}
	apiTestCall(t, state, mem, 16, 3, ptr, 3)
	apiTestCall(t, state, mem, 65, apiFailure, ptr, 3, 5)
	apiTestCall(t, state, mem, 65, 3, ptr, 3, 6)
	apiTestCall(t, state, mem, 65, 3, ptr, 3, 100)
	apiTestCall(t, state, mem, 65, apiFailure, ptr, 3, 0)
	apiTestCall(t, state, mem, 65, apiFailure, ptr, 3, apiFailure)
	apiTestCall(t, state, mem, 16, apiFailure, ptr, 0)
	if state.offset != 1 {
		t.Fatal("find modified input cursor")
	}
	long := apiTestBytes(t, mem, bytes.Repeat([]byte{'a'}, 1025))
	apiTestCall(t, state, mem, 16, apiFailure, long, 1025)
	state = &apiState{input: make([]byte, 8192)}
	state.input[4095], state.input[4096] = 'A', 'B'
	needle := apiTestBytes(t, mem, []byte("AB"))
	apiTestCall(t, state, mem, 16, apiFailure, needle, 2)
	state.offset = 1
	apiTestCall(t, state, mem, 16, 4095, needle, 2)
	if state.offset != 1 {
		t.Fatal("windowed find modified cursor")
	}
}

func TestAPIOutputIsOwnedAndBounded(t *testing.T) {
	mem := &memory{}
	x := apiTestBytes(t, mem, []byte("x"))
	yz := apiTestBytes(t, mem, []byte("yz"))
	state := &apiState{input: []byte("abc"), offset: 1}
	apiTestCall(t, state, mem, 3, 1, x, 1)
	apiTestCall(t, state, mem, 3, 2, yz, 2)
	apiTestCall(t, state, mem, 3, 0, yz, 0)
	if err := mem.write(yz, []byte("!!")); err != nil {
		t.Fatal(err)
	}
	if string(state.output) != "xyz" || state.offset != 1 {
		t.Fatalf("output %q, offset %d", state.output, state.offset)
	}
	state.output = make([]byte, maxAPIOutput-1)
	apiTestCall(t, state, mem, 3, 1, x, 1)
	_, err := state.invoke(context.Background(), 3, []uint64{x, 1}, mem, func(uint64) error { return nil })
	if !errors.Is(err, ErrLimit) || len(state.output) != maxAPIOutput {
		t.Fatalf("output cap: %v, length %d", err, len(state.output))
	}
	uninitialized, err := mem.alloc(2, false)
	if err != nil {
		t.Fatal(err)
	}
	state.output = nil
	_, err = state.invoke(context.Background(), 3, []uint64{uninitialized, 2}, mem, func(uint64) error { return nil })
	if !errors.Is(err, ErrRuntime) || len(state.output) != 0 {
		t.Fatalf("uninitialized output: %v, %q", err, state.output)
	}
}

func TestAPIVirusNameCStringOwnershipAndBounds(t *testing.T) {
	mem := &memory{}
	ptr := apiTestBytes(t, mem, []byte("Example.Malware\x00"))
	state := &apiState{}
	// Upstream ignores the length after the VM validates its pointer span.
	apiTestCall(t, state, mem, 5, 0, ptr, 1)
	if err := mem.write(ptr, []byte("X")); err != nil {
		t.Fatal(err)
	}
	if state.virusName != "Example.Malware" {
		t.Fatalf("virus name was not copied: %q", state.virusName)
	}
	boundary := apiTestBytes(t, mem, []byte(strings.Repeat("A", 256)+"\x00"))
	apiTestCall(t, state, mem, 5, 0, boundary, 256)
	if len(state.virusName) != 256 {
		t.Fatal("256-byte virus name rejected")
	}
	for _, tc := range []struct {
		name string
		data []byte
		want error
	}{
		{"too long", []byte(strings.Repeat("B", 257) + "\x00"), ErrLimit},
		{"unterminated", []byte("abc"), ErrRuntime},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := apiTestBytes(t, mem, tc.data)
			_, err := state.invoke(context.Background(), 5, []uint64{p, uint64(len(tc.data))}, mem, func(uint64) error { return nil })
			if !errors.Is(err, tc.want) || state.virusName != strings.Repeat("A", 256) {
				t.Fatalf("name failure: %v, name length %d", err, len(state.virusName))
			}
		})
	}
}

func TestAPIDiagnosticsDisabledReturnProfile(t *testing.T) {
	mem := &memory{}
	ptr := apiTestBytes(t, mem, []byte("ab\x00cd"))
	state := &apiState{}
	apiTestCall(t, state, mem, 6, 0, ptr, 1)
	apiTestCall(t, state, mem, 7, 0, apiFailure)
	apiTestCall(t, state, mem, 54, 0, ptr, 5)
	apiTestCall(t, state, mem, 55, 0, ptr, 5)
	apiTestCall(t, state, mem, 54, apiFailure, ptr, 0)
	apiTestCall(t, state, mem, 55, apiFailure, ptr, 0)
	want := []string{"ab", "-1", "ab", "ab\x00cd"}
	if !reflect.DeepEqual(state.debug, want) {
		t.Fatalf("diagnostics %#v; want %#v", state.debug, want)
	}
	state.debugBytes = maxAPIDebug - 3
	apiTestCall(t, state, mem, 54, 0, ptr, 2)
	apiTestCall(t, state, mem, 7, 0, 0)
	if state.debugBytes != maxAPIDebug || len(state.debug) != 5 || !state.debugTruncated {
		t.Fatalf("diagnostic cap: bytes %d, entries %d, truncated %t", state.debugBytes, len(state.debug), state.debugTruncated)
	}
	for _, id := range []int{6, 54, 55} {
		_, err := (&apiState{}).invoke(context.Background(), id, []uint64{0, 0}, mem, func(uint64) error { return nil })
		if !errors.Is(err, ErrRuntime) {
			t.Fatalf("API %d null pointer: %v", id, err)
		}
	}
}

func TestAPICancellationAndSharedBudgetPreserveState(t *testing.T) {
	mem := &memory{}
	ptr := apiTestBytes(t, mem, []byte("...."))
	state := &apiState{input: []byte("abc")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := state.invoke(ctx, 2, []uint64{ptr, 3}, mem, func(uint64) error { return nil })
	if !errors.Is(err, context.Canceled) || state.offset != 0 {
		t.Fatalf("cancellation: %v, offset %d", err, state.offset)
	}
	budget := uint64(1) // the call's step, not the bytes it would read
	charge := func(n uint64) error {
		if n > budget {
			return ErrLimit
		}
		budget -= n
		return nil
	}
	_, err = state.invoke(context.Background(), 2, []uint64{ptr, 3}, mem, charge)
	if !errors.Is(err, ErrLimit) || state.offset != 0 {
		t.Fatalf("budget: %v, offset %d", err, state.offset)
	}
	got, err := mem.read(ptr, 4)
	if err != nil || string(got) != "...." {
		t.Fatalf("failed read modified memory: %q, %v", got, err)
	}
	state.input = make([]byte, 8192)
	budget = 100
	_, err = state.invoke(context.Background(), 16, []uint64{ptr, 1}, mem, charge)
	if !errors.Is(err, ErrLimit) || state.offset != 0 {
		t.Fatalf("search budget: %v, offset %d", err, state.offset)
	}
}

func TestAPIRejectUnknownAndBadArity(t *testing.T) {
	state := &apiState{}
	for _, tc := range []struct {
		id   int
		args []uint64
		want error
	}{
		{999, nil, ErrUnsupported},
		{1, nil, ErrRuntime}, {1, []uint64{1, 2, 3}, ErrRuntime},
	} {
		_, err := state.invoke(context.Background(), tc.id, tc.args, nil, func(uint64) error { return nil })
		if !errors.Is(err, tc.want) {
			t.Fatalf("API %d: %v; want %v", tc.id, err, tc.want)
		}
	}
}

func TestAPIArithmeticKnownValues(t *testing.T) {
	state := &apiState{}
	for _, tc := range []struct {
		id   int
		args []uint64
		want uint64
	}{
		{46, []uint64{1, 0}, math.MaxInt32},
		{46, []uint64{1, 1}, 0},
		{46, []uint64{2, 1}, 1 << 26},
		{46, []uint64{1, 2}, 0xfc000000},
		{46, []uint64{2, 2}, 0},
		{46, []uint64{3, 1}, 106365033},
		{47, []uint64{0, apiFailure, 1}, math.MaxInt32},
		{47, []uint64{0, 3, 1}, 0},
		{47, []uint64{2, 10, 3}, 3072},
		{47, []uint64{0xfffffffe, 3, 1}, 0xfffffff8},
		{47, []uint64{2, apiFailure, 1}, 1},
		{47, []uint64{2, apiFailure, apiFailure}, apiFailure},
		{47, []uint64{1, 1, math.MaxInt32}, math.MaxInt32},
		{47, []uint64{1, 1, 0x80000000}, 0x80000000},
		{48, []uint64{1, 0, 1}, math.MaxInt32},
		{48, []uint64{0, 1, 1}, 1},
		{48, []uint64{1, 1, 1000}, 2718},
		{48, []uint64{apiFailure, 1, 1000}, 368},
		{49, []uint64{1, 0, 1}, math.MaxInt32},
		{49, []uint64{0, 1, 1000}, 0},
		{49, []uint64{1, 1, 1000}, 841},
		{49, []uint64{apiFailure, 1, 1000}, 0xfffffcb7},
		{50, []uint64{1, 0, 1}, math.MaxInt32},
		{50, []uint64{0, 1, 1000}, 1000},
		{50, []uint64{1, 1, 1000}, 540},
	} {
		apiTestCall(t, state, nil, tc.id, tc.want, tc.args...)
	}
}

func TestAPIArithmeticUndefinedCastsFail(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   int
		args []uint64
	}{
		{"log of zero", 46, []uint64{0, 1}},
		{"log positive overflow", 46, []uint64{math.MaxUint32, 1}},
		{"power positive overflow", 47, []uint64{2, 31, 1}},
		{"power infinity", 47, []uint64{2, math.MaxInt32, 1}},
		{"zero times infinity", 47, []uint64{2, math.MaxInt32, 0}},
		{"unsigned exponential negative", 48, []uint64{0, 1, apiFailure}},
		{"exponential infinity", 48, []uint64{math.MaxInt32, 1, 1}},
		{"exponential positive overflow", 48, []uint64{1, 1, math.MaxInt32}},
		{"cos signed positive overflow", 50, []uint64{314159265, 100000000, 0x80000000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&apiState{}).invoke(context.Background(), tc.id, tc.args, nil, func(uint64) error { return nil })
			if !errors.Is(err, ErrRuntime) {
				t.Fatalf("got %v; want runtime error", err)
			}
		})
	}
}

func TestAPIRoundingBoundaries(t *testing.T) {
	for _, tc := range []struct {
		value    float64
		unsigned bool
		want     uint64
		fail     bool
	}{
		{0.5, false, 1, false}, {-0.5, false, apiFailure, false},
		{0.499, false, 0, false}, {-0.499, false, 0, false},
		{math.MaxInt32, false, math.MaxInt32, false},
		{math.MinInt32, false, 0x80000000, false},
		{math.MaxInt32 + 0.5, false, 0, true},
		{math.MinInt32 - 0.5, false, 0, true},
		{math.MaxUint32, true, math.MaxUint32, false},
		{math.MaxUint32 + 0.5, true, 0, true},
		{-0.499, true, 0, false}, {-0.5, true, 0, true},
		{math.Inf(1), false, 0, true}, {math.Inf(-1), false, 0, true},
		{math.NaN(), true, 0, true},
	} {
		got, err := apiRoundedInteger(tc.value, tc.unsigned)
		if tc.fail {
			if !errors.Is(err, ErrRuntime) {
				t.Fatalf("round(%v, %v): %v", tc.value, tc.unsigned, err)
			}
		} else if err != nil || got != tc.want {
			t.Fatalf("round(%v, %v): %#x, %v; want %#x", tc.value, tc.unsigned, got, err, tc.want)
		}
	}
}

func TestAPIRuntimeLocationIsDiagnostic(t *testing.T) {
	state := &apiState{}
	apiTestCall(t, state, nil, 42, 0, (123<<8)|45)
	apiTestCall(t, state, nil, 42, 0, apiFailure)
	want := []string{"Bytecode runtime error at line 123, col 45", "Bytecode runtime error at line 4294967295, col 255"}
	if !reflect.DeepEqual(state.debug, want) {
		t.Fatalf("location diagnostics %#v", state.debug)
	}
}

func TestAPIDiagnosticCaptureTruncatesSilently(t *testing.T) {
	// With debugging disabled, upstream debug APIs return their normal values
	// and have no observable effect. A full capture buffer drops later
	// messages, but still validates operands and charges their work.
	mem := &memory{}
	text := apiTestBytes(t, mem, []byte("abcdef\x00"))
	unterminated := apiTestBytes(t, mem, []byte("abc"))
	uninitialized, err := mem.alloc(4, false)
	if err != nil {
		t.Fatal(err)
	}
	full := func() *apiState {
		return &apiState{debug: []string{"kept"}, debugBytes: maxAPIDebug - 1}
	}
	for _, tc := range []struct {
		id   int
		args []uint64
		want uint64
	}{
		{6, []uint64{text, 1}, 0}, {7, []uint64{12345}, 0}, {42, []uint64{0x101}, 0},
		{54, []uint64{text, 7}, 0}, {55, []uint64{text, 7}, 0},
		{54, []uint64{text, 0}, apiFailure}, {55, []uint64{text, 0}, apiFailure},
	} {
		state := full()
		work := uint64(0)
		got, err := state.invoke(context.Background(), tc.id, tc.args, mem, func(n uint64) error { work += n; return nil })
		if err != nil || got != tc.want || !reflect.DeepEqual(state.debug, []string{"kept"}) || state.debugBytes != maxAPIDebug-1 {
			t.Fatalf("API %d with full capture = %#x, %v; entries %q, bytes %d", tc.id, got, err, state.debug, state.debugBytes)
		}
		if tc.want == 0 && (!state.debugTruncated || work < 2) {
			t.Fatalf("API %d dropped a message without marking truncation or charging work (%d)", tc.id, work)
		}
	}
	// A string walk longer than the remaining capture space is not a limit.
	long := apiTestBytes(t, mem, append(bytes.Repeat([]byte{'x'}, maxAPIDebug+10), 0))
	apiTestCall(t, full(), mem, 6, 0, long, 1)
	for _, tc := range []struct {
		id   int
		args []uint64
	}{
		{6, []uint64{0, 0}}, {6, []uint64{unterminated, 3}}, {6, []uint64{uninitialized, 4}},
		{54, []uint64{uninitialized, 4}}, {55, []uint64{uninitialized, 4}}, {54, []uint64{0, 1}},
	} {
		_, err := full().invoke(context.Background(), tc.id, tc.args, mem, func(uint64) error { return nil })
		if !errors.Is(err, ErrRuntime) {
			t.Fatalf("API %d(%v) with full capture: %v; want runtime fault", tc.id, tc.args, err)
		}
	}
	// Work is still charged against the shared budget once capture is full.
	_, err = full().invoke(context.Background(), 6, []uint64{long, 1}, mem, func(n uint64) error {
		if n > 0 {
			return ErrLimit
		}
		return nil
	})
	if !errors.Is(err, ErrLimit) {
		t.Fatalf("full capture skipped work accounting: %v", err)
	}
}
