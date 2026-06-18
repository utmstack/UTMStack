package services

import (
	"crypto/tls"
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

		resp, err := client.Get(baseURL + "/api/ping")
		if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 202) {
			if intent >= 10 {
				return fmt.Errorf("backend start failure: %v", err)
			}
		} else {
			break
		}
	}

	return nil
}

func SetBaseURL(hostname string) error {
	containerID, err := getPostgresContainerID()
	if err != nil {
		return err
	}

	baseURL := fmt.Sprintf("https://%s.utmstack.com", hostname)
	query := fmt.Sprintf("UPDATE public.utm_configuration_parameter SET conf_param_value='%s' WHERE conf_param_short='utmstack.mail.baseUrl';", baseURL)

	return execPsql(containerID, "utmstack", query)
}
