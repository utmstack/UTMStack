package bytecode

import (
	"context"
	"errors"
	"testing"
)

func TestAPIHex2UI(t *testing.T) {
	state := &apiState{}
	apiTestCall(t, state, nil, 52, 0x0a, '0', 'a')
	apiTestCall(t, state, nil, 52, 0x7f, '7', 'F')
	apiTestCall(t, state, nil, 52, 0xffffff80, '8', '0')
	apiTestCall(t, state, nil, 52, 0xffffffff, 'F', 'f')
	apiTestCall(t, state, nil, 52, apiFailure, 'x', '1')
	apiTestCall(t, state, nil, 52, apiFailure, '1', '?')
	apiTestCall(t, state, nil, 52, 0, 0x130, 0x130) // uint32 inputs truncate to bytes
}

func TestAPIEntropyBuffer(t *testing.T) {
	state := &apiState{}
	var mem memory
	mono := apiTestBytes(t, &mem, []byte("aaaa"))
	apiTestCall(t, state, &mem, 56, 0, mono, 4)
	balanced := apiTestBytes(t, &mem, []byte("abab"))
	apiTestCall(t, state, &mem, 56, 1<<26, balanced, 4)
	apiTestCall(t, state, &mem, 56, apiFailure, balanced, 0)
	if _, err := state.invoke(context.Background(), 56, []uint64{0, 4}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("invalid entropy pointer: %v", err)
	}
	var charged uint64
	_, err := state.invoke(context.Background(), 56, []uint64{balanced, 4}, &mem, func(n uint64) error { charged += n; return nil })
	if err != nil || charged < 2 {
		t.Fatalf("entropy work: %v, charged %d", err, charged)
	}
}
