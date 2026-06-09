package setup

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/branding"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/network"
	"github.com/utmstack/UTMStack/installer/services"
	"github.com/utmstack/UTMStack/installer/system"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Apply(version string, updating bool) (string, error) {
	cnf := config.GetConfig()

	fmt.Println("Generating Stack configuration...")
	stack := docker.GetStackConfig()

	// Check distro (always needed for later steps)
	distro, err := system.CheckDistro()
	if err != nil {
		return "", err
	}

	if utils.GetLock(202501131200, stack.LocksDir) {
		fmt.Print("Checking system requirements")
		if err := system.CheckCPU(config.RequiredMinCPUCores); err != nil {
			return "", err
		}
		if err := system.CheckDisk(config.RequiredMinDiskSpace); err != nil {
			return "", err
		}

		// Check verifying prerequisites for AirGap mode
		if !config.ConnectedToInternet {
			fmt.Println(" [OK]")
			fmt.Print("AirGap mode detected - verifying prerequisites...")
			if err := system.VerifyAirGapPrerequisites(); err != nil {
				return "", fmt.Errorf("AirGap prerequisites check failed: %w", err)
			}
		}

		if err := utils.SetLock(202501131200, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(1, stack.LocksDir) {
		fmt.Print("Generating certificates")
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return "", err
		}
		if err := utils.SetLock(1, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(2, stack.LocksDir) {
		fmt.Printf("Preparing system to run %s", branding.Name())
		if err := system.PrepareSystem(distro); err != nil {
			return "", err
		}
		if err := utils.SetLock(2, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(202402081552, stack.LocksDir) {
		fmt.Printf("Preparing kernel to run %s", branding.Name())
		if err := system.PrepareKernel(); err != nil {
			return "", err
		}
		if err := utils.SetLock(202402081552, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(202402081553, stack.LocksDir) {
		fmt.Print("Configuring VLAN")
		iface, err := utils.GetMainIface(cnf.MainServer)
		if err != nil {
			return "", err
		}
		// Check AirGap
		if config.ConnectedToInternet {
			if err := network.InstallVlan(distro); err != nil {
				return "", err
			}
		}

		if err := network.ConfigureVLAN(iface, distro); err != nil {
			return "", err
		}
		if err := utils.SetLock(202402081553, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(3, stack.LocksDir) {
		fmt.Print("Configuring Docker")
		// Check AirGap
		if !config.ConnectedToInternet {
			fmt.Println(" [SKIPPED] (AirGap mode detected, skipping Docker installation)")
		} else {
			if err := docker.InstallDocker(distro); err != nil {
				return "", err
			}
		}
		if err := utils.SetLock(3, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(4, stack.LocksDir) {
		fmt.Print("Initializing Swarm")
		// Check AirGap
		if !config.ConnectedToInternet {
			mainIP, err := utils.GetMainIPInAirGapMode()
			if err != nil {
				return "", err
			}
			if err := docker.InitSwarm(mainIP); err != nil {
				return "", err
			}

		} else {
			mainIP, err := utils.GetMainIP()
			if err != nil {
				return "", err
			}
			if err := docker.InitSwarm(mainIP); err != nil {
				return "", err
			}
		}

		if err := utils.SetLock(4, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if installed, err := branding.InstallLicense(cnf.UpdatesFolder); err != nil {
		fmt.Printf("Warning: could not install bundled license: %v\n", err)
	} else if installed {
		fmt.Println("Bundled license installed.")
	}

	time.Sleep(10 * time.Second)

	err = docker.StackUP(version)
	if err != nil {
		return "", err
	}

	fmt.Print("Installing reverse proxy. This may take a while.")
	if config.ConnectedToInternet {
		if err := network.InstallNginx(distro); err != nil {
			return "", err
		}
	}

	if err := network.ConfigureNginx(stack, distro); err != nil {
		return "", err
	}

	fmt.Println(" [OK]")

	if utils.GetLock(5, stack.LocksDir) {
		fmt.Print("Installing Administration Tools")
		// Check AirGap
		if config.ConnectedToInternet {
			if err := system.InstallTools(distro); err != nil {
				return "", err
			}
		}

		if err := utils.SetLock(5, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(6, stack.LocksDir) {
		fmt.Printf("Initializing %s and AgentManager databases", branding.Name())
		for i := 0; i < 10; i++ {
			if err := services.InitPgUtmstack(cnf); err != nil {
				if i > 8 {
					return "", err
				}
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}

		if err := utils.SetLock(6, stack.LocksDir); err != nil {
			return "", err
		}

		fmt.Println(" [OK]")
	}

	if utils.GetLock(7, stack.LocksDir) {
		fmt.Print("Initializing OpenSearch. This may take a while.")
		if err := services.InitOpenSearch(); err != nil {
			return "", err
		}

		if err := utils.SetLock(7, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	if !updating {
		fmt.Print("Waiting for Backend to be ready. This may take a while.")

		if err := services.Backend(); err != nil {
			return "", err
		}
	}

	fmt.Println(" [OK]")

	if branding.IsCustom() && utils.GetLock(202606091200, stack.LocksDir) {
		fmt.Printf("Applying %s branding", branding.Name())
		if err := branding.Seed("https://127.0.0.1", cnf.InternalKey); err != nil {
			fmt.Printf(" [WARNING]\nerror seeding branding: %v\n", err)
		} else {
			if err := utils.SetLock(202606091200, stack.LocksDir); err != nil {
				return "", err
			}
			fmt.Println(" [OK]")
		}
	}

	if utils.GetLock(9, stack.LocksDir) {
		fmt.Print("Generating Base URL")
		if err := services.SetBaseURL(cnf.ServerName); err != nil {
			return "", err
		}

		if err := utils.SetLock(9, stack.LocksDir); err != nil {
			return "", err
		}
		fmt.Println(" [OK]")
	}

	// if utils.GetLock(10, stack.LocksDir) {
	// 	fmt.Print("Sending sample logs")
	// 	if err := SendSampleData(); err != nil {
	// 		fmt.Printf("error sending sample data: %v", err)
	// 	}

	// 	if err := utils.SetLock(10, stack.LocksDir); err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(" [OK]")
	// }

	return cnf.Password, nil
}
