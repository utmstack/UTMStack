package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/log-input/config"
)

const (
	maxBodyBytes = 10 << 20 // 10 MB
	maxBatchSize = 1000
)

type ingestResponse struct {
	Accepted int           `json:"accepted"`
	Failed   []failedEntry `json:"failed"`
}

type failedEntry struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

func allErrorsAreNATSDown(errs []error) bool {
	if len(errs) == 0 {
		return false
	}
	for _, err := range errs {
		if !isNATSDown(err) {
			return false
		}
	}
	return true
}

func isNATSDown(err error) bool {
	return errors.Is(err, nats.ErrConnectionClosed) ||
		errors.Is(err, nats.ErrNoResponders)
}

type HTTPServer struct {
	srv      *http.Server
	resolver tenantResolver
	pub      Publisher
	cfg      *config.Config
}

func NewHTTPServer(cfg *config.Config, resolver tenantResolver, pub Publisher) (*HTTPServer, error) {
	tlsCfg, err := loadTLSConfig(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, catcher.Error("cannot load TLS config for HTTP server", err, map[string]any{
			"process": processName,
			"cert":    cfg.CertFile,
		})
	}

	h := &HTTPServer{
		resolver: resolver,
		pub:      pub,
		cfg:      cfg,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handle(w, r)
	})

	h.srv = &http.Server{
		Addr:         cfg.HTTPListenAddr,
		Handler:      mux,
		TLSConfig:    tlsCfg,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 35 * time.Second,
	}

	return h, nil
}

func (h *HTTPServer) Serve() error {
	catcher.Info("http ingest listening", map[string]any{
		"process": processName,
		"addr":    h.cfg.HTTPListenAddr,
	})
	return h.srv.ListenAndServeTLS("", "")
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func (h *HTTPServer) handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			_ = catcher.Error("panic in http ingest handler", fmt.Errorf("%v", rec), map[string]any{
				"process": processName,
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}()

	logs, err := parseBody(w, r)
	if err != nil {
		return
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	creds := Credentials{
		APIKey:   r.Header.Get("Utm-Api-Key"),
		ConnKey:  r.Header.Get("X-Connector-Key"),
		ConnType: r.Header.Get("X-Connector-Type"),
		ClientIP: clientIP,
	}
	if connIDStr := r.Header.Get("X-Connector-Id"); connIDStr != "" {
		creds.ConnID, err = strconv.ParseUint(connIDStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-Connector-Id is not a valid uint64"})
			return
		}
	}

	tenant, err := resolveAuth(r.Context(), h.resolver, creds)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	ctx := WithTenant(r.Context(), tenant)

	var (
		accepted int
		failed   []failedEntry
		pubErrs  []error
	)

	for i, l := range logs {
		applyDefaults(ctx, h.cfg, l)

		if pubErr := h.pub.Publish(ctx, l); pubErr != nil {
			failed = append(failed, failedEntry{Index: i, Reason: pubErr.Error()})
			pubErrs = append(pubErrs, pubErr)
			continue
		}
		accepted++
	}

	if accepted == 0 && len(failed) > 0 && allErrorsAreNATSDown(pubErrs) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "broker unavailable"})
		return
	}

	if failed == nil {
		failed = []failedEntry{}
	}
	writeJSON(w, http.StatusOK, ingestResponse{Accepted: accepted, Failed: failed})
}

func parseBody(w http.ResponseWriter, r *http.Request) ([]*plugins.Log, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var buf []byte
	{
		dec := json.NewDecoder(r.Body)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "body exceeds 10 MB limit"})
				return nil, err
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return nil, err
		}
		buf = []byte(raw)
	}

	var shapeProbe struct {
		Logs *[]json.RawMessage `json:"logs"`
	}
	if err := json.Unmarshal(buf, &shapeProbe); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return nil, err
	}

	if shapeProbe.Logs != nil {
		rawLogs := *shapeProbe.Logs
		if len(rawLogs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty batch"})
			return nil, fmt.Errorf("empty batch")
		}
		if len(rawLogs) > maxBatchSize {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("batch exceeds %d logs", maxBatchSize),
			})
			return nil, fmt.Errorf("batch exceeds %d logs", maxBatchSize)
		}
		logs := make([]*plugins.Log, 0, len(rawLogs))
		for i, rawEntry := range rawLogs {
			var l plugins.Log
			if err := json.Unmarshal(rawEntry, &l); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": fmt.Sprintf("log[%d]: invalid JSON: %s", i, err.Error()),
				})
				return nil, err
			}
			logs = append(logs, &l)
		}
		return logs, nil
	}

	var l plugins.Log
	if err := json.Unmarshal(buf, &l); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return nil, err
	}
	return []*plugins.Log{&l}, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
