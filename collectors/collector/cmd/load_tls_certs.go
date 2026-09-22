package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

var loadTLSCertsCmd = &cobra.Command{
	Use:   "load-tls-certs <cert> <key> [ca]",
	Short: "Load TLS certificates for integration listeners",
	Long: `Validate and install TLS certificate and private key for use by
integration listeners (TCP/TLS and HTTPS). An optional CA certificate can
be provided; if omitted, the leaf certificate is used as the CA.

Files are copied to the collector's certs/ directory:
  - integration.crt  (certificate)
  - integration.key  (private key)
  - integration_ca.crt  (CA certificate)

Examples:
  utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key
  utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key /path/to/ca.crt`,
	Args:    cobra.RangeArgs(2, 3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		certPath := args[0]
		keyPath := args[1]
		caPath := ""
		if len(args) == 3 {
			caPath = args[2]
		}

		fmt.Println("Loading TLS certificates ...")

		src := utils.CertificateFiles{
			CertPath: certPath,
			KeyPath:  keyPath,
			CAPath:   caPath,
		}
		dest := utils.CertificateFiles{
			CertPath: config.IntegrationCertPath,
			KeyPath:  config.IntegrationKeyPath,
			CAPath:   config.IntegrationCAPath,
		}

		if err := utils.LoadUserCertificatesWithStruct(src, dest); err != nil {
			fmt.Printf("Error loading TLS certificates: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("TLS certificates loaded successfully:")
		fmt.Printf("  Certificate : %s\n", config.IntegrationCertPath)
		fmt.Printf("  Private key : %s\n", config.IntegrationKeyPath)
		fmt.Printf("  CA cert     : %s\n", config.IntegrationCAPath)
		time.Sleep(2 * time.Second)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadTLSCertsCmd)
}
