package modules

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/session"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/parser"
	"github.com/utmstack/UTMStack/agent/utils"
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
}

func GetNetflowModule() *NetflowModule {
	netflowOnce.Do(func() {
		netflowModule = &NetflowModule{
			Parser:    parser.GetParser("netflow"),
			DataType:  "netflow",
			IsEnabled: false,
			Decoders:  make(map[string]*netflow.Decoder),
		}
	})
	return netflowModule
}

func (m *NetflowModule) EnablePort(proto string, enableTLS bool) error {
	if enableTLS {
		return fmt.Errorf("TLS not supported for NetFlow protocol")
	}

	if proto == "udp" && !m.IsEnabled {
		utils.Logger.Info("Server %s listening in port: %s protocol: UDP", m.DataType, config.ProtoPorts[config.DataTypeNetflow].UDP)
		m.IsEnabled = true

		port, err := strconv.Atoi(config.ProtoPorts[config.DataTypeNetflow].UDP)
		if err != nil {
			utils.Logger.ErrorF("error converting port to int: %v", err)
			return err
		}

		m.Listener, err = net.ListenUDP("udp", &net.UDPAddr{
			Port: port,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			utils.Logger.ErrorF("error listening netflow: %v", err)
			return err
		}

		m.CTX, m.Cancel = context.WithCancel(context.Background())

		buffer := make([]byte, 2048)

		go func() {
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

						var netOpErr *net.OpError
						ok := errors.As(err, &netOpErr)
						if ok && netOpErr.Timeout() {
							continue
						}

						utils.Logger.ErrorF("error connecting with netflow listener: %v", err)
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
						utils.Logger.ErrorF("error decoding NetFlow message: %v", err)
						continue
					}

					err = m.Parser.ProcessData(parser.NetflowObject{
						Remote:  addr.String(),
						Message: message,
					}, "", logservice.LogQueue)
					if err != nil {
						utils.Logger.ErrorF("error parsing netflow: %v", err)
					}
				}
			}
		}()
		return nil
	}
	return fmt.Errorf("NetFlow only supports UDP protocol")
}

func (m *NetflowModule) DisablePort(proto string) {
	if proto == "udp" && m.IsEnabled {
		utils.Logger.Info("Server %s closed in port: %s protocol: UDP", m.DataType, config.ProtoPorts[config.DataTypeNetflow].UDP)
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

func (m *NetflowModule) SetNewPort(_ string, _ string) {
	//TODO: implement this function
}

func (m *NetflowModule) GetPort(proto string) string {
	switch proto {
	case "udp":
		return config.ProtoPorts[config.DataTypeNetflow].UDP
	default:
		return ""
	}
}
