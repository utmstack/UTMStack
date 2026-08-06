package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/playground/config"
	"github.com/utmstack/UTMStack/plugins/playground/internal/runner"
)

var nameAllowlist = regexp.MustCompile(`^[a-zA-Z0-9_-]+\.yaml$`)

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, Response{Status: "ok"})
}

func makeWriteHandler(r *runner.Runner, kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.Lock()
		defer r.Unlock()

		var body WriteFileRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if !nameAllowlist.MatchString(body.Name) {
			writeError(w, http.StatusBadRequest,
				`invalid name: must match ^[a-zA-Z0-9_-]+\.yaml$`)
			return
		}

		var destDir string
		if kind == "filter" {
			destDir = filepath.Join(r.WorkDir(), "pipeline", "filters")
		} else {
			destDir = filepath.Join(r.WorkDir(), "rules")
		}

		destPath := filepath.Join(destDir, body.Name)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to ensure directory: %v", err))
			return
		}
		if err := runner.WriteFileAtomic(destDir, destPath, []byte(body.Content), 0644); err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to write file: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, Response{Status: "ok"})
	}
}

func makeCopyHandler(r *runner.Runner, kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.Lock()
		defer r.Unlock()

		var src, dst string
		if kind == "filter" {
			src = filepath.Join(config.ProductionWorkDir, "pipeline", "filters")
			dst = filepath.Join(config.WorkDir, "pipeline", "filters")
		} else {
			src = filepath.Join(config.ProductionWorkDir, "rules")
			dst = filepath.Join(config.WorkDir, "rules")
		}

		if err := runner.ClearDir(dst); err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to clear destination directory: %v", err))
			return
		}
		if err := runner.CopyDir(src, dst); err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to copy from production: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, Response{Status: "ok"})
	}
}

func makeDeleteHandler(r *runner.Runner, kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.Lock()
		defer r.Unlock()

		var dir string
		if kind == "filter" {
			dir = filepath.Join(config.WorkDir, "pipeline", "filters")
		} else {
			dir = filepath.Join(config.WorkDir, "rules")
		}

		if err := runner.DeleteFilesInDir(dir); err != nil {
			writeError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to delete files: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, Response{Status: "ok"})
	}
}

func makeRunHandler(r *runner.Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		req.Body = http.MaxBytesReader(w, req.Body, config.MaxRunBodyBytes)

		var log plugins.Log
		if err := json.NewDecoder(req.Body).Decode(&log); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeError(w, http.StatusRequestEntityTooLarge,
					"request body too large (max 1 MiB)")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if log.TenantId == "" {
			log.TenantId = config.DefaultTenant
		}

		result, err := r.Run(req.Context(), &log)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				"internal error: "+err.Error())
			return
		}

		data, err := json.Marshal(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				"failed to marshal result: "+err.Error())
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Status: "ok",
			Data:   json.RawMessage(data),
		})
	}
}

func makeStatusHandler(r *runner.Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		state := r.State()
		if !state.Active {
			writeJSON(w, http.StatusOK, Response{Status: "idle"})
			return
		}

		uuidData, _ := json.Marshal(map[string]string{"uuid": state.UUID})
		writeJSON(w, http.StatusOK, Response{
			Status: "running",
			Data:   json.RawMessage(uuidData),
		})
	}
}
