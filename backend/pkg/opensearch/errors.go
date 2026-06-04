package opensearch

import "fmt"

// ErrOSRequest is returned when the OpenSearch HTTP request itself fails.
var ErrOSRequest = fmt.Errorf("opensearch request error")

// ErrOSDecode is returned when the OpenSearch response body cannot be decoded.
var ErrOSDecode = fmt.Errorf("opensearch decode error")

// WrapOS wraps an error with an operation name for context.
func WrapOS(err error, op string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
