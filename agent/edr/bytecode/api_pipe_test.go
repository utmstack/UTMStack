package bytecode

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestAPIBufferPipeFixedCursorAndLifetime(t *testing.T) {
	var mem memory
	s := &apiState{}
	apiTestCall(t, s, &mem, 30, 0, 4)
	apiTestCall(t, s, &mem, 32, 0, 0)
	apiTestCall(t, s, &mem, 35, 4, 0)
	write, err := s.invoke(context.Background(), 36, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil || write == 0 {
		t.Fatalf("write_get = %#x, %v", write, err)
	}
	if err := mem.write(write, []byte("ABC")); err != nil {
		t.Fatal(err)
	}
	apiTestCall(t, s, &mem, 37, 0, 0, 3)
	apiTestCall(t, s, &mem, 32, 3, 0)
	apiTestCall(t, s, &mem, 35, 1, 0)
	read, err := s.invoke(context.Background(), 33, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil || read != write {
		t.Fatalf("read_get = %#x, %v; want %#x", read, err, write)
	}
	data, err := mem.read(read, 3)
	if err != nil || string(data) != "ABC" {
		t.Fatalf("pipe data = %q, %v", data, err)
	}
	apiTestCall(t, s, &mem, 34, 0, 0, 2)
	apiTestCall(t, s, &mem, 32, 1, 0)
	apiTestCall(t, s, &mem, 37, 0, 0, 100) // overrun clamps at capacity
	apiTestCall(t, s, &mem, 35, 0, 0)
	apiTestCall(t, s, &mem, 34, 0, 0, 100)
	apiTestCall(t, s, &mem, 35, 4, 0) // reset only after whole buffer consumed
	if got, err := mem.read(write, 4); err != nil || !bytes.Equal(got[:3], []byte("ABC")) {
		t.Fatalf("backing buffer = %q, %v", got, err)
	}
	apiTestCall(t, s, &mem, 38, 0, 0)
	if err := mem.check(write, 1, false); !errors.Is(err, ErrRuntime) || mem.used != 0 {
		t.Fatalf("retired pointer = %v, used = %d", err, mem.used)
	}
	if _, err := s.invoke(context.Background(), 32, []uint64{0}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("use after done = %v", err)
	}
	apiTestCall(t, s, &mem, 30, 1, 1) // IDs do not reuse the retired slot
	if err := mem.check(write, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("stale pointer revived: %v", err)
	}
}

func TestAPIBufferPipeFileWindowAndReadOnly(t *testing.T) {
	var mem memory
	input := bytes.Repeat([]byte{'X'}, int(pipeFileWindow())+3)
	input[1] = 'Q'
	s := &apiState{input: input}
	apiTestCall(t, s, &mem, 31, 0, 1)
	apiTestCall(t, s, &mem, 32, uint64(pipeFileWindow()), 0)
	apiTestCall(t, s, &mem, 35, 0, 0)
	apiTestCall(t, s, &mem, 36, 0, 0, 1)
	apiTestCall(t, s, &mem, 37, apiFailure, 0, 1)
	get := func() uint64 {
		t.Helper()
		ptr, err := s.invoke(context.Background(), 33, []uint64{0, 1}, &mem, func(uint64) error { return nil })
		if err != nil || ptr == 0 {
			t.Fatalf("read_get = %#x, %v", ptr, err)
		}
		return ptr
	}
	first := get()
	if got, err := mem.read(first, 1); err != nil || string(got) != "Q" {
		t.Fatalf("file pipe byte = %q, %v", got, err)
	}
	if err := mem.write(first, []byte{'Z'}); !errors.Is(err, ErrRuntime) {
		t.Fatalf("file pointer is writable: %v", err)
	}
	apiTestCall(t, s, &mem, 34, 0, 0, uint64(pipeFileWindow()))
	if second := get(); second != first+uint64(pipeFileWindow()) {
		t.Fatalf("file pointer identity lost: %#x, %#x", first, second)
	}
	apiTestCall(t, s, &mem, 32, 2, 0)
	apiTestCall(t, s, &mem, 34, 0, 0, 999)
	apiTestCall(t, s, &mem, 32, 0, 0)
	apiTestCall(t, s, &mem, 33, 0, 0, 1)
	apiTestCall(t, s, &mem, 38, 0, 0)
	if got, err := mem.read(first, 1); err != nil || string(got) != "Q" || mem.used != 0 {
		t.Fatalf("file mapping after pipe done = %q, %v, used %d", got, err, mem.used)
	}
	apiTestCall(t, s, &mem, 31, apiFailure, uint64(len(input)))
}

func TestAPIBufferPipeInvalidAndLimits(t *testing.T) {
	var mem memory
	s := &apiState{input: []byte("abc")}
	for _, id := range []int{32, 33, 35, 36} {
		args := []uint64{0xffffffff}
		if id == 33 || id == 36 {
			args = append(args, 1)
		}
		apiTestCall(t, s, &mem, id, 0, args...)
	}
	for _, id := range []int{34, 37, 38} {
		args := []uint64{0xffffffff}
		if id != 38 {
			args = append(args, 1)
		}
		apiTestCall(t, s, &mem, id, apiFailure, args...)
	}
	apiTestCall(t, s, &mem, 30, apiFailure, uint64(maxAllocation)+1)
	apiTestCall(t, s, &mem, 30, apiFailure, 0)
	apiTestCall(t, s, &mem, 30, 0, 1)
	apiTestCall(t, s, &mem, 33, 0, 0, 0)
	apiTestCall(t, s, &mem, 36, 0, 0, 2)
	apiTestCall(t, s, &mem, 34, apiFailure, 0, 1)
	apiTestCall(t, s, &mem, 37, 0, 0, 999)
	apiTestCall(t, s, &mem, 37, 0, 0, 1)
	apiTestCall(t, s, &mem, 35, 0, 0)
	if _, err := s.invoke(context.Background(), 37, []uint64{0, ^uint64(0)}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("write cursor overflow = %v", err)
	}
	if _, err := s.invoke(context.Background(), 30, []uint64{1}, &mem, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) || len(s.pipes) != 1 {
		t.Fatalf("work limit = %v, pipes %d", err, len(s.pipes))
	}
	apiTestCall(t, s, &mem, 30, 1, 3)
	apiTestCall(t, s, &mem, 37, 0, 1, 2)
	apiTestCall(t, s, &mem, 34, 0, 1, 1)
	if _, err := s.invoke(context.Background(), 34, []uint64{1, ^uint64(0)}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("read cursor overflow = %v", err)
	}
	s.pipes = make([]bufferPipe, maxPipes)
	if _, err := s.invoke(context.Background(), 31, []uint64{0}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrLimit) {
		t.Fatalf("pipe count limit = %v", err)
	}
}

func TestAPIBufferPipeSerializedFileRead(t *testing.T) {
	m := memoryFixture{
		types:  []string{memoryType(1, 2, 32, 32), memoryType(1, 3, 65, 32, 32)},
		apiMax: 33,
		apis:   []string{memoryAPI(31, 69, "buffer_pipe_new_fromfile"), memoryAPI(33, 70, "buffer_pipe_read_get")},
		functions: []compatFunction{{locals: []int{32, 65, 8}, result: 8, blocks: [][]string{{
			memoryAPICall(0, 31, compatConstant(1, 32)),
			compatInst(65, 1, 33, compatFixed(2, 1), compatNum(33), compatNum(0), compatConstant(1, 32)),
			compatInst(8, 2, 39, compatNum(1)), compatRet(8, compatNum(2)),
		}}}},
	}
	result, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, []byte("aQz"), 1000)
	if err != nil || result.Value != 'Q' {
		t.Fatalf("serialized file pipe = %+v, %v", result, err)
	}
}
