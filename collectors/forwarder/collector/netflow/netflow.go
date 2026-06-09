package netflow

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
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/configwatcher"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/schema"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

const (
	cacheCleanupInterval = 5 * time.Minute
	cacheTTL             = 30 * time.Minute
)

// templateSystem implements netflow.NetFlowTemplateSystem for goflow2
type templateSystem struct {
	templates map[uint16]map[uint32]map[uint16]interface{}
	lastUsed  time.Time
	mu        sync.RWMutex
}

func newTemplateSystem() *templateSystem {
	return &templateSystem{
		templates: make(map[uint16]map[uint32]map[uint16]interface{}),
		lastUsed:  time.Now(),
	}
}

// legacyDecoderEntry wraps a decoder with its last used timestamp
type legacyDecoderEntry struct {
	decoder  *tehmaze.Decoder
	lastUsed time.Time
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

// NetflowCollector manages the single netflow UDP listener.
// It reads the config file periodically and reconciles port state internally.
type NetflowCollector struct {
	dataType       string
	legacyDecoders map[string]*legacyDecoderEntry
	templateSystem map[string]*templateSystem
	listener       *net.UDPConn
	ctx            context.Context
	cancel         context.CancelFunc
	isEnabled      bool
	port           string
	mu             sync.RWMutex
	queue          chan *plugins.Log
}

// New creates a new NetflowCollector.
func New() *NetflowCollector {
	return &NetflowCollector{
		dataType:       "netflow",
		legacyDecoders: make(map[string]*legacyDecoderEntry),
		templateSystem: make(map[string]*templateSystem),
		isEnabled:      false,
		port:           config.ProtoPorts[config.DataTypeNetflow].UDP,
	}
}

func (nc *NetflowCollector) Name() string {
	return "netflow"
}

func (nc *NetflowCollector) Stop() {
	nc.disablePort()
}

// Start begins watching for configuration changes using fsnotify.
// It performs an initial reconciliation and then reacts to config file changes.
func (nc *NetflowCollector) Start(ctx context.Context, queue chan *plugins.Log) {
	nc.queue = queue

	// Start cache cleanup goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in netflow cache cleanup: %v", r)
			}
		}()

		ticker := time.NewTicker(cacheCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				nc.cleanupStaleEntries()
			}
		}
	}()

	configwatcher.Watch(ctx, "netflow collector", nc.reconcile)
}

func (nc *NetflowCollector) reconcile() {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		utils.Logger.ErrorF("error reading collector config: %v", err)
		return
	}

	integration, ok := cnf.Integrations["netflow"]
	if !ok {
		return
	}

	cfgEnabled := integration.UDP.IsListen
	cfgPort := integration.UDP.Port

	nc.mu.RLock()
	isListening := nc.isEnabled
	currentPort := nc.port
	nc.mu.RUnlock()

	needKill := false
	needStart := false

	if isListening && !cfgEnabled {
		needKill = true
	} else if !isListening && cfgEnabled {
		needStart = true
	} else if isListening && cfgEnabled && currentPort != cfgPort {
		needKill = true
		needStart = true
	}

	if needKill {
		nc.disablePort()
		if needStart {
			time.Sleep(200 * time.Millisecond)
		}
	}

	if cfgPort != "" {
		nc.mu.Lock()
		nc.port = cfgPort
		nc.mu.Unlock()
	}

	if needStart {
		nc.enablePort()
	}
}

func (nc *NetflowCollector) enablePort() {
	nc.mu.Lock()
	if nc.isEnabled {
		nc.mu.Unlock()
		return
	}

	port, err := strconv.Atoi(nc.port)
	if err != nil {
		nc.mu.Unlock()
		utils.Logger.ErrorF("error converting port to int: %v", err)
		return
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		nc.mu.Unlock()
		utils.Logger.ErrorF("error listening netflow: %v", err)
		return
	}

	nc.isEnabled = true
	nc.listener = listener
	nc.ctx, nc.cancel = context.WithCancel(context.Background())
	nc.mu.Unlock()

	utils.Logger.Info("Server %s listening in port: %s protocol: UDP", nc.dataType, nc.port)

	buffer := make([]byte, 65535)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in netflow listener: %v", r)
			}
		}()

		for {
			select {
			case <-nc.ctx.Done():
				return
			default:
				nc.listener.SetDeadline(time.Now().Add(1 * time.Second))

				length, addr, err := nc.listener.ReadFromUDP(buffer)
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

				packetData := buffer[:length]
				packetInfo, validationErr := validateNetflowPacket(packetData)
				if validationErr != nil {
					utils.Logger.ErrorF("invalid NetFlow packet from %s (length: %d bytes): %v", addr.String(), length, validationErr)
					continue
				}

				var message interface{}

				switch packetInfo.version {
				case 5:
					msg, err := netflowlegacy.DecodeMessage(bytes.NewBuffer(packetData))
					if err != nil {
						utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
						continue
					}
					message = msg

				case 9, 10:
					ts := nc.getOrCreateTemplateSystem(addr.String())
					msg, err := netflow.DecodeMessage(bytes.NewBuffer(packetData), ts)
					if err != nil {
						if !strings.Contains(err.Error(), "template not found") {
							utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
						}
						continue
					}
					message = msg

				case 1, 6, 7:
					d := nc.getOrCreateLegacyDecoder(addr.String())
					msg, err := d.Read(bytes.NewBuffer(packetData))
					if err != nil {
						utils.Logger.ErrorF("error decoding %s message from %s: %v", packetInfo.versionName, addr.String(), err)
						nc.removeLegacyDecoder(addr.String())
						continue
					}
					message = msg

				default:
					utils.Logger.ErrorF("unsupported NetFlow version %d from %s", packetInfo.version, addr.String())
					continue
				}

				err = processMessage(addr.String(), message, nc.queue)
				if err != nil {
					utils.Logger.ErrorF("error parsing netflow: %v", err)
				}
			}
		}
	}()
}

func (nc *NetflowCollector) disablePort() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if nc.isEnabled {
		utils.Logger.Info("Server %s closed in port: %s protocol: UDP", nc.dataType, nc.port)
		nc.cancel()
		nc.listener.Close()
		nc.isEnabled = false
	}
}

func (nc *NetflowCollector) getOrCreateTemplateSystem(addr string) *templateSystem {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if ts, ok := nc.templateSystem[addr]; ok {
		ts.lastUsed = time.Now()
		return ts
	}
	ts := newTemplateSystem()
	nc.templateSystem[addr] = ts
	return ts
}

func (nc *NetflowCollector) getOrCreateLegacyDecoder(addr string) *tehmaze.Decoder {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if entry, ok := nc.legacyDecoders[addr]; ok {
		entry.lastUsed = time.Now()
		return entry.decoder
	}
	s := session.New()
	d := tehmaze.NewDecoder(s)
	nc.legacyDecoders[addr] = &legacyDecoderEntry{
		decoder:  d,
		lastUsed: time.Now(),
	}
	return d
}

func (nc *NetflowCollector) removeLegacyDecoder(addr string) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	delete(nc.legacyDecoders, addr)
}

// cleanupStaleEntries removes entries that haven't been used recently
func (nc *NetflowCollector) cleanupStaleEntries() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	now := time.Now()

	// Cleanup legacy decoders
	for addr, entry := range nc.legacyDecoders {
		if now.Sub(entry.lastUsed) > cacheTTL {
			delete(nc.legacyDecoders, addr)
			utils.Logger.Info("Removed stale legacy decoder for %s", addr)
		}
	}

	// Cleanup template systems
	for addr, ts := range nc.templateSystem {
		if now.Sub(ts.lastUsed) > cacheTTL {
			delete(nc.templateSystem, addr)
			utils.Logger.Info("Removed stale template system for %s", addr)
		}
	}
}

// --- Packet validation ---

type netflowPacketInfo struct {
	version     uint16
	count       uint16
	minSize     int
	versionName string
}

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
		info.minSize = 24 + int(count)*48
	case 5:
		info.versionName = "NetFlow v5"
		info.minSize = 24 + int(count)*48
	case 6:
		info.versionName = "NetFlow v6"
		info.minSize = 24 + int(count)*52
	case 7:
		info.versionName = "NetFlow v7"
		info.minSize = 24 + int(count)*52
	case 9:
		info.versionName = "NetFlow v9"
		info.minSize = 20
	case 10:
		info.versionName = "IPFIX"
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
