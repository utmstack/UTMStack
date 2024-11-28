package main

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/installer/updater"
)

func main() {
	fmt.Println("### UTMStack Installer ###")
	switch os.Args[1] {
	case "updater":
		updater.RunService()
	default:
		err := Install()
		if err != nil {
			fmt.Println("")
			fmt.Println(err)
		}
	}
}
