package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	go_sdk "github.com/threatwinds/go-sdk"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var localLogsChannel chan *go_sdk.Log

type PluginConfig struct {
	ServerName   string `yaml:"serverName"`
	InternalKey  string `yaml:"internalKey"`
	AgentManager string `yaml:"agentManager"`
	Backend      string `yaml:"backend"`
	Logstash     string `yaml:"logstash"`
	CertsFolder  string `yaml:"certsFolder"`
}

func main() {
	mode := os.Getenv("MODE")
	if mode == "manager" {
		os.Exit(0)
	}

	CheckAgentManagerHealth()

	autService := NewLogAuthService()
	go autService.SyncAuth()

	middlewares := NewMiddlewares(autService)

	cert, key, err := loadCerts()
	if err != nil {
		go_sdk.Logger().ErrorF("failed to load certificates: %v", err)
		os.Exit(1)
	}

	cpu := runtime.NumCPU()

	localLogsChannel = make(chan *go_sdk.Log, cpu*100)

	for i := 0; i < cpu; i++ {
		go sendLog()
	}

	go startHTTPServer(middlewares, cert, key)
	startGRPCServer(middlewares, cert, key)
}

func loadCerts() (string, string, error) {
	cnf, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	certPath := filepath.Join(cnf.CertsFolder, utmCertFileName)
	keyPath := filepath.Join(cnf.CertsFolder, utmCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return certPath, keyPath, nil
}
