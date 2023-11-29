package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	maxConnectionAttempts = 3
	initialReconnectDelay = 10 * time.Second
	maxReconnectDelay     = 60 * time.Second
)

// DownloadFile downloads a file from a URL and saves it to disk. Returns an error on failure.
func DownloadFile(url string, fileName string) error {
	connectionAttemps := 0
	reconnectDelay := initialReconnectDelay

	var resp *http.Response
	var err error

	for {
		if connectionAttemps >= maxConnectionAttempts {
			return fmt.Errorf("error downloading file after %d attemps: %v", maxConnectionAttempts, err)
		}
		resp, err = http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			connectionAttemps++
			time.Sleep(reconnectDelay)
			reconnectDelay = IncrementReconnectDelay(reconnectDelay, maxReconnectDelay)
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
