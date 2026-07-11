package handler

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
)

func TestDirectorRewritesPathAndHeaders(t *testing.T) {
	// Setup instance config
	cfg := &instanceconfig.InstanceConfig{
		Server:      "https://example.com",
		InstanceID:  "test-id-123",
		InstanceKey: "test-key-456",
	}

	// Create a test request
	req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/threat-intel/entity/foo?param=value", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("id", "old-id")
	req.Header.Set("key", "old-key")

	// Create a reverse proxy handler
	h := NewReverseProxyHandler("/proxy/api/analytics/v1/entity/foo/details")

	// Parse the target URL
	targetURL, err := url.Parse("https://cm.example.com")
	if err != nil {
		t.Fatal(err)
	}

	// Get the director function
	director := h.directorFunc(cfg, targetURL)

	// Apply director
	director(req)

	// Verify URL rewrite
	if req.URL.Scheme != "https" {
		t.Fatalf("expected scheme https, got %s", req.URL.Scheme)
	}
	if req.URL.Host != "cm.example.com" {
		t.Fatalf("expected host cm.example.com, got %s", req.URL.Host)
	}
	if req.URL.Path != "/proxy/api/analytics/v1/entity/foo/details" {
		t.Fatalf("expected path /proxy/api/analytics/v1/entity/foo/details, got %s", req.URL.Path)
	}
	if !strings.Contains(req.URL.RawQuery, "param=value") {
		t.Fatalf("expected query param preserved, got %s", req.URL.RawQuery)
	}

	// Verify headers
	if req.Header.Get("Authorization") != "" {
		t.Fatal("Authorization header should be deleted")
	}
	if req.Header.Get("id") != "test-id-123" {
		t.Fatalf("expected id header test-id-123, got %s", req.Header.Get("id"))
	}
	if req.Header.Get("key") != "test-key-456" {
		t.Fatalf("expected key header test-key-456, got %s", req.Header.Get("key"))
	}
}
