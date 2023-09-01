package main

import (
	"fmt"
	"os"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func main() {
	fmt.Println("### UTMStack Installer ###")
	fmt.Println("Checking system requirements")
	if err := utils.CheckDistro("ubuntu"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var config = new(Config)
	err := config.Get()
	if err != nil {
		fmt.Println("creating new config file because: ", err)

		config.Branch = "v10"
		config.Password = utils.GenerateSecret(16)
		config.InternalKey = utils.GenerateSecret(32)
		config.DataDir = "/utmstack"
		config.ServerType = "aio"
	}

	mainIP, err := utils.GetMainIP()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sName, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.ServerName = sName
	config.MainServer = mainIP

	err = config.Set()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch config.ServerType {
	case "probe":
		err := Probe(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "aio":
		err := Master(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown server type, try with probe or aio")
		os.Exit(1)
	}
}
