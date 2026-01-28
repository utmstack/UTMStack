package modules

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/netsampler/goflow2/decoders/netflow"
	"github.com/netsampler/goflow2/decoders/netflowlegacy"
	tehmaze "github.com/tehmaze/netflow"
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

// templateSystem implements netflow.NetFlowTemplateSystem for goflow2
type templateSystem struct {
	templates map[uint16]map[uint32]map[uint16]interface{}
	mu        sync.RWMutex
}

func newTemplateSystem() *templateSystem {
	return &templateSystem{
		templates: make(map[uint16]map[uint32]map[uint16]interface{}),
	}
}

func (t *templateSystem) GetTemplate(version uint16, obsDomainId uint32, templateId uint16) (interface{}, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if versionMap, ok := t.templates[version]; ok {
		if domainMap, ok := versionMap[obsDomainId]; ok {
			if template, ok := domainMap[templateId]; ok {
				return template, nil
			}
		}
	}
	return nil, fmt.Errorf("template not found: version=%d, obsDomainId=%d, templateId=%d", version, obsDomainId, templateId)
}

func (t *templateSystem) AddTemplate(version uint16, obsDomainId uint32, template interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.templates[version]; !ok {
		t.templates[version] = make(map[uint32]map[uint16]interface{})
	}
	if _, ok := t.templates[version][obsDomainId]; !ok {
		t.templates[version][obsDomainId] = make(map[uint16]interface{})
	}

	// Extract template ID based on type
	var templateId uint16
	switch tmpl := template.(type) {
	case netflow.TemplateRecord:
		templateId = tmpl.TemplateId
	case netflow.IPFIXOptionsTemplateRecord:
		templateId = tmpl.TemplateId
	case netflow.NFv9OptionsTemplateRecord:
		templateId = tmpl.TemplateId
	default:
		return
	}

	t.templates[version][obsDomainId][templateId] = template
}

type NetflowModule struct {
	DataType        string
	Parser          parser.Parser
	LegacyDecoders  map[string]*tehmaze.Decoder  // For v1, v6, v7 (tehmaze/netflow)
	TemplateSystem  map[string]*templateSystem   // For v5, v9, IPFIX (goflow2)
	Listener        *net.UDPConn
	CTX             context.Context
	Cancel          context.CancelFunc
	IsEnabled       bool
	mu              sync.RWMutex
}

func GetNetflowModule() *NetflowModule {
	netflowOnce.Do(func() {
		netflowModule = &NetflowModule{
			Parser:         parser.GetParser("netflow"),
			DataType:       "netflow",
			IsEnabled:      false,
			LegacyDecoders: make(map[string]*tehmaze.Decoder),
			TemplateSystem: make(map[string]*templateSystem),
		}
	})
	return netflowModule
}

func (m *NetflowModule) getOrCreateTemplateSystem(addr string) *templateSystem {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ts, ok := m.TemplateSystem[addr]; ok {
		return ts
	}
	ts := newTemplateSystem()
	m.TemplateSystem[addr] = ts
	return ts
}

func (m *NetflowModule) getOrCreateLegacyDecoder(addr string) *tehmaze.Decoder {
	m.mu.Lock()
	defer m.mu.Unlock()

	if d, ok := m.LegacyDecoders[addr]; ok {
		return d
	}
	s := session.New()
	d := tehmaze.NewDecoder(s)
	m.LegacyDecoders[addr] = d
	return d
}

func (m *NetflowModule) removeLegacyDecoder(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.LegacyDecoders, addr)
}

func (m *NetflowModule) EnablePort(proto string, enableTLS bool) error {
	if enableTLS {
		return fmt.Errorf("TLS not supported for NetFlow protocol")
	}

	if proto == "udp" && !m.IsEnabled {
		port, err := strconv.Atoi(config.ProtoPorts[config.DataTypeNetflow].UDP)
		if err != nil {
			utils.Logger.ErrorF("error converting port to int: %v", err)
			return err
		}

		listener, err := net.ListenUDP("udp", &net.UDPAddr{
			Port: port,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			utils.Logger.ErrorF("error listening netflow: %v", err)
			return err
		}

		m.IsEnabled = true
		m.Listener = listener
		m.CTX, m.Cancel = context.WithCancel(context.Background())

		utils.Logger.Info("Server %s listening in port: %s protocol: UDP", m.DataType, config.ProtoPorts[config.DataTypeNetflow].UDP)

		buffer := make([]byte, 65535)

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

					// Validate packet structure before attempting to decode
					packetData := buffer[:length]
					packetInfo, validationErr := validateNetflowPacket(packetData)
					if validationErr != nil {
						utils.Logger.ErrorF("invalid NetFlow packet from %s (length: %d bytes): %v", addr.String(), length, validationErr)
						continue
					}

					var message interface{}

					// Use hybrid approach: goflow2 for v5/v9/IPFIX, tehmaze for v1/v6/v7
					switch packetInfo.version {
					case 5:
						// Use goflow2 for NetFlow v5
						msg, err := netflowlegacy.DecodeMessage(bytes.NewBuffer(packetData))
						if err != nil {
							utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
							continue
						}
						message = msg

					case 9, 10:
						// Use goflow2 for NetFlow v9 and IPFIX
						ts := m.getOrCreateTemplateSystem(addr.String())
						msg, err := netflow.DecodeMessage(bytes.NewBuffer(packetData), ts)
						if err != nil {
							// Template not found is expected when data arrives before template
							// This is normal NetFlow v9/IPFIX behavior, don't log as error
							if !strings.Contains(err.Error(), "template not found") {
								utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
							}
							continue
						}
						message = msg

					case 1, 6, 7:
						// Use tehmaze/netflow for legacy versions (v1, v6, v7)
						d := m.getOrCreateLegacyDecoder(addr.String())
						msg, err := d.Read(bytes.NewBuffer(packetData))
						if err != nil {
							utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
							m.removeLegacyDecoder(addr.String())
							continue
						}
						message = msg

					default:
						utils.Logger.ErrorF("unsupported NetFlow version %d from %s", packetInfo.version, addr.String())
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

// netflowPacketInfo contains basic information about a NetFlow packet for validation
type netflowPacketInfo struct {
	version     uint16
	count       uint16
	minSize     int
	versionName string
}

// validateNetflowPacket checks if a NetFlow packet has valid structure before decoding
// Returns packet info if valid, error otherwise
func validateNetflowPacket(data []byte) (*netflowPacketInfo, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("packet too small: %d bytes (minimum 4 bytes for version and count)", len(data))
	}

	version := binary.BigEndian.Uint16(data[0:2])
	count := binary.BigEndian.Uint16(data[2:4])

	info := &netflowPacketInfo{
		version: version,
		count:   count,
	}

	switch version {
	case 1:
		info.versionName = "NetFlow v1"
		info.minSize = 24 + int(count)*48 // header (24) + records (48 each)
	case 5:
		info.versionName = "NetFlow v5"
		info.minSize = 24 + int(count)*48 // header (24) + records (48 each)
	case 6:
		info.versionName = "NetFlow v6"
		info.minSize = 24 + int(count)*52 // header (24) + records (52 each)
	case 7:
		info.versionName = "NetFlow v7"
		info.minSize = 24 + int(count)*52 // header (24) + records (52 each)
	case 9:
		info.versionName = "NetFlow v9"
		// NetFlow v9 header is 20 bytes, minimum packet size is just the header
		info.minSize = 20
	case 10:
		info.versionName = "IPFIX"
		// IPFIX header is 16 bytes, field at offset 2-4 is the total message length
		info.minSize = 16
		ipfixLength := binary.BigEndian.Uint16(data[2:4])
		if int(ipfixLength) != len(data) {
			return nil, fmt.Errorf("IPFIX length mismatch: header says %d bytes, received %d bytes", ipfixLength, len(data))
		}
		return info, nil
	default:
		return nil, fmt.Errorf("unsupported NetFlow version: %d", version)
	}

	if len(data) < info.minSize {
		return nil, fmt.Errorf("%s packet too small: received %d bytes, minimum expected %d bytes (count=%d)",
			info.versionName, len(data), info.minSize, count)
	}

	return info, nil
}
