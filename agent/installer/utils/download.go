package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/utmstack/UTMStack/agent/installer/models"
)

func DownloadFileByChunks(url string, headers map[string]string, fileName string, skipTls bool) (string, error) {
	var version string
	out, err := os.Create(fileName)
	if err != nil {
		return version, fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	client := &http.Client{}
	if skipTls {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	const chunkSize = 5
	partindex := 1

	var bar *progressbar.ProgressBar

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s&partIndex=%d&partSize=%d", url, partindex, chunkSize), nil)
		if err != nil {
			return version, fmt.Errorf("error creating new request: %v", err)
		}
		for key, value := range headers {
			req.Header.Add(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return version, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			return version, fmt.Errorf("expected status %d; got %d", http.StatusPartialContent, resp.StatusCode)
		}

		var response models.DependencyUpdateResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return version, fmt.Errorf("error decoding response: %v", err)
		}
		resp.Body.Close()

		if response.FileContent == nil {
			return version, fmt.Errorf("no file content in response")
		}

		if _, err := out.Write(response.FileContent); err != nil {
			return version, fmt.Errorf("error writing to file: %v", err)
		}

		if bar == nil {
			bar = progressbar.NewOptions64(
				response.TotalSize,
				progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", filepath.Base(fileName))),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "=",
					SaucerHead:    ">",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(50),
				progressbar.OptionSetPredictTime(false),
				progressbar.OptionShowCount(),
				progressbar.OptionShowIts(),
			)
		}

		bar.Add(5 * 1024 * 1024)

		partindex++
		if response.IsLastPart {
			version = response.Version
			break
		}
	}
	fmt.Println("")
	return version, nil
}
