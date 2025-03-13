package modules

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/parser"
	"github.com/utmstack/UTMStack/agent/utils"
)

type SyslogModule struct {
	DataType    string
	TCPListener listenerTCP
	UDPListener listenerUDP
	Parser      parser.Parser
}

type listenerTCP struct {
	Listener  net.Listener
	CTX       context.Context
	Cancel    context.CancelFunc
	IsEnabled bool
	Port      string
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

func (m *SyslogModule) EnablePort(proto string) {
	switch proto {
	case "tcp":
		go m.enableTCP()
	case "udp":
		go m.enableUDP()
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
		m.TCPListener.IsEnabled = true

		listener, err := net.Listen("tcp", "0.0.0.0"+":"+m.TCPListener.Port)
		if err != nil {
			utils.Logger.ErrorF("error listening TCp in port %s: %v", m.TCPListener.Port, err)
			return
		}

		tcpListener, ok := listener.(*net.TCPListener)
		if !ok {
			utils.Logger.ErrorF("Could not assert to *net.TCPListener")
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
					tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
					conn, err := m.TCPListener.Listener.Accept()
					if err != nil {
						if errors.Is(err, net.ErrClosed) {
							return
						}

						netOpErr, ok := err.(*net.OpError)
						if ok && netOpErr.Timeout() {
							continue
						}

						utils.Logger.ErrorF("error connecting with tcp listener: %v", err)
						continue
					}
					go m.handleConnectionTCP(conn)
				}
			}
		}()
		select {}
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
			utils.Logger.ErrorF("Could not assert to *net.UDPConn")
			return
		}

		m.UDPListener.Listener = listener
		m.UDPListener.CTX, m.UDPListener.Cancel = context.WithCancel(context.Background())

		buffer := make([]byte, 1024)
		msgChannel := make(chan string)

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
					messageWithIP := config.GetMessageFormated(remoteAddr, string(buffer[:n]))
					msgChannel <- messageWithIP
				}
			}
		}()
	}
}

func (m *SyslogModule) disableTCP() {
	if m.TCPListener.IsEnabled && m.TCPListener.Port != "" {
		utils.Logger.Info("Server %s closed in port: %s protocol: TCP", m.DataType, m.TCPListener.Port)
		m.TCPListener.Cancel()
		m.TCPListener.Listener.Close()
		m.TCPListener.IsEnabled = false
	}
}

func (m *SyslogModule) disableUDP() {
	if m.UDPListener.IsEnabled && m.UDPListener.Port != "" {
		utils.Logger.Info("Server %s closed in port: %s protocol: UDP", m.DataType, m.UDPListener.Port)
		m.UDPListener.Cancel()
		m.UDPListener.Listener.Close()
		m.UDPListener.IsEnabled = false
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

	msgChannel := make(chan string)
	go m.handleMessageTCP(msgChannel)

	for {
		select {
		case <-m.TCPListener.CTX.Done():
			return
		default:
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || err.(net.Error).Timeout() {
					return
				}
				utils.Logger.ErrorF("error reading tcp data: %v", err)
				return
			}
			message = config.GetMessageFormated(remoteAddr, message)
			msgChannel <- message
		}
	}
}

func (m *SyslogModule) handleMessageTCP(logsChannel chan string) {
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				if m.Parser != nil {
					logs, err := m.Parser.ProcessData(logBatch)
					if err != nil {
						utils.Logger.ErrorF("error parsing data: %v", err)
						continue
					}
					for typ, bulk := range logs {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: bulk,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  m.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}

		case message := <-logsChannel:
			message = strings.TrimSuffix(message, "\n")
			message, _, err := validations.ValidateString(message, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
			}
			logBatch = append(logBatch, message)

			if len(logBatch) == config.BatchCapacity {
				if m.Parser != nil {
					logs, err := m.Parser.ProcessData(logBatch)
					if err != nil {
						utils.Logger.ErrorF("error parsing data: %v", err)
						continue
					}
					for typ, bulk := range logs {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: bulk,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  m.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}
		}
	}
}

func (m *SyslogModule) handleConnectionUDP(logsChannel chan string) {
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				if m.Parser != nil {
					logs, err := m.Parser.ProcessData(logBatch)
					if err != nil {
						utils.Logger.ErrorF("error parsing data: %v", err)
						continue
					}
					for typ, bulk := range logs {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: bulk,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  m.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}

		case <-m.UDPListener.CTX.Done():
			return

		case message := <-logsChannel:
			message = strings.TrimSuffix(message, "\n")
			message, _, err := validations.ValidateString(message, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
			}
			logBatch = append(logBatch, message)

			if len(logBatch) == config.BatchCapacity {
				if m.Parser != nil {
					logs, err := m.Parser.ProcessData(logBatch)
					if err != nil {
						utils.Logger.ErrorF("error parsing data: %v", err)
						continue
					}
					for typ, bulk := range logs {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: bulk,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  m.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}
		}
	}
}
