package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DoReq[response any](url string, data []byte, method string, headers map[string]string) (response, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return *new(response), http.StatusInternalServerError, fmt.Errorf("error creating request: %w", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return *new(response), http.StatusInternalServerError, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return *new(response), resp.StatusCode, fmt.Errorf("error reading response body: %w; response body: %s", err, string(body))
	}

	var result response

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(response), resp.StatusCode, fmt.Errorf("error unmarshalling response: %w; response body: %s", err, string(body))
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, fmt.Errorf("unexpected status code: %d; response body: %s", resp.StatusCode, string(body))
	}

	return result, resp.StatusCode, nil
}
