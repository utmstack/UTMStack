package syslog

import (
	"context"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/collector/configwatcher"
	"github.com/utmstack/UTMStack/agent/collector/schema"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

// SyslogCollector manages all syslog instances. It reads the config file
// periodically and reconciles port state internally.
type SyslogCollector struct {
	instances map[string]*syslogInstance
	mu        sync.RWMutex
	queue     chan *plugins.Log
}

// New creates a new SyslogCollector.
func New() *SyslogCollector {
	return &SyslogCollector{
		instances: make(map[string]*syslogInstance),
	}
}

func (sc *SyslogCollector) Name() string {
	return "syslog"
}

func (sc *SyslogCollector) Stop() {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	for _, inst := range sc.instances {
		inst.disableTCP()
		inst.disableUDP()
	}
}

// Start begins watching for configuration changes using fsnotify.
// It performs an initial reconciliation and then reacts to config file changes.
func (sc *SyslogCollector) Start(ctx context.Context, queue chan *plugins.Log) {
	sc.queue = queue
	configwatcher.Watch(ctx, "syslog collector", sc.reconcile)
}

func (sc *SyslogCollector) reconcile() {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		utils.Logger.ErrorF("error reading collector config: %v", err)
		return
	}

	for intType, integration := range cnf.Integrations {
		// Only handle syslog-type integrations
		if config.ValidateModuleType(intType) != "syslog" {
			continue
		}

		inst := sc.getOrCreateInstance(intType)

		sc.reconcileProto(inst, intType, "tcp", integration)
		sc.reconcileProto(inst, intType, "udp", integration)
	}
}

func (sc *SyslogCollector) getOrCreateInstance(dataType string) *syslogInstance {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if inst, ok := sc.instances[dataType]; ok {
		return inst
	}

	inst := &syslogInstance{
		DataType: dataType,
		TCPListener: listenerTCP{
			listenerState: listenerState{
				IsEnabled: false,
				Port:      config.ProtoPorts[config.DataType(dataType)].TCP,
			},
		},
		UDPListener: listenerUDP{
			listenerState: listenerState{
				IsEnabled: false,
				Port:      config.ProtoPorts[config.DataType(dataType)].UDP,
			},
		},
	}
	sc.instances[dataType] = inst
	return inst
}

func (sc *SyslogCollector) reconcileProto(inst *syslogInstance, intType string, proto string, integration schema.Integration) {
	var cfgEnabled bool
	var cfgPort string
	var cfgTLS bool

	switch proto {
	case "tcp":
		cfgEnabled = integration.TCP.IsListen
		cfgPort = integration.TCP.Port
		cfgTLS = integration.TCP.TLSEnabled
	case "udp":
		cfgEnabled = integration.UDP.IsListen
		cfgPort = integration.UDP.Port
	}

	isListening := inst.isPortListen(proto)
	currentPort := inst.getPort(proto)

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

	changeAllowed := true
	if cfgPort != "" && currentPort != cfgPort {
		changeAllowed = schema.ValidatePortChange(cfgPort)
	}

	if needKill {
		inst.disablePort(proto)
		if needStart {
			time.Sleep(200 * time.Millisecond)
		}
	}

	if changeAllowed {
		inst.setNewPort(proto, cfgPort)
		if needStart {
			enableTLS := proto == "tcp" && cfgTLS
			err := inst.enablePort(proto, enableTLS, sc.queue)
			if err != nil {
				utils.Logger.ErrorF("error enabling port for %s %s: %v", intType, proto, err)
			}
		}
	} else {
		utils.Logger.Info("port %s is out of valid range %s-%s", cfgPort, config.PortRangeMin, config.PortRangeMax)
		sc.writeConfigFromInstances()
	}
}

// writeConfigFromInstances writes the current live state back to the config file (rollback).
func (sc *SyslogCollector) writeConfigFromInstances() {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Read current config to preserve FileIntegrations
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		cnf = schema.CollectorConfig{
			Integrations: make(map[string]schema.Integration),
		}
	}

	// Update only the syslog integrations
	for _, inst := range sc.instances {
		cnf.Integrations[inst.DataType] = schema.Integration{
			TCP: schema.Port{
				IsListen: inst.isPortListen("tcp"),
				Port:     inst.getPort("tcp"),
			},
			UDP: schema.Port{
				IsListen: inst.isPortListen("udp"),
				Port:     inst.getPort("udp"),
			},
		}
	}
	if err := schema.WriteCollectorConfig(&cnf); err != nil {
		utils.Logger.ErrorF("error fixing collector config: %v", err)
	}
}
