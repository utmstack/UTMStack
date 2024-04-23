package modules

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/session"
	"github.com/threatwinds/logger"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/parser"
)

var (
	netflowModule *NetflowModule
	netflowOnce   sync.Once
)

type NetflowModule struct {
	DataType  string
	Parser    parser.Parser
	Decoders  map[string]*netflow.Decoder
	Listener  *net.UDPConn
	CTX       context.Context
	Cancel    context.CancelFunc
	IsEnabled bool
	h         *logger.Logger
}

func GetNetflowModule(h *logger.Logger) *NetflowModule {
	netflowOnce.Do(func() {
		netflowModule = &NetflowModule{
			Parser:    parser.GetParser("netflow"),
			DataType:  "netflow",
			IsEnabled: false,
			h:         h,
			Decoders:  make(map[string]*netflow.Decoder),
		}
	})
	return netflowModule
}

func (m *NetflowModule) EnablePort(proto string) {
	if proto == "udp" && !m.IsEnabled {
		m.h.Info("Server %s listening in port: %s protocol: UDP", m.DataType, configuration.ProtoPorts[configuration.LogTypeNetflow].UDP)
		m.IsEnabled = true

		port, err := strconv.Atoi(configuration.ProtoPorts[configuration.LogTypeNetflow].UDP)
		if err != nil {
			m.h.ErrorF("error converting port to int: %v", err)
			return
		}

		m.Listener, err = net.ListenUDP("udp", &net.UDPAddr{
			Port: port,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			m.h.ErrorF("error listening netflow: %v", err)
			return
		}

		m.CTX, m.Cancel = context.WithCancel(context.Background())

		buffer := make([]byte, 2048)
		msgChannel := make(chan []string)

		go m.handleConnection(msgChannel)
		go func() {
			defer m.Listener.Close()
			for {
				select {
				case <-m.CTX.Done():
					return
				default:
					m.Listener.SetDeadline(time.Now().Add(1 * time.Second))

					length, addr, err := m.Listener.ReadFromUDP(buffer)
					if err != nil {
						if errors.Is(err, net.ErrClosed) {
							return
						}

						netOpErr, err := err.(*net.OpError)
						if err && netOpErr.Timeout() {
							continue
						}

						m.h.ErrorF("error connecting with netflow listener: %v", err)
						continue
					}

					d, found := m.Decoders[addr.String()]
					if !found {
						s := session.New()
						d = netflow.NewDecoder(s)
						m.Decoders[addr.String()] = d
					}

					message, err := d.Read(bytes.NewBuffer(buffer[:length]))
					if err != nil {
						m.h.ErrorF("error decoding NetFlow message: %v", err)
						continue
					}

					logs, err := m.Parser.ProcessData(parser.NetflowObject{
						Remote:  addr.String(),
						Message: message,
					}, m.h)
					if err != nil {
						m.h.ErrorF("error parsing netflow: %v", err)
					}
					for _, bulk := range logs {
						msgChannel <- bulk
					}
				}
			}
		}()
	}
}

func (m *NetflowModule) DisablePort(proto string) {
	if proto == "udp" && m.IsEnabled {
		m.h.Info("Server %s closed in port: %s protocol: UDP", m.DataType, configuration.ProtoPorts[configuration.LogTypeNetflow].UDP)
		m.Cancel()
		m.Listener.Close()
		m.IsEnabled = false
	}
}

func (m *NetflowModule) GetDataType() string {
	return m.DataType
}

func (m *NetflowModule) IsPortListen(proto string) bool {
	switch proto {
	case "udp":
		return m.IsEnabled
	default:
		return false
	}
}

func (m *NetflowModule) SetNewPort(proto string, port string) {
}

func (m *NetflowModule) GetPort(proto string) string {
	switch proto {
	case "udp":
		return configuration.ProtoPorts[configuration.LogTypeNetflow].UDP
	default:
		return ""
	}
}

func (m *NetflowModule) handleConnection(logsChannel chan []string) {
	logBatch := []string{}
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			if len(logBatch) > 0 {
				logservice.LogQueue <- logservice.LogPipe{
					Src:  m.DataType,
					Logs: logBatch,
				}
				logBatch = []string{}
			}
		case <-m.CTX.Done():
			return
		case messages := <-logsChannel:
			for _, message := range messages {
				msg, _, err := validations.ValidateString(message, false)
				if err != nil {
					m.h.ErrorF("error validating string: %v: message: %s", err, message)
				}
				logBatch = append(logBatch, msg)
			}

			if len(logBatch) >= configuration.BatchCapacity {
				logservice.LogQueue <- logservice.LogPipe{
					Src:  m.DataType,
					Logs: logBatch,
				}
				logBatch = []string{}
			}
		}
	}
}
