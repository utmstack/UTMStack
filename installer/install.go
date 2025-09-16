package main

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/network"
	"github.com/utmstack/UTMStack/installer/services"
	"github.com/utmstack/UTMStack/installer/system"
	"github.com/utmstack/UTMStack/installer/updater"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Install() error {
	fmt.Println("### Installing UTMStack ###")
	go updater.MonitorConnection(config.GetCMServer(), 30*time.Second, 3, &config.ConnectedToInternet)

	isInstalledAlready, err := utils.CheckIfServiceIsInstalled("UTMStackComponentsUpdater")
	if err != nil {
		return fmt.Errorf("error checking if service is installed: %v", err)
	}

	if isInstalledAlready {
		fmt.Println("UTMStack is already installed. If you want to re-install it, please remove the service UTMStackComponentsUpdater first.")
		return nil
	}

	cnf := config.GetConfig()

	fmt.Print("Checking system requirements")
	distro, err := system.CheckDistro()
	if err != nil {
		return err
	}
	if err := system.CheckCPU(config.RequiredMinCPUCores); err != nil {
		return err
	}
	if err := system.CheckDisk(config.RequiredMinDiskSpace); err != nil {
		return err
	}
	fmt.Println(" [OK]")

	// Check verifying prerequisites for AirGap mode
	if !config.ConnectedToInternet {
		fmt.Println("AirGap mode detected - verifying prerequisites...")
		if err := system.VerifyAirGapPrerequisites(); err != nil {
			return fmt.Errorf("AirGap prerequisites check failed: %w", err)
		}
		fmt.Println(" [OK]")
	}

	fmt.Println("Generating Stack configuration...")
	stack := docker.GetStackConfig()

	if utils.GetLock(1, stack.LocksDir) {
		fmt.Print("Generating certificates")
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return err
		}
		if err := utils.SetLock(1, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(2, stack.LocksDir) {
		fmt.Print("Preparing system to run UTMStack")
		if err := system.PrepareSystem(distro); err != nil {
			return err
		}
		if err := utils.SetLock(2, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(202402081552, stack.LocksDir) {
		fmt.Print("Preparing kernel to run UTMStack")
		if err := system.PrepareKernel(); err != nil {
			return err
		}
		if err := utils.SetLock(202402081552, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(202402081553, stack.LocksDir) {
		fmt.Print("Configuring VLAN")
		iface, err := utils.GetMainIface(cnf.MainServer)
		if err != nil {
			return err
		}
		// Check AirGap
		if config.ConnectedToInternet {
			if err := network.InstallVlan(distro); err != nil {
				return err
			}
		}

		if err := network.ConfigureVLAN(iface, distro); err != nil {
			return err
		}
		if err := utils.SetLock(202402081553, stack.LocksDir); err != nil {
			return err
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
				return err
			}
		}
		if err := utils.SetLock(3, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(4, stack.LocksDir) {
		fmt.Print("Initializing Swarm")
		// Check AirGap
		if !config.ConnectedToInternet {
			mainIP, err := utils.GetMainIPInAirGapMode()
			if err != nil {
				return err
			}
			if err := docker.InitSwarm(mainIP); err != nil {
				return err
			}

		} else {
			mainIP, err := utils.GetMainIP()
			if err != nil {
				return err
			}
			if err := docker.InitSwarm(mainIP); err != nil {
				return err
			}
		}

		if err := utils.SetLock(4, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if !utils.GetLock(5, stack.LocksDir) && utils.GetLock(202407051241, stack.LocksDir) {
		fmt.Print("Removing old services")
		if err := docker.RemoveServices([]string{
			"utmstack_aws",
			"utmstack_bitdefender",
			"utmstack_correlation",
			"utmstack_filebrowser",
			"utmstack_log-auth-proxy",
			"utmstack_logstash",
			"utmstack_mutate",
			"utmstack_office365",
			"utmstack_sophos",
			"utmstack_socai",
		}); err != nil {
			return err
		}
		if err := utils.SetLock(202407051241, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	time.Sleep(10 * time.Second)
	fmt.Print("Getting UTMStack Latest Version")
	version, err := updater.GetVersion()
	if err != nil {
		return err
	}
	fmt.Println(" [OK]")

	fmt.Printf("Installing UTMStack version %s-%s. This may take a while.\n", version.Version, version.Edition)
	err = docker.StackUP(version.Version + "-" + version.Edition)
	if err != nil {
		return err
	}

	fmt.Print("Installing reverse proxy. This may take a while.")
	if config.ConnectedToInternet {
		if err := network.InstallNginx(distro); err != nil {
			return err
		}
	}

	if err := network.ConfigureNginx(cnf, stack, distro); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	if utils.GetLock(5, stack.LocksDir) {
		fmt.Print("Installing Administration Tools")
		// Check AirGap
		if config.ConnectedToInternet {
			if err := system.InstallTools(distro); err != nil {
				return err
			}
		}

		if err := utils.SetLock(5, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(6, stack.LocksDir) {
		fmt.Print("Initializing UTMStack and AgentManager databases")
		for i := 0; i < 10; i++ {
			if err := services.InitPgUtmstack(cnf); err != nil {
				if i > 8 {
					return err
				}
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}

		if err := utils.SetLock(6, stack.LocksDir); err != nil {
			return err
		}

		fmt.Println(" [OK]")
	}

	if utils.GetLock(202311301747, stack.LocksDir) {
		fmt.Print("Initializing User Auditor database")
		for i := 0; i < 10; i++ {
			if err := services.InitPgUserAuditor(cnf); err != nil {
				if i > 8 {
					return err
				}
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}

		if err := utils.SetLock(202311301747, stack.LocksDir); err != nil {
			return err
		}

		fmt.Println(" [OK]")
	}

	if utils.GetLock(7, stack.LocksDir) {
		fmt.Print("Initializing OpenSearch. This may take a while.")
		if err := services.InitOpenSearch(); err != nil {
			return err
		}

		if err := utils.SetLock(7, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	fmt.Print("Waiting for Backend to be ready. This may take a while.")

	if err := services.Backend(); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	if utils.GetLock(8, stack.LocksDir) {
		fmt.Print("Generating Connection Key")
		if err := services.RegenerateKey(cnf.InternalKey); err != nil {
			return err
		}

		if err := utils.SetLock(8, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(9, stack.LocksDir) {
		fmt.Print("Generating Base URL")
		if err := services.SetBaseURL(cnf.Password, cnf.ServerName); err != nil {
			return err
		}

		if err := utils.SetLock(9, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	fmt.Print("Installing Updater Service")
	updater.InstallService()
	fmt.Println(" [OK]")

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

	fmt.Println("Running post installation scripts. This may take a while.")

	if err := docker.PostInstallation(); err != nil {
		return err
	}

	fmt.Println("Installation fisnished successfully. We have generated a configuration file for you, please do not modify or remove it. You can find it at /root/utmstack.yml.")
	fmt.Println("You can also use it to re-install your stack in case of a disaster or changes in your hardware. Just run the installer again.")
	fmt.Println("You can access to your Web-GUI at https://<your-server-ip> using admin as your username")
	fmt.Printf("Web-GUI default password for admin: %s \n", cnf.Password)
	fmt.Println("You can also access to your Web-based Administration Interface at https://<your-server-ip>:9090 using your Linux system credentials.")
	fmt.Println("Detailed installation logs can be found at /var/log/utmstack-installer.log")

	fmt.Println("### Thanks for using UTMStack ###")

	return nil
}
