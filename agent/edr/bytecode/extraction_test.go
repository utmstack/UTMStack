package bytecode

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func extractionFixture(terminal string) memoryFixture {
	marker := []byte("HARMLESS-CHILD")
	values := make([]uint64, len(marker))
	for i, b := range marker {
		values[i] = uint64(b)
	}
	return memoryFixture{
		types: []string{
			memoryType(4, len(marker), 8), memoryType(5, 1, 69),
			memoryType(1, 3, 32, 65, 32), memoryType(1, 2, 32, 32),
		},
		globals: []string{memoryGlobal(69, values...), memoryGlobal(65, 0, 1)},
		apiMax:  22,
		apis:    []string{memoryAPI(3, 71, "write"), memoryAPI(22, 72, "extract_new")},
		functions: []compatFunction{{locals: []int{32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len(marker)), 32)),
			memoryAPICall(1, 22, compatConstant(7, 32)),
			terminal,
		}}}},
	}
}

func extractionProgram(t *testing.T, terminal string) *Program {
	t.Helper()
	return memoryParse(t, extractionFixture(terminal))
}

func TestExtractNewScansAndRecyclesOutput(t *testing.T) {
	p := extractionProgram(t, compatRet(32, compatNum(1)))
	var calls int
	got, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("benign input"), EmptyMatchContext(), func(data []byte, id int32) (string, error) {
		calls++
		if id != -1 || !bytes.Equal(data, []byte("HARMLESS-CHILD")) {
			t.Errorf("child %d: %q", id, data)
		}
		return "Child.Hash.Detected", nil
	}, 1000)
	if err != nil || calls != 1 || got.Value != 1 || got.ChildName != "Child.Hash.Detected" || len(got.Output) != 0 {
		t.Fatalf("extract_new result %+v, err %v, calls %d", got, err, calls)
	}
	without, err := p.RunWithMatchContext(context.Background(), 0, nil, []byte("benign input"), EmptyMatchContext(), 1000)
	if !errors.Is(err, ErrExtraction) || without.ChildName != "" {
		t.Fatalf("missing child scanner: %+v, %v", without, err)
	}
}

func TestExtractedDetectionSurvivesLaterVMFault(t *testing.T) {
	p := extractionProgram(t, "T"+compatFixed(45, 2))
	got, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("benign input"), EmptyMatchContext(), func([]byte, int32) (string, error) {
		return "Child.Hash.Detected", nil
	}, 1000)
	if !errors.Is(err, ErrRuntime) || got.ChildName != "Child.Hash.Detected" {
		t.Fatalf("lost child detection: %+v, %v", got, err)
	}
}

func TestPendingOutputIsScannedAfterVMFault(t *testing.T) {
	f := extractionFixture("T" + compatFixed(45, 2))
	f.functions[0].blocks[0] = []string{
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len("HARMLESS-CHILD")), 32)),
		"T" + compatFixed(45, 2),
	}
	var calls int
	got, err := memoryParse(t, f).RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("benign input"), EmptyMatchContext(), func(data []byte, id int32) (string, error) {
		calls++
		if id != -1 || !bytes.Equal(data, []byte("HARMLESS-CHILD")) {
			t.Errorf("pending child %d: %q", id, data)
		}
		return "Child.Hash.Detected", nil
	}, 1000)
	if !errors.Is(err, ErrRuntime) || calls != 1 || got.ChildName != "Child.Hash.Detected" {
		t.Fatalf("lost pending child after VM fault: %+v, %v, calls %d", got, err, calls)
	}
}

func TestExtractNewStartsAnotherOutputFile(t *testing.T) {
	f := extractionFixture(compatRet(32, compatNum(1)))
	f.functions[0].blocks[0] = []string{
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len("HARMLESS-CHILD")), 32)),
		memoryAPICall(1, 22, compatConstant(7, 32)),
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len("HARMLESS-CHILD")), 32)),
		memoryAPICall(1, 22, compatConstant(8, 32)),
		compatRet(32, compatNum(1)),
	}
	var ids []int32
	got, err := memoryParse(t, f).RunWithMatchContextAndScanner(context.Background(), 0, nil, nil, EmptyMatchContext(), func(data []byte, id int32) (string, error) {
		if !bytes.Equal(data, []byte("HARMLESS-CHILD")) {
			t.Errorf("child %d: %q", id, data)
		}
		ids = append(ids, id)
		return "", nil
	}, 1000)
	if err != nil || got.Value != 0 || len(ids) != 2 || ids[0] != -1 || ids[1] != 7 {
		t.Fatalf("extracted files: %+v, %v, ids %v", got, err, ids)
	}
}

func TestExtractNewIDBelongsToNextFile(t *testing.T) {
	f := extractionFixture(compatRet(32, compatNum(1)))
	f.functions[0].blocks[0] = []string{
		memoryAPICall(1, 22, compatConstant(7, 32)),
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len("HARMLESS-CHILD")), 32)),
		compatRet(32, compatNum(1)),
	}
	var id int32 = -1
	got, err := memoryParse(t, f).RunWithMatchContextAndScanner(context.Background(), 0, nil, nil, EmptyMatchContext(), func(data []byte, nextID int32) (string, error) {
		if !bytes.Equal(data, []byte("HARMLESS-CHILD")) {
			t.Errorf("child data: %q", data)
		}
		id = nextID
		return "", nil
	}, 1000)
	if err != nil || id != 7 || got.Value != uint64(0) {
		t.Fatalf("final file ID: %+v, %v, id %d", got, err, id)
	}
}

func TestImplicitOutputFlushScansChild(t *testing.T) {
	p := memoryParse(t, memoryIOProgram())
	input := []byte("UTMStack.Test\x00abcdefghi")
	var seen []byte
	got, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(data []byte, id int32) (string, error) {
		if id != -1 {
			t.Errorf("implicit child ID = %d", id)
		}
		seen = bytes.Clone(data)
		return "", nil
	}, 1000)
	if err != nil || !bytes.Equal(seen, memoryIOExpected(input).Output) || len(got.Output) != 0 {
		t.Fatalf("implicit flush: %+v, %v, child %q", got, err, seen)
	}
}
