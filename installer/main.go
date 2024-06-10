package main

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/installer/utils"
	"github.com/utmstack/UTMStack/installer/types"
)

func main() {
	fmt.Println("### UTMStack Installer ###")
	fmt.Println("Checking system requirements")

	args := os.Args

	var update bool

	for _, arg := range args {
		if arg == "update" {
			update = true
		}
	}

	if err := utils.CheckDistro("ubuntu"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var config = new(types.Config)
	config.Get()

	mainIP, err := utils.GetMainIP()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.MainServer = mainIP

	sName, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.ServerName = sName

	if config.ServerType != "aio" &&
		config.ServerType != "cloud" {
		config.ServerType = "aio"
	}

	if config.Password == "" {
		config.Password = utils.GenerateSecret(16)
	}

	if config.InternalKey == "" {
		config.InternalKey = utils.GenerateSecret(32)
	}

	if config.Branch != "v10-dev" &&
		config.Branch != "v10-qa" &&
		config.Branch != "v10-rc" {
		config.Branch = "v10"
	}

	if config.DataDir == "" {
		config.DataDir = "/utmstack"
	}

	err = config.Set()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch config.ServerType {
	case "aio":
		err := Master(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "cloud":
		err := Cloud(config, update)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println(`unknown server type, try with "aio" or "cloud"`)
		os.Exit(1)
	}
}
