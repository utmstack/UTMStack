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
	Use:   "enable-integration <name> [port] <protocol>",
	Short: "Enable a log integration",
	Long: `Enable a log integration by name and protocol.

The port is optional for known integrations — the Forwarder uses the built-in
default port for that data type. For a new/custom data type you must pass a port
so it can be created.

Protocol: tcp | udp | tls | http | https

Examples:
  utmstack_forwarder enable-integration firewall-cisco-asa udp
  utmstack_forwarder enable-integration firewall-cisco-asa tcp
  utmstack_forwarder enable-integration firewall-cisco-asa tls
  utmstack_forwarder enable-integration my-app 8080 http
  utmstack_forwarder enable-integration my-app 9000 udp`,
	Args:    cobra.RangeArgs(2, 3),
	PreRunE: requireInstalled,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var port, proto string
		if len(args) == 3 {
			port = args[1]
			proto = strings.ToLower(args[2])
		} else {
			proto = strings.ToLower(args[1])
		}

		// Validate protocol
		switch proto {
		case "tcp", "udp", "tls", "http", "https":
		default:
			return fmt.Errorf("invalid protocol %q: must be tcp, udp, tls, http or https", proto)
		}

		entry, known := config.ResolveDataType(name)

		// Resolve the port from the catalog when omitted. Known integrations carry
		// their built-in default; a new/custom data type must provide one.
		if port == "" {
			switch proto {
			case "udp":
				port = entry.UDP
			case "tcp", "tls":
				port = entry.TCP
			case "http", "https":
				port = entry.Port
			}
			if port == "" {
				if !known {
					return fmt.Errorf("port is required for new integration %q: it has no built-in default", name)
				}
				return fmt.Errorf("integration %q has no default %s port; pass one explicitly", name, proto)
			}
		}

		// If the DataType doesn't exist, create it automatically
		if _, err := collectorpkg.ResolveOrCreateDataType(name, proto, port); err != nil {
			return err
		}
		if !known {
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

	// Defaults for a generic HTTP integration.
	path, bind, auth, sigHeader := "/logs", "127.0.0.1", enableAuth, enableSigHeader

	// Built-in HTTP integrations (e.g. github) ship their own path/bind/auth so the
	// user doesn't have to pass them; an explicit --auth still wins.
	if spec, ok := config.HTTPPorts[config.DataType(name)]; ok {
		path, bind = spec.Path, spec.Bind
		if auth == "" {
			auth = spec.Auth
		}
		if spec.SignatureHeader != "" {
			sigHeader = spec.SignatureHeader
		}
	}

	opts := collectorpkg.HTTPIntegrationOptions{
		Proto:           proto,
		Port:            port,
		Path:            path,
		Bind:            bind,
		Auth:            auth,
		SignatureHeader: sigHeader,
	}
	if _, err := collectorpkg.EnableHTTPIntegration(name, opts); err != nil {
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
