package modules

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/parser"
	"github.com/utmstack/UTMStack/agent/utils"
)

const (
	MinBufferSize         = 480
	RecommendedBufferSize = 2048
	MaxBufferSize         = 8192
	UDPBufferSize         = 2048
)

type FramingMethod int

const (
	FramingNewline FramingMethod = iota
	FramingOctetCounting
)

type SyslogModule struct {
	DataType    string
	TCPListener listenerTCP
	UDPListener listenerUDP
	Parser      parser.Parser
}

type listenerTCP struct {
	Listener   net.Listener
	CTX        context.Context
	Cancel     context.CancelFunc
	IsEnabled  bool
	Port       string
	TLSEnabled bool
}

type listenerUDP struct {
	Listener  net.PacketConn
	CTX       context.Context
	Cancel    context.CancelFunc
	IsEnabled bool
	Port      string
}

func GetSyslogModule(dataType string, protoPorts config.ProtoPort) *SyslogModule {
	return &SyslogModule{
		DataType: dataType,
		TCPListener: listenerTCP{
			IsEnabled: false,
			Port:      protoPorts.TCP,
		},
		UDPListener: listenerUDP{
			IsEnabled: false,
			Port:      protoPorts.UDP,
		},
		Parser: parser.GetParser(dataType),
	}
}

func (m *SyslogModule) GetDataType() string {
	return m.DataType
}

func (m *SyslogModule) IsPortListen(proto string) bool {
	switch proto {
	case "tcp":
		return m.TCPListener.IsEnabled
	case "udp":
		return m.UDPListener.IsEnabled
	default:
		return false
	}
}

func (m *SyslogModule) SetNewPort(proto string, port string) {
	// validate port by dataType, ranges allowed and ports in use
	switch proto {
	case "tcp":
		m.TCPListener.Port = port
	case "udp":
		m.UDPListener.Port = port
	}
}

func (m *SyslogModule) GetPort(proto string) string {
	switch proto {
	case "tcp":
		return m.TCPListener.Port
	case "udp":
		return m.UDPListener.Port
	default:
		return ""
	}
}

func (m *SyslogModule) EnablePort(proto string, enableTLS bool) error {
	switch proto {
	case "tcp":
		if enableTLS {
			if !utils.CheckIfPathExist(config.IntegrationCertPath) || !utils.CheckIfPathExist(config.IntegrationKeyPath) {
				return fmt.Errorf("TLS certificates not found. Please load certificates first")
			}
		}

		m.TCPListener.TLSEnabled = enableTLS
		go m.enableTCP()
		return nil
	case "udp":
		if enableTLS {
			return fmt.Errorf("TLS not supported for UDP protocol")
		}
		go m.enableUDP()
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (m *SyslogModule) DisablePort(proto string) {
	switch proto {
	case "tcp":
		m.disableTCP()
	case "udp":
		m.disableUDP()
	}
}

func (m *SyslogModule) enableTCP() {
	if !m.TCPListener.IsEnabled && m.TCPListener.Port != "" {
		utils.Logger.Info("Server %s listening in port: %s protocol: TCP", m.DataType, m.TCPListener.Port)
		if m.TCPListener.TLSEnabled {
			utils.Logger.Info("Server %s TLS enabled in port: %s protocol: TCP", m.DataType, m.TCPListener.Port)
		}
		m.TCPListener.IsEnabled = true

		listener, err := net.Listen("tcp", "0.0.0.0:"+m.TCPListener.Port)
		if err != nil {
			utils.Logger.ErrorF("error listening TCP in port %s: %v", m.TCPListener.Port, err)
			return
		}

		m.TCPListener.Listener = listener
		m.TCPListener.CTX, m.TCPListener.Cancel = context.WithCancel(context.Background())

		go func() {
			defer func() {
				err = m.TCPListener.Listener.Close()
				if err != nil {
					utils.Logger.ErrorF("error closing tcp listener: %v", err)
				}
			}()
			for {
				select {
				case <-m.TCPListener.CTX.Done():
					return
				default:
					conn, err := m.TCPListener.Listener.Accept()
					if err != nil {
						if errors.Is(err, net.ErrClosed) {
							return
						}

						var netOpErr *net.OpError
						ok := errors.As(err, &netOpErr)
						if ok && netOpErr.Timeout() {
							continue
						}

						utils.Logger.ErrorF("error connecting with tcp listener: %v", err)
						continue
					}

					// Connection handling based on TLS configuration
					if m.TCPListener.TLSEnabled {
						go m.handleTLSConnection(conn)
					} else {
						go m.handleConnectionTCP(conn)
					}
				}
			}
		}()
	}
}

func (m *SyslogModule) enableUDP() {
	if !m.UDPListener.IsEnabled && m.UDPListener.Port != "" {
		utils.Logger.Info("Server %s listening in port: %s protocol: UDP\n", m.DataType, m.UDPListener.Port)
		m.UDPListener.IsEnabled = true

		listener, err := net.ListenPacket("udp", "0.0.0.0"+":"+m.UDPListener.Port)
		if err != nil {
			utils.Logger.ErrorF("error listening UDP in port %s: %v", m.UDPListener.Port, err)
			return
		}

		udpListener, ok := listener.(*net.UDPConn)
		if !ok {
			utils.Logger.ErrorF("could not assert to *net.UDPConn")
			return
		}

		m.UDPListener.Listener = listener
		m.UDPListener.CTX, m.UDPListener.Cancel = context.WithCancel(context.Background())

		buffer := make([]byte, UDPBufferSize)
		msgChannel := make(chan config.MSGDS)

		go m.handleConnectionUDP(msgChannel)

		go func() {
			defer func() {
				err = m.UDPListener.Listener.Close()
				if err != nil {
					utils.Logger.ErrorF("error closing udp listener: %v", err)
				}
			}()
			for {
				select {
				case <-m.UDPListener.CTX.Done():
					return
				default:
					udpListener.SetDeadline(time.Now().Add(time.Second * 1))

					n, add, err := listener.ReadFrom(buffer)
					if err != nil {
						if errors.Is(err, net.ErrClosed) {
							return
						}

						var netOpErr *net.OpError
						ok := errors.As(err, &netOpErr)
						if ok && netOpErr.Timeout() {
							continue
						}

						utils.Logger.ErrorF("error connecting with udp listener: %v", err)
						continue
					}
					remoteAddr := add.String()
					remoteAddr, _, err = net.SplitHostPort(remoteAddr)
					if err != nil {
						utils.Logger.ErrorF("error getting remote addr: %v", err)
						continue
					}
					if remoteAddr == "127.0.0.1" {
						remoteAddr, err = os.Hostname()
						if err != nil {
							utils.Logger.ErrorF("error getting hostname: %v\n", err)
							continue
						}
					}
					msgChannel <- config.MSGDS{
						DataSource: remoteAddr,
						Message:    string(buffer[:n]),
					}
				}
			}
		}()
	}
}

func (m *SyslogModule) disableTCP() {
	if m.TCPListener.IsEnabled && m.TCPListener.Port != "" {
		utils.Logger.Info("Server %s closed in port: %s protocol: TCP", m.DataType, m.TCPListener.Port)

		if m.TCPListener.Listener != nil {
			if err := m.TCPListener.Listener.Close(); err != nil {
				utils.Logger.ErrorF("error closing TCP listener: %v", err)
			}
		}

		m.TCPListener.Cancel()
		m.TCPListener.IsEnabled = false
	}
}

func (m *SyslogModule) disableUDP() {
	if m.UDPListener.IsEnabled && m.UDPListener.Port != "" {
		utils.Logger.Info("Server %s closed in port: %s protocol: UDP", m.DataType, m.UDPListener.Port)

		if m.UDPListener.Listener != nil {
			if err := m.UDPListener.Listener.Close(); err != nil {
				utils.Logger.ErrorF("error closing UDP listener: %v", err)
			}
		}

		m.UDPListener.Cancel()
		m.UDPListener.IsEnabled = false
	}
}

// detectFramingMethod detects the syslog framing method by peeking at the first byte
func detectFramingMethod(reader *bufio.Reader) (FramingMethod, error) {
	firstByte, err := reader.Peek(1)
	if err != nil {
		utils.Logger.ErrorF("failed to peek first byte for framing detection: %v", err)
		return 0, fmt.Errorf("failed to peek first byte: %w", err)
	}

	if firstByte[0] >= '0' && firstByte[0] <= '9' {
		return FramingOctetCounting, nil
	}

	if firstByte[0] == '<' {
		return FramingNewline, nil
	}

	utils.Logger.ErrorF("unknown framing method detected, first byte: 0x%02x", firstByte[0])
	return 0, fmt.Errorf("unknown framing method, first byte: 0x%02x", firstByte[0])
}

// readOctetCountingFrame reads a syslog message using octet counting framing method
func readOctetCountingFrame(reader *bufio.Reader) (string, error) {
	lengthStr, err := reader.ReadString(' ')
	if err != nil {
		utils.Logger.ErrorF("failed to read message length in octet counting frame: %v", err)
		return "", fmt.Errorf("failed to read message length: %w", err)
	}

	lengthStr = strings.TrimSuffix(lengthStr, " ")
	msgLen, err := strconv.Atoi(lengthStr)
	if err != nil {
		utils.Logger.ErrorF("invalid message length '%s' in octet counting frame: %v", lengthStr, err)
		return "", fmt.Errorf("invalid message length '%s': %w", lengthStr, err)
	}

	if msgLen < 1 {
		utils.Logger.ErrorF("message length %d is too small (minimum 1 byte)", msgLen)
		return "", fmt.Errorf("message length %d is too small (minimum 1)", msgLen)
	}
	if msgLen > MaxBufferSize {
		utils.Logger.ErrorF("message length %d exceeds maximum %d bytes", msgLen, MaxBufferSize)
		return "", fmt.Errorf("message length %d exceeds maximum %d", msgLen, MaxBufferSize)
	}

	msgBytes := make([]byte, msgLen)
	_, err = io.ReadFull(reader, msgBytes)
	if err != nil {
		utils.Logger.ErrorF("failed to read %d byte message body: %v", msgLen, err)
		return "", fmt.Errorf("failed to read %d byte message body: %w", msgLen, err)
	}

	return string(msgBytes), nil
}

// readNewlineFrame reads a syslog message using newline-delimited framing method
func readNewlineFrame(reader *bufio.Reader) (string, error) {
	message, err := reader.ReadString('\n')
	if err != nil {
		utils.Logger.ErrorF("failed to read newline-delimited message: %v", err)
		return "", fmt.Errorf("failed to read newline-delimited message: %w", err)
	}
	return message, nil
}

// readSyslogMessage reads a syslog message with automatic framing detection
func readSyslogMessage(reader *bufio.Reader) (string, error) {
	method, err := detectFramingMethod(reader)
	if err != nil {
		return "", err
	}

	switch method {
	case FramingOctetCounting:
		return readOctetCountingFrame(reader)
	case FramingNewline:
		return readNewlineFrame(reader)
	default:
		utils.Logger.ErrorF("unsupported framing method: %d", method)
		return "", fmt.Errorf("unsupported framing method: %d", method)
	}
}

func (m *SyslogModule) handleConnectionTCP(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	remoteAddr := c.RemoteAddr().String()

	var err error
	remoteAddr, _, err = net.SplitHostPort(remoteAddr)
	if err != nil {
		utils.Logger.ErrorF("error spliting host and port: %v", err)
	}

	if remoteAddr == "127.0.0.1" {
		remoteAddr, err = os.Hostname()
		if err != nil {
			utils.Logger.ErrorF("error getting hostname: %v\n", err)
		}
	}

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

	msgChannel := make(chan config.MSGDS)
	go m.handleMessageTCP(msgChannel)

	for {
		select {
		case <-m.TCPListener.CTX.Done():
			return
		default:
			message, err := readSyslogMessage(reader)
			if err != nil {
				if err == io.EOF {
					utils.Logger.Info("TCP connection closed by %s", remoteAddr)
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					utils.Logger.Info("TCP connection timeout from %s", remoteAddr)
					return
				}
				utils.Logger.ErrorF("error reading syslog message from %s: %v", remoteAddr, err)
				return
			}
			msgChannel <- config.MSGDS{
				DataSource: remoteAddr,
				Message:    message,
			}
		}
	}
}

func (m *SyslogModule) handleTLSConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	remoteAddr, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		utils.Logger.ErrorF("error splitting host and port: %v", err)
		remoteAddr = "unknown"
	}

	if remoteAddr == "127.0.0.1" {
		if hostname, err := os.Hostname(); err == nil {
			remoteAddr = hostname
		}
	}

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
	// Keep a reasonable read timeout instead of removing it entirely
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(tlsConn)
	msgChannel := make(chan config.MSGDS)
	go m.handleMessageTCP(msgChannel)

	for {
		select {
		case <-m.TCPListener.CTX.Done():
			return
		default:
			// Set read timeout for each message
			conn.SetDeadline(time.Now().Add(30 * time.Second))
			message, err := readSyslogMessage(reader)
			if err != nil {
				if err == io.EOF {
					utils.Logger.Info("TLS connection closed by %s", remoteAddr)
					return
				}
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					utils.Logger.Info("TLS connection timeout from %s", remoteAddr)
					return
				}
				utils.Logger.ErrorF("error reading TLS data from %s: %v", remoteAddr, err)
				return
			}
			msgChannel <- config.MSGDS{
				DataSource: remoteAddr,
				Message:    message,
			}
		}
	}
}

func (m *SyslogModule) handleMessageTCP(logsChannel chan config.MSGDS) {
	for {
		select {
		case <-m.TCPListener.CTX.Done():
			return

		case msgDS := <-logsChannel:
			message := msgDS.Message
			message = strings.TrimSuffix(message, "\n")
			message, _, err := entities.ValidateString(message, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
				continue
			}

			if m.Parser != nil {
				err := m.Parser.ProcessData(message, msgDS.DataSource, logservice.LogQueue)
				if err != nil {
					utils.Logger.ErrorF("error parsing data: %v", err)
					continue
				}
			} else {
				logservice.LogQueue <- &plugins.Log{
					DataType:   m.DataType,
					DataSource: msgDS.DataSource,
					Raw:        message,
				}
			}

		}
	}
}

func (m *SyslogModule) handleConnectionUDP(logsChannel chan config.MSGDS) {
	for {
		select {
		case <-m.UDPListener.CTX.Done():
			return

		case msgDS := <-logsChannel:
			message := msgDS.Message
			message = strings.TrimSuffix(message, "\n")
			message, _, err := entities.ValidateString(message, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
				continue
			}

			if m.Parser != nil {
				err := m.Parser.ProcessData(message, msgDS.DataSource, logservice.LogQueue)
				if err != nil {
					utils.Logger.ErrorF("error parsing data: %v", err)
					continue
				}
			} else {
				logservice.LogQueue <- &plugins.Log{
					DataType:   m.DataType,
					DataSource: msgDS.DataSource,
					Raw:        message,
				}
			}
		}
	}
}
