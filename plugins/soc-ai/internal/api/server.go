package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/agent"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/queue"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

// AgentTaskRequest is the body for the live operations endpoint. Page is a
// human-readable description of where the user currently is in the app (route +
// any focused entity), forwarded so the agent can pick tools and craft navigation.
type AgentTaskRequest struct {
	Task string `json:"task"`
	Page string `json:"page"`
	// Lang is the user's interface language code (en/es/pt/…) so the agent
	// replies in that language regardless of the message language.
	Lang string `json:"lang"`
}

// AnalyzeRequest represents the request body for manual alert analysis
type AnalyzeRequest struct {
	schema.AlertFields
}

// AnalyzeResponse represents the response for the analyze endpoint
type AnalyzeResponse struct {
	Status  string `json:"status"`
	AlertID string `json:"alertId,omitempty"`
	Message string `json:"message,omitempty"`
}

// StartHTTPServer starts the HTTP API server for manual alert submission
func StartHTTPServer() {
	mux := http.NewServeMux()

	// Health check endpoint (no auth - for docker health checks)
	mux.HandleFunc("/health", handleHealth)

	// Protected endpoints (require X-Internal-Key header)
	mux.HandleFunc("/api/v1/analyze", authMiddleware(handleAnalyze))
	mux.HandleFunc("/api/v1/agent/task", authMiddleware(handleAgentTask))

	addr := ":" + config.HTTP_API_PORT
	catcher.Info("Starting HTTP API server", map[string]any{
		"process": "plugin_com.utmstack.soc-ai",
		"address": addr,
	})

	if err := http.ListenAndServe(addr, mux); err != nil {
		catcher.Error("HTTP server failed", err, map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
		})
	}
}

// authMiddleware validates the X-Internal-Key header for protected endpoints
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cfg := config.GetConfig()
		if cfg == nil || cfg.InternalKey == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(AnalyzeResponse{
				Status:  "error",
				Message: "Service not configured",
			})
			return
		}

		key := r.Header.Get("X-Internal-Key")
		if key == "" || key != cfg.InternalKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(AnalyzeResponse{
				Status:  "error",
				Message: "Unauthorized: Invalid or missing X-Internal-Key header",
			})
			return
		}

		next(w, r)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	// Check if module is active
	if config.GetConfig() == nil || !config.GetConfig().ModuleActive {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "SOC-AI module is not active",
		})
		return
	}

	// Parse request body
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Alert ID is required",
		})
		return
	}

	// Enqueue for processing
	if !queue.EnqueueManual(&req.AlertFields) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			AlertID: req.Id,
			Message: "Queue is full. Try again later.",
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(AnalyzeResponse{
		Status:  "queued",
		AlertID: req.Id,
		Message: "Alert queued for LLM analysis",
	})
}

func handleAgentTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	ag := agent.Current()
	if ag == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "SOC-AI agent is not configured")
		return
	}

	var req AgentTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Task == "" {
		writeJSONError(w, http.StatusBadRequest, `Body must be {"task": "..."}`)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.WriteHeader(http.StatusOK)

	var mu sync.Mutex
	sink := func(ev agent.Event) {
		mu.Lock()
		defer mu.Unlock()
		data, err := json.Marshal(ev)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Kind, data)
		flusher.Flush()
	}

	maxIters := 0
	var capabilities []string
	if cfg := config.GetConfig(); cfg != nil {
		maxIters = cfg.MaxToolIterations
		capabilities = cfg.Capabilities
	}

	_, _ = ag.Run(r.Context(), agent.RunTask{
		System:        agent.OpsPrompt(req.Page, req.Lang, capabilities),
		Input:         req.Task,
		EnabledGroups: capabilities,
		MaxIters:      maxIters,
	}, sink)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(AnalyzeResponse{Status: "error", Message: msg})
}
