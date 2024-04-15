package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/utmstack/UTMStack/bdgz/configuration"
	"github.com/utmstack/UTMStack/bdgz/server"
	"github.com/utmstack/UTMStack/bdgz/utils"
	"github.com/utmstack/config-client-go/types"
)

var (
	mutex        = &sync.Mutex{}
	moduleConfig = types.ConfigurationSection{}
)

func main() {
	h := utils.GetLogger()
	path, err := utils.GetMyPath()
	if err != nil {
		h.Fatal("failed to get current path: %v", err)
	}

	// Generate Certificates
	certsPath := filepath.Join(path, "certs")
	err = utils.CreatePathIfNotExist(certsPath)
	if err != nil {
		h.Fatal("error creating path: %s", err)
	}

	err = utils.GenerateCerts(certsPath)
	if err != nil {
		h.Fatal("error generating certificates: %v", err)
	}

	server.ServerUp(&moduleConfig, certsPath, h)
	go configuration.ConfigureModules(&moduleConfig, mutex, h)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}
