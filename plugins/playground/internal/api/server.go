package api

import (
	"encoding/json"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/playground/config"
	"github.com/utmstack/UTMStack/plugins/playground/internal/runner"
)

type Response struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type WriteFileRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func StartHTTPServer(addr string, r *runner.Runner, keyFn func() string) {
	mux := http.NewServeMux()

	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(keyFn, h)
	}

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /playground/filters", auth(makeWriteHandler(r, "filter")))
	mux.HandleFunc("POST /playground/rules", auth(makeWriteHandler(r, "rule")))
	mux.HandleFunc("POST /playground/filters/copy", auth(makeCopyHandler(r, "filter")))
	mux.HandleFunc("POST /playground/rules/copy", auth(makeCopyHandler(r, "rule")))
	mux.HandleFunc("DELETE /playground/filters", auth(makeDeleteHandler(r, "filter")))
	mux.HandleFunc("DELETE /playground/rules", auth(makeDeleteHandler(r, "rule")))
	mux.HandleFunc("POST /playground/run", auth(makeRunHandler(r)))
	mux.HandleFunc("GET /playground/status", auth(makeStatusHandler(r)))

	catcher.Info("Starting playground HTTP API server", map[string]any{
		"process": config.ProcessName,
		"address": addr,
	})

	if err := http.ListenAndServe(addr, mux); err != nil {
		_ = catcher.Error("playground HTTP server failed", err, map[string]any{
			"process": config.ProcessName,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, Response{Status: "error", Message: msg})
}
