package utils

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

// Retry configuration
const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
	maxRetries     = 10
	backoffFactor  = 2.0
	jitterFactor   = 0.1 // 10% jitter
)

// ConnectionChecker verifies connectivity to a URL with exponential backoff
func ConnectionChecker(url string) error {
	checkConn := func() error {
		if err := checkConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	return retryWithBackoff(checkConn, "connection failed")
}

// checkConnection performs a single connection check using the pooled client
func checkConnection(url string) error {
	initClients() // Ensure pooled clients are initialized

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(f func() error, retryOnError string) error {
	var lastErr error
	backoff := initialBackoff
	errorLogged := false

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := f()
		if err == nil {
			return nil
		}

		// Only retry if error matches the expected pattern
		if !containsError(err, retryOnError) {
			return err
		}

		lastErr = err

		// Log only on first retry attempt
		if !errorLogged {
			_ = catcher.Error("Connection error, retrying with backoff...", err, map[string]any{
				"process":     "plugin_com.utmstack.soc-ai",
				"max_retries": maxRetries,
			})
			errorLogged = true
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			// Add jitter to prevent thundering herd
			jitter := time.Duration(float64(backoff) * jitterFactor * (rand.Float64()*2 - 1))
			sleepTime := backoff + jitter

			time.Sleep(sleepTime)

			// Exponential backoff with cap
			backoff = min(time.Duration(float64(backoff)*backoffFactor), maxBackoff)
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %v", maxRetries, lastErr)
}

// containsError checks if an error message contains any of the specified strings
func containsError(e error, args ...string) bool {
	if e == nil {
		return false
	}
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
