package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
)

// DoReq makes a request to the specified URL with the specified data, method and headers.
func DoReq[response any](url string, data []byte, method string, headers map[string]string, config *tls.Config) (response, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return *new(response), http.StatusInternalServerError, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	transp := &http.Transport{
		TLSClientConfig: config,
	}

	client := &http.Client{Transport: transp}

	resp, err := client.Do(req)
	if err != nil {
		return *new(response), http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return *new(response), http.StatusInternalServerError, err
	}

	var result response

	err = json.Unmarshal(body, &result)
	if err != nil {
		return *new(response), http.StatusInternalServerError, err
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return *new(response), http.StatusInternalServerError, err
	}

	return result, resp.StatusCode, nil
}
