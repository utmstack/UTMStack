package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func DoReq[response any](url string, data []byte, method string, headers map[string]string) (response, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return *new(response), http.StatusInternalServerError, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	//transCfg := &http.Transport{
	//	TLSClientConfig: &tls.Config{},
	//}
	client := &http.Client{}

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
		return result, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}
