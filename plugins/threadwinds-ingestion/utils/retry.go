package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

type RetryConfig struct {
	MaxRetries        int           // Maximum number of retry attempts (-1 for unlimited)
	InitialBackoff    time.Duration // Initial wait time before the first retry attempt
	MaxBackoff        time.Duration // Maximum wait time between retry attempts (upper limit for exponential backoff)
	BackoffMultiplier float64       // Growth factor for exponential backoff (0 = fixed wait, >1 = exponential growth)
	LogInterval       int           // Log every N attempts (0 = log only once, >0 = log periodically)
	ErrorFilter       []string      // List of error message substrings to match for retry (nil = retry all errors)
	StopOnMismatch    bool          // If true, return error when it doesn't match ErrorFilter; if false, continue without retrying
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        -1,
		InitialBackoff:    5 * time.Second,
		MaxBackoff:        2 * time.Minute,
		BackoffMultiplier: 2.0,
		LogInterval:       10,
		ErrorFilter:       nil,
		StopOnMismatch:    false,
	}
}

func ConnectionRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        -1,
		InitialBackoff:    3 * time.Second,
		MaxBackoff:        3 * time.Second,
		BackoffMultiplier: 0,
		LogInterval:       0,
		ErrorFilter:       []string{"connection failed"},
		StopOnMismatch:    true,
	}
}

func Retry(f func() error, operationName string, config RetryConfig) error {
	attempt := 0
	currentBackoff := config.InitialBackoff
	errorLogged := false

	retryType := "infinite retry"
	if config.MaxRetries >= 0 {
		retryType = fmt.Sprintf("max %d retries", config.MaxRetries)
	}

	if config.LogInterval > 0 {
		catcher.Info(fmt.Sprintf("Starting %s with %s", operationName, retryType), map[string]any{
			"initial_backoff": config.InitialBackoff.String(),
			"max_backoff":     config.MaxBackoff.String(),
		})
	}

	for {
		attempt++
		err := f()

		if err == nil {
			return nil
		}

		if len(config.ErrorFilter) > 0 && !matchesErrorFilter(err, config.ErrorFilter) {
			if config.StopOnMismatch {
				return err
			}
			continue
		}

		if config.MaxRetries >= 0 && attempt > config.MaxRetries {
			_ = catcher.Error(fmt.Sprintf("%s failed after %d attempts", operationName, attempt-1), err, map[string]any{
				"max_retries": config.MaxRetries,
			})
			return err
		}

		shouldLog := config.LogInterval == 0 && !errorLogged ||
			config.LogInterval > 0 && (attempt == 1 || attempt%config.LogInterval == 0)

		if shouldLog {
			logMsg := fmt.Sprintf("%s failed, will retry...", operationName)
			if config.MaxRetries < 0 {
				logMsg = fmt.Sprintf("%s failed, will retry indefinitely...", operationName)
			}
			_ = catcher.Error(logMsg, err, map[string]any{
				"attempt":       attempt,
				"next_retry_in": currentBackoff.String(),
			})
			errorLogged = true
		}

		time.Sleep(currentBackoff)

		if config.BackoffMultiplier > 0 {
			nextBackoff := time.Duration(float64(currentBackoff) * config.BackoffMultiplier)
			currentBackoff = min(nextBackoff, config.MaxBackoff)
		}
	}
}

func matchesErrorFilter(err error, filters []string) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	for _, filter := range filters {
		if strings.Contains(errMsg, filter) {
			return true
		}
	}
	return false
}

func ConnectionChecker(url string) error {
	checkConn := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := checkConnection(url, ctx); err != nil {
			return fmt.Errorf("connection failed")
		}
		return nil
	}

	return Retry(checkConn, "connection check", ConnectionRetryConfig())
}

func checkConnection(url string, ctx context.Context) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = catcher.Error("error closing response body", closeErr, nil)
		}
	}()

	return nil
}
