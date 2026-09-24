package service

import (
	"fmt"
	"os"

	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/kardianos/service"
)

func InstallService() {
	// Ensure the quarantine store exists and is root-only (0700) before the
	// service starts — it holds live malware and must not be readable by
	// ordinary users. MkdirAll only applies the mode to freshly-created dirs,
	// so chmod it explicitly to also correct a pre-existing dir with loose
	// perms (a prior install or a manual mkdir).
	if err := os.MkdirAll(edrconfig.QuarantineDir, 0o700); err != nil {
		fmt.Println("Error creating quarantine directory:", err)
		os.Exit(1)
	}
	if err := os.Chmod(edrconfig.QuarantineDir, 0o700); err != nil {
		fmt.Println("Error setting quarantine directory permissions:", err)
		os.Exit(1)
	}

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
	// On Linux, kardianos/service Install() already runs `systemctl enable`
	// (service_systemd_linux.go), so the service survives a reboot.
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
