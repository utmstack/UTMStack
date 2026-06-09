package main

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/installer/branding"
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
				fmt.Printf("\nerror installing %s: %v", branding.Name(), err)
				os.Exit(1)
			}

		case "--run", "-r":
			updater.RunService()

		case "--version", "-v":
			version, err := updater.GetVersion()
			if err != nil {
				fmt.Printf("\nerror getting %s version: %v", branding.Name(), err)
				os.Exit(1)
			}

			fmt.Printf("%s version: %s, edition: %s\n", branding.Name(), version.Version, version.Edition)

		case "--uninstall", "-u":
			err := Uninstall()
			if err != nil {
				fmt.Printf("\nerror uninstalling %s: %v", branding.Name(), err)
				os.Exit(1)
			}

		default:
			help()
		}
	} else {
		err := Install()
		if err != nil {
			fmt.Printf("\nerror installing %s: %v", branding.Name(), err)
			os.Exit(1)
		}
	}
}

func help() {
	name := branding.Name()
	fmt.Printf("### %s ###\n", name)
	fmt.Println("Usage: installer <argument>")
	fmt.Println("Arguments:")
	fmt.Println("  --help, -h                            Show this help")
	fmt.Printf("  --install, -i                         Install %s\n", name)
	fmt.Printf("  --uninstall, -u                       Uninstall %s\n", name)
	fmt.Printf("  --version, -v                         Show %s version\n", name)
}
