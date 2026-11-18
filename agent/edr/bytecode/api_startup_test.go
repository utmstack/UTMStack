package bytecode

import (
	"context"
	"errors"
	"testing"
)

func TestStartupDisableHostAPIs(t *testing.T) {
	var mem memory
	reason := apiTestBytes(t, &mem, []byte("inert test reason\x00"))
	s := &apiState{kind: 1}
	apiTestCall(t, s, &mem, 74, 0, reason, 17, 0)
	apiTestCall(t, s, &mem, 74, 1, reason, 17, 1)
	apiTestCall(t, s, &mem, 73, 1, reason, 17, 0)
	apiTestCall(t, s, &mem, 73, 2, reason, 17, 7)
	apiTestCall(t, s, &mem, 74, 2, reason, 17, 1)
	if s.disableStatus != 2 {
		t.Fatalf("disable status = %d", s.disableStatus)
	}
	apiTestCall(t, &apiState{kind: 256}, &mem, 73, apiFailure, reason, 17, 1)
	apiTestCall(t, &apiState{kind: 256}, &mem, 74, apiFailure, reason, 17, 1)
	if _, err := s.invoke(context.Background(), 73, []uint64{0, 17, 1}, &mem,
		func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("invalid reason pointer: %v", err)
	}
}
