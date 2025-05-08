package processor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/aws/configuration"
	"github.com/utmstack/UTMStack/aws/utils"
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
	var logStrings []string
	for _, log := range data {
		body, err := json.Marshal(log)
		if err != nil {
			utils.Logger.ErrorF("error encoding log to JSON: %v", err)
			continue
		}
		logStrings = append(logStrings, string(body))
	}

	if len(logStrings) == 0 {
		return nil
	}

	var logs string
	for _, str := range logStrings {
		logs += str + configuration.UTMLogSeparator
	}

	url := fmt.Sprintf(configuration.LogstashEndpoint, configuration.GetLogstashHost(), configuration.GetLogstashPort())

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(logs))
	if err != nil {
		return utils.Logger.ErrorF("error creating request: %v", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			utils.Logger.ErrorF("error sending logs with error: %v", err.Error())
		}
		return utils.Logger.ErrorF("error sending logs: %v", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.Logger.ErrorF("error sending logs with http code %d", resp.StatusCode)
	}

	utils.Logger.Info("successfully sent %d logs to Logstash", len(logStrings))
	return nil
}
