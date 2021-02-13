package main

import (
	"fmt"
	"os"

	"github.com/UTMStack/installer/utils"
	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("", "UTMStack installer")
	removeCmd := parser.NewCommand("remove", "Uninstall UTMStack")

	masterCmd := parser.NewCommand("master", "Install master server")
	masterDataDir := masterCmd.String("", "datadir", &argparse.Options{Required: true, Help: "Data directory"})
	masterPass := masterCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Password to access the database. Please use a secure password"})
	fqdn := masterCmd.String("", "fqdn", &argparse.Options{Required: true, Help: "Full qualified domain name, example: utmmaster.utmstack.com"})
	customerName := masterCmd.String("", "customer-name", &argparse.Options{Required: true, Help: "Your name, example: John Doe"})
	customerEmail := masterCmd.String("", "customer-email", &argparse.Options{Required: true, Help: "A valid email address to send important notifications about the system health. Example: john@doe.com"})

	probeCmd := parser.NewCommand("probe", "Install probe")
	probeDataDir := probeCmd.String("", "datadir", &argparse.Options{Required: true, Help: "Data directory"})
	probePass := probeCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Password to access the database"})
	host := probeCmd.String("", "host", &argparse.Options{Required: true, Help: "Master server address"})

	if len(os.Args) == 1 {
		tui()
	} else if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
	} else if removeCmd.Happened() {
		utils.Check(utils.Uninstall("cli"))
	} else if masterCmd.Happened() {
		utils.Check(utils.InstallMaster("cli", *masterDataDir, *masterPass, *fqdn, *customerName, *customerEmail))
	} else if probeCmd.Happened() {
		utils.Check(utils.InstallProbe("cli", *probeDataDir, *probePass, *host))
	}
}
