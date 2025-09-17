package server

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

var (
	reconnectDelay = 5 * time.Second
)

func loadCerts() (tls.Certificate, error) {
	certsFolderConfig := ""
	for {
		pluginConfig := plugins.PluginCfg("com.utmstack", false)
		if !pluginConfig.Exists() {
			_ = catcher.Error("plugin configuration not found", nil, nil)
			time.Sleep(reconnectDelay)
			continue
		}

		certsFolderConfig = pluginConfig.Get("certsFolder").String()
		break
	}

	certsFolder, err := utils.MkdirJoin(certsFolderConfig)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("cannot create certificates directory: %v", err)
	}

	certPath := certsFolder.FileJoin("utm.crt")
	keyPath := certsFolder.FileJoin("utm.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return tls.Certificate{}, fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return tls.Certificate{}, fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return tls.LoadX509KeyPair(certPath, keyPath)
}
