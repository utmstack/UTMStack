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
			_ = catcher.Error("configuration not found", nil, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
			http.Error(w, "Configuration not found", http.StatusInternalServerError)
			return
		}

		if conf.ModuleActive {
			if r.Header.Get("authorization") == "" {
				message := "401 Missing Authorization Header"
				_ = catcher.Error("missing authorization header", nil, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
				j, _ := json.Marshal(message)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(j)
				if err != nil {
					_ = catcher.Error("cannot write response", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
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
				_ = catcher.Error("invalid authentication credentials", nil, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
				j, _ := json.Marshal(message)
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write(j)
				if err != nil {
					_ = catcher.Error("cannot write response", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
				}
				return
			}

			var newBody schema.BodyEvents
			err := json.NewDecoder(r.Body).Decode(&newBody)
			if err != nil {
				_ = catcher.Error("error decoding body", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
				return
			}

			events := newBody.Events
			CreateMessage(conf, events)

			j, _ := json.Marshal("HTTP 200 OK")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(j)
			if err != nil {
				_ = catcher.Error("cannot write response", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
			}
		} else {
			_ = catcher.Error("bitdefender module disabled", nil, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
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
		_ = catcher.Error("error loading certificates", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
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
	}

	server := &http.Server{
		Addr:           ":" + config.BitdefenderGZPort,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,
	}

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		_ = catcher.Error("server stopped unexpectedly", err, map[string]any{"process": "plugin_com.utmstack.bitdefender"})
	}
}
