package api

import (
	"net/http"
)

func authMiddleware(keyFn func() string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := keyFn()
		if key == "" {
			writeError(w, http.StatusServiceUnavailable, "service not configured")
			return
		}

		clientKey := r.Header.Get("X-Internal-Key")
		if clientKey == "" || clientKey != key {
			writeError(w, http.StatusUnauthorized, "unauthorized: invalid or missing X-Internal-Key")
			return
		}

		next(w, r)
	}
}
