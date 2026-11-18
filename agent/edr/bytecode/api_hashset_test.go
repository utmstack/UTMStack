package bytecode

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestAPIHashsetLifecycle(t *testing.T) {
	state := &apiState{}
	apiTestCall(t, state, nil, 24, 0)
	apiTestCall(t, state, nil, 24, 1)
	apiTestCall(t, state, nil, 29, 1, 0)
	apiTestCall(t, state, nil, 25, 0, 0, 0)
	apiTestCall(t, state, nil, 25, 0, 0, 0) // duplicate add succeeds
	apiTestCall(t, state, nil, 25, 0, 0, math.MaxUint32)
	apiTestCall(t, state, nil, 27, 1, 0, 0)
	apiTestCall(t, state, nil, 27, 1, 0, math.MaxUint32)
	apiTestCall(t, state, nil, 27, 0, 1, 0) // sets do not share keys
	apiTestCall(t, state, nil, 29, 0, 0)
	apiTestCall(t, state, nil, 26, apiFailure, 0, 123)
	apiTestCall(t, state, nil, 26, 0, 0, 0)
	apiTestCall(t, state, nil, 26, 0, 0, math.MaxUint32)
	apiTestCall(t, state, nil, 29, 1, 0)
	apiTestCall(t, state, nil, 28, 0, 0)
	apiTestCall(t, state, nil, 29, 1, math.MaxUint32) // invalid ID
	if _, err := state.invoke(context.Background(), 27, []uint64{0, 0}, nil, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("use after done: %v", err)
	}
	apiTestCall(t, state, nil, 28, 0, 1)
	apiTestCall(t, state, nil, 24, 1) // last slot is reused
	apiTestCall(t, state, nil, 29, 1, 1)
	apiTestCall(t, &apiState{}, nil, 24, 0) // no cross-run state
}

func TestAPIHashsetResourceBounds(t *testing.T) {
	state := &apiState{}
	apiTestCall(t, state, nil, 24, 0)
	state.hashKeys = maxHashKeys
	apiTestCall(t, state, nil, 27, 0, 0, 1)
	if _, err := state.invoke(context.Background(), 25, []uint64{0, 1}, nil, func(uint64) error { return nil }); !errors.Is(err, ErrLimit) {
		t.Fatalf("key limit: %v", err)
	}
	state = &apiState{hashsets: make([]map[uint32]struct{}, maxHashsets)}
	if _, err := state.invoke(context.Background(), 24, nil, nil, func(uint64) error { return nil }); !errors.Is(err, ErrLimit) {
		t.Fatalf("set limit: %v", err)
	}
}

func TestProgramHashsetCalls(t *testing.T) {
	program := memoryFixture{
		types: []string{
			memoryType(1, 1, 32),
			memoryType(1, 3, 32, 32, 32),
			memoryType(1, 2, 32, 32),
		},
		apiMax: 29,
		apis: []string{
			memoryAPI(24, 69, "hashset_new"),
			memoryAPI(25, 70, "hashset_add"),
			memoryAPI(27, 70, "hashset_contains"),
			memoryAPI(28, 71, "hashset_done"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 24),
			memoryAPICall(1, 25, compatNum(0), compatConstant(42, 32)),
			memoryAPICall(2, 27, compatNum(0), compatConstant(42, 32)),
			memoryAPICall(3, 28, compatNum(0)),
			compatRet(32, compatNum(2)),
		}}}},
	}
	result, err := memoryParse(t, program).RunWithInput(context.Background(), 0, nil, nil, 1000)
	if err != nil || result.Value != 1 {
		t.Fatalf("hashset program: %+v, %v", result, err)
	}
}
