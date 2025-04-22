package server

import (
	"encoding/json"
	"github.com/threatwinds/go-sdk/catcher"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/utmstack/UTMStack/plugins/bitdefender/configuration"
	"github.com/utmstack/UTMStack/plugins/bitdefender/schema"
	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
	"github.com/utmstack/config-client-go/types"
)

func GetLogs(config *types.ConfigurationSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.ModuleActive {
			if r.Header.Get("authorization") == "" {
				message := "401 Missing Authorization Header"
				_ = catcher.Error("missing authorization header", nil, map[string]any{})
				j, _ := json.Marshal(message)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(j)
				if err != nil {
					_ = catcher.Error("cannot write response", err, nil)
				}
				return
			}

			var isAuth bool
			for _, groupConf := range config.ConfigurationGroups {
				if utils.GenerateAuthCode(groupConf.Configurations[0].ConfValue) == r.Header.Get("authorization") {
					isAuth = true
				}
			}
			if !isAuth {
				message := "401 Invalid Authentication Credentials"
				_ = catcher.Error("invalid authentication credentials", nil, map[string]any{})
				j, _ := json.Marshal(message)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(j)
				if err != nil {
					_ = catcher.Error("cannot write response", err, nil)
				}
				return
			}

			var newBody schema.BodyEvents
			err := json.NewDecoder(r.Body).Decode(&newBody)
			if err != nil {
				_ = catcher.Error("error decoding body", err, map[string]any{})
				return
			}

			events := newBody.Events
			CreateMessage(config, events)

			j, _ := json.Marshal("HTTP 200 OK")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(j)
			if err != nil {
				_ = catcher.Error("cannot write response", err, nil)
			}
		} else {
			_ = catcher.Error("bitdefender module disabled", nil, map[string]any{})
		}
	}
}

func StartServer(cnf *types.ConfigurationSection, cert string, key string) {
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/api", GetLogs(cnf)).Methods("POST")
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Server is up and running"))
	}).Methods("GET")

	server := &http.Server{
		Addr:           ":" + configuration.BitdefenderGZPort,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		err := server.ListenAndServeTLS(cert, key)
		if err != nil {
			_ = catcher.Error("error creating server", err, map[string]any{})
		}
	}()
}
