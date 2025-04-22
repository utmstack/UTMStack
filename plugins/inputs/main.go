package main

import (
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"os"
	"runtime"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var localLogsChannel chan *plugins.Log

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	CheckAgentManagerHealth()

	autService := NewLogAuthService()
	go autService.SyncAuth()

	middlewares := NewMiddlewares(autService)

	cert, key, err := loadCerts()
	if err != nil {
		_ = catcher.Error("cannot load certificates", err, nil)
		os.Exit(1)
	}

	cpu := runtime.NumCPU()

	localLogsChannel = make(chan *plugins.Log, cpu*100)

	for i := 0; i < cpu; i++ {
		go sendLog()
	}

	go startHTTPServer(middlewares, cert, key)
	startGRPCServer(middlewares, cert, key)
}

func loadCerts() (string, string, error) {
	certsFolderPath := plugins.PluginCfg("com.utmstack", false).Get("certsFolder").String()

	certsFolder, err := utils.MkdirJoin(certsFolderPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot create certificates directory: %v", err)
	}

	certPath := certsFolder.FileJoin(utmCertFileName)
	keyPath := certsFolder.FileJoin(utmCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return certPath, keyPath, nil
}
