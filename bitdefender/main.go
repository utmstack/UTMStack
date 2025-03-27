package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

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
		utils.Logger.Fatal("failed to get current path: %v", err)
	}

	certsPath := filepath.Join(path, "certs")
	err = utils.CreatePathIfNotExist(certsPath)
	if err != nil {
		utils.Logger.Fatal("error creating path: %s", err)
	}

	err = utils.GenerateCerts(certsPath)
	if err != nil {
		utils.Logger.Fatal("error generating certificates: %v", err)
	}

	server.ServerUp(&moduleConfig, certsPath)
	go configuration.ConfigureModules(&moduleConfig, mutex)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}
