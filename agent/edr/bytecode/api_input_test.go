package bytecode

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIInputSwitchFileSizePipesAndPointers(t *testing.T) {
	var mem memory
	size, err := mem.alloc(4, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.store(size, 4, 5); err != nil {
		t.Fatal(err)
	}
	mem.protect(size)
	s := &apiState{input: []byte("ABCDE"), offset: 4, fileSizePtr: size}
	apiTestCall(t, s, &mem, 31, 0, 0)
	originalPointer, err := s.invoke(context.Background(), 33, []uint64{0, 1}, &mem, func(uint64) error { return nil })
	if err != nil || originalPointer == 0 || originalPointer != s.filePtr {
		t.Fatalf("file pipe did not borrow the original input: %#x, %v", originalPointer, err)
	}
	output := apiTestBytes(t, &mem, []byte("xy"))
	apiTestCall(t, s, &mem, 71, apiFailure, 1)
	apiTestCall(t, s, &mem, 3, 2, output, 2)
	apiTestCall(t, s, &mem, 71, 0, 1)
	apiTestCall(t, s, &mem, 71, 0, 2) // all nonzero values select output
	if s.offset != 4 || string(s.input) != "xy" {
		t.Fatalf("switch changed cursor or chose wrong file: offset %d, input %q", s.offset, s.input)
	}
	if got, err := mem.load(size, 4); err != nil || got != 2 {
		t.Fatalf("switched filesize = %d, %v", got, err)
	}
	apiTestCall(t, s, &mem, 2, 0, output, 1) // old offset is now past EOF
	apiTestCall(t, s, &mem, 17, 'x', 0)
	apiTestCall(t, s, &mem, 32, 2, 0) // same pipe cursor, current file
	extractedPointer, err := s.invoke(context.Background(), 33, []uint64{0, 1}, &mem, func(uint64) error { return nil })
	if err != nil || extractedPointer == 0 || extractedPointer == originalPointer {
		t.Fatalf("extracted file pointer = %#x, %v", extractedPointer, err)
	}
	if data, err := mem.read(extractedPointer, 1); err != nil || string(data) != "x" {
		t.Fatalf("extracted pointer = %q, %v", data, err)
	}
	apiTestCall(t, s, &mem, 4, 0, 0, 0)
	apiTestCall(t, s, &mem, 2, 1, output, 1)
	if data, err := mem.read(output, 1); err != nil || string(data) != "x" {
		t.Fatalf("read from extracted file = %q, %v", data, err)
	}
	apiTestCall(t, s, &mem, 71, 0, 0)
	apiTestCall(t, s, &mem, 71, 0, 0)
	if got, err := mem.load(size, 4); err != nil || got != 5 {
		t.Fatalf("restored filesize = %d, %v", got, err)
	}
	if err := mem.check(extractedPointer, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("retired extracted fmap pointer is still valid: %v", err)
	}
	if data, err := mem.read(originalPointer, 1); err != nil || string(data) != "A" {
		t.Fatalf("original fmap pointer = %q, %v", data, err)
	}
	apiTestCall(t, s, &mem, 17, 'A', 0)
	if s.offset != 1 {
		t.Fatalf("input switch unexpectedly reset offset to %d", s.offset)
	}
}

func TestAPIInputSwitchBudgetAndUnsupportedActiveExtraction(t *testing.T) {
	var mem memory
	output := apiTestBytes(t, &mem, []byte("xy"))
	s := &apiState{input: []byte("AB"), scanning: true, scanExtracted: func([]byte, int32) (string, error) { return "", nil }}
	apiTestCall(t, s, &mem, 3, 2, output, 2)
	charges := 0
	_, err := s.invoke(context.Background(), 71, []uint64{1}, &mem, func(n uint64) error {
		charges++
		if charges == 2 && n > 0 {
			return ErrLimit
		}
		return nil
	})
	if !errors.Is(err, ErrLimit) || charges != 2 || s.extractedInput || string(s.input) != "AB" {
		t.Fatalf("failed switch mutated state: %+v, %v", s, err)
	}
	apiTestCall(t, s, &mem, 71, 0, 1)
	if _, err := s.invoke(context.Background(), 22, []uint64{7}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("extract_new with active output fmap = %v", err)
	}
	if !bytes.Equal(s.output, []byte("xy")) {
		t.Fatalf("unsupported extraction lost pending child: %q", s.output)
	}
}

func TestSerializedInputSwitchReadsExtractedOutput(t *testing.T) {
	m := memoryFixture{
		types: []string{
			memoryType(4, 2, 8), memoryType(5, 1, 69),
			memoryType(1, 3, 32, 65, 32), memoryType(1, 2, 32, 32),
			memoryType(1, 2, 32, 32),
		},
		globals: []string{memoryGlobal(69, 'x', 'y'), memoryGlobal(65, 0, 1)},
		apiMax:  71,
		apis: []string{
			memoryAPI(3, 71, "write"), memoryAPI(71, 72, "input_switch"),
			memoryAPI(17, 73, "file_byteat"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(2, 32)),
			memoryAPICall(1, 71, compatConstant(1, 32)),
			memoryAPICall(2, 17, compatConstant(0, 32)),
			compatRet(32, compatNum(2)),
		}}}},
	}
	result, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, []byte("ABCDE"), 1000)
	if err != nil || result.Value != 'x' || string(result.Output) != "xy" {
		t.Fatalf("serialized input switch = %+v, %v", result, err)
	}
	// The VM's special filesize global must follow the active fmap while
	// remaining protected against bytecode writes.
	m.globalMax = globalFileSize
	m.globals = append(m.globals, memoryGlobal(67, 0, globalFileSize))
	m.functions[0].blocks[0][2] = compatInst(32, 2, 39, memoryGlobalRef(3))
	result, err = memoryParse(t, m).RunWithInput(context.Background(), 0, nil, []byte("ABCDE"), 1000)
	if err != nil || result.Value != 2 {
		t.Fatalf("switched filesize global = %+v, %v", result, err)
	}
	m.functions[0].blocks[0] = []string{
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(2, 32)),
		memoryAPICall(1, 71, compatConstant(1, 32)),
		memoryAPICall(1, 71, compatConstant(0, 32)),
		compatInst(32, 2, 39, memoryGlobalRef(3)),
		compatRet(32, compatNum(2)),
	}
	result, err = memoryParse(t, m).RunWithInput(context.Background(), 0, nil, []byte("ABCDE"), 1000)
	if err != nil || result.Value != 5 {
		t.Fatalf("restored filesize global = %+v, %v", result, err)
	}
}

func TestWriteInputSwitchOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_INPUT_SWITCH_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_INPUT_SWITCH_ORACLE_DIR for a harmless native ClamAV differential")
	}
	const label = "UTM.InputSwitch"
	m := engineLevelProgram(17, "file_byteat", 'x', label)
	m.kind = 256
	m.logical = label + ";Engine:56-255,Target:0;0;4142"
	m.types = append(m.types, memoryType(1, 2, 32, 32))
	m.globals = append(m.globals,
		memoryGlobal(69, stringComponents("xy", 16)...), memoryGlobal(65, 0, 3))
	m.apiMax = 71
	m.apis = []string{
		memoryAPI(5, 72, "setvirusname"), memoryAPI(3, 72, "write"),
		memoryAPI(71, 73, "input_switch"), memoryAPI(17, 73, "file_byteat"),
	}
	m.functions = []compatFunction{{locals: []int{32, 32, 32, 1, 32}, result: 32, blocks: [][]string{
		{
			memoryAPICall(0, 3, memoryGlobalRef(4), compatConstant(2, 32)),
			memoryAPICall(1, 71, compatConstant(1, 32)),
			memoryAPICall(2, 17, compatConstant(0, 32)),
			compatInst(1, 3, 21, compatNum(32), compatNum(2), compatConstant('x', 32)),
			"T" + compatFixed(17, 2) + compatNum(3) + compatNum(1) + compatNum(2),
		},
		{memoryAPICall(4, 5, memoryGlobalRef(2), compatConstant(uint64(len(label)), 32)), compatRet(32, compatNum(2))},
		{compatRet(32, compatNum(2))},
	}}}
	p := memoryParse(t, m)
	result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("AB"), EmptyMatchContext(),
		func(data []byte, _ int32) (string, error) {
			if string(data) != "xy" {
				t.Errorf("extracted child = %q", data)
			}
			return "", nil
		}, 1000)
	if err != nil || result.VirusName != label || result.Value != 'x' {
		t.Fatalf("Go input-switch oracle = %+v, %v", result, err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"input-switch.cbc": m.encode(),
		"sample-ab.bin":    append([]byte("AB"), bytes.Repeat([]byte{'Q'}, 1024)...),
		"sample-ac.bin":    append([]byte("AC"), bytes.Repeat([]byte{'Q'}, 1024)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
