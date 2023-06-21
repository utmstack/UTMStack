package main

import (
	"crypto/tls"
	"net/http"
	"time"
)

func Backend() error {
	baseURL := "https://127.0.0.1/"

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
