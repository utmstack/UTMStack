package bytecode

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
)

func hookProgram(t *testing.T, kind uint64, trigger string, m memoryFixture) *Program {
	t.Helper()
	p := memoryParse(t, m)
	p.Kind, p.LogicalSignature = kind, trigger
	return p
}

func TestValidateHook(t *testing.T) {
	plain := memoryFixture{functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}}
	for _, kind := range []uint64{257, 259, 260, 261, 262} {
		if err := hookProgram(t, kind, "Hook;Engine:51-255,Target:1;0;4142", plain).ValidateHook(216); err != nil {
			t.Fatalf("kind %d: %v", kind, err)
		}
	}
	for _, kind := range []uint64{0, 256, 258} {
		if err := hookProgram(t, kind, "Hook;Engine:51-255,Target:1;0;4142", plain).ValidateHook(216); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("kind %d: %v; want ErrUnsupported", kind, err)
		}
	}
	for _, kind := range []uint64{257, 259, 260, 261, 262} {
		if err := hookProgram(t, kind, "", plain).ValidateHook(216); err != nil {
			t.Fatalf("kind %d without a trigger: %v", kind, err)
		}
	}
	// A PDF hook runs on every PDF object, without a trigger: its line is
	// empty or holds only a name, as the official ones do.
	for _, line := range []string{"", "BC.Pdf.Test-1"} {
		if err := hookProgram(t, 258, line, plain).ValidateHook(216); err != nil {
			t.Fatalf("PDF hook %q: %v", line, err)
		}
	}
	p := hookProgram(t, 257, "Hook;Engine:51-255,Target:1;0;4142", plain)
	p.MinFunctionality = 217
	if err := p.ValidateHook(216); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("functionality window: %v", err)
	}
	// A declared write loads; a scanning run that calls it stops (see
	// TestScanningRunStopsAtWrite).
	writer := memoryFixture{types: []string{memoryType(1, 3, 32, 65, 32)}, apiMax: 3, apis: []string{memoryAPI(3, 69, "write")},
		functions: plain.functions}
	if err := hookProgram(t, 259, "Hook;Engine:51-255,Target:1;0;4142", writer).ValidateHook(216); err != nil {
		t.Fatalf("write declared: %v", err)
	}
}

func TestKindReadsZero(t *testing.T) {
	// The reference never sets the kind pointer: __clambc_kind reads 0 for
	// every program, whatever its kind.
	m := memoryFixture{globalMax: 0x8001, globals: []string{memoryGlobal(66, 0, 0x8001)}, functions: []compatFunction{{locals: []int{16}, result: 16, blocks: [][]string{{
		compatInst(16, 0, 39, memoryGlobalRef(1)), compatRet(16, compatNum(0)),
	}}}}}
	for _, kind := range []uint64{256, 257, 259} {
		p := memoryParse(t, m)
		p.Kind = kind
		got, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), EmptyMatchContext(), 100)
		if err != nil || got.Value != 0 {
			t.Fatalf("kind %d: %+v, %v; want 0", kind, got, err)
		}
	}
}

func TestHookRecordIsCopied(t *testing.T) {
	record := make([]byte, 648)
	binary.LittleEndian.PutUint16(record[264:], 0x10b)
	binary.LittleEndian.PutUint32(record[632:], 0x80)
	binary.LittleEndian.PutUint32(record[4:], 0x200)
	pe := &PEInfo{EntryPoint: 0x400, HeaderSize: 0x1000, Record: record}
	for offset, want := range map[uint64]uint64{264: 0x10b, 632: 0x80, 4: 0x200} {
		matches := MatchContext{PE: pe, Hook: true}
		got, err := memoryParse(t, peDataProgram(offset)).RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), matches, 100)
		if err != nil || got.Value != want {
			t.Fatalf("offset %d: %+v, %v; want %#x", offset, got, err, want)
		}
	}
	// Without a hook record the logical-path record is unchanged.
	matches := EmptyMatchContext()
	matches.PE = &PEInfo{EntryPoint: 0x400}
	if got, err := memoryParse(t, peDataProgram(4)).RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), matches, 100); err != nil || got.Value != 0x400 {
		t.Fatalf("logical record: %+v, %v", got, err)
	}
}

func TestHookMatchContextIsAllZero(t *testing.T) {
	// A hook program gets zero counts and zero offsets, not the missing
	// marker; any other content is refused. A logical context still needs
	// the missing marker in its empty slots.
	p := memoryParse(t, peDataProgram(4))
	if _, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), MatchContext{Hook: true}, 100); err != nil {
		t.Fatalf("zero hook context: %v", err)
	}
	for _, mutate := range []func(*MatchContext){
		func(c *MatchContext) { c.Counts[3] = 1 },
		func(c *MatchContext) { c.Offsets[0] = MissingMatchOffset },
	} {
		c := MatchContext{Hook: true}
		mutate(&c)
		if _, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), c, 100); !errors.Is(err, ErrRuntime) {
			t.Fatalf("hook context with content: %v; want ErrRuntime", err)
		}
	}
	if _, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), MatchContext{}, 100); !errors.Is(err, ErrRuntime) {
		t.Fatalf("zero logical context: %v; want ErrRuntime", err)
	}
}

// A generic run keeps written output; a run on a scanned object (with a
// match context) stops at the first write, because the reference would scan
// the extracted output as a new object.
func TestScanningRunStopsAtWrite(t *testing.T) {
	p := memoryParse(t, memoryIOProgram())
	input := []byte("UTMStack.Test\x00abcdefghi")
	if got, err := p.RunWithInput(context.Background(), 0, nil, input, 1000); err != nil || len(got.Output) == 0 {
		t.Fatalf("generic run: %+v, %v", got, err)
	}
	got, err := p.RunWithMatchContext(context.Background(), 0, nil, input, EmptyMatchContext(), 1000)
	if !errors.Is(err, ErrExtraction) || !failedEmpty(got, 1000) {
		t.Fatalf("scanning run: %+v, %v; want ErrExtraction", got, err)
	}
}

func TestExtractNewStopsWithoutChildScanner(t *testing.T) {
	p := memoryParse(t, memoryFixture{
		types:  []string{memoryType(1, 2, 32, 32)},
		apiMax: 22,
		apis:   []string{memoryAPI(22, 69, "extract_new")},
		functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 22, compatConstant(0xffffffff, 32)),
			compatRet(32, compatNum(0)),
		}}}},
	})
	p.Kind, p.LogicalSignature = 257, "Hook;Engine:51-255,Target:1;0;4142"
	if err := p.ValidateHook(216); err != nil {
		t.Fatal(err)
	}
	for _, matches := range []MatchContext{EmptyMatchContext(), {Hook: true}} {
		got, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("MZ"), matches, 1000)
		if !errors.Is(err, ErrExtraction) || !failedEmpty(got, 1000) {
			t.Fatalf("extract_new without child scanner: %+v, %v", got, err)
		}
	}
}
