package syslog

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

// resolveRemoteAddr extracts the IP from addr and replaces localhost with hostname.
func resolveRemoteAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		utils.Logger.ErrorF("error splitting host and port: %v", err)
		return "unknown"
	}
	if host == "127.0.0.1" {
		if hostname, err := os.Hostname(); err == nil {
			return hostname
		}
	}
	return host
}

// readLoop reads syslog messages from reader and sends them to msgChannel.
// If conn is provided, it sets a deadline before each read (used for TLS).
func (inst *syslogInstance) readLoop(ctx context.Context, reader *bufio.Reader, remoteAddr string, msgChannel chan MSGDS, conn net.Conn) {
	connType := "TCP"
	if conn != nil {
		connType = "TLS"
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if conn != nil {
				conn.SetDeadline(time.Now().Add(30 * time.Second))
			}
			message, err := readSyslogMessage(reader)
			if err != nil {
				if err == io.EOF {
					utils.Logger.Info("%s connection closed by %s", connType, remoteAddr)
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					utils.Logger.Info("%s connection timeout from %s", connType, remoteAddr)
					return
				}
				utils.Logger.ErrorF("error reading %s data from %s: %v", connType, remoteAddr, err)
				return
			}
			msgChannel <- MSGDS{
				DataSource: remoteAddr,
				Message:    message,
			}
		}
	}
}

func (inst *syslogInstance) handleConnectionTCP(c net.Conn, queue chan *plugins.Log) {
	defer c.Close()
	reader := bufio.NewReader(c)
	remoteAddr := resolveRemoteAddr(c.RemoteAddr().String())

	// Detect and reject TLS connections when TLS is disabled
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	firstBytes := make([]byte, 3)
	n, err := reader.Read(firstBytes)
	if err != nil {
		utils.Logger.ErrorF("error reading initial bytes from %s: %v", remoteAddr, err)
		return
	}

	// TLS handshake starts with: 0x16 (22 decimal) for TLS 1.0-1.3
	if n >= 1 && firstBytes[0] == 0x16 {
		utils.Logger.ErrorF("TLS connection rejected from %s: TLS is disabled, only plain text connections accepted", remoteAddr)
		return
	}

	// Reset deadline and create a new reader that includes the read bytes
	c.SetReadDeadline(time.Time{})
	reader = bufio.NewReader(io.MultiReader(strings.NewReader(string(firstBytes[:n])), reader))

	msgChannel := make(chan MSGDS)
	defer close(msgChannel)
	go inst.handleMessage(inst.TCPListener.CTX, msgChannel, queue)

	inst.readLoop(inst.TCPListener.CTX, reader, remoteAddr, msgChannel, nil)
}

func (inst *syslogInstance) handleTLSConnection(conn net.Conn, queue chan *plugins.Log) {
	defer conn.Close()

	remoteAddr := resolveRemoteAddr(conn.RemoteAddr().String())

	tlsConfig, err := utils.LoadIntegrationTLSConfig(
		config.IntegrationCertPath,
		config.IntegrationKeyPath,
	)
	if err != nil {
		utils.Logger.ErrorF("error loading TLS config: %v", err)
		return
	}

	tlsConn := tls.Server(conn, tlsConfig)

	conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err := tlsConn.Handshake(); err != nil {
		utils.Logger.ErrorF("TLS handshake failed from %s: %v", remoteAddr, err)
		return
	}

	reader := bufio.NewReader(tlsConn)
	msgChannel := make(chan MSGDS)
	defer close(msgChannel)
	go inst.handleMessage(inst.TCPListener.CTX, msgChannel, queue)

	inst.readLoop(inst.TCPListener.CTX, reader, remoteAddr, msgChannel, conn)
}

// handleMessage processes messages from the channel and sends them to the queue.
func (inst *syslogInstance) handleMessage(ctx context.Context, logsChannel chan MSGDS, queue chan *plugins.Log) {
	for {
		select {
		case <-ctx.Done():
			return
		case msgDS, ok := <-logsChannel:
			if !ok {
				return // channel closed
			}
			message := strings.TrimSuffix(msgDS.Message, "\n")
			message, _, err := entities.ValidateString(message, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case queue <- &plugins.Log{
				DataType:   inst.DataType,
				DataSource: msgDS.DataSource,
				Raw:        message,
			}:
			}
		}
	}
}
