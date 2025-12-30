package utils

import (
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	retryInitialBackoff    = 5 * time.Second
	retryMaxBackoff        = 2 * time.Minute
	retryBackoffMultiplier = 2.0
	retryLogInterval       = 10
)

func InfiniteRetry(f func() error, operationName string) {
	attempt := 0
	currentBackoff := retryInitialBackoff

	catcher.Info(fmt.Sprintf("Starting %s with infinite retry and exponential backoff", operationName), map[string]any{
		"initial_backoff": retryInitialBackoff.String(),
		"max_backoff":     retryMaxBackoff.String(),
	})

	for {
		attempt++
		err := f()

		if err == nil {
			catcher.Info(fmt.Sprintf("%s completed successfully", operationName), map[string]any{
				"attempts": attempt,
			})
			return
		}

		if attempt == 1 || attempt%retryLogInterval == 0 {
			_ = catcher.Error(fmt.Sprintf("%s failed, will retry indefinitely...", operationName), err, map[string]any{
				"attempt":       attempt,
				"next_retry_in": currentBackoff.String(),
			})
		}

		time.Sleep(currentBackoff)

		currentBackoff = min(time.Duration(float64(currentBackoff)*retryBackoffMultiplier), retryMaxBackoff)
	}
}
