package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/plugins/bitdefender/configuration"
	"github.com/utmstack/UTMStack/plugins/bitdefender/schema"
	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
	"github.com/utmstack/config-client-go/types"
)

func GetBDGZLogs(config *types.ConfigurationSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helpers.Logger().Info("New group of events received")
		if config.ModuleActive {
			if r.Header.Get("authorization") == "" {
				messag := "401 Missing Authorization Header"
				helpers.Logger().ErrorF(messag)
				j, _ := json.Marshal(messag)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(j)
				return
			}

			var isAuth bool
			for _, groupConf := range config.ConfigurationGroups {
				if utils.GenerateAuthCode(groupConf.Configurations[0].ConfValue) == r.Header.Get("authorization") {
					isAuth = true
				}
			}
			if !isAuth {
				messag := "401 Invalid Authentication Credentials"
				helpers.Logger().ErrorF(messag)
				j, _ := json.Marshal(messag)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(j)
				return
			}

			var newBody schema.BodyEvents
			err := json.NewDecoder(r.Body).Decode(&newBody)
			if err != nil {
				helpers.Logger().ErrorF("error to decode body: %v", err)
				return
			}

			events := newBody.Events
			CreateMessage(config, events)

			j, _ := json.Marshal("HTTP 200 OK")
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		} else {
			helpers.Logger().ErrorF("Bitdefender module disabled")
		}
	}
}

func ServerUp(cnf *types.ConfigurationSection, cert string, key string) {

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/api", (GetBDGZLogs(cnf))).Methods("POST")
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Server is up and running"))
	}).Methods("GET")

	server := &http.Server{
		Addr:           ":" + configuration.GetConfig().BdgzPort,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		helpers.Logger().Info("Listening in port %s...\n", configuration.GetConfig().BdgzPort)
		err := server.ListenAndServeTLS(cert, key)
		if err != nil {
			helpers.Logger().ErrorF("error creating server: %v", err)
		}
	}()
}
