// edr/internal/feeds/client_test.go
package feeds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAccumulativeSendsAuthHeadersAndPath(t *testing.T) {
	var gotPath, gotKey, gotSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("api-key")
		gotSecret = r.Header.Get("api-secret")
		w.Write([]byte("payload"))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, APIKey: "k", APISecret: "s", HTTP: srv.Client()}
	b, err := c.FetchAccumulative(context.Background(), "level1", "ip")
	if err != nil || string(b) != "payload" {
		t.Fatalf("body = %q, err %v", b, err)
	}
	if gotPath != "/api/feeds/v1/download/list/level1/accumulative/ip" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotKey != "k" || gotSecret != "s" {
		t.Fatalf("auth headers = %q/%q", gotKey, gotSecret)
	}
}

func TestFetchAccumulativeErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	if _, err := c.FetchAccumulative(context.Background(), "level1", "ip"); err == nil {
		t.Fatal("expected error on 503")
	}
}

func TestFetchCredentialsReadsBackendConfigKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Key") != "ik" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v1/config/utmstack.tw.apiKey":
			w.Write([]byte(`{"value":"the-key"}`))
		case "/api/v1/config/utmstack.tw.apiSecret":
			w.Write([]byte(`{"value":"the-secret"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	k, s, err := FetchCredentials(context.Background(), srv.URL, "ik", srv.Client())
	if err != nil || k != "the-key" || s != "the-secret" {
		t.Fatalf("k=%q s=%q err=%v", k, s, err)
	}
}
