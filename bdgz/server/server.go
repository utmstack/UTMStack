package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/bdgz/constants"
	"github.com/utmstack/UTMStack/bdgz/schema"
	"github.com/utmstack/UTMStack/bdgz/utils"
	"github.com/utmstack/config-client-go/types"
)

var syslogHelper EpsSyslogHelper

// GetBDGZLogs gets the Bitdefender Api Push logs and sends them to the syslog server
func GetBDGZLogs(config *types.ConfigurationSection, h *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Info("New group of events received")
		// Check if the Bitdefender Module is active
		if config.ModuleActive {
			//Check if the authorization exist
			if r.Header.Get("authorization") == "" {
				messag := "401 Missing Authorization Header"
				h.ErrorF(messag)
				j, _ := json.Marshal(messag)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(j)
				return
			}

			//Check if the authorization is valid
			var isAuth bool
			for _, groupConf := range config.ConfigurationGroups {
				if utils.GenerateAuthCode(groupConf.Configurations[0].ConfValue) == r.Header.Get("authorization") {
					isAuth = true
				}
			}
			if !isAuth {
				messag := "401 Invalid Authentication Credentials"
				h.ErrorF(messag)
				j, _ := json.Marshal(messag)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(j)
				return
			}

			// Decode the request body
			var newBody schema.BodyEvents
			err := json.NewDecoder(r.Body).Decode(&newBody)
			if err != nil {
				h.ErrorF("error to decode body: %v", err)
				return
			}

			// Process the events and send them to the syslog server
			events := newBody.Events
			syslogHelper.SentToSyslog(config, events, h)

			// Return a successful HTTP response
			j, _ := json.Marshal("HTTP 200 OK")
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		} else {
			h.ErrorF("Bitdefender module disabled")
		}
	}
}

// ServerUp raises the connector that will receive the data and process it so that it is sent to the syslog server
func ServerUp(cnf *types.ConfigurationSection, certsPath string, h *logger.Logger) {
	syslogHelper.Init()

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/api", (GetBDGZLogs(cnf, h))).Methods("POST")
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Server is up and running"))
	}).Methods("GET")

	server := &http.Server{
		Addr:           ":" + constants.GetConnectorPort(),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		h.Info("Listening in port %s...\n", constants.GetConnectorPort())
		err := server.ListenAndServeTLS(filepath.Join(certsPath, "server.crt"), filepath.Join(certsPath, "server.key"))
		if err != nil {
			h.ErrorF("%v", err)
		}
		//Close connection with syslogServer
		syslogHelper.clientSyslog.Close()
	}()
}
