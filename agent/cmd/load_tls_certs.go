package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

var loadTLSCertsCmd = &cobra.Command{
	Use:   "load-tls-certs <certificate_path> <private_key_path> [ca_certificate_path]",
	Short: "Load your own TLS certificates (RECOMMENDED for production)",
	Long: `Load your own TLS certificates.

Examples:
  utmstack_agent load-tls-certs /path/to/server.crt /path/to/server.key /path/to/ca.crt
  utmstack_agent load-tls-certs /path/to/server.crt /path/to/server.key  # Without CA`,
	Args:    cobra.RangeArgs(2, 3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		userCertPath := args[0]
		userKeyPath := args[1]
		var userCAPath string
		if len(args) > 2 {
			userCAPath = args[2]
		}

		fmt.Println("Loading user TLS certificates ...")

		fmt.Print("Validating certificate files ... ")
		if err := utils.ValidateIntegrationCertificates(userCertPath, userKeyPath); err != nil {
			fmt.Printf("\nError: Invalid certificate files: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[OK]")

		fmt.Print("Installing certificates ... ")
		src := utils.CertificateFiles{
			CertPath: userCertPath,
			KeyPath:  userKeyPath,
			CAPath:   userCAPath,
		}
		dest := utils.CertificateFiles{
			CertPath: config.IntegrationCertPath,
			KeyPath:  config.IntegrationKeyPath,
			CAPath:   config.IntegrationCAPath,
		}
		if err := utils.LoadUserCertificatesWithStruct(src, dest); err != nil {
			fmt.Printf("\nError loading certificates: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[OK]")

		fmt.Println("TLS certificates loaded successfully!")
		time.Sleep(5 * time.Second)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadTLSCertsCmd)
}
