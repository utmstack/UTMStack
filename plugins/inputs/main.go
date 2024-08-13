package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var localLogsChannel chan *plugins.Log
var localNotificationsChannel chan *plugins.Message

type PluginConfig struct {
	ServerName   string `yaml:"server_name"`
	InternalKey  string `yaml:"internal_key"`
	AgentManager string `yaml:"agent_manager"`
	Backend      string `yaml:"backend"`
	Logstash     string `yaml:"logstash"`
	CertsFolder  string `yaml:"certs_folder"`
}

func main() {
	autService := NewLogAuthService()
	go autService.SyncAuth()

	middlewares := NewMiddlewares(autService)

	cert, key, err := loadCerts()
	if err != nil {
		helpers.Logger().ErrorF("failed to load certificates: %v", err)
		os.Exit(1)
	}

	cpu := runtime.NumCPU()

	localLogsChannel = make(chan *plugins.Log, cpu*100)
	localNotificationsChannel = make(chan *plugins.Message, cpu*100)

	for i := 0; i < cpu; i++ {
		go sendLog()
		go sendNotification()
	}

	go startHTTPServer(middlewares, cert, key)
	startGRPCServer(middlewares, cert, key)
}

func loadCerts() (string, string, error) {
	cnf, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
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