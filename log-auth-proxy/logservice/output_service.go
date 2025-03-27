package logservice

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/panelservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/utils"
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

type LogOutputService struct {
	Connections map[config.LogType]string
	Mutex       sync.Mutex
	Ticker      *time.Ticker
	Client      *http.Client
}

func NewLogOutputService() *LogOutputService {
	connections, _ := getServiceMap()
	return &LogOutputService{
		Connections: connections,
		Client:      &http.Client{Transport: transport, Timeout: 2 * time.Second},
	}
}

func (out *LogOutputService) SendLog(logType config.LogType, logData string) {
	out.Mutex.Lock()
	defer out.Mutex.Unlock()
	port, err := out.getConnectionPort(logType)
	if err != nil {
		utils.Logger.ErrorF("error getting connection port: %v", err)
		return
	}
	singleLog := logData + config.UTMLogSeparator
	out.sendLogsToLogstash(port, singleLog)
}

func (out *LogOutputService) SendBulkLog(logType config.LogType, logDataArray []string) {
	out.Mutex.Lock()
	defer out.Mutex.Unlock()

	var logs string
	for _, str := range logDataArray {
		logs += str + config.UTMLogSeparator
	}

	port, err := out.getConnectionPort(logType)
	if err != nil {
		utils.Logger.ErrorF("error getting connection port: %v", err)
		return
	}

	out.sendLogsToLogstash(port, logs)
}

func (out *LogOutputService) getConnectionPort(logType config.LogType) (string, error) {
	port, existLogType := out.Connections[logType]
	if !existLogType {
		portGeneric, isGenericUp := out.Connections[config.Generic]
		if !isGenericUp {
			return "", fmt.Errorf("neither %s or %s connections are available", logType, config.Generic)
		}
		port = portGeneric
	}

	if port == "" {
		return "", fmt.Errorf("connection is nil for service: %s", logType)
	}
	return port, nil
}

func (out *LogOutputService) sendLogsToLogstash(port string, logs string) {
	url := fmt.Sprintf(config.LogstashPipelinesEndpoint, config.LogstashHost(), port)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(logs))
	if err != nil {
		utils.Logger.ErrorF("error creating request: %v", err)
	}

	resp, err := out.Client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			utils.Logger.ErrorF("error sending logs with error: %v", err.Error())
		}
		return
	}
	if resp.StatusCode != http.StatusOK {
		utils.Logger.ErrorF("error sending logs with http code %d", resp.StatusCode)
		return
	}
}

func getServiceMap() (map[config.LogType]string, error) {
	logTypesMap := make(map[config.LogType]string)
	pipelines, err := panelservice.GetPipelines()
	if err != nil {
		return logTypesMap, errors.New("unable to get the pipelines: " + err.Error())
	}
	for _, pipeline := range pipelines {
		logTypesMap[pipeline.InputId] = pipeline.Port
	}
	return logTypesMap, nil
}

func (out *LogOutputService) SyncOutputs() {
	out.Ticker = time.NewTicker(60 * time.Second)
	go func() {
		for range out.Ticker.C {
			serviceMap, err := getServiceMap()
			if err != nil {
				utils.Logger.ErrorF("error getting service map: %v", err)
				continue
			}
			out.Mutex.Lock()
			out.Connections = serviceMap
			out.Mutex.Unlock()
		}
	}()
}
