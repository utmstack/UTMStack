package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/bitdefender/configuration"
	"github.com/utmstack/UTMStack/bitdefender/server"
	"github.com/utmstack/UTMStack/bitdefender/utils"
	"github.com/utmstack/config-client-go/types"
)

var (
	mutex        = &sync.Mutex{}
	moduleConfig = types.ConfigurationSection{}
)

func main() {
	path, err := utils.GetMyPath()
	if err != nil {
		catcher.Error("failed to get current path", err, nil)
		os.Exit(1)
	}

	certsPath := filepath.Join(path, "certs")
	err = utils.CreatePathIfNotExist(certsPath)
	if err != nil {
		catcher.Error("error creating path", err, nil)
		os.Exit(1)
	}

	err = utils.GenerateCerts(certsPath)
	if err != nil {
		catcher.Error("error generating certificates", err, nil)
		os.Exit(1)
	}

	server.ServerUp(&moduleConfig, certsPath)
	go configuration.ConfigureModules(&moduleConfig, mutex)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}
