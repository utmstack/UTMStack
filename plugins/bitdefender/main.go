package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/utmstack/UTMStack/plugins/bitdefender/configuration"
	"github.com/utmstack/UTMStack/plugins/bitdefender/processor"
	"github.com/utmstack/UTMStack/plugins/bitdefender/server"
	"github.com/utmstack/config-client-go/types"
)

var (
	mutex        = &sync.Mutex{}
	moduleConfig = types.ConfigurationSection{}
)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	cert, key, err := loadCerts()
	if err != nil {
		_ = catcher.Error("cannot load certificates", err, nil)
		os.Exit(1)
	}

	server.StartServer(&moduleConfig, cert, key)
	go configuration.ConfigureModules(&moduleConfig, mutex)

	go processor.ProcessLogs()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}

func loadCerts() (string, string, error) {
	utmConfig := plugins.PluginCfg("com.utmstack", false)
	certsFolder, err := utils.MkdirJoin(utmConfig.Get("certsFolder").String())
	if err != nil {
		return "", "", fmt.Errorf("cannot create certificates directory: %v", err)
	}

	certPath := certsFolder.FileJoin(configuration.UtmCertFileName)
	keyPath := certsFolder.FileJoin(configuration.UtmCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return certPath, keyPath, nil
}
