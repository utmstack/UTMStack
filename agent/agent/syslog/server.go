package syslog

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/filters"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
)

type listenerTCP struct {
	Listener net.Listener
	CTX      context.Context
	Cancel   context.CancelFunc
}

type listenerUDP struct {
	Listener net.PacketConn
	CTX      context.Context
	Cancel   context.CancelFunc
}

type listenerTLS struct {
	Listener net.Listener
	CTX      context.Context
	Cancel   context.CancelFunc
}

type SyslogServer struct {
	TCPListener    listenerTCP
	UDPListener    listenerUDP
	TLSListener    listenerTLS
	DataType       string
	IsEnabled      bool
	IsListen       bool
	protoPorts     configuration.ProtoPort
	addr           string
	tlsConfig      *tls.Config
	ChannelHandler chan logservice.LogPipe
}

func NewSyslogServer(dataType string, isEnabled bool, protoPorts configuration.ProtoPort, addr string) SyslogServer {
	return SyslogServer{
		DataType:   dataType,
		IsEnabled:  isEnabled,
		protoPorts: protoPorts,
		addr:       addr,
	}
}

func (s *SyslogServer) SetHandler(channel chan logservice.LogPipe) {
	s.ChannelHandler = channel
}

func (s *SyslogServer) Listen(h *holmes.Logger) {
	if s.protoPorts.UDP != "" {
		h.Info("Server %s listening in port: %s protocol: UDP\n", s.DataType, s.protoPorts.UDP)
		go s.listenUDP(h)
	}

	if s.protoPorts.TCP != "" {
		h.Info("Server %s listening in port: %s protocol: TCP\n", s.DataType, s.protoPorts.TCP)
		go s.listenTCP(h)
	}

	if s.protoPorts.TLS != "" {
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(configuration.GetCaPath())
		if err != nil {
			h.FatalError("%s", err)
		}
		caCertPool.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(configuration.GetCertPath(), configuration.GetKeyPath())
		if err != nil {
			h.FatalError("%s", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    caCertPool,
		}
		s.tlsConfig = tlsConfig

		h.Info("Server %s listening in port: %s protocol: TCP(TLS)\n", s.DataType, s.protoPorts.TLS)
		go s.listenTLS(h)
	}
}

func (s *SyslogServer) Kill(h *holmes.Logger) {
	if s.protoPorts.UDP != "" {
		h.Info("Server %s closed in port: %s protocol: UDP\n", s.DataType, s.protoPorts.UDP)
		s.UDPListener.Cancel()
		s.UDPListener.Listener.Close()
	}

	if s.protoPorts.TCP != "" {
		h.Info("Server %s closed in port: %s protocol: TCP\n", s.DataType, s.protoPorts.TCP)
		s.TCPListener.Cancel()
		s.TCPListener.Listener.Close()
	}

	if s.protoPorts.TLS != "" {
		h.Info("Server %s closed in port: %s protocol: TCP(TLS)\n", s.DataType, s.protoPorts.TLS)
		s.TLSListener.Cancel()
		s.TLSListener.Listener.Close()
	}
}

func (s *SyslogServer) listenUDP(h *holmes.Logger) {
	listener, err := net.ListenPacket("udp", s.addr+":"+s.protoPorts.UDP)
	if err != nil {
		h.Error("%v", err)
		return
	}

	udpListener, ok := listener.(*net.UDPConn)
	if !ok {
		h.Error("Could not assert to *net.UDPConn")
		return
	}

	s.UDPListener.Listener = listener
	s.UDPListener.CTX, s.UDPListener.Cancel = context.WithCancel(context.Background())

	buffer := make([]byte, 1024)
	msgChannel := make(chan string)

	go s.handleConnectionUDP(msgChannel, h)
	go func() {
		defer s.UDPListener.Listener.Close()
		for {
			select {
			case <-s.UDPListener.CTX.Done():
				return
			default:
				udpListener.SetDeadline(time.Now().Add(time.Second * 1))

				n, add, err := listener.ReadFrom(buffer)
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}

					netOpErr, ok := err.(*net.OpError)
					if ok && netOpErr.Timeout() {
						continue
					}

					h.Error("%v", err)
					continue
				}
				remoteAddr := add.String()
				remoteAddr, _, err = net.SplitHostPort(remoteAddr)
				if err != nil {
					h.Error("%v", err)
					continue
				}
				if remoteAddr == "127.0.0.1" {
					remoteAddr, err = os.Hostname()
					if err != nil {
						h.Error("error getting hostname: %v\n", err)
						continue
					}
				}
				messageWithIP := "[utm_stack_agent_ds=" + remoteAddr + "]-" + string(buffer[:n])
				msgChannel <- messageWithIP
			}
		}
	}()
}

func (s *SyslogServer) listenTCP(h *holmes.Logger) {
	listener, err := net.Listen("tcp", s.addr+":"+s.protoPorts.TCP)
	if err != nil {
		h.Error("%v", err)
		return
	}

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		h.Error("Could not assert to *net.TCPListener")
		return
	}

	s.TCPListener.Listener = listener
	s.TCPListener.CTX, s.TCPListener.Cancel = context.WithCancel(context.Background())

	go func() {
		defer s.TCPListener.Listener.Close()
		for {
			select {
			case <-s.TCPListener.CTX.Done():
				return
			default:
				tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
				conn, err := s.TCPListener.Listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}

					netOpErr, ok := err.(*net.OpError)
					if ok && netOpErr.Timeout() {
						continue
					}

					h.Error("%v", err)
					continue
				}

				go s.handleConnectionTCP(conn, h)
			}
		}
	}()
}

func (s *SyslogServer) listenTLS(h *holmes.Logger) {
	listener, err := tls.Listen("tcp", s.addr+":"+s.protoPorts.TLS, s.tlsConfig)
	if err != nil {
		h.FatalError("%s", err)
	}

	s.TLSListener.Listener = listener
	s.TLSListener.CTX, s.TLSListener.Cancel = context.WithCancel(context.Background())

	go func() {
		defer s.TLSListener.Listener.Close()
		for {
			select {
			case <-s.TLSListener.CTX.Done():
				return
			default:
				conn, err := s.TLSListener.Listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}

					netOpErr, ok := err.(*net.OpError)
					if ok && netOpErr.Timeout() {
						continue
					}

					h.Error("%v", err)
					continue
				}

				go s.handleConnectionTLS(conn, h)
			}
		}
	}()
}

func (s *SyslogServer) handleConnectionUDP(logsChannel chan string, h *holmes.Logger) {
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				if s.DataType == string(configuration.LogTypeCiscoGeneric) {
					ciscoData, err := filters.ProcessCiscoData(logBatch)
					if err != nil {
						h.Error("error processing cisco data: %v", err)
						continue
					}
					for typ, logB := range ciscoData {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: logB,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  s.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}

		case <-s.UDPListener.CTX.Done():
			return

		case message := <-logsChannel:
			message = strings.TrimSuffix(message, "\n")
			logBatch = append(logBatch, message)

			if len(logBatch) == configuration.BatchCapacity {
				if s.DataType == string(configuration.LogTypeCiscoGeneric) {
					ciscoData, err := filters.ProcessCiscoData(logBatch)
					if err != nil {
						h.Error("error processing cisco data: %v", err)
						continue
					}
					for typ, logB := range ciscoData {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: logB,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  s.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}
		}
	}
}

func (s *SyslogServer) handleConnectionTCP(c net.Conn, h *holmes.Logger) {
	defer c.Close()
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)
	reader := bufio.NewReader(c)
	remoteAddr := c.RemoteAddr().String()

	var err error
	remoteAddr, _, err = net.SplitHostPort(remoteAddr)
	if err != nil {
		h.FatalError("error spliting host and port: ", err)
	}
	if remoteAddr == "127.0.0.1" {
		remoteAddr, err = os.Hostname()
		if err != nil {
			h.FatalError("error getting hostname: %v\n", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				if s.DataType == string(configuration.LogTypeCiscoGeneric) {
					ciscoData, err := filters.ProcessCiscoData(logBatch)
					if err != nil {
						h.Error("error processing cisco data: %v", err)
						continue
					}
					for typ, logB := range ciscoData {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: logB,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  s.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}
		case <-s.TCPListener.CTX.Done():
			return

		default:
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || err.(net.Error).Timeout() {
					return
				}
				h.Error("%v", err)
				return
			}
			message = strings.TrimSuffix(message, "\n")
			message = "[utm_stack_agent_ds=" + remoteAddr + "]-" + message
			logBatch = append(logBatch, message)

			if len(logBatch) == configuration.BatchCapacity {
				if s.DataType == string(configuration.LogTypeCiscoGeneric) {
					ciscoData, err := filters.ProcessCiscoData(logBatch)
					if err != nil {
						h.Error("error processing cisco data: %v", err)
						continue
					}
					for typ, logB := range ciscoData {
						logservice.LogQueue <- logservice.LogPipe{
							Src:  typ,
							Logs: logB,
						}
					}
				} else {
					logservice.LogQueue <- logservice.LogPipe{
						Src:  s.DataType,
						Logs: logBatch,
					}
				}
				logBatch = []string{}
			}
		}

	}
}

func (s *SyslogServer) handleConnectionTLS(c net.Conn, h *holmes.Logger) {
	defer c.Close()
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)
	reader := bufio.NewReader(c)
	remoteAddr := c.RemoteAddr().String()

	var err error
	remoteAddr, _, err = net.SplitHostPort(remoteAddr)
	if err != nil {
		h.FatalError("error spliting host and port: ", err)
	}
	if remoteAddr == "127.0.0.1" {
		remoteAddr, err = os.Hostname()
		if err != nil {
			h.FatalError("error getting hostname: %v\n", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				logservice.LogQueue <- logservice.LogPipe{
					Src:  s.DataType,
					Logs: logBatch,
				}
				logBatch = []string{}
			}
		case <-s.TLSListener.CTX.Done():
			return

		default:
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || err.(net.Error).Timeout() {
					return
				}
				h.Error("%v", err)
				return
			}
			message = strings.TrimSuffix(message, "\n")
			message = "[utm_stack_agent_ds=" + remoteAddr + "]-" + message
			logBatch = append(logBatch, message)

			if len(logBatch) == 5 {
				logservice.LogQueue <- logservice.LogPipe{
					Src:  s.DataType,
					Logs: logBatch,
				}
				logBatch = []string{}
			}
		}

	}
}
