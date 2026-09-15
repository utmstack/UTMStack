package agent

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorEventText(t *testing.T) {
	if got := errorEventText(ErrLLMRateLimited); got != ErrLLMRateLimited.Error() {
		t.Errorf("rate limit error: got %q, want the specific message %q", got, ErrLLMRateLimited.Error())
	}

	// Wrapped, still detected.
	wrapped := fmt.Errorf("completing: %w", ErrLLMRateLimited)
	if got := errorEventText(wrapped); got != ErrLLMRateLimited.Error() {
		t.Errorf("wrapped rate limit error: got %q, want %q", got, ErrLLMRateLimited.Error())
	}

	// Anything else collapses to the generic message — it may carry URLs,
	// auth headers or other internals that must not reach the client.
	other := errors.New("dial tcp 10.0.0.5:443: connection refused")
	if got := errorEventText(other); got != genericErrorMsg {
		t.Errorf("opaque error: got %q, want the generic message %q", got, genericErrorMsg)
	}
}
