package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url string, headers map[string]string, fileName string, path string, skipTls bool) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating new request: %v", err)
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	if !skipTls {
		tlsConfig, err := LoadTLSCredentials(filepath.Join(GetMyPath(), "certs", "utm.crt"))
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
	}

	out, err := os.Create(filepath.Join(path, fileName))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	return nil
}
