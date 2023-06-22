package main

import (
	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func Master(c *Config) error {
	if err := utils.CheckMem(8); err != nil {
		return err
	}

	if err := utils.CheckCPU(4); err != nil {
		return err
	}

	if err := utils.CheckDisk(256); err != nil {
		return err
	}

	var stack = new(StackConfig)
	if err := stack.Populate(c); err != nil {
		return err
	}

	if utils.GetStep() < 1 {
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return err
		}

		if err := utils.SetStep(1); err != nil {
			return err
		}
	}

	if utils.GetStep() < 2 {
		if err := PrepareSystem(); err != nil {
			return err
		}

		if err := utils.SetStep(2); err != nil {
			return err
		}
	}

	if utils.GetStep() < 3 {
		if err := InstallDocker(); err != nil {
			return err
		}

		if err := utils.SetStep(3); err != nil {
			return err
		}
	}

	if utils.GetStep() < 4 {
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
	}

	if err := StackUP(c, stack); err != nil {
		return err
	}

	if utils.GetStep() < 6 {
		if err := InstallTools(); err != nil {
			return err
		}

		if err := utils.SetStep(6); err != nil {
			return err
		}
	}

	if utils.GetStep() < 7 {
		if err := InitOpenSearch(); err != nil {
			return err
		}

		if err := utils.SetStep(7); err != nil {
			return err
		}
	}

	if utils.GetStep() < 8 {
		if err := InitPostgres(c); err != nil {
			return err
		}

		if err := utils.SetStep(8); err != nil {
			return err
		}
	}

	if err := Backend(); err != nil {
		return err
	}

	return nil
}
