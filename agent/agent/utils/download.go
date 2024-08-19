package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/utmstack/UTMStack/agent/agent/models"
)

func DownloadFileByChunks(url string, headers map[string]string, fileName string, skipTls bool) (string, bool, error) {
	var version string
	client := &http.Client{}
	if skipTls {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	const chunkSize = 5
	partindex := 1

	var out *os.File
	var fileCreated bool = false

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s&partIndex=%d&partSize=%d", url, partindex, chunkSize), nil)
		if err != nil {
			return version, false, fmt.Errorf("error creating new request: %v", err)
		}
		for key, value := range headers {
			req.Header.Add(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return version, false, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
			return version, false, fmt.Errorf("error response: %v", resp.Status)
		}

		var response models.DependencyUpdateResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return version, false, fmt.Errorf("error decoding response: %v", err)
		}
		resp.Body.Close()

		needUpdate, err := processUpdateResponse(response.Message)
		if err != nil {
			return version, false, err
		}

		if !needUpdate {
			return response.Version, false, nil
		}

		if response.FileContent == nil {
			return version, false, fmt.Errorf("no file content in response")
		}

		if !fileCreated {
			out, err = os.Create(fileName)
			if err != nil {
				return version, false, fmt.Errorf("error creating file: %v", err)
			}
			defer out.Close()
			fileCreated = true
		}

		if _, err := out.Write(response.FileContent); err != nil {
			return version, false, fmt.Errorf("error writing to file: %v", err)
		}

		if partindex == response.NParts {
			return response.Version, true, nil
		}
		partindex++
	}
}

func processUpdateResponse(updateMessage string) (bool, error) {
	switch {
	case updateMessage == "dependency not found", strings.Contains(updateMessage, "error getting dependency file"):
		return false, fmt.Errorf("error getting dependency file: %v", updateMessage)
	case updateMessage == "dependency already up to date":
		return false, nil
	case updateMessage == "dependency update available":
		return true, nil
	default:
		return false, fmt.Errorf("error processing update response: %s", updateMessage)
	}
}
