package main

import (
	"fmt"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func Master(c *Config) error {
	if err := utils.CheckMem(8); err != nil {
		return err
	}

	if err := utils.CheckCPU(4); err != nil {
		return err
	}

	if err := utils.CheckDisk(200); err != nil {
		return err
	}

	fmt.Println("Checking system requirements [OK]")

	fmt.Println("Generating Stack configuration")

	var stack = new(StackConfig)
	if err := stack.Populate(c); err != nil {
		return err
	}

	fmt.Println("Generating Stack configuration [OK]")

	if utils.GetStep() < 1 {
		fmt.Println("Generating certificates")
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return err
		}

		if err := utils.SetStep(1); err != nil {
			return err
		}
		fmt.Println("Generating certificates [OK]")
	}

	if utils.GetStep() < 2 {
		fmt.Println("Preparing system to run UTMStack")
		if err := PrepareSystem(); err != nil {
			return err
		}

		if err := utils.SetStep(2); err != nil {
			return err
		}
		fmt.Println("Preparing system to run UTMStack [OK]")
	}

	if utils.GetStep() < 3 {
		fmt.Println("Installing Docker")
		if err := InstallDocker(); err != nil {
			return err
		}

		if err := utils.SetStep(3); err != nil {
			return err
		}
		fmt.Println("Installing Docker [OK]")
	}

	if utils.GetStep() < 4 {
		fmt.Println("Initializing Swarm")
		mainIP, err := utils.GetMainIP()
		if err != nil {
			return err
		}

		if err := InitSwarm(mainIP); err != nil {
			return err
		}

		if err := utils.SetStep(4); err != nil {
			return err
		}
		fmt.Println("Initializing Swarm [OK]")
	}

	fmt.Println("Installing Stack. This may take a while")

	if err := StackUP(c, stack); err != nil {
		return err
	}

	fmt.Println("Installing Stack [OK]")

	if utils.GetStep() < 6 {
		fmt.Println("Installing Administration Tools")
		if err := InstallTools(); err != nil {
			return err
		}

		if err := utils.SetStep(6); err != nil {
			return err
		}
		fmt.Println("Installing Administration Tools [OK]")
	}

	if utils.GetStep() < 7 {
		fmt.Println("Initializing OpenSearch. This may take a while")
		if err := InitOpenSearch(); err != nil {
			return err
		}

		if err := utils.SetStep(7); err != nil {
			return err
		}
		fmt.Println("Initializing OpenSearch [OK]")
	}

	if utils.GetStep() < 8 {
		fmt.Println("Initializing PostgreSQL")
		if err := InitPostgres(c); err != nil {
			return err
		}

		if err := utils.SetStep(8); err != nil {
			return err
		}
		fmt.Println("Initializing PostgreSQL [OK]")
	}

	fmt.Println("Initializing Web-GUI. This may take a while")

	if err := Backend(); err != nil {
		return err
	}

	fmt.Println("Initializing Web-GUI [OK]")

	fmt.Println("Installation fisnished successfully. We have generated a configuration file for you, please do not modify or remove it. You can find it at /root/utmstack.yml.")
	fmt.Println("You can also use it to re-install your stack in case of a disaster or changes in your hardware. Just run the installer again.")
	fmt.Println("You can access to your Web-GUI at https://<your-server-ip>:443 using admin as your username and the password in the configuration file.")
	fmt.Println("You can also access to your Web-based Administration Interface at https://<your-server-ip>:9090 using your Linux system credentials.")

	fmt.Println("### Thanks for using UTMStack ###")

	return nil
}
