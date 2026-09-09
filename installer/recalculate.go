package main

import (
	"fmt"

	"github.com/utmstack/UTMStack/installer/branding"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/updater"
)

func RecalculateMemory() error {
	fmt.Printf("### Recalculating %s memory allocation ###\n", branding.Name())

	version, err := updater.GetVersion()
	if err != nil {
		return err
	}

	fmt.Print("Balancing memory against current system resources")
	if err := docker.RecalculateMemory(); err != nil {
		return fmt.Errorf("error recalculating memory: %v", err)
	}
	fmt.Println(" [OK]")

	fmt.Print("Redeploying stack with new memory limits. This may take a while.")
	if err := docker.StackUP(version.Version); err != nil {
		return fmt.Errorf("error redeploying stack: %v", err)
	}
	fmt.Println(" [OK]")

	fmt.Println("Memory allocation recalculated and applied successfully.")
	return nil
}
