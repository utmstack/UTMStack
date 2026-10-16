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

func ApplyMemory(dryRun bool) error {
	fmt.Printf("### Applying %s memory allocation ###\n", branding.Name())

	stack := docker.GetStackConfig()
	rsrcs, err := stack.ApplyMemory()
	if err != nil {
		return err
	}

	fmt.Printf("  %-22s %6s\n", "Service", "Memory(MB)")
	for _, svc := range docker.Services {
		fmt.Printf("  %-22s %6d\n", svc.Name, rsrcs[svc.Name].AssignedMemory)
	}
	for name := range rsrcs {
		known := false
		for _, svc := range docker.Services {
			if svc.Name == name {
				known = true
				break
			}
		}
		if !known {
			fmt.Printf("  warning: unknown service %q in memory allocation, ignored\n", name)
		}
	}

	if dryRun {
		fmt.Println("\nDry run: stack was not redeployed. Re-run without --dry-run to apply.")
		return nil
	}

	version, err := updater.GetVersion()
	if err != nil {
		return err
	}

	fmt.Print("\nRedeploying stack with the new memory limits. This may take a while.")
	if err := docker.StackUP(version.Version); err != nil {
		return fmt.Errorf("error redeploying stack: %v", err)
	}
	fmt.Println(" [OK]")

	fmt.Println("Memory allocation applied and stack redeployed successfully.")
	return nil
}
