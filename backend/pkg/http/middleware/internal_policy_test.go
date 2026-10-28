package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func runWithPolicy(t *testing.T, p InternalPolicy, path, remote string) int {
	t.Helper()
	engine := gin.New()
	engine.Use(Authenticate(nil, nil, testInternalKey, nil, WithInternalPolicy(p)))
	engine.Any("/*any", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(InternalKeyHeader, testInternalKey)
	req.RemoteAddr = remote + ":1234"
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code
}

func TestTheInternalKeyOnlyWorksOnListedRoutes(t *testing.T) {
	p := InternalPolicy{Paths: []string{"/api/v1/mcp", "/api/v1/internal"}}

	cases := map[string]int{
		"/api/v1/mcp":                      http.StatusOK,
		"/api/v1/mcp/tools":                http.StatusOK,
		"/api/v1/internal/alert-tag-rules": http.StatusOK,
		"/api/v1/mcpx":                     http.StatusUnauthorized,
		"/api/v1/users":                    http.StatusUnauthorized,
		"/api/v1/storage/tiering":          http.StatusUnauthorized,
	}
	for path, want := range cases {
		if got := runWithPolicy(t, p, path, "10.0.0.5"); got != want {
			t.Errorf("%s = %d, want %d", path, got, want)
		}
	}
}

func TestTheInternalKeyCanBeLimitedToANetwork(t *testing.T) {
	nets, err := ParseCIDRs("10.0.0.0/8, 172.16.0.0/12")
	if err != nil {
		t.Fatal(err)
	}
	p := InternalPolicy{Paths: []string{"/api/v1/mcp"}, CIDRs: nets}

	if got := runWithPolicy(t, p, "/api/v1/mcp", "10.1.2.3"); got != http.StatusOK {
		t.Errorf("inside = %d", got)
	}
	if got := runWithPolicy(t, p, "/api/v1/mcp", "203.0.113.9"); got != http.StatusUnauthorized {
		t.Errorf("outside = %d", got)
	}
}

func TestABadCIDRIsAnError(t *testing.T) {
	if _, err := ParseCIDRs("10.0.0.0/8,nope"); err == nil {
		t.Error("accepted")
	}
}
