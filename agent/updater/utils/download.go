package utils

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/quantfall/holmes"
)

const (
	reconnectDelay = 5 * time.Minute
)

// DownloadFile downloads a file from a URL and saves it to disk. Returns an error on failure.
func DownloadFile(url string, fileName string, utmLogger *holmes.Logger) error {
	var resp *http.Response
	var err error

	for {
		resp, err = http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			utmLogger.Error("error downloading file from %s: %v", url, err)
			time.Sleep(reconnectDelay)
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
