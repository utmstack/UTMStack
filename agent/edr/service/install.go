package service

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

func InstallService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}
	if err := s.Install(); err != nil {
		fmt.Println("Error installing service:", err)
		os.Exit(1)
	}
}

func UninstallService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}
	_ = s.Stop()
	if err := s.Uninstall(); err != nil {
		fmt.Println("Error uninstalling service:", err)
		os.Exit(1)
	}
}
