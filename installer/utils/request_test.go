package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoReqReportsStatusBeforeDecoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"usage report failed"}`))
	}))
	defer srv.Close()

	_, status, err := DoReq[[]string](srv.URL, nil, http.MethodGet, nil, nil)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d", status)
	}
	if err == nil || !strings.Contains(err.Error(), "unexpected status code: 500") || strings.Contains(err.Error(), "unmarshalling") {
		t.Fatalf("error should report the status, got %v", err)
	}
}

func TestDoReqDecodesOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["a","b"]`))
	}))
	defer srv.Close()

	got, status, err := DoReq[[]string](srv.URL, nil, http.MethodGet, nil, nil)
	if err != nil || status != 200 || len(got) != 2 {
		t.Fatalf("got %v %d %v", got, status, err)
	}
}
