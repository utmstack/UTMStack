package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	gosdk "github.com/threatwinds/go-sdk"
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
	mode := gosdk.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	cert, key, err := loadCerts()
	if err != nil {
		_ = gosdk.Error("cannot load certificates", err, nil)
		os.Exit(1)
	}

	server.ServerUp(&moduleConfig, cert, key)
	go configuration.ConfigureModules(&moduleConfig, mutex)

	go processor.ProcessLogs()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}

func loadCerts() (string, string, error) {
	utmConfig := gosdk.PluginCfg("com.utmstack", false)
	certsFolder := utmConfig.Get("certsFolder").String()

	certPath := filepath.Join(certsFolder, configuration.UtmCertFileName)
	keyPath := filepath.Join(certsFolder, configuration.UtmCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return certPath, keyPath, nil
}
