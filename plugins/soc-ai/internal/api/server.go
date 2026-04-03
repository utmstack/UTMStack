package api

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/queue"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

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

// Server metrics
var (
	requestsReceived int64
	requestsQueued   int64
	requestsFailed   int64
)

// StartHTTPServer starts the HTTP API server for manual alert submission
func StartHTTPServer() {
	mux := http.NewServeMux()

	// Health check endpoint (no auth - for docker health checks)
	mux.HandleFunc("/health", handleHealth)

	// Protected endpoints (require X-Internal-Key header)
	mux.HandleFunc("/api/v1/analyze", authMiddleware(handleAnalyze))
	mux.HandleFunc("/api/v1/metrics", authMiddleware(handleMetrics))

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
			atomic.AddInt64(&requestsFailed, 1)
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
	atomic.AddInt64(&requestsReceived, 1)
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST
	if r.Method != http.MethodPost {
		atomic.AddInt64(&requestsFailed, 1)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	// Check if module is active
	if config.GetConfig() == nil || !config.GetConfig().ModuleActive {
		atomic.AddInt64(&requestsFailed, 1)
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
		atomic.AddInt64(&requestsFailed, 1)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Id == "" {
		atomic.AddInt64(&requestsFailed, 1)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			Message: "Alert ID is required",
		})
		return
	}

	// Enqueue for processing
	if !queue.EnqueueManual(&req.AlertFields) {
		atomic.AddInt64(&requestsFailed, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Status:  "error",
			AlertID: req.Id,
			Message: "Queue is full. Try again later.",
		})
		return
	}

	atomic.AddInt64(&requestsQueued, 1)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(AnalyzeResponse{
		Status:  "queued",
		AlertID: req.Id,
		Message: "Alert queued for LLM analysis",
	})
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"requests_received": atomic.LoadInt64(&requestsReceived),
		"requests_queued":   atomic.LoadInt64(&requestsQueued),
		"requests_failed":   atomic.LoadInt64(&requestsFailed),
	})
}
