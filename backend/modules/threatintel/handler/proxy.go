package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/threatintel/internal"
)

// ReverseProxyHandler creates a reverse proxy that maps a local route to a CM proxy target.
type ReverseProxyHandler struct {
	targetPath string
}

// NewReverseProxyHandler creates a handler for a specific route mapping.
func NewReverseProxyHandler(targetPath string) *ReverseProxyHandler {
	return &ReverseProxyHandler{targetPath: targetPath}
}

// Handle is the Gin handler that proxies the request.
func (h *ReverseProxyHandler) Handle(c *gin.Context) {
	cfg := internal.Get()
	if cfg == nil || cfg.Server == "" || cfg.InstanceID == "" || cfg.InstanceKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "instance not configured"})
		return
	}

	targetURL, err := url.Parse(cfg.Server)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "invalid CM server URL"})
		return
	}

	// Create reverse proxy with custom director
	proxy := &httputil.ReverseProxy{
		Director: h.directorFunc(cfg, targetURL),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"upstream service error"}`))
		},
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// directorFunc returns a Director function that rewrites the request.
func (h *ReverseProxyHandler) directorFunc(cfg *internal.InstanceConfig, targetURL *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		// Set scheme and host from CM proxy
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host

		// Rewrite path to the target on CM proxy
		req.URL.Path = h.targetPath
		// Preserve query string (already in req.URL.RawQuery)

		req.RequestURI = "" // must be cleared for the reverse proxy
		req.Host = targetURL.Host

		// Clear auth headers
		req.Header.Del("Authorization")
		req.Header.Del("id")
		req.Header.Del("key")

		// Set instance credentials
		req.Header.Set("id", cfg.InstanceID)
		req.Header.Set("key", cfg.InstanceKey)
	}
}

// HandleUsageEndpoint is a special handler for the /usage endpoint which has a different prefix structure.
func (h *ReverseProxyHandler) HandleUsageEndpoint(c *gin.Context) {
	cfg := internal.Get()
	if cfg == nil || cfg.Server == "" || cfg.InstanceID == "" || cfg.InstanceKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "instance not configured"})
		return
	}

	// Build the full target URL for usage endpoint (no /proxy/api prefix, just /proxy/usage)
	fullURL := fmt.Sprintf("%s%s", cfg.Server, h.targetPath)
	targetURL, err := url.Parse(fullURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "invalid target URL"})
		return
	}

	// Create and execute the request
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to create request"})
		return
	}

	allowed := map[string]struct{}{
		"Accept":                    {},
		"Content-Type":              {},
		"User-Agent":                {},
		"X-Request-Id":              {},
		"X-Correlation-Id":          {},
	}

	// Copy headers
	for k, values := range c.Request.Header {
		canonical := textproto.CanonicalMIMEHeaderKey(k)
		if _, ok := allowed[canonical]; ok {
			for _, v := range values {
				req.Header.Add(canonical, v)
			}
		}
	}

	// Set instance credentials
	req.Header.Set("id", cfg.InstanceID)
	req.Header.Set("key", cfg.InstanceKey)

	// Execute request
	client := &http.Client{
		Timeout: 20* time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service error"})
		return
	}
	defer resp.Body.Close()

	// Copy response
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Header(k, vv)
		}
	}
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
