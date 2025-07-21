package processor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/sophos/configuration"
	"github.com/utmstack/UTMStack/sophos/utils"
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
			utils.Logger.ErrorF("error encoding log to JSON: %v", err)
			continue
		}
		if err := sendLogs(body); err != nil {
			utils.Logger.ErrorF("error sending logs to logstach: %v", err)
			continue
		}
	}
	return nil
}

func sendLogs(log []byte) error {
	url := fmt.Sprintf(configuration.LogstashEndpoint, configuration.GetLogstashHost(), configuration.GetLogstashPort())

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(log))
	if err != nil {
		return utils.Logger.ErrorF("error creating request: %v", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return utils.Logger.ErrorF("error sending logs: %v", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.Logger.ErrorF("error sending logs with http code %d", resp.StatusCode)
	}
	return nil
}
