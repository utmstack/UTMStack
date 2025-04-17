package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// DoReq sends an HTTP request and processes the response.
//
// This function sends an HTTP request to the specified URL with the given
// method, data, and headers. It returns the response body unmarshalled into
// the specified response type, the HTTP status code, and an error if any
// occurred during the process.
//
// Type Parameters:
//   - response: The type into which the response body will be unmarshalled.
//
// Parameters:
//   - url: The URL to which the request is sent.
//   - data: The request payload as a byte slice.
//   - method: The HTTP method to use for the request (e.g., "GET", "POST").
//   - headers: A map of headers to include in the request.
//
// Returns:
//   - response: The response body unmarshalled into the specified type.
//   - int: The HTTP status code of the response.
//   - error: An error if any occurred during the request or response
//     processing, otherwise nil.
func DoReq[response any](url string, data []byte, method string, headers map[string]string, transCfg *http.Transport) (response, int, error) {
	var result response
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return result, http.StatusInternalServerError, fmt.Errorf("error creating request: %w", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}
	if transCfg != nil {
		client.Transport = transCfg
	}

	resp, err := client.Do(req)
	if err != nil {
		return result, http.StatusInternalServerError, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, resp.StatusCode, fmt.Errorf("error reading response body: %w; response body: %s", err, string(body))
	}

	if len(body) != 0 {
		if err = json.Unmarshal(body, &result); err != nil {
			return result, resp.StatusCode, fmt.Errorf("error unmarshalling response: %w; response body: %s", err, string(body))
		}
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, fmt.Errorf("unexpected status code: %d; response body: %s", resp.StatusCode, string(body))
	}

	return result, resp.StatusCode, nil
}

// Download downloads the content from the specified URL and saves it to the specified file.
// It returns an error if any error occurs during the process.
//
// Parameters:
//   - url: The URL from which to download the content.
//   - file: The path to the file where the content should be saved.
//
// Returns:
//   - error: An error object if an error occurs, otherwise nil.
func Download(url, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
