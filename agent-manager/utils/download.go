package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Debug ********************************************************************************************************************
const (
	reconnectDelay  = 10 * time.Second
	attemptsAllowed = 3
)

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

// Debug ********************************************************************************************************************

func DownloadFileByChucks(url string, fileName string) error {
	const chunkSize = 5 * 1024 * 1024
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	client := &http.Client{}
	offset := int64(0)

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("error creating new request: %v", err)
		}
		byteRange := fmt.Sprintf("bytes=%d-%d", offset, offset+chunkSize-1)
		req.Header.Add("Range", byteRange)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			break
		}

		if resp.StatusCode != http.StatusPartialContent {
			return fmt.Errorf("expected status %d; got %d", http.StatusPartialContent, resp.StatusCode)
		}

		written, err := io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("io.Copy: %v", err)
		}

		offset += written
	}

	ALogger.Info("Downloaded correctly: %s", fileName)
	return nil
}
