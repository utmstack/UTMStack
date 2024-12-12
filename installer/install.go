package main

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/updater"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Install() error {
	cnf := config.GetConfig()

	fmt.Print("Checking system requirements")
	if err := utils.CheckDistro(config.RequiredDistro); err != nil {
		return err
	}
	if err := utils.CheckCPU(config.RequiredMinCPUCores); err != nil {
		return err
	}
	if err := utils.CheckDisk(config.RequiredMinDiskSpace); err != nil {
		return err
	}
	fmt.Println("[OK]")

	fmt.Print("Generating Stack configuration")
	stack := config.GetStackConfig()
	fmt.Println(" [OK]")

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
		if err := PrepareSystem(); err != nil {
			return err
		}
		if err := utils.SetLock(2, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(202402081552, stack.LocksDir) {
		fmt.Print("Preparing kernel to run UTMStack")
		if err := PrepareSystem(); err != nil {
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
		if err := InstallVlan(); err != nil {
			return err
		}
		if err := ConfigureVLAN(iface); err != nil {
			return err
		}
		if err := utils.SetLock(202402081553, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(3, stack.LocksDir) {
		fmt.Print("Installing Docker")
		if err := InstallDocker(); err != nil {
			return err
		}
		if err := utils.SetLock(3, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(4, stack.LocksDir) {
		fmt.Print("Initializing Swarm")
		mainIP, err := utils.GetMainIP()
		if err != nil {
			return err
		}
		if err := InitSwarm(mainIP); err != nil {
			return err
		}
		if err := utils.SetLock(4, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if !utils.GetLock(5, stack.LocksDir) && utils.GetLock(202407051241, stack.LocksDir) {
		fmt.Print("Removing old services")
		if err := utils.RemoveServices([]string{
			"utmstack_aws",
			"utmstack_bitdefender",
			"utmstack_correlation",
			"utmstack_filebrowser",
			"utmstack_log-auth-proxy",
			"utmstack_logstash",
			"utmstack_mutate",
			"utmstack_office365",
			"utmstack_sophos",
		}); err != nil {
			return err
		}
		if err := utils.SetLock(202407051241, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	fmt.Println("Downloading Dependencies and Base Configurations:")
	if err := updater.GetUpdaterClient().CheckUpdate(true, false); err != nil {
		return err
	}

	fmt.Println("Installing Stack. This may take a while.")

	if err := StackUP(cnf, stack); err != nil {
		return err
	}

	fmt.Print("Installing reverse proxy. This may take a while.")

	if err := InstallNginx(); err != nil {
		return err
	}

	if err := ConfigureNginx(cnf, stack); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	if utils.GetLock(5, stack.LocksDir) {
		fmt.Print("Installing Administration Tools")
		if err := InstallTools(); err != nil {
			return err
		}

		if err := utils.SetLock(5, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(6, stack.LocksDir) {
		fmt.Print("Initializing UTMStack and AgentManager databases")
		for i := 0; i < 10; i++ {
			if err := InitPgUtmstack(cnf); err != nil {
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
			if err := InitPgUserAuditor(cnf); err != nil {
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
		if err := InitOpenSearch(); err != nil {
			return err
		}

		if err := utils.SetLock(7, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	fmt.Print("Waiting for Backend to be ready. This may take a while.")

	if err := Backend(); err != nil {
		return err
	}

	fmt.Println(" [OK]")

	if utils.GetLock(8, stack.LocksDir) {
		fmt.Print("Generating Connection Key")
		if err := RegenerateKey(cnf.InternalKey); err != nil {
			return err
		}

		if err := utils.SetLock(8, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(9, stack.LocksDir) {
		fmt.Print("Generating Base URL")
		if err := SetBaseURL(cnf.Password, cnf.ServerName); err != nil {
			return err
		}

		if err := utils.SetLock(9, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	if utils.GetLock(10, stack.LocksDir) {
		fmt.Print("Sending sample logs")
		if err := SendSampleData(); err != nil {
			fmt.Printf("error sending sample data: %v", err)
		}

		if err := utils.SetLock(10, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println(" [OK]")
	}

	fmt.Println("Running post installation scripts. This may take a while.")

	if err := PostInstallation(); err != nil {
		return err
	}

	fmt.Println("Installing Updater Service")
	updater.InstallService()

	fmt.Println("Installation fisnished successfully. We have generated a configuration file for you, please do not modify or remove it. You can find it at /root/utmstack.yml.")
	fmt.Println("You can also use it to re-install your stack in case of a disaster or changes in your hardware. Just run the installer again.")
	fmt.Println("You can access to your Web-GUI at https://<your-server-ip> using admin as your username")
	fmt.Printf("Web-GUI default password for admin: %s \n", cnf.Password)
	fmt.Println("You can also access to your Web-based Administration Interface at https://<your-server-ip>:9090 using your Linux system credentials.")
	fmt.Println("Detailed installation logs can be found at /var/log/utmstack-installer.log")

	fmt.Println("### Thanks for using UTMStack ###")

	return nil
}
