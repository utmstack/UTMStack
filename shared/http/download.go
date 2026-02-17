// Package http provides HTTP utility functions.
package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadOptions configures a download request.
type DownloadOptions struct {
	// Headers to include in the request
	Headers map[string]string
	// SkipTLSVerify skips TLS certificate validation
	SkipTLSVerify bool
	// Timeout for the download (default: 5 minutes)
	Timeout time.Duration
}

// DefaultOptions returns default download options.
func DefaultOptions() DownloadOptions {
	return DownloadOptions{
		Headers:       make(map[string]string),
		SkipTLSVerify: false,
		Timeout:       5 * time.Minute,
	}
}

// Download downloads a file from the given URL to the specified destination.
func Download(url, destDir, filename string, opts DownloadOptions) error {
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.SkipTLSVerify,
		},
	}

	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	destPath := filepath.Join(destDir, filename)
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// DownloadFile is a convenience function that downloads a file with common options.
func DownloadFile(url string, headers map[string]string, filename, destDir string, skipTLSVerify bool) error {
	opts := DownloadOptions{
		Headers:       headers,
		SkipTLSVerify: skipTLSVerify,
		Timeout:       5 * time.Minute,
	}
	return Download(url, destDir, filename, opts)
}
