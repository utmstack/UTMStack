package branding

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/installer/utils"
)

func InstallLicense(updatesFolder string) (bool, error) {
	src := LicensePath()
	if !utils.CheckIfPathExist(src) {
		return false, nil
	}
	dst := filepath.Join(updatesFolder, "LICENSE")
	if utils.CheckIfPathExist(dst) {
		return false, nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return false, err
	}
	return true, nil
}

var assetFiles = map[string]string{
	"logo":        "logo",
	"logoDark":    "logo-dark",
	"favicon":     "favicon",
	"reportLogo":  "report-logo",
	"reportCover": "report-cover",
}

var imageExts = []string{".svg", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".ico"}

func findAsset(base string) string {
	for _, ext := range imageExts {
		p := filepath.Join(Dir(), base+ext)
		if utils.CheckIfPathExist(p) {
			return p
		}
	}
	return ""
}

func Seed(baseURL, internalKey string) error {
	if !IsCustom() {
		return nil
	}
	b := Get()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fields := map[string]string{
		"enabled":     "true",
		"productName": b.ProductName,
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}

	for slot, base := range assetFiles {
		p := findAsset(base)
		if p == "" {
			continue
		}
		if err := addFile(w, slot, p); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v1/branding/seed", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-Internal-Key", internalKey)

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// mimeForExt maps an image extension to its MIME type (for data URIs).
func mimeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

// LogoDataURI returns the shipped logo encoded as a base64 data URI, so it can
// be embedded directly in the nginx maintenance page (which is served by the
// host nginx while the backend — and its /uploads — is down). Returns false
// when no logo was shipped.
func LogoDataURI() (string, bool) {
	p := findAsset("logo")
	if p == "" {
		return "", false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", false
	}
	return "data:" + mimeForExt(filepath.Ext(p)) + ";base64," + base64.StdEncoding.EncodeToString(data), true
}

func addFile(w *multipart.Writer, field, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	fw, err := w.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, f)
	return err
}
