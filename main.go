package main

import (
	"fmt"
	"os"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func main() {
	config := Config{}
	err := config.Get()
	if err != nil {
		fmt.Println("creating new config file because of: ", err)
		config.Branch = "v10"
		config.Password = utils.GenerateSecret(32)
		config.MainServer = ""
		config.DataDir = "/utmstack"
		err = config.Set()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := utils.CheckDistro("ubuntu"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch config.MainServer {
	case "":
		err := Master(&config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		err := Probe(&config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
