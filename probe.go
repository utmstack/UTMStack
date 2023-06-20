package main

import (
	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func Probe(c *Config) error {
	if err := utils.CheckMem(1); err != nil {
		return err
	}

	if err := utils.CheckCPU(1); err != nil {
		return err
	}

	return nil
}
