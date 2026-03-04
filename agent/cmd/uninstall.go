package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	pb "github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/collector"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/dependency"
	"github.com/utmstack/UTMStack/agent/serv"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall the UTMStackAgent service",
	Args:    cobra.NoArgs,
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling UTMStackAgent service ...")

		fmt.Print("Stopping UTMStackUpdater service... ")
		updaterPath := filepath.Join(fs.GetExecutablePath(), dependency.UpdaterFile(""))
		if fs.Exists(updaterPath) {
			err := exec.Run(updaterPath, fs.GetExecutablePath(), "uninstall")
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
			} else {
				fmt.Println("[OK]")
			}
			time.Sleep(2 * time.Second)
		} else {
			fmt.Println("[SKIPPED - not found]")
		}

		cnf, err := config.GetCurrentConfig()
		if err != nil {
			fmt.Println("Error getting config: ", err)
			os.Exit(1)
		}
		if err = pb.DeleteAgent(cnf); err != nil {
			utils.Logger.ErrorF("error deleting agent: %v", err)
		}
		if err = collector.UninstallAll(); err != nil {
			fmt.Printf("error uninstalling collectors: %v\n", err)
			os.Exit(1)
		}
		os.Remove(config.ConfigurationFile)

		serv.UninstallService()

		fmt.Println("[OK]")
		fmt.Println("UTMStackAgent service uninstalled correctly")
		os.Exit(0)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
