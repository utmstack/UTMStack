package bytecode

import (
	"context"
	"errors"
	"testing"
)

func TestTraceAPIsUseDisabledProfileAfterPointerValidation(t *testing.T) {
	var mem memory
	ptr := apiTestBytes(t, &mem, []byte("abc\x00"))
	state := &apiState{}
	for id := 9; id <= 14; id++ {
		apiTestCall(t, state, &mem, id, 0, ptr, 4)
		for _, args := range [][2]uint64{{0, 1}, {ptr + 4, 0}, {ptr, 5}} {
			if _, err := state.invoke(context.Background(), id, args[:], &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
				t.Fatalf("API %d pointer %v: %v", id, args, err)
			}
		}
	}
}

func TestTraceAPIsSerializedProgram(t *testing.T) {
	m := memoryFixture{
		types: []string{
			memoryType(4, 4, 8), memoryType(1, 3, 32, 65, 32),
		},
		globals: []string{memoryGlobal(69, 'a', 'b', 'c', 0), memoryGlobal(65, 0, 1)},
		apiMax:  14,
		apis:    []string{memoryAPI(9, 70, "trace_directory"), memoryAPI(14, 70, "trace_ptr")},
		functions: []compatFunction{{locals: []int{32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 9, memoryGlobalRef(2), compatConstant(4, 32)),
			memoryAPICall(1, 14, memoryGlobalRef(2), compatConstant(4, 32)),
			compatRet(32, compatNum(1)),
		}}}},
	}
	result, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, nil, 1000)
	if err != nil || result.Value != 0 {
		t.Fatalf("trace-disabled program = %+v, %v", result, err)
	}
}
