package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/threatwinds/logger"
)

const (
	reconnectDelay  = 10 * time.Second
	attemptsAllowed = 3
)

// DownloadFile downloads a file from a URL and saves it to disk. Returns an error on failure.
func DownloadFile(url string, fileName string, utmLogger *logger.Logger) error {
	var resp *http.Response
	var err error
	var attempts = 0

	for {
		resp, err = http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			if attempts >= attemptsAllowed {
				return fmt.Errorf("error downloading file from %s", url)
			}
			utmLogger.ErrorF("error downloading file from %s", url)
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
