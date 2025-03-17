package serv

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

func InstallService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println("\nError creating new service: ", err)
		os.Exit(1)
	}
	err = newService.Install()
	if err != nil {
		fmt.Println("\nError installing new service: ", err)
		os.Exit(1)
	}

	err = newService.Start()
	if err != nil {
		fmt.Println("\nError starting new service: ", err)
		os.Exit(1)
	}
}
