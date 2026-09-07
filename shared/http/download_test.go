package http

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadAndVerify_MatchingChecksum(t *testing.T) {
	content := []byte("binary contents")
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agent":
			w.Write(content)
		case "/agent.sha256":
			w.Write([]byte(hexSum + "  agent\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	destDir := t.TempDir()
	verified, err := DownloadAndVerify(server.URL+"/agent", destDir, "agent", DownloadOptions{})
	if err != nil {
		t.Fatalf("DownloadAndVerify() error = %v, want nil", err)
	}
	if !verified {
		t.Fatalf("verified = false, want true")
	}

	got, err := os.ReadFile(filepath.Join(destDir, "agent"))
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("downloaded content = %q, want %q", got, content)
	}
	if _, err := os.Stat(filepath.Join(destDir, "agent.sha256.tmp")); !os.IsNotExist(err) {
		t.Fatalf("checksum sidecar temp file was not cleaned up")
	}
}

func TestDownloadAndVerify_NoChecksumPublished(t *testing.T) {
	content := []byte("binary contents")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/agent" {
			w.Write(content)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	destDir := t.TempDir()
	verified, err := DownloadAndVerify(server.URL+"/agent", destDir, "agent", DownloadOptions{})
	if err != nil {
		t.Fatalf("DownloadAndVerify() error = %v, want nil (missing checksum must not fail the download)", err)
	}
	if verified {
		t.Fatalf("verified = true, want false when no checksum was published")
	}
	if _, err := os.Stat(filepath.Join(destDir, "agent")); err != nil {
		t.Fatalf("downloaded file missing even though the download itself succeeded: %v", err)
	}
}

func TestDownloadAndVerify_MismatchedChecksum(t *testing.T) {
	content := []byte("binary contents")
	wrongSum := sha256.Sum256([]byte("something else entirely"))
	hexSum := hex.EncodeToString(wrongSum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agent":
			w.Write(content)
		case "/agent.sha256":
			w.Write([]byte(hexSum))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	destDir := t.TempDir()
	verified, err := DownloadAndVerify(server.URL+"/agent", destDir, "agent", DownloadOptions{})
	if err == nil {
		t.Fatalf("DownloadAndVerify() error = nil, want a checksum mismatch error")
	}
	if verified {
		t.Fatalf("verified = true, want false on mismatch")
	}
}

func TestDownloadAndVerify_MalformedChecksum(t *testing.T) {
	content := []byte("binary contents")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agent":
			w.Write(content)
		case "/agent.sha256":
			w.Write([]byte("not-a-valid-hex-digest"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	destDir := t.TempDir()
	verified, err := DownloadAndVerify(server.URL+"/agent", destDir, "agent", DownloadOptions{})
	if err != nil {
		t.Fatalf("DownloadAndVerify() error = %v, want nil (malformed checksum must degrade gracefully)", err)
	}
	if verified {
		t.Fatalf("verified = true, want false for a malformed checksum")
	}
}
