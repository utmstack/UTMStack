package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/database"
	"github.com/utmstack/UTMStack/agent/models"
)

var cleanLogsCmd = &cobra.Command{
	Use:     "clean-logs",
	Short:   "Clean old logs from the database",
	Args:    cobra.NoArgs,
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Cleaning old logs ...")
		db := database.GetDB()
		datR, err := agent.GetDataRetention()
		if err != nil {
			fmt.Println("Error getting retention: ", err)
			os.Exit(1)
		}
		_, err = db.DeleteOld(models.Log{}, datR)
		if err != nil {
			fmt.Println("Error cleaning logs: ", err)
			os.Exit(1)
		}
		fmt.Println("Logs cleaned correctly")
		time.Sleep(5 * time.Second)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanLogsCmd)
}
