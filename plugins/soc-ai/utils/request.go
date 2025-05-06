package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func DoParseReq[response any](url string, data []byte, method string, headers map[string]string, timeoutInSec int) (response, int, error) {
	body, status, err := DoReq(url, data, method, headers, timeoutInSec)
	if err != nil {
		return *new(response), status, fmt.Errorf("error reading response body: %v", err)
	}

	var result response

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(response), http.StatusInternalServerError, fmt.Errorf("error decoding response: %v", err)
	}

	if status != http.StatusAccepted && status != http.StatusOK {
		return result, status, fmt.Errorf("status code '%d' received '%s' while sending '%s'", status, body, data)
	}

	return result, status, nil
}

func DoReq(url string, data []byte, method string, headers map[string]string, timeoutInSec int) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutInSec) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, http.StatusInternalServerError, fmt.Errorf("request timed out: %v: %s", err, data)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("error performing request: %v: %s", err, data)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error reading response body: %v", err)
	}

	return body, resp.StatusCode, nil
}
