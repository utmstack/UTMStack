package logservice

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/model"
	"github.com/utmstack/UTMStack/log-auth-proxy/panelservice"
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
	Connections map[model.LogType]string
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

func (out *LogOutputService) SendLog(logType model.LogType, logData string) {
	out.Mutex.Lock()
	defer out.Mutex.Unlock()
	port, err := out.getConnectionPort(logType)
	if err != nil {
		log.Println(err)
		return
	}
	singleLog := logData + config.UTMLogSeparator
	out.sendLogsToLogstash(port, singleLog)
}

func (out *LogOutputService) SendBulkLog(logType model.LogType, logDataArray []string) {
	out.Mutex.Lock()
	defer out.Mutex.Unlock()

	var logs string
	for _, str := range logDataArray {
		logs += str + config.UTMLogSeparator
	}

	port, err := out.getConnectionPort(logType)
	if err != nil {
		log.Println(err)
		return
	}

	out.sendLogsToLogstash(port, logs)
}

func (out *LogOutputService) getConnectionPort(logType model.LogType) (string, error) {
	port, existLogType := out.Connections[logType]
	if !existLogType {
		portGeneric, isGenericUp := out.Connections[model.Generic]
		if !isGenericUp {
			return "", fmt.Errorf("neither %s or %s connections are available", logType, model.Generic)
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
		log.Printf("error creating request: %v", err.Error())
	}

	resp, err := out.Client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			log.Printf("error sending logs with error: %v", err.Error())
		}
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("error sending logs with http code %d", resp.StatusCode)
		return
	}
}

func getServiceMap() (map[model.LogType]string, error) {
	logTypesMap := make(map[model.LogType]string)
	pipelines, err := panelservice.GetPipelines()
	if err != nil {
		return logTypesMap, errors.New("unable to get the pipelines")
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
				fmt.Println("error getting service map:", err)
				continue
			}
			out.Mutex.Lock()
			out.Connections = serviceMap
			out.Mutex.Unlock()
		}
	}()
}
