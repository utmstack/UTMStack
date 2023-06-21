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

	if utils.GetVersion() < 10 {
		if err := utils.GenerateCerts(stack.Cert); err != nil {
			return err
		}

		if err := InstallDocker(); err != nil {
			return err
		}

		mainIP, err := utils.GetMainIP()
		if err != nil {
			return err
		}

		if err := InitSwarm(mainIP); err != nil {
			return err
		}
	}

	if err := StackUP(c, stack); err != nil {
		return err
	}

	if utils.GetVersion() < 10 {
		if err := InitPostgres(c); err != nil {
			return err
		}

		if err := InitOpenSearch(); err != nil {
			return err
		}
	}

	if err := Backend(); err != nil {
		return err
	}

	if err := utils.SetVersion(10); err != nil {
		return err
	}

	return nil
}
