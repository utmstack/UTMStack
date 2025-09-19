package processor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/office365/configuration"
)

var transport = &http.Transport{
	MaxIdleConns:          100,
	IdleConnTimeout:       2 * time.Second,
	ResponseHeaderTimeout: 2 * time.Second,
	ForceAttemptHTTP2:     true,
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

var client = &http.Client{Transport: transport, Timeout: 2 * time.Second}

func SendToLogstash(data []TransformedLog) *logger.Error {
	for _, str := range data {
		body, err := json.Marshal(str)
		if err != nil {
			catcher.Error("error encoding log to JSON", err, map[string]any{})
			continue
		}
		if err := sendLogs(body); err != nil {
			catcher.Error("error sending logs to logstach", err, map[string]any{})
			continue
		}
	}
	return nil
}

func sendLogs(log []byte) error {
	url := fmt.Sprintf(configuration.LogstashEndpoint, configuration.GetLogstashHost(), configuration.GetLogstashPort())

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(log))
	if err != nil {
		return catcher.Error("error creating request", err, map[string]any{})
	}

	resp, err := client.Do(req)
	if err != nil {
		return catcher.Error("error sending logs: %v", err, map[string]any{})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return catcher.Error("error sending logs with http code", nil, map[string]any{"status_code": resp.StatusCode})
	}
	return nil
}
