package updates

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent-manager/config"
)

func TestEDRMirrorRouteServesStaticFiles(t *testing.T) {
	dir := t.TempDir()
	orig := config.EDRMirrorFolder
	config.EDRMirrorFolder = dir
	defer func() { config.EDRMirrorFolder = orig }()
	if err := os.WriteFile(filepath.Join(dir, "status.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	r := newRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/private/edr/status.json", nil))
	if w.Code != 200 || w.Body.String() != `{"ok":true}` {
		t.Fatalf("code=%d body=%q", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/private/edr/missing", nil))
	if w.Code != 404 {
		t.Fatalf("missing file: code=%d", w.Code)
	}
}
