package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var localLogsChannel chan *plugins.Log

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "worker" {
		return
	}

	CheckAgentManagerHealth()

	autService := NewLogAuthService()
	go func() {
		autService.SyncAuth()
	}()

	middlewares := NewMiddlewares(autService)

	// Retry logic for loading certificates
	maxRetries := 3
	retryDelay := 2 * time.Second
	var cert, key string
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		cert, key, err = loadCerts()
		if err == nil {
			break
		}

		_ = catcher.Error("cannot load certificates, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when loading certificates", err, nil)
			return
		}
	}

	cpu := runtime.NumCPU()

	localLogsChannel = make(chan *plugins.Log, cpu*100)

	for i := 0; i < cpu; i++ {
		go sendLog()
	}

	go startHTTPServer(middlewares, cert, key)
	_ = startGRPCServer(middlewares, cert, key)
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
