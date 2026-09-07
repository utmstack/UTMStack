package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var requestClient = &http.Client{Timeout: 60 * time.Second}

// Exists because the go-sdk utils.DoReq helper cannot surface response headers.
func doReqWithHeaders[T any](link string, body []byte, method string, headers map[string]string) (T, http.Header, int, error) {
	var result T

	var payload io.Reader
	if len(body) > 0 {
		payload = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, link, payload)
	if err != nil {
		return result, nil, 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := requestClient.Do(req)
	if err != nil {
		return result, nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, resp.Header, resp.StatusCode, err
	}

	return result, resp.Header, resp.StatusCode, nil
}
