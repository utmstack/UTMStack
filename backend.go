package main

import (
	"bytes"
	"crypto/tls"
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
	if err !=nil{
		return err
	}

	err = resp.Body.Close()

	return err
}
