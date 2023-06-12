package main

import (
	"fmt"
	"os"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
	"github.com/akamensky/argparse"
)

var tag string

func main() {
	tag = os.Getenv("TAG")
	if tag == "" {
		tag = "v9"
	}

	utils.CheckErr(utils.CheckDistro("ubuntu"))

	parser := argparse.NewParser("", "UTMStack installer")
	removeCmd := parser.NewCommand("remove", "Uninstall UTMStack")

	masterCmd := parser.NewCommand("master", "Install Master")
	masterDataDir := "/utmstack"
	masterPass := masterCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Master password. Please use a secure password"})

	probeCmd := parser.NewCommand("probe", "Install Probe")
	probeDataDir := "/utmstack"
	probePass := probeCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Master password. Generated when the master was installed"})
	host := probeCmd.String("", "host", &argparse.Options{Required: true, Help: "Master server IP or FQDN"})

	if len(os.Args) == 1 {
		tui()
	} else if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
	} else if removeCmd.Happened() {
		utils.CheckErr(utils.Uninstall("cli"))
	} else if masterCmd.Happened() {
		utils.CheckErr(utils.InstallMaster("cli", masterDataDir, *masterPass, tag))
		fmt.Println("Successfully installed.")
	} else if probeCmd.Happened() {
		utils.CheckErr(utils.InstallProbe("cli", probeDataDir, *probePass, *host, tag))
		fmt.Println("Successfully installed.")
	}
}
