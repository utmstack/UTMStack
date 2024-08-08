package main

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

func Master(c *types.Config) error {
	if err := utils.CheckCPU(2); err != nil {
		return err
	}

	if err := utils.CheckDisk(30); err != nil {
		return err
	}

	fmt.Println("Generating Stack configuration")

	var stack = new(types.StackConfig)
	if err := stack.Populate(c); err != nil {
		return err
	}

	fmt.Println("Checking system requirements [OK]")

	fmt.Println("Generating Stack configuration [OK]")

	if utils.GetLock(1, stack.LocksDir) {
		fmt.Println("Generating certificates")
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return err
		}

		if err := utils.SetLock(1, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Generating certificates [OK]")
	}

	if utils.GetLock(2, stack.LocksDir) {
		fmt.Println("Preparing system to run UTMStack")
		if err := PrepareSystem(); err != nil {
			return err
		}

		if err := utils.SetLock(2, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Preparing system to run UTMStack [OK]")
	}

	if utils.GetLock(202402081552, stack.LocksDir) {
		fmt.Println("Preparing kernel to run UTMStack")
		if err := PrepareSystem(); err != nil {
			return err
		}

		if err := utils.SetLock(202402081552, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Preparing kernel to run UTMStack [OK]")
	}

	if utils.GetLock(202402081553, stack.LocksDir) {
		fmt.Println("Configuring VLAN")
		iface, err := utils.GetMainIface(c.MainServer)
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
		fmt.Println("Configuring VLAN [OK]")
	}

	if utils.GetLock(3, stack.LocksDir) {
		fmt.Println("Installing Docker")
		if err := InstallDocker(); err != nil {
			return err
		}

		if err := utils.SetLock(3, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Installing Docker [OK]")
	}

	if utils.GetLock(4, stack.LocksDir) {
		fmt.Println("Initializing Swarm")
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
		fmt.Println("Initializing Swarm [OK]")
	}

	fmt.Println("Downloading Plugins and Base Configurations")

	if err := Downloads(c, stack); err != nil {
		return err
	}

	fmt.Println("Downloading Plugins and Base Configurations [OK]")

	if !utils.GetLock(5, stack.LocksDir) && utils.GetLock(202407051241, stack.LocksDir) {
		fmt.Println("Removing old services")

		if err := utils.RemoveServices([]string{
			"utmstack_log-auth-proxy",
			"utmstack_mutate",
			"utmstack_filebrowser",
			"utmstack_correlation",
			"utmstack_logstash",
		}); err != nil {
			return err
		}

		if err := utils.SetLock(202407051241, stack.LocksDir); err != nil {
			return err
		}

		fmt.Println("Removing old services [OK]")
	}

	fmt.Println("Installing Stack. This may take a while.")

	if err := StackUP(c, stack); err != nil {
		return err
	}

	fmt.Println("Installing Stack [OK]")

	fmt.Println("Installing reverse proxy. This may take a while.")

	if err := InstallNginx(); err != nil {
		return err
	}

	if err := ConfigureNginx(c, stack); err != nil {
		return err
	}

	fmt.Println("Installing reverse proxy [OK]")

	if utils.GetLock(5, stack.LocksDir) {
		fmt.Println("Installing Administration Tools")
		if err := InstallTools(); err != nil {
			return err
		}

		if err := utils.SetLock(5, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Installing Administration Tools [OK]")
	}

	if utils.GetLock(6, stack.LocksDir) {
		fmt.Println("Initializing UTMStack and AgentManager databases")
		for i := 0; i < 10; i++ {
			if err := InitPgUtmstack(c); err != nil {
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

		fmt.Println("Initializing UTMStack and AgentManager databases [OK]")
	}

	if utils.GetLock(202311301747, stack.LocksDir) {
		fmt.Println("Initializing User Auditor database")
		for i := 0; i < 10; i++ {
			if err := InitPgUserAuditor(c); err != nil {
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

		fmt.Println("Initializing User Auditor database [OK]")
	}

	if utils.GetLock(7, stack.LocksDir) {
		fmt.Println("Initializing OpenSearch. This may take a while.")
		if err := InitOpenSearch(); err != nil {
			return err
		}

		if err := utils.SetLock(7, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Initializing OpenSearch [OK]")
	}

	fmt.Println("Waiting for Backend to be ready. This may take a while.")

	if err := Backend(); err != nil {
		return err
	}

	fmt.Println("Waiting for Backend to be ready [OK]")

	if utils.GetLock(8, stack.LocksDir) {
		fmt.Println("Generating Connection Key")
		if err := RegenerateKey(c.InternalKey); err != nil {
			return err
		}

		if err := utils.SetLock(8, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Generating Connection Key [OK]")
	}

	if utils.GetLock(9, stack.LocksDir) {
		fmt.Println("Generating Base URL")
		if err := SetBaseURL(c.Password, c.ServerName); err != nil {
			return err
		}

		if err := utils.SetLock(9, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Generating Base URL [OK]")
	}

	if utils.GetLock(10, stack.LocksDir) {
		fmt.Println("Sending sample logs")
		if err := SendSampleData(); err != nil {
			return err
		}

		if err := utils.SetLock(10, stack.LocksDir); err != nil {
			return err
		}
		fmt.Println("Sending sample logs [OK]")
	}

	fmt.Println("Running post installation scripts. This may take a while.")

	if err := PostInstallation(); err != nil {
		return err
	}

	fmt.Println("Running post installation scripts [OK]")

	fmt.Println("Installation fisnished successfully. We have generated a configuration file for you, please do not modify or remove it. You can find it at /root/utmstack.yml.")
	fmt.Println("You can also use it to re-install your stack in case of a disaster or changes in your hardware. Just run the installer again.")
	fmt.Println("You can access to your Web-GUI at https://<your-server-ip> using admin as your username")
	fmt.Printf("Web-GUI default password for admin: %s \n", c.Password)
	fmt.Println("You can also access to your Web-based Administration Interface at https://<your-server-ip>:9090 using your Linux system credentials.")

	fmt.Println("### Thanks for using UTMStack ###")

	return nil
}
