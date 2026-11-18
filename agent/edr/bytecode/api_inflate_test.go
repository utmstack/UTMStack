package bytecode

import (
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"math/rand"
	"testing"
)

func TestAPIInflatePipesAcrossInputAndOutputWindows(t *testing.T) {
	raw := make([]byte, 20<<10)
	if _, err := rand.New(rand.NewSource(185)).Read(raw); err != nil {
		t.Fatal(err)
	}
	for _, wrapped := range []bool{false, true} {
		packed := packedForInflate(t, raw, flate.DefaultCompression, wrapped)
		bits := uint64(0xfffffff1) // int32(-15) in the bytecode wire ABI
		if wrapped {
			bits = 15
		}
		var mem memory
		s := &apiState{input: packed}
		apiTestCall(t, s, &mem, 31, 0, 0)
		apiTestCall(t, s, &mem, 30, 1, 128)
		apiTestCall(t, s, &mem, 39, 0, 0, 1, bits)
		var output []byte
		ended := false
		for step := 0; step < 1000; step++ {
			status, err := s.invoke(context.Background(), 40, []uint64{0}, &mem, func(uint64) error { return nil })
			if err != nil {
				t.Fatalf("wrapped=%t step=%d: %v", wrapped, step, err)
			}
			available := s.pipeReadAvail(&s.pipes[1])
			if available > 0 {
				ptr, err := s.invoke(context.Background(), 33, []uint64{1, uint64(available)}, &mem, func(uint64) error { return nil })
				if err != nil || ptr == 0 {
					t.Fatalf("read_get = %#x, %v", ptr, err)
				}
				chunk, err := mem.read(ptr, int(available))
				if err != nil {
					t.Fatal(err)
				}
				output = append(output, chunk...)
				apiTestCall(t, s, &mem, 34, 0, 1, uint64(available))
			}
			if status == inflateStreamEnd {
				ended = true
				break
			}
			if status != inflateOK {
				t.Fatalf("wrapped=%t step=%d status=%d", wrapped, step, status)
			}
		}
		if !ended || !bytes.Equal(output, raw) || s.pipes[0].read != uint32(len(packed)) {
			t.Fatalf("wrapped=%t end=%t output=%d/%d input=%d/%d", wrapped, ended, len(output), len(raw), s.pipes[0].read, len(packed))
		}
		apiTestCall(t, s, &mem, 41, apiFailure, 0) // stream end called done
	}
}

func TestAPIInflateInvalidAndCorruptStatus(t *testing.T) {
	var mem memory
	s := &apiState{input: packedForInflate(t, []byte("safe"), flate.DefaultCompression, true)}
	apiTestCall(t, s, &mem, 39, apiFailure, 0, 1, 15) // no pipes
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 39, apiFailure, 0, 1, 99)
	apiTestCall(t, s, &mem, 40, apiFailure, 0)
	apiTestCall(t, s, &mem, 39, 1, 0, 1, 15)
	apiTestCall(t, s, &mem, 41, 0, 1)
	apiTestCall(t, s, &mem, 40, apiFailure, 1)
	s.input[len(s.input)-1] ^= 1
	apiTestCall(t, s, &mem, 39, 2, 0, 1, 15)
	if got, err := s.invoke(context.Background(), 40, []uint64{2}, &mem, func(uint64) error { return nil }); err != nil || got != 0xfffffffd {
		t.Fatalf("corrupt stream status = %#x, %v", got, err)
	}
	if _, err := s.invoke(context.Background(), 39, []uint64{0, 1, 15}, &mem, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) {
		t.Fatalf("work budget = %v", err)
	}
}

func TestAPIInflateSyncMarkerAcrossFilePipeWindows(t *testing.T) {
	child := []byte("RECOVERED-PIPE-CHILD")
	window := int(pipeFileWindow())
	input := append([]byte{0x06}, bytes.Repeat([]byte{0x7f}, window-3)...)
	input = append(input, 0, 0, 0xff, 0xff)
	input = append(input, packedForInflate(t, child, flate.DefaultCompression, false)...)
	var mem memory
	s := &apiState{input: input}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 128)
	apiTestCall(t, s, &mem, 39, 0, 0, 1, 0xfffffff1)
	apiTestCall(t, s, &mem, 40, 0xfffffffd, 0) // marker prefix was read
	if s.pipes[0].read != uint32(window) || s.pipes[1].write != 0 {
		t.Fatalf("first window cursors: read=%d write=%d", s.pipes[0].read, s.pipes[1].write)
	}
	apiTestCall(t, s, &mem, 40, inflateStreamEnd, 0)
	if s.pipes[0].read != uint32(len(input)) || s.pipeReadAvail(&s.pipes[1]) != uint32(len(child)) {
		t.Fatalf("recovered cursors: read=%d/%d output=%d/%d", s.pipes[0].read, len(input), s.pipeReadAvail(&s.pipes[1]), len(child))
	}
	ptr, err := s.invoke(context.Background(), 33, []uint64{1, uint64(len(child))}, &mem, func(uint64) error { return nil })
	if err != nil || ptr == 0 {
		t.Fatalf("output pointer = %#x, %v", ptr, err)
	}
	got, err := mem.read(ptr, len(child))
	if err != nil || !bytes.Equal(got, child) {
		t.Fatalf("recovered output = %q, %v", got, err)
	}
}

func TestAPIInflatePresetDictionaryStatus(t *testing.T) {
	input := append([]byte{0x78, 0xbb, 1, 2, 3, 4}, packedForInflate(t, []byte("dict"), flate.DefaultCompression, false)...)
	var mem memory
	s := &apiState{input: input}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 39, 0, 0, 1, 15)
	apiTestCall(t, s, &mem, 40, inflateNeedDict, 0)
	if s.pipes[0].read != 6 || s.pipes[1].write != 0 {
		t.Fatalf("dictionary cursors: read=%d write=%d", s.pipes[0].read, s.pipes[1].write)
	}
	apiTestCall(t, s, &mem, 40, inflateNeedDict, 0)
	apiTestCall(t, s, &mem, 41, 0, 0)
}

func TestAPIInflateSerializedDispatch(t *testing.T) {
	input := packedForInflate(t, []byte("bytecode inflate"), flate.DefaultCompression, true)
	m := memoryFixture{
		types:  []string{memoryType(1, 2, 32, 32), memoryType(1, 4, 32, 32, 32, 32)},
		apiMax: 40,
		apis: []string{
			memoryAPI(30, 69, "buffer_pipe_new"), memoryAPI(31, 69, "buffer_pipe_new_fromfile"),
			memoryAPI(39, 70, "inflate_init"), memoryAPI(40, 69, "inflate_process"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 31, compatConstant(0, 32)),
			memoryAPICall(1, 30, compatConstant(64, 32)),
			memoryAPICall(2, 39, compatNum(0), compatNum(1), compatConstant(15, 32)),
			memoryAPICall(3, 40, compatNum(2)), compatRet(32, compatNum(3)),
		}}}},
	}
	result, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, input, 1000)
	if err != nil || result.Value != inflateStreamEnd {
		t.Fatalf("serialized inflate = %+v, %v", result, err)
	}
}

func TestAPIInflateSerializedOutputReachesChildScanner(t *testing.T) {
	child := []byte("HARMLESS-CHILD")
	input := packedForInflate(t, child, flate.DefaultCompression, true)
	m := memoryFixture{
		types: []string{
			memoryType(1, 2, 32, 32),         // 69: i32(i32)
			memoryType(1, 4, 32, 32, 32, 32), // 70: i32(i32,i32,i32)
			memoryType(1, 3, 32, 32, 32),     // 71: i32(i32,i32)
			memoryType(1, 3, 65, 32, 32),     // 72: i8*(i32,i32)
			memoryType(1, 3, 32, 65, 32),     // 73: i32(i8*,i32)
		},
		apiMax: 40,
		apis: []string{
			memoryAPI(3, 73, "write"), memoryAPI(30, 69, "buffer_pipe_new"),
			memoryAPI(31, 69, "buffer_pipe_new_fromfile"),
			memoryAPI(32, 69, "buffer_pipe_read_avail"),
			memoryAPI(33, 72, "buffer_pipe_read_get"),
			memoryAPI(39, 70, "inflate_init"), memoryAPI(40, 69, "inflate_process"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32, 32, 65, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 31, compatConstant(0, 32)),
			memoryAPICall(1, 30, compatConstant(64, 32)),
			memoryAPICall(2, 39, compatNum(0), compatNum(1), compatConstant(15, 32)),
			memoryAPICall(3, 40, compatNum(2)),
			memoryAPICall(4, 32, compatNum(1)),
			compatInst(65, 5, 33, compatFixed(2, 1), compatNum(33), compatNum(1), compatNum(4)),
			memoryAPICall(6, 3, compatNum(5), compatNum(4)),
			compatRet(32, compatNum(3)),
		}}}},
	}
	var scanned int
	result, err := memoryParse(t, m).RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(data []byte, id int32) (string, error) {
		scanned++
		if id != -1 || !bytes.Equal(data, child) {
			t.Errorf("decoded child = %q, id=%d", data, id)
		}
		return "Child.Hash.Detected", nil
	}, 1000)
	if err != nil || result.Value != inflateStreamEnd || result.ChildName != "Child.Hash.Detected" || scanned != 1 {
		t.Fatalf("result=%+v, err=%v, scanned=%d", result, err, scanned)
	}
}
