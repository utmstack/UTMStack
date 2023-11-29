package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"
)

func generateSample(severity string, seq int) Log {
	return Log{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		DataSource: "sample-host",
		DataType:   "sample-data",
		Log: map[string]interface{}{
			"logType":     "Example data",
			"vendor":      "UTMStack",
			"severity":    severity,
			"description": "This is an example log used for automatic UTMStack test alert generation",
			"src_user":    "user-1",
			"dest_user":   "user-2",
			"src_host":    "dns.google",
			"dest_host":   "host-2.utmstack.com",
			"src_port":    24300,
			"dest_port":   8080,
			"src_ip":      "8.8.8.8",
			"dest_ip":     "10.10.10.10",
			"proto":       "tcp",
			"sequence":    seq,
		},
	}
}

func SendSampleData() error {
	severities := []string{
		"High",
		"Medium",
		"Low",
	}

	const baseURL = "http://127.0.0.1:10000"

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	for i := 0; i <= 10; i++ {
		for _, s := range severities {
			log := generateSample(s, i)

			jLog, err := json.Marshal(log)
			if err != nil {
				return err
			}

			req, err := http.NewRequest("POST", baseURL+"/v1/newlog", bytes.NewBuffer(jLog))
			if err != nil {
				return err
			}

			client := &http.Client{Transport: transCfg}

			var resp *http.Response

			for intent := 0; intent <= 10; intent++ {
				resp, err = client.Do(req)
				if err != nil {
					if intent >= 10 {
						return err
					}
					time.Sleep(1 * time.Minute)
				} else {
					break
				}
			}

			err = resp.Body.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Log struct {
	Timestamp  string                 `json:"@timestamp"`
	DataSource string                 `json:"dataSource"`
	DataType   string                 `json:"dataType"`
	Log        map[string]interface{} `json:"logx"`
}
