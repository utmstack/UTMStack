package main

import (
	"fmt"

	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/updater"
)

func RecalculateMemory() error {
	fmt.Println("### Recalculating UTMStack memory allocation ###")

	version, err := updater.GetVersion()
	if err != nil {
		return fmt.Errorf("error getting UTMStack version: %v", err)
	}

	if err := docker.RecalculateMemory(); err != nil {
		return fmt.Errorf("error recalculating memory: %v", err)
	}

	fmt.Println("Redeploying stack with the new memory allocation...")
	if err := docker.StackUP(version.Version); err != nil {
		return fmt.Errorf("error redeploying stack: %v", err)
	}

	fmt.Println("Memory allocation recalculated and stack redeployed successfully.")
	return nil
}
