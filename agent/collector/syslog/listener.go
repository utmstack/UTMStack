package syslog

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

// listenerState holds common state for TCP and UDP listeners.
type listenerState struct {
	CTX       context.Context
	Cancel    context.CancelFunc
	IsEnabled bool
	Port      string
}

// disable cancels the context and marks the listener as disabled.
// The caller must close the actual listener and hold the appropriate lock.
func (ls *listenerState) disable() {
	if ls.Cancel != nil {
		ls.Cancel()
	}
	ls.IsEnabled = false
}

type listenerTCP struct {
	listenerState
	Listener   net.Listener
	TLSEnabled bool
}

type listenerUDP struct {
	listenerState
	Listener net.PacketConn
}

type syslogInstance struct {
	DataType    string
	TCPListener listenerTCP
	UDPListener listenerUDP
	mu          sync.RWMutex
}

// --- Port query methods ---

func (inst *syslogInstance) isPortListen(proto string) bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	switch proto {
	case "tcp":
		return inst.TCPListener.IsEnabled
	case "udp":
		return inst.UDPListener.IsEnabled
	default:
		return false
	}
}

func (inst *syslogInstance) getPort(proto string) string {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	switch proto {
	case "tcp":
		return inst.TCPListener.Port
	case "udp":
		return inst.UDPListener.Port
	default:
		return ""
	}
}

func (inst *syslogInstance) setNewPort(proto string, port string) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	switch proto {
	case "tcp":
		inst.TCPListener.Port = port
	case "udp":
		inst.UDPListener.Port = port
	}
}

// --- Port enable/disable ---

func (inst *syslogInstance) enablePort(proto string, enableTLS bool, queue chan *plugins.Log) error {
	switch proto {
	case "tcp":
		if enableTLS {
			if !fs.Exists(config.IntegrationCertPath) || !fs.Exists(config.IntegrationKeyPath) {
				return fmt.Errorf("TLS certificates not found. Please load certificates first")
			}
		}
		inst.TCPListener.TLSEnabled = enableTLS
		go inst.enableTCP(queue)
		return nil
	case "udp":
		if enableTLS {
			return fmt.Errorf("TLS not supported for UDP protocol")
		}
		go inst.enableUDP(queue)
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (inst *syslogInstance) disablePort(proto string) {
	switch proto {
	case "tcp":
		inst.disableTCP()
	case "udp":
		inst.disableUDP()
	}
}

func (inst *syslogInstance) enableTCP(queue chan *plugins.Log) {
	inst.mu.Lock()
	if inst.TCPListener.IsEnabled || inst.TCPListener.Port == "" {
		inst.mu.Unlock()
		return
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+inst.TCPListener.Port)
	if err != nil {
		inst.mu.Unlock()
		utils.Logger.ErrorF("error listening TCP in port %s: %v", inst.TCPListener.Port, err)
		return
	}

	inst.TCPListener.IsEnabled = true
	inst.TCPListener.Listener = listener
	inst.TCPListener.CTX, inst.TCPListener.Cancel = context.WithCancel(context.Background())
	inst.mu.Unlock()

	utils.Logger.Info("Server %s listening in port: %s protocol: TCP", inst.DataType, inst.TCPListener.Port)
	if inst.TCPListener.TLSEnabled {
		utils.Logger.Info("Server %s TLS enabled in port: %s protocol: TCP", inst.DataType, inst.TCPListener.Port)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in TCP listener for %s: %v", inst.DataType, r)
			}
			err = inst.TCPListener.Listener.Close()
			if err != nil {
				utils.Logger.ErrorF("error closing tcp listener: %v", err)
			}
		}()
		for {
			select {
			case <-inst.TCPListener.CTX.Done():
				return
			default:
				conn, err := inst.TCPListener.Listener.Accept()
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

				if inst.TCPListener.TLSEnabled {
					go inst.handleTLSConnection(conn, queue)
				} else {
					go inst.handleConnectionTCP(conn, queue)
				}
			}
		}
	}()
}

func (inst *syslogInstance) enableUDP(queue chan *plugins.Log) {
	inst.mu.Lock()
	if inst.UDPListener.IsEnabled || inst.UDPListener.Port == "" {
		inst.mu.Unlock()
		return
	}

	listener, err := net.ListenPacket("udp", "0.0.0.0:"+inst.UDPListener.Port)
	if err != nil {
		inst.mu.Unlock()
		utils.Logger.ErrorF("error listening UDP in port %s: %v", inst.UDPListener.Port, err)
		return
	}

	udpListener, ok := listener.(*net.UDPConn)
	if !ok {
		inst.mu.Unlock()
		utils.Logger.ErrorF("could not assert to *net.UDPConn")
		listener.Close()
		return
	}

	inst.UDPListener.IsEnabled = true
	inst.UDPListener.Listener = listener
	inst.UDPListener.CTX, inst.UDPListener.Cancel = context.WithCancel(context.Background())
	inst.mu.Unlock()

	utils.Logger.Info("Server %s listening in port: %s protocol: UDP", inst.DataType, inst.UDPListener.Port)

	buffer := make([]byte, UDPBufferSize)
	msgChannel := make(chan models.MSGDS)

	go inst.handleMessage(inst.UDPListener.CTX, msgChannel, queue)

	go func() {
		defer close(msgChannel)
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in UDP listener for %s: %v", inst.DataType, r)
			}
			if err := inst.UDPListener.Listener.Close(); err != nil {
				utils.Logger.ErrorF("error closing udp listener: %v", err)
			}
		}()
		for {
			select {
			case <-inst.UDPListener.CTX.Done():
				return
			default:
				udpListener.SetDeadline(time.Now().Add(time.Second * 1))

				n, addr, err := listener.ReadFrom(buffer)
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

				remoteAddr := resolveRemoteAddr(addr.String())
				msgChannel <- models.MSGDS{
					DataSource: remoteAddr,
					Message:    string(buffer[:n]),
				}
			}
		}
	}()
}

func (inst *syslogInstance) disableTCP() {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if !inst.TCPListener.IsEnabled {
		return
	}

	utils.Logger.Info("Server %s closed in port: %s protocol: TCP", inst.DataType, inst.TCPListener.Port)

	if inst.TCPListener.Listener != nil {
		if err := inst.TCPListener.Listener.Close(); err != nil {
			utils.Logger.ErrorF("error closing TCP listener: %v", err)
		}
	}
	inst.TCPListener.disable()
}

func (inst *syslogInstance) disableUDP() {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	if !inst.UDPListener.IsEnabled {
		return
	}

	utils.Logger.Info("Server %s closed in port: %s protocol: UDP", inst.DataType, inst.UDPListener.Port)

	if inst.UDPListener.Listener != nil {
		if err := inst.UDPListener.Listener.Close(); err != nil {
			utils.Logger.ErrorF("error closing UDP listener: %v", err)
		}
	}
	inst.UDPListener.disable()
}
