package http

import (
	"io"
	stdhttp "net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"
)

func (h *HTTPCollector) newHandler(inst *httpInstance) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodPost {
			stdhttp.Error(w, "method not allowed", stdhttp.StatusMethodNotAllowed)
			return
		}

		r.Body = stdhttp.MaxBytesReader(w, r.Body, 8<<20)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			if strings.Contains(err.Error(), "http: request body too large") {
				stdhttp.Error(w, "payload too large", stdhttp.StatusRequestEntityTooLarge)
			} else {
				stdhttp.Error(w, "bad request", stdhttp.StatusBadRequest)
			}
			return
		}

		if err := validateAuth(inst, r, bodyBytes); err != nil {
			stdhttp.Error(w, "unauthorized", stdhttp.StatusUnauthorized)
			return
		}

		select {
		case h.queue <- &plugins.Log{
			Id:         uuid.NewString(),
			DataType:   inst.dataType,
			DataSource: inst.dataType,
			Raw:        string(bodyBytes),
		}:
		default:
			w.Header().Set("Retry-After", "5")
			stdhttp.Error(w, "service unavailable", stdhttp.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(stdhttp.StatusNoContent)
	})
}
