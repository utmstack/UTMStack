package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	reconnectDelay  = 10 * time.Second
	attemptsAllowed = 3
)

// DownloadFile downloads a file from a URL and saves it to disk. Returns an error on failure.
func DownloadFile(url string, fileName string) error {
	var resp *http.Response
	var err error
	var attempts = 0

	client := &http.Client{}

	for {
		urlUniq := fmt.Sprintf("%s?%d", url, time.Now().UnixNano())
		req, err := http.NewRequest("GET", urlUniq, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")

		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			if attempts >= attemptsAllowed {
				return fmt.Errorf("error downloading file from %s", url)
			}
			ALogger.ErrorF("error downloading file from %s", url)
			time.Sleep(reconnectDelay)
			attempts++
			continue
		}
		break
	}
	defer resp.Body.Close()

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
