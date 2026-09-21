// edr/internal/serve/serve.go
// Package serve exposes the mirror tree over plain HTTP for self-signed
// deployments. On such installs freshclam cannot validate the server's TLS
// certificate, so the agent fetches the SAME mirror tree over plain HTTP at
// the SAME /private/edr/ path layout that agentmanager serves over HTTPS.
// Integrity of this public, signed data is guaranteed by ClamAV's own CVD
// signatures, so serving it over plain HTTP is intentional and safe.
package serve

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Handler serves the files under mirrorDir at the URL prefix /private/edr/.
// Directory requests (a trailing "/", including the bare prefix) return 404
// rather than a browsable listing — agents fetch named files only, so the
// mirror's tree layout is not exposed. Requests outside the prefix 404.
func Handler(mirrorDir string) http.Handler {
	fs := http.StripPrefix("/private/edr/", http.FileServer(http.Dir(mirrorDir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})
}

// Serve runs an HTTP server bound to addr that exposes mirrorDir. It shuts
// down gracefully when ctx is cancelled and returns nil on a clean shutdown.
func Serve(ctx context.Context, addr, mirrorDir string) error {
	srv := &http.Server{Addr: addr, Handler: Handler(mirrorDir)}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
