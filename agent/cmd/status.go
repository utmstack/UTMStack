package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/agent"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show per-source log collection health (received/persisted/failed, last event)",
	Args:    cobra.NoArgs,
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshot, err := agent.ReadStatus()
		if err != nil {
			fmt.Println("No status snapshot available yet (service may not be running): ", err)
			os.Exit(1)
		}

		fmt.Printf("Status as of %s\n\n", snapshot.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
		if len(snapshot.Sources) == 0 {
			fmt.Println("No sources have reported activity yet.")
			return nil
		}

		names := make([]string, 0, len(snapshot.Sources))
		for name := range snapshot.Sources {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Printf("%-24s %10s %10s %14s %s\n", "SOURCE", "RECEIVED", "PERSISTED", "PERSIST_FAIL", "LAST EVENT")
		for _, name := range names {
			s := snapshot.Sources[name]
			last := "never"
			if !s.LastEventAt.IsZero() {
				last = s.LastEventAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("%-24s %10d %10d %14d %s\n", name, s.Received, s.Persisted, s.PersistFailed, last)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
