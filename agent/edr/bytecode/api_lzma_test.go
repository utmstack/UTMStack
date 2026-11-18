package bytecode

import (
	"bytes"
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

const lzmaTestChild = "HARMLESS-CHILD"

func lzmaTestPacked(t *testing.T) []byte {
	return lzmaTestPackedConfig(t, true)
}

func lzmaTestPackedConfig(t *testing.T, knownSize bool) []byte {
	t.Helper()
	var b bytes.Buffer
	cfg := lzma.WriterConfig{DictCap: 1 << 16}
	if knownSize {
		cfg.Size = int64(len(lzmaTestChild))
	}
	w, err := cfg.NewWriter(&b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(lzmaTestChild)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestAPILZMAUnknownLengthEOS(t *testing.T) {
	packed := lzmaTestPackedConfig(t, false)
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 102, 0, 0, 1)
	apiTestCall(t, s, &mem, 103, 2, 0)
	if s.pipes[0].read != uint32(len(packed)) || s.pipes[1].write != uint32(len(lzmaTestChild)) {
		t.Fatalf("unknown-length cursors: in=%d/%d out=%d", s.pipes[0].read, len(packed), s.pipes[1].write)
	}
}

func TestAPILZMACompleteStreamAndBoundaries(t *testing.T) {
	packed := lzmaTestPacked(t)
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 102, apiFailure, 0, 1)
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 102, 0, 0, 1)
	if s.pipes[0].read != 13 {
		t.Fatalf("init consumed %d bytes, want 13", s.pipes[0].read)
	}
	apiTestCall(t, s, &mem, 103, 2, 0)
	if s.pipes[0].read != uint32(len(packed)) || s.pipes[1].write != uint32(len(lzmaTestChild)) {
		t.Fatalf("pipe cursors: input=%d/%d output=%d", s.pipes[0].read, len(packed), s.pipes[1].write)
	}
	if got, err := mem.read(s.pipes[1].ptr, len(lzmaTestChild)); err != nil || string(got) != lzmaTestChild {
		t.Fatalf("decoded %q: %v", got, err)
	}
	apiTestCall(t, s, &mem, 104, 0, 0)
	apiTestCall(t, s, &mem, 104, apiFailure, 0)
	if s.lzmaDict != 0 {
		t.Fatalf("dictionary budget retained: %d", s.lzmaDict)
	}

	for _, tc := range []struct {
		name  string
		input []byte
		cap   uint64
	}{
		{"short-output", packed, 5},
		{"short-input", packed[:len(packed)-3], 64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mem memory
			s := &apiState{input: tc.input}
			apiTestCall(t, s, &mem, 31, 0, 0)
			apiTestCall(t, s, &mem, 30, 1, tc.cap)
			apiTestCall(t, s, &mem, 102, 0, 0, 1)
			status, err := s.invoke(context.Background(), 103, []uint64{0}, &mem, func(uint64) error { return nil })
			if err != nil || status != 0 || s.pipes[0].read <= 13 || s.pipes[1].write == 0 {
				t.Fatalf("partial stream: status=%d err=%v, in=%d, out=%d", status, err, s.pipes[0].read, s.pipes[1].write)
			}
		})
	}
}

func TestAPILZMALeavesTrailingInput(t *testing.T) {
	for _, knownSize := range []bool{true, false} {
		packed := append(lzmaTestPackedConfig(t, knownSize), 'X')
		var mem memory
		s := &apiState{input: packed}
		apiTestCall(t, s, &mem, 31, 0, 0)
		apiTestCall(t, s, &mem, 30, 1, 64)
		apiTestCall(t, s, &mem, 102, 0, 0, 1)
		apiTestCall(t, s, &mem, 103, 2, 0)
		if s.pipes[0].read != uint32(len(packed)-1) || s.pipeReadAvail(&s.pipes[0]) != 1 {
			t.Fatalf("known size %t: compressed input cursor %d/%d", knownSize, s.pipes[0].read, len(packed))
		}
		ptr, err := s.invoke(context.Background(), 33, []uint64{0, 1}, &mem, func(uint64) error { return nil })
		if err != nil || ptr == 0 {
			t.Fatalf("known size %t: trailing pointer %#x, %v", knownSize, ptr, err)
		}
		if got, err := mem.read(ptr, 1); err != nil || string(got) != "X" {
			t.Fatalf("known size %t: trailing byte %q, %v", knownSize, got, err)
		}
	}
}

func TestAPILZMAMemoryPipeAndInvalidHeader(t *testing.T) {
	packed := lzmaTestPacked(t)
	var mem memory
	s := &apiState{}
	apiTestCall(t, s, &mem, 30, 0, uint64(len(packed)))
	if err := mem.write(s.pipes[0].ptr, packed); err != nil {
		t.Fatal(err)
	}
	apiTestCall(t, s, &mem, 37, 0, 0, uint64(len(packed)))
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 102, 0, 0, 1)
	apiTestCall(t, s, &mem, 103, 2, 0)
	if s.pipes[0].read != 0 || s.pipes[0].write != 0 {
		t.Fatalf("memory pipe not reset: %+v", s.pipes[0])
	}
	if got, err := mem.read(s.pipes[1].ptr, len(lzmaTestChild)); err != nil || string(got) != lzmaTestChild {
		t.Fatalf("decoded %q: %v", got, err)
	}

	bad := bytes.Clone(packed)
	bad[0] = 225
	var badMem memory
	badState := &apiState{input: bad}
	apiTestCall(t, badState, &badMem, 31, 0, 0)
	apiTestCall(t, badState, &badMem, 30, 1, 64)
	apiTestCall(t, badState, &badMem, 102, 1, 0, 1)
	if badState.pipes[0].read != 13 {
		t.Fatalf("bad header consumed %d", badState.pipes[0].read)
	}
	apiTestCall(t, badState, &badMem, 104, 0, 0)

	short := &apiState{input: packed[:12]}
	var shortMem memory
	apiTestCall(t, short, &shortMem, 31, 0, 0)
	apiTestCall(t, short, &shortMem, 30, 1, 64)
	apiTestCall(t, short, &shortMem, 102, apiFailure, 0, 1)
	if short.pipes[0].read != 0 || len(short.lzmas) != 0 {
		t.Fatal("short header changed state")
	}
}

func lzmaSerializedProgram() memoryFixture {
	m := bzip2SerializedProgram()
	m.apiMax = 107
	m.apis[len(m.apis)-3] = memoryAPI(102, 70, "lzma_init")
	m.apis[len(m.apis)-2] = memoryAPI(103, 69, "lzma_process")
	m.apis[len(m.apis)-1] = memoryAPI(104, 69, "lzma_done")
	for i := range m.functions[0].blocks {
		for j, line := range m.functions[0].blocks[i] {
			switch line {
			case memoryAPICall(2, 105, compatNum(0), compatNum(1)):
				m.functions[0].blocks[i][j] = memoryAPICall(2, 102, compatNum(0), compatNum(1))
			case memoryAPICall(3, 106, compatNum(2)):
				m.functions[0].blocks[i][j] = memoryAPICall(3, 103, compatNum(2))
			case memoryAPICall(7, 107, compatNum(2)):
				m.functions[0].blocks[i][j] = memoryAPICall(7, 104, compatNum(2))
			case compatInst(1, 8, 21, compatNum(32), compatNum(3), compatConstant(4, 32)):
				m.functions[0].blocks[i][j] = compatInst(1, 8, 21, compatNum(32), compatNum(3), compatConstant(2, 32))
			}
		}
	}
	return m
}

func TestAPILZMASerializedChildScan(t *testing.T) {
	input := append([]byte("AB"), lzmaTestPacked(t)...)
	p := memoryParse(t, lzmaSerializedProgram())
	var scanned int
	result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(child []byte, id int32) (string, error) {
		scanned++
		if id != -1 || !bytes.Equal(child, []byte(lzmaTestChild)) {
			t.Errorf("child=%q id=%d", child, id)
		}
		return "Child.LZMA", nil
	}, 1000)
	if err != nil || result.Value != 2 || result.ChildName != "Child.LZMA" || scanned != 1 {
		t.Fatalf("serialized child result=%+v err=%v scans=%d", result, err, scanned)
	}
}

func lzmaTrailingProgram() memoryFixture {
	m := lzmaSerializedProgram()
	m.functions[0].blocks[1][1] = compatInst(1, 10, 21, compatNum(32), compatNum(9), compatConstant(1, 32))
	return m
}

func TestAPILZMASerializedTrailingChildScan(t *testing.T) {
	p := memoryParse(t, lzmaTrailingProgram())
	for _, knownSize := range []bool{true, false} {
		for _, suffix := range []string{"X", ""} {
			input := append(append([]byte("AB"), lzmaTestPackedConfig(t, knownSize)...), suffix...)
			var scanned int
			result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(child []byte, id int32) (string, error) {
				scanned++
				if id != -1 || !bytes.Equal(child, []byte(lzmaTestChild)) {
					t.Errorf("child=%q id=%d", child, id)
				}
				return "Child.LZMA", nil
			}, 1000)
			wantScans := 0
			if suffix != "" {
				wantScans = 1
			}
			if err != nil || result.Value != 2 || scanned != wantScans {
				t.Fatalf("known size %t suffix %q: result=%+v err=%v scans=%d", knownSize, suffix, result, err, scanned)
			}
		}
	}
}

func TestWriteLZMAOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_LZMA_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_LZMA_ORACLE_DIR for a harmless ClamAV differential")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := lzmaSerializedProgram()
	m.kind = 256
	m.logical = "UTM.LZMA;Engine:56-255,Target:0;0;4142"
	for name, data := range map[string][]byte{
		"program.cbc":           m.encode(),
		"child.ndb":             []byte("Child.LZMA:0:*:" + hex.EncodeToString([]byte(lzmaTestChild)) + "\n"),
		"sample-ab.bin":         append([]byte("AB"), lzmaTestPacked(t)...),
		"sample-ac.bin":         append([]byte("AC"), lzmaTestPacked(t)...),
		"sample-ab-unknown.bin": append([]byte("AB"), lzmaTestPackedConfig(t, false)...),
		"sample-ac-unknown.bin": append([]byte("AC"), lzmaTestPackedConfig(t, false)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWriteLZMATrailingOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_LZMA_TRAILING_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_LZMA_TRAILING_ORACLE_DIR for a harmless ClamAV differential")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := lzmaTrailingProgram()
	m.kind = 256
	m.logical = "UTM.LZMA.Trailing;Engine:56-255,Target:0;0;4142"
	for name, data := range map[string][]byte{
		"program.cbc":          m.encode(),
		"child.ndb":            []byte("Child.LZMA:0:*:" + hex.EncodeToString([]byte(lzmaTestChild)) + "\n"),
		"known-trailing.bin":   append(append([]byte("AB"), lzmaTestPacked(t)...), 'X'),
		"known-plain.bin":      append([]byte("AB"), lzmaTestPacked(t)...),
		"unknown-trailing.bin": append(append([]byte("AB"), lzmaTestPackedConfig(t, false)...), 'X'),
		"unknown-plain.bin":    append([]byte("AB"), lzmaTestPackedConfig(t, false)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
