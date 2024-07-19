package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DoReq[response any](url string, data []byte, method string, headers map[string]string, skipTlsVerification bool) (response, int, error) {
	var result response

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return result, http.StatusInternalServerError, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}
	if skipTlsVerification {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, fmt.Errorf("while sending request to %s received status code: %d and response body: %s", url, resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}

	return result, resp.StatusCode, nil
}
