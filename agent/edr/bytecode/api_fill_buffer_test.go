package bytecode

import (
	"context"
	"errors"
	"testing"
)

func TestAPIFillBufferRefillsAndPreservesReferenceEOFOrder(t *testing.T) {
	var mem memory
	ptr := apiTestBytes(t, &mem, []byte("abCDefgh"))
	s := &apiState{input: []byte("123456")}
	apiTestCall(t, s, &mem, 21, 8, ptr, 8, 4, 2, 99) // fill is ignored
	if got, err := mem.read(ptr, 8); err != nil || string(got) != "CD123456" || s.offset != 6 {
		t.Fatalf("refill: %q, %v, offset %d", got, err, s.offset)
	}
	// The reference checks file EOF before shifting the unread suffix.
	apiTestCall(t, s, &mem, 21, 0, ptr, 8, 8, 3, 1)
	if got, _ := mem.read(ptr, 8); string(got) != "CD123456" {
		t.Fatalf("EOF changed buffer: %q", got)
	}
	s.input, s.offset = []byte("XY"), 0
	apiTestCall(t, s, &mem, 21, 7, ptr, 8, 8, 3, 8)
	if got, _ := mem.read(ptr, 7); string(got) != "23456XY" || s.offset != 2 {
		t.Fatalf("short refill: %q, offset %d", got, s.offset)
	}
	// A full buffer leads to read(0), which returns 0 even with input left.
	s.input, s.offset = []byte("Z"), 0
	apiTestCall(t, s, &mem, 21, 0, ptr, 8, 8, 0, 8)
	if s.offset != 0 {
		t.Fatalf("zero-capacity read advanced input: %d", s.offset)
	}
}

func TestAPIFillBufferRejectsBadStateWithoutCleanResult(t *testing.T) {
	var mem memory
	ptr := apiTestBytes(t, &mem, []byte("abcdefgh"))
	s := &apiState{input: []byte("XYZ")}
	for _, args := range [][]uint64{
		{ptr, 0, 0, 0, 1}, {ptr, 8, 9, 0, 1}, {ptr, 8, 2, 3, 1},
	} {
		apiTestCall(t, s, &mem, 21, apiFailure, args...)
		if got, _ := mem.read(ptr, 8); string(got) != "abcdefgh" || s.offset != 0 {
			t.Fatalf("invalid args %v changed state: %q offset %d", args, got, s.offset)
		}
	}
	if _, err := s.invoke(context.Background(), 21, []uint64{ptr + 1, 8, 0, 0, 8}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("out-of-bounds buffer: %v", err)
	}
	var partial memory
	part, err := partial.alloc(8, false)
	if err != nil || partial.write(part, []byte("abcd")) != nil {
		t.Fatal(err)
	}
	if _, err := s.invoke(context.Background(), 21, []uint64{part, 8, 8, 4, 8}, &partial, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) || s.offset != 0 {
		t.Fatalf("uninitialized unread suffix: %v, offset %d", err, s.offset)
	}
	if _, err := s.invoke(context.Background(), 21, []uint64{ptr, 8, 0, 0, 8}, &mem, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) || s.offset != 0 {
		t.Fatalf("work limit: %v, offset %d", err, s.offset)
	}
}

func TestAPIFillBufferWireDeclaration(t *testing.T) {
	for _, tc := range []struct {
		name  string
		proto string
		want  error
	}{
		{"fill_buffer", memoryType(1, 6, 32, 65, 32, 32, 32, 32), nil},
		{"fill_buffer", memoryType(1, 5, 32, 65, 32, 32, 32), ErrMalformed},
		{"read", memoryType(1, 6, 32, 65, 32, 32, 32, 32), ErrMalformed},
	} {
		m := memoryFixture{types: []string{tc.proto}, apiMax: 21,
			apis:      []string{memoryAPI(21, 69, tc.name)},
			functions: []compatFunction{{result: 32, blocks: [][]string{{compatRet(32, compatConstant(0, 32))}}}}}
		if _, err := Parse(m.encode()); !errors.Is(err, tc.want) {
			t.Fatalf("%s parse = %v; want %v", tc.name, err, tc.want)
		}
	}
	program := memoryFixture{
		types:  append(memoryArrayTypes(8, 8), memoryType(1, 6, 32, 65, 32, 32, 32, 32)),
		apiMax: 21,
		apis:   []string{memoryAPI(21, 71, "fill_buffer")},
		functions: []compatFunction{{locals: []int{0x8045, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(1, 21, compatNum(0), compatConstant(8, 32), compatConstant(0, 32), compatConstant(0, 32), compatConstant(8, 32)),
			compatRet(32, compatNum(1)),
		}}}},
	}
	result, err := memoryParse(t, program).RunWithInput(context.Background(), 0, nil, []byte("abcdef"), 1000)
	if err != nil || result.Value != 6 {
		t.Fatalf("serialized fill_buffer program: %+v, %v", result, err)
	}
}
