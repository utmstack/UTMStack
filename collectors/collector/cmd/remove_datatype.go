package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

var removeDataTypeCmd = &cobra.Command{
	Use:   "remove-datatype <name>",
	Short: "Remove a user-defined DataType from the catalog",
	Long: `Remove a user-defined DataType that was previously added with 'enable-integration --create'.

Built-in DataTypes cannot be removed.
The DataType must not be referenced by any active integration in the collector config.`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.RemoveUserDataType(name); err != nil {
			return fmt.Errorf("failed to remove data type: %w", err)
		}
		fmt.Printf("Data type %q removed successfully.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeDataTypeCmd)
}
