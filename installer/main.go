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
			if len(os.Args) < 3 {
				fmt.Println("\nPlease provide organization ID")
				os.Exit(1)
			}

			err := Install(os.Args[2])
			if err != nil {
				fmt.Printf("\nerror installing UTMStack: %v", err)
				os.Exit(1)
			}

		case "--run", "-r":
			updater.RunService()

		case "--version", "-v":
			version := updater.GetVersion()
			if version.Version == "" || version.Edition == "" {
				fmt.Println("UTMStack version not found")
				os.Exit(1)
			}
			v := fmt.Sprintf("%s-%s", version.Version, version.Edition)
			fmt.Println("UTMStack version:", v)

		default:
			help()
		}
	} else {
		help()
	}
}

func help() {
	fmt.Println("### UTMStack ###")
	fmt.Println("Usage: utmstack <argument>")
	fmt.Println("Arguments:")
	fmt.Println("  --help, -h              Show this help")
	fmt.Println("  --install, -i <ID>      Install UTMStack (Requires organization ID)")
	fmt.Println("  --version, -v           Show UTMStack version")
}
