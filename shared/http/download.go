// Package http provides HTTP utility functions.
package http

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func DownloadAndVerify(url, destDir, filename string, opts DownloadOptions) (verified bool, err error) {
	if err := Download(url, destDir, filename, opts); err != nil {
		return false, err
	}

	checksumFilename := filename + ".sha256.tmp"
	if err := Download(url+".sha256", destDir, checksumFilename, opts); err != nil {
		return false, nil
	}
	checksumPath := filepath.Join(destDir, checksumFilename)
	defer os.Remove(checksumPath)

	raw, err := os.ReadFile(checksumPath)
	if err != nil {
		return false, nil
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return false, nil
	}
	expected := strings.ToLower(fields[0])
	if len(expected) != sha256.Size*2 {
		return false, nil
	}

	f, err := os.Open(filepath.Join(destDir, filename))
	if err != nil {
		return false, fmt.Errorf("error opening downloaded file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, fmt.Errorf("error hashing downloaded file: %w", err)
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return false, fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filename, expected, actual)
	}

	return true, nil
}

func DownloadFileAndVerify(url string, headers map[string]string, filename, destDir string, skipTLSVerify bool) (verified bool, err error) {
	opts := DownloadOptions{
		Headers:       headers,
		SkipTLSVerify: skipTLSVerify,
		Timeout:       5 * time.Minute,
	}
	return DownloadAndVerify(url, destDir, filename, opts)
}
