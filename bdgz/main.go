package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/bdgz/configuration"
	"github.com/utmstack/UTMStack/bdgz/server"
	"github.com/utmstack/UTMStack/bdgz/utils"
	"github.com/utmstack/config-client-go/types"
)

var (
	h            = holmes.New("debug", "BDGZ_Integration")
	mutex        = &sync.Mutex{}
	moduleConfig = types.ConfigurationSection{}
)

func main() {
	path, err := utils.GetMyPath()
	if err != nil {
		h.FatalError("failed to get current path: %v", err)
	}

	// Generate Certificates
	certsPath := filepath.Join(path, "certs")
	err = utils.CreatePathIfNotExist(certsPath)
	if err != nil {
		h.FatalError("error creating path: %s", err)
	}

	err = utils.GenerateCerts(certsPath)
	if err != nil {
		h.FatalError("error generating certificates: %v", err)
	}

	server.ServerUp(&moduleConfig, certsPath, h)
	go configuration.ConfigureModules(&moduleConfig, mutex, h)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	<-sigCh
}
