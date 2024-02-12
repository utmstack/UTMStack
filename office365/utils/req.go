package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/threatwinds/logger"
)

func DoReq[response any](url string, data []byte, method string, headers map[string]string) (response, int, *logger.Error) {
	var result response

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return result, http.StatusInternalServerError, Logger.ErrorF(http.StatusInternalServerError, err.Error())
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, http.StatusInternalServerError, Logger.ErrorF(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, http.StatusInternalServerError, Logger.ErrorF(http.StatusInternalServerError, err.Error())
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, Logger.ErrorF(resp.StatusCode, "while sending request to %s received status code: %d and response body: %s", url, resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, http.StatusInternalServerError, Logger.ErrorF(http.StatusInternalServerError, err.Error())
	}

	return result, resp.StatusCode, nil
}
