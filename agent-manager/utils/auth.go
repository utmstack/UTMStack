package utils

import (
	"crypto/tls"
	"net/http"
	"strings"
)

func IsConnectionKeyValid(panelUrl string, token string) bool {
	requestBody := strings.NewReader(token)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}
	resp, err := client.Post(panelUrl, "application/json", requestBody)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func IsKeyPairValid(key string, id uint, cache map[uint]string) (string, bool) {
	for agentId, agentKey := range cache {
		if key == agentKey && id == agentId {
			return agentKey, true
		}
	}
	return "", false
}
