package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

func Backend() error {
	baseURL := "https://127.0.0.1"

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transCfg}

	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		_, err := client.Get(baseURL + "/api/ping")
		if err != nil {
			if intent >= 10 {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func RegenerateKey(internal string) error {
	baseURL := "https://127.0.0.1"

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	req, err := http.NewRequest("GET", baseURL+"/api/federation-service/generateApiToken", bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	req.Header.Add("Utm-Internal-Key", internal)

	client := &http.Client{Transport: transCfg}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = resp.Body.Close()

	return err
}

func SetBaseURL(password, hostname string) error {
	// Connecting to utmstack
	psqlconn := fmt.Sprintf("host=localhost port=5432 user=postgres password=%s sslmode=disable database=utmstack", password)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return err
	}

	// Close connection when finish
	defer db.Close()

	// Check connection status
	err = db.Ping()
	if err != nil {
		return err
	}

	// Set Base URL
	_, err = db.Exec(`UPDATE public.utm_configuration_parameter SET conf_param_value=$1 WHERE conf_param_short='utmstack.mail.baseUrl';`, fmt.Sprintf("https://%s.utmstack.com", hostname))
	if err != nil {
		return err
	}

	return err
}
