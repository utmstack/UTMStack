package server

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/gorilla/mux"
	"github.com/utmstack/UTMStack/plugins/bitdefender/config"
	"github.com/utmstack/UTMStack/plugins/bitdefender/schema"
	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
)

func GetLogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := config.GetConfig()
		if conf == nil {
			_ = catcher.Error("configuration not found", nil, nil)
			http.Error(w, "Configuration not found", http.StatusInternalServerError)
			return
		}

		if conf.ModuleActive {
			if r.Header.Get("authorization") == "" {
				message := "401 Missing Authorization Header"
				_ = catcher.Error("missing authorization header", nil, nil)
				j, _ := json.Marshal(message)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(j)
				if err != nil {
					_ = catcher.Error("cannot write response", err, nil)
				}
				return
			}

			var isAuth bool
			for _, groupConf := range conf.ModuleGroups {
				moduleConfig := config.GetBDGZModuleConfig(groupConf)
				if utils.GenerateAuthCode(moduleConfig.ConnectionKey) == r.Header.Get("authorization") {
					isAuth = true
				}
			}
			if !isAuth {
				message := "401 Invalid Authentication Credentials"
				_ = catcher.Error("invalid authentication credentials", nil, nil)
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
				_ = catcher.Error("error decoding body", err, nil)
				return
			}

			events := newBody.Events
			CreateMessage(conf, events)

			j, _ := json.Marshal("HTTP 200 OK")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(j)
			if err != nil {
				_ = catcher.Error("cannot write response", err, nil)
			}
		} else {
			_ = catcher.Error("bitdefender module disabled", nil, nil)
		}
	}
}

func StartServer() {
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/api", GetLogs()).Methods("POST")
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Server is up and running"))
	}).Methods("GET")

	loadedCerts, err := loadCerts()
	if err != nil {
		_ = catcher.Error("error loading certificates", err, nil)
		return
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{loadedCerts},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},

		PreferServerCipherSuites: true,
	}

	server := &http.Server{
		Addr:           ":" + config.BitdefenderGZPort,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,
	}

	go func() {
		maxRetries := 3
		retryDelay := 2 * time.Second

		for retry := 0; retry < maxRetries; retry++ {
			err := server.ListenAndServeTLS("", "")
			if err == nil {
				return
			}

			_ = catcher.Error("error creating server, retrying", err, map[string]any{
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
			} else {
				_ = catcher.Error("all retries failed when creating server", err, nil)
			}
		}
	}()
}
