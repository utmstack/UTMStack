package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func Master(c *Config) error {
	if err := utils.CheckMem(8); err != nil {
		return err
	}

	if err := utils.CheckCPU(4); err != nil {
		return err
	}

	baseURL := "https://127.0.0.1/"

	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: transCfg}

		_, err := client.Get(baseURL + "/api/ping")

		if err != nil && intent <= 9 {
			continue
		} else if err == nil {
			break
		}

		return err
	}

	return nil
}
