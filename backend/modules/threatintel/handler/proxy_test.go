package handler

import (
	"errors"
	"net"
	"net/http"
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

// The bound has to be on the response headers, not on the whole exchange: the
// AI chat route streams its completion and a total timeout would cut off a
// reply that is arriving perfectly well.
func TestProxyTransportBoundsASilentUpstream(t *testing.T) {
	if proxyTransport.ResponseHeaderTimeout == 0 {
		t.Fatal("no ResponseHeaderTimeout: an upstream that goes silent holds the goroutine indefinitely")
	}
	if proxyTransport.TLSHandshakeTimeout == 0 {
		t.Error("no TLSHandshakeTimeout")
	}
}

func TestUpstreamErrorDistinguishesTimeout(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"timeout", &net.DNSError{IsTimeout: true}, http.StatusGatewayTimeout},
		{"refused", errors.New("connection refused"), http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeUpstreamError(rec, tc.err)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if rec.Body.Len() == 0 {
				t.Error("empty body; the caller gets no reason")
			}
		})
	}
}
