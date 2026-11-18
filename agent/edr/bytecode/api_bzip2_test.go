package bytecode

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const bzip2TestHex = "425a6831314159265359d8672502000002960000022e66180020003100d34d0401a32213cc05f144af177245385090d8672502"
const bzip2TestChild = "HARMLESS-CHILD"

func bzip2TestPacked(t *testing.T) []byte {
	t.Helper()
	b, err := hex.DecodeString(bzip2TestHex)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAPIBzip2CompleteStreamAndBoundaries(t *testing.T) {
	packed := bzip2TestPacked(t)
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 105, apiFailure, 0, 1) // no pipes
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 105, 0, 0, 1)
	apiTestCall(t, s, &mem, 106, 4, 0)
	if s.pipes[0].read != uint32(len(packed)) || s.pipeReadAvail(&s.pipes[1]) != uint32(len(bzip2TestChild)) {
		t.Fatalf("pipe cursors after stream end: in=%d/%d out=%d/%d", s.pipes[0].read, len(packed), s.pipeReadAvail(&s.pipes[1]), len(bzip2TestChild))
	}
	ptr, err := s.invoke(context.Background(), 33, []uint64{1, uint64(len(bzip2TestChild))}, &mem, func(uint64) error { return nil })
	if err != nil || ptr == 0 {
		t.Fatalf("output pointer: %#x %v", ptr, err)
	}
	if got, err := mem.read(ptr, len(bzip2TestChild)); err != nil || string(got) != bzip2TestChild {
		t.Fatalf("output: %q %v", got, err)
	}
	apiTestCall(t, s, &mem, 106, apiFailure, 0)
	apiTestCall(t, s, &mem, 107, 0, 0)
	apiTestCall(t, s, &mem, 107, apiFailure, 0)

	// Partial windows preserve real progress and do not claim stream end.
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
			apiTestCall(t, s, &mem, 105, 0, 0, 1)
			status, err := s.invoke(context.Background(), 106, []uint64{0}, &mem, func(uint64) error { return nil })
			if err != nil || status != bzipOK || s.pipes[0].read == 0 || s.pipes[1].write == 0 {
				t.Fatalf("partial stream: %v, in=%d, out=%d", err, s.pipes[0].read, s.pipes[1].write)
			}
		})
	}
}

func TestBzip2FirstStreamBitAlignment(t *testing.T) {
	for _, tc := range []struct {
		name, packed, want string
	}{
		{"empty-byte-aligned", "425a683117724538509000000000", ""},
		{"one-byte-offset-three", "425a683131415926535981b02d8b00000004002000200021184682ee48a70a12103605b160", "A"},
		{"two-byte-offset-four", "425a683131415926535951137d0b000000040030002000210082b17724538509051137d0b0", "AB"},
		{"range-offset-six", "425a6831314159265359707e66af00000060007fffffffa000318c0013000130a04c4c98098007043217c31d1ea55f2d8d73e11fe53a56d7c675be745dc914e14241c1f99abc", string([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			input := append(bytes.Clone(packed), 'X', 'Y')
			decoded := make([]byte, 64)
			stream := new(streamBzip2)
			consumed, produced, status, err := stream.process(input, decoded, nil)
			if err != nil || status != bzipEnd || consumed != len(packed) || string(decoded[:produced]) != tc.want {
				t.Fatalf("decoded %q, consumed %d/%d, status %d, error %v", decoded[:produced], consumed, len(packed), status, err)
			}
		})
	}
}

func TestAPIBzip2OutputWindowsAndMemoryReservation(t *testing.T) {
	packed := bzip2TestPacked(t)
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 5)
	apiTestCall(t, s, &mem, 105, 0, 0, 1)
	var decoded []byte
	for calls := 0; calls < 20; calls++ {
		status, err := s.invoke(context.Background(), 106, []uint64{0}, &mem, func(uint64) error { return nil })
		if err != nil || status != bzipOK && status != bzipEnd {
			t.Fatalf("process: %d %v", status, err)
		}
		n := s.pipeReadAvail(&s.pipes[1])
		data, err := mem.read(s.pipes[1].ptr+uint64(s.pipes[1].read), int(n))
		if err != nil {
			t.Fatal(err)
		}
		decoded = append(decoded, data...)
		apiTestCall(t, s, &mem, 34, 0, 1, uint64(n))
		if status == bzipEnd {
			if string(decoded) != bzip2TestChild || s.pipes[0].read != uint32(len(packed)) || s.bzip2Memory != 400000 {
				t.Fatalf("completion: %q cursor=%d reservation=%d", decoded, s.pipes[0].read, s.bzip2Memory)
			}
			apiTestCall(t, s, &mem, 107, 0, 0)
			if s.bzip2Memory != 0 || s.bzip2s[0].stream != nil {
				t.Fatal("done retains block reservation")
			}
			return
		}
	}
	t.Fatal("did not complete")
}

func TestAPIBzip2SharedMemoryLimit(t *testing.T) {
	var mem memory
	s := &apiState{input: []byte("BZh9"), bzip2Memory: maxBzip2Memory - 3599999}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 1)
	apiTestCall(t, s, &mem, 105, 0, 0, 1)
	_, err := s.invoke(context.Background(), 106, []uint64{0}, &mem, func(uint64) error { return nil })
	if !errors.Is(err, ErrLimit) || s.bzip2Memory != maxBzip2Memory-3599999 || s.bzip2s[0].reserved != 0 || s.pipes[0].read != 0 {
		t.Fatalf("reservation failure: %v total=%d reserved=%d read=%d", err, s.bzip2Memory, s.bzip2s[0].reserved, s.pipes[0].read)
	}
	apiTestCall(t, s, &mem, 107, 0, 0)
	if s.bzip2Memory != maxBzip2Memory-3599999 {
		t.Fatal("done released another stream's memory")
	}
}

func TestAPIBzip2LeavesTrailingInput(t *testing.T) {
	packed := append(bzip2TestPacked(t), 'X')
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 105, 0, 0, 1)
	apiTestCall(t, s, &mem, 106, 4, 0)
	if s.pipes[0].read != uint32(len(packed)-1) || s.pipeReadAvail(&s.pipes[0]) != 1 {
		t.Fatalf("compressed input cursor: %d/%d", s.pipes[0].read, len(packed))
	}
	ptr, err := s.invoke(context.Background(), 33, []uint64{0, 1}, &mem, func(uint64) error { return nil })
	if err != nil || ptr == 0 {
		t.Fatalf("trailing byte pointer = %#x, %v", ptr, err)
	}
	if got, err := mem.read(ptr, 1); err != nil || string(got) != "X" {
		t.Fatalf("trailing byte = %q, %v", got, err)
	}
}

func TestAPIBzip2MemoryPipeAndEmptyStream(t *testing.T) {
	for _, tc := range []struct {
		name, packed string
		want         string
	}{
		{"content", bzip2TestHex, bzip2TestChild},
		{"empty", "425a683117724538509000000000", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			var mem memory
			s := &apiState{}
			apiTestCall(t, s, &mem, 30, 0, uint64(len(packed)))
			if err := mem.write(s.pipes[0].ptr, packed); err != nil {
				t.Fatal(err)
			}
			apiTestCall(t, s, &mem, 37, 0, 0, uint64(len(packed)))
			apiTestCall(t, s, &mem, 30, 1, 64)
			apiTestCall(t, s, &mem, 105, 0, 0, 1)
			apiTestCall(t, s, &mem, 106, 4, 0)
			if s.pipes[0].read != 0 || s.pipes[0].write != 0 || s.pipes[1].write != uint32(len(tc.want)) {
				t.Fatalf("pipe cursors after completion: %+v %+v", s.pipes[0], s.pipes[1])
			}
			if got, err := mem.read(s.pipes[1].ptr, len(tc.want)); err != nil || string(got) != tc.want {
				t.Fatalf("decoded %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func bzip2SerializedProgram() memoryFixture {
	m := memoryFixture{
		types: []string{
			memoryType(1, 2, 32, 32),     // 69: i32(i32)
			memoryType(1, 3, 32, 32, 32), // 70: i32(i32,i32)
			memoryType(1, 3, 65, 32, 32), // 71: i8*(i32,i32)
			memoryType(1, 3, 32, 65, 32), // 72: i32(i8*,i32)
		},
		apiMax: 107,
		apis: []string{
			memoryAPI(3, 72, "write"), memoryAPI(30, 69, "buffer_pipe_new"),
			memoryAPI(31, 69, "buffer_pipe_new_fromfile"),
			memoryAPI(32, 69, "buffer_pipe_read_avail"),
			memoryAPI(33, 71, "buffer_pipe_read_get"),
			memoryAPI(105, 70, "bzip2_init"), memoryAPI(106, 69, "bzip2_process"),
			memoryAPI(107, 69, "bzip2_done"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32, 32, 65, 32, 32, 1, 32, 1}, result: 32, blocks: [][]string{
			{
				memoryAPICall(0, 31, compatConstant(2, 32)),
				memoryAPICall(1, 30, compatConstant(64, 32)),
				memoryAPICall(2, 105, compatNum(0), compatNum(1)),
				memoryAPICall(3, 106, compatNum(2)),
				compatInst(1, 8, 21, compatNum(32), compatNum(3), compatConstant(4, 32)),
				"T" + compatFixed(17, 2) + compatNum(8) + compatNum(1) + compatNum(3),
			},
			{
				memoryAPICall(9, 32, compatNum(0)),
				compatInst(1, 10, 21, compatNum(32), compatNum(9), compatConstant(0, 32)),
				"T" + compatFixed(17, 2) + compatNum(10) + compatNum(2) + compatNum(3),
			},
			{
				memoryAPICall(4, 32, compatNum(1)),
				compatInst(65, 5, 33, compatFixed(2, 1), compatNum(33), compatNum(1), compatNum(4)),
				memoryAPICall(6, 3, compatNum(5), compatNum(4)),
				memoryAPICall(7, 107, compatNum(2)),
				compatRet(32, compatNum(3)),
			},
			{compatRet(32, compatNum(3))},
		}}},
	}
	return m
}

func TestAPIBzip2SerializedChildScan(t *testing.T) {
	input := append([]byte("AB"), bzip2TestPacked(t)...)
	p := memoryParse(t, bzip2SerializedProgram())
	var scanned int
	result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(child []byte, id int32) (string, error) {
		scanned++
		if id != -1 || !bytes.Equal(child, []byte(bzip2TestChild)) {
			t.Errorf("child=%q id=%d", child, id)
		}
		return "Child.Bzip2", nil
	}, 1000)
	if err != nil || result.Value != 4 || result.ChildName != "Child.Bzip2" || scanned != 1 {
		t.Fatalf("serialized child result=%+v err=%v scans=%d", result, err, scanned)
	}
}

func bzip2TrailingProgram() memoryFixture {
	m := bzip2SerializedProgram()
	// Scan the decoded child only when one byte remains in the source pipe.
	m.functions[0].blocks[1][1] = compatInst(1, 10, 21, compatNum(32), compatNum(9), compatConstant(1, 32))
	return m
}

func TestAPIBzip2SerializedTrailingChildScan(t *testing.T) {
	p := memoryParse(t, bzip2TrailingProgram())
	for _, tc := range []struct {
		name, suffix string
		wantScans    int
	}{
		{"trailing-byte", "X", 1},
		{"no-trailing-byte", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := append(append([]byte("AB"), bzip2TestPacked(t)...), tc.suffix...)
			var scanned int
			result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(child []byte, id int32) (string, error) {
				scanned++
				if id != -1 || !bytes.Equal(child, []byte(bzip2TestChild)) {
					t.Errorf("child=%q id=%d", child, id)
				}
				return "Child.Bzip2", nil
			}, 1000)
			if err != nil || result.Value != 4 || scanned != tc.wantScans {
				t.Fatalf("result=%+v err=%v scans=%d, want %d", result, err, scanned, tc.wantScans)
			}
		})
	}
}

func TestWriteBzip2OracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_BZIP2_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_BZIP2_ORACLE_DIR for a harmless ClamAV differential")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := bzip2SerializedProgram()
	m.kind = 256
	m.logical = "UTM.Bzip2;Engine:56-255,Target:0;0;4142"
	for name, data := range map[string][]byte{
		"program.cbc":   m.encode(),
		"child.ndb":     []byte("Child.Bzip2:0:*:" + hex.EncodeToString([]byte(bzip2TestChild)) + "\n"),
		"sample-ab.bin": append([]byte("AB"), bzip2TestPacked(t)...),
		"sample-ac.bin": append([]byte("AC"), bzip2TestPacked(t)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWriteBzip2TrailingOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_BZIP2_TRAILING_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_BZIP2_TRAILING_ORACLE_DIR for a harmless ClamAV differential")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := bzip2TrailingProgram()
	m.kind = 256
	m.logical = "UTM.Bzip2.Trailing;Engine:56-255,Target:0;0;4142"
	for name, data := range map[string][]byte{
		"program.cbc":    m.encode(),
		"child.ndb":      []byte("Child.Bzip2:0:*:" + hex.EncodeToString([]byte(bzip2TestChild)) + "\n"),
		"sample-ab.bin":  append(append([]byte("AB"), bzip2TestPacked(t)...), 'X'),
		"sample-raw.bin": append([]byte("AB"), bzip2TestPacked(t)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
