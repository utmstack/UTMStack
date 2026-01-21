package main

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/installer/updater"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {

		case "--help", "-h":
			help()

		case "--install", "-i":
			err := Install()
			if err != nil {
				fmt.Printf("\nerror installing UTMStack: %v", err)
				os.Exit(1)
			}

		case "--run", "-r":
			updater.RunService()

		case "--version", "-v":
			version, err := updater.GetVersion()
			if err != nil {
				fmt.Printf("\nerror getting UTMStack version: %v", err)
				os.Exit(1)
			}

			fmt.Printf("UTMStack version: %s, edition: %s\n", version.Version, version.Edition)

		case "--uninstall", "-u":
			err := Uninstall()
			if err != nil {
				fmt.Printf("\nerror uninstalling UTMStack: %v", err)
				os.Exit(1)
			}

		default:
			help()
		}
	} else {
		err := Install()
		if err != nil {
			fmt.Printf("\nerror installing UTMStack: %v", err)
			os.Exit(1)
		}
	}
}

func help() {
	fmt.Println("### UTMStack ###")
	fmt.Println("Usage: installer <argument>")
	fmt.Println("Arguments:")
	fmt.Println("  --help, -h                            Show this help")
	fmt.Println("  --install, -i                         Install UTMStack")
	fmt.Println("  --uninstall, -u                       Uninstall UTMStack")
	fmt.Println("  --version, -v                         Show UTMStack version")
}
