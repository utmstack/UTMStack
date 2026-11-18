// edr/internal/serve/serve_test.go
package serve

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandlerServesMirrorTree(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "signatures"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "signatures", "daily.cvd"), []byte("CVD-BYTES"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "status.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(Handler(tmp))
	defer ts.Close()

	cases := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{"cvd", "/private/edr/signatures/daily.cvd", http.StatusOK, "CVD-BYTES"},
		{"status", "/private/edr/status.json", http.StatusOK, ""},
		{"missing", "/private/edr/missing", http.StatusNotFound, ""},
		{"root_dir_no_listing", "/private/edr/", http.StatusNotFound, ""},
		{"subdir_no_listing", "/private/edr/signatures/", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.wantCode {
				t.Fatalf("GET %s: status = %d, want %d", tc.path, resp.StatusCode, tc.wantCode)
			}
			if tc.wantBody != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != tc.wantBody {
					t.Fatalf("GET %s: body = %q, want %q", tc.path, body, tc.wantBody)
				}
			}
		})
	}
}
