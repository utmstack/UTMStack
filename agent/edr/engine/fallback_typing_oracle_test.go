package engine

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// These harmless custom-feed cases were compared with ClamAV 1.5.4 using
// only the two NDB rules below (no FTM): head, JSP and pyc detect; header,
// html-with-attributes and the near-miss pyc header remain clean.
func TestBuiltInTypeMarkersMatchClamAV154Oracle(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	const htmlMarker = "HARMLESS_V171_HTML_MARKER"
	const pythonMarker = "HARMLESS_V171_PYTHON_MARKER"
	rules := map[string][]byte{
		"html.ndb":  []byte("Native.HTML.Marker:3:*:" + hex.EncodeToString([]byte("harmless_v171_html_marker")) + "\n"),
		"other.ndb": []byte("Native.Python.Marker:14:*:" + hex.EncodeToString([]byte(pythonMarker)) + "\n"),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("fallback-typing-154")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, input, want string
	}{
		{"head", "<head></head><body>" + htmlMarker + "</body>", "Native.HTML.Marker"},
		{"header", "<header></header><body>" + htmlMarker + "</body>", ""},
		{"html attributes", "<html lang=en><body>" + htmlMarker + "</body></html>", ""},
		{"jsp", "<%@ LANGUAGE = VBScript %><body>" + htmlMarker + "</body>", "Native.HTML.Marker"},
		{"python", string([]byte{0x2d, 0x0d, 0x0d, 0x0a, 0x80}) + pythonMarker, "Native.Python.Marker"},
		{"near pyc", string([]byte{0x2d, 0x0d, 0x0d, 0x0b, 0x80}) + pythonMarker, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clean, name, err := e.Scan(context.Background(), []byte(tc.input))
			if err != nil || clean != (tc.want == "") || name != tc.want {
				t.Fatalf("scan clean=%v name=%q err=%v, want %q", clean, name, err, tc.want)
			}
		})
	}
}
