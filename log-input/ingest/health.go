package ingest

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

// Health is a plain HTTP endpoint on its own port. It is deliberately not the
// gRPC health service: that port is behind the auth interceptors, so a probe
// would have to carry credentials to be told the process is alive.
type Health struct {
	srv     *http.Server
	serving atomic.Bool
}

func NewHealth(addr string) *Health {
	h := &Health{}
	h.serving.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if !h.serving.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("draining"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	h.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return h
}

func (h *Health) Serve() {
	if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		_ = catcher.Error("the health endpoint stopped", err, map[string]any{
			"process": processName,
			"addr":    h.srv.Addr,
		})
	}
}

// Draining makes the probe fail while the streams in flight finish, so nothing
// new is routed to a replica on its way out.
func (h *Health) Draining() {
	h.serving.Store(false)
}

func (h *Health) Close(ctx context.Context) {
	_ = h.srv.Shutdown(ctx)
}
