package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

var enableIntegrationCmd = &cobra.Command{
	Use:   "enable-integration <name> <port> <protocol>",
	Short: "Enable a log integration",
	Long: `Enable a log integration by name, port and protocol.

Protocol: tcp | udp | tls | http | https

If the DataType does not exist it will be created automatically.

Examples:
  utmstack_forwarder enable-integration firewall-cisco-asa 514 udp
  utmstack_forwarder enable-integration firewall-cisco-asa 1470 tcp
  utmstack_forwarder enable-integration firewall-cisco-asa 1470 tls
  utmstack_forwarder enable-integration my-app 8080 http
  utmstack_forwarder enable-integration my-app 8443 https`,
	Args:    cobra.ExactArgs(3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		port := args[1]
		proto := strings.ToLower(args[2])

		// Validate protocol
		switch proto {
		case "tcp", "udp", "tls", "http", "https":
		default:
			return fmt.Errorf("invalid protocol %q: must be tcp, udp, tls, http or https", proto)
		}

		// If the DataType doesn't exist, create it automatically
		if _, ok := config.ResolveDataType(name); !ok {
			kind := "syslog"
			switch proto {
			case "http", "https":
				kind = proto
			case "netflow":
				kind = "netflow"
			}
			udp, tcp := "", ""
			switch proto {
			case "udp":
				udp = port
			case "tcp", "tls":
				tcp = port
			}
			if err := config.AddUserDataType(name, kind, udp, tcp); err != nil {
				return fmt.Errorf("failed to create data type %q: %w", name, err)
			}
			fmt.Printf("New data type %q created.\n", name)
		}

		// Enable
		switch proto {
		case "http", "https":
			return enableHTTPIntegration(name, port, proto)
		default:
			return enableNetworkIntegration(name, port, proto)
		}
	},
}

func enableHTTPIntegration(name, port, proto string) error {
	fmt.Printf("Enabling %s integration %q on port %s...\n", proto, name, port)

	opts := collectorpkg.HTTPIntegrationOptions{
		Proto:           proto,
		Port:            port,
		Path:            "/logs",
		Bind:            "127.0.0.1",
		Auth:            enableAuth,
		SignatureHeader: enableSigHeader,
	}
	if err := collectorpkg.EnableHTTPIntegration(name, opts); err != nil {
		return fmt.Errorf("error enabling %s integration: %w", proto, err)
	}

	if enableAuth != "" {
		fmt.Printf("Integration %q enabled (%s port %s, auth: %s)\n", name, proto, port, enableAuth)
	} else {
		fmt.Printf("Integration %q enabled (%s port %s, no auth)\n", name, proto, port)
	}
	time.Sleep(2 * time.Second)
	return nil
}

func enableNetworkIntegration(name, port, proto string) error {
	useTLS := proto == "tls"
	networkProto := proto
	if useTLS {
		networkProto = "tcp" // tls is tcp with TLS enabled
	}
	fmt.Printf("Enabling %s integration %q on port %s...\n", proto, name, port)

	actualPort, err := collectorpkg.ChangeIntegrationStatus(name, networkProto, true, useTLS)
	if err != nil {
		return fmt.Errorf("error enabling integration: %w", err)
	}

	if useTLS {
		fmt.Printf("Integration %q enabled on tcp port %s (TLS)\n", name, actualPort)
	} else {
		fmt.Printf("Integration %q enabled on %s port %s\n", name, proto, actualPort)
	}
	time.Sleep(2 * time.Second)
	return nil
}

var (
	enableAuth      string // "" | "bearer" | "hmac"
	enableSigHeader string // e.g. "X-Hub-Signature-256"
)

func init() {
	enableIntegrationCmd.Flags().StringVar(&enableAuth, "auth", "", "HTTP/HTTPS auth mode: bearer | hmac (optional, default: no auth)")
	enableIntegrationCmd.Flags().StringVar(&enableSigHeader, "signature-header", "X-Hub-Signature-256", "Header name carrying the HMAC signature (hmac auth only)")
	rootCmd.AddCommand(enableIntegrationCmd)
}
