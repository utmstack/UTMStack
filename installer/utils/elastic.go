package utils

import (
	"fmt"
	"net/http"
	"time"
)

// CheckIndexExists checks if the given Elasticsearch index exists by sending an HTTP GET
// request to the provided URL. It returns true if the index exists, false otherwise.
func CheckIndexExists(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// The index exists.
		return true, nil
	case http.StatusNotFound:
		// The index does not exist.
		return false, nil
	default:
		// For any other status code, return an error.
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// retrying every 5 seconds up to a total duration of 1 minute.
func CheckIndexExistsWithRetry(url string) (bool, error) {
	timeout := 5 * time.Minute
	interval := 5 * time.Second
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		exists, err := CheckIndexExists(url)
		if err == nil {
			return exists, nil
		}
		lastErr = err
		time.Sleep(interval)
	}
	return false, fmt.Errorf("OpenSearch status check failed after retries: %v", lastErr)
}
