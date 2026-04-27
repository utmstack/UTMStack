package config

import (
	"fmt"
	"io"
	"net/http"
	sync "sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
	"golang.org/x/sync/errgroup"
)

var (
	configServer *ConfigServer
	configOnce   sync.Once
)

type pluginConnection struct {
	stream ConfigService_StreamConfigServer
	done   <-chan struct{}
}

type failedModule struct {
	moduleName  string
	pluginType  PluginType
	lastAttempt time.Time
	retries     int
	lastError   string
}

type Decrypter func(*ConfigurationSection) error

type ConfigServer struct {
	UnimplementedConfigServiceServer

	mu            sync.RWMutex
	plugins       map[PluginType][]*pluginConnection
	cache         map[PluginType]*ConfigurationSection
	failedModules map[PluginType]*failedModule
	failedMu      sync.RWMutex
	decrypt       Decrypter
}

func GetConfigServer() *ConfigServer {
	configOnce.Do(func() {
		configServer = &ConfigServer{
			plugins:       make(map[PluginType][]*pluginConnection),
			cache:         make(map[PluginType]*ConfigurationSection),
			failedModules: make(map[PluginType]*failedModule),
		}
	})
	return configServer
}

func (s *ConfigServer) SetDecrypter(d Decrypter) {
	s.decrypt = d
}

func (s *ConfigServer) runDecrypter(section *ConfigurationSection) error {
	if s.decrypt == nil {
		return nil
	}
	return s.decrypt(section)
}

func (s *ConfigServer) GetModuleGroup(moduleName PluginType) *ConfigurationSection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	section, exists := s.cache[moduleName]
	if !exists {
		catcher.Error("module group not found", nil, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		},
		)
		return nil
	}

	return section
}

func (s *ConfigServer) StreamConfig(stream ConfigService_StreamConfigServer) error {
	ctx := stream.Context()
	var pluginType PluginType
	conn := &pluginConnection{stream: stream, done: ctx.Done()}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch payload := msg.Payload.(type) {
		case *BiDirectionalMessage_PluginInit:
			pluginType = payload.PluginInit.Type
			catcher.Info(fmt.Sprintf("Plugin (%s) connected", pluginType), map[string]any{"process": "plugin_com.utmstack.modules-config"})

			s.mu.Lock()
			s.plugins[pluginType] = append(s.plugins[pluginType], conn)
			s.mu.Unlock()

			s.mu.RLock()
			section := s.cache[pluginType]
			s.mu.RUnlock()
			if section != nil {
				_ = stream.Send(&BiDirectionalMessage{
					Payload: &BiDirectionalMessage_Config{
						Config: section,
					},
				})
			}

			go s.monitorDisconnect(pluginType, conn)

		default:
			catcher.Error("unexpected message type", nil, map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"message_type": fmt.Sprintf("%T", payload),
			},
			)
		}
	}

	return nil
}

func (s *ConfigServer) monitorDisconnect(t PluginType, conn *pluginConnection) {
	<-conn.done
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.plugins[t]
	updated := []*pluginConnection{}
	for _, c := range list {
		if c != conn {
			updated = append(updated, c)
		}
	}
	s.plugins[t] = updated
}

func (s *ConfigServer) NotifyUpdate(moduleName string, section *ConfigurationSection) {
	pluginType, exists := AllModules[moduleName]
	if !exists {
		catcher.Error("unknown module name", nil, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		})
		return
	}

	connections := s.updateCache(pluginType, section)
	s.clearFailedModule(pluginType, moduleName)
	s.notifyConnectedPlugins(connections, section, moduleName)
}

func (s *ConfigServer) fetchModuleConfig(backend, moduleName, internalKey string) (*ConfigurationSection, int, error) {
	url := fmt.Sprintf("%s/api/utm-modules/moduleDetails?nameShort=%s&serverId=1", backend, moduleName)
	// Remove after testing, before release to production
	catcher.Info("fetchModuleConfig: requesting", map[string]any{
		"process":        "plugin_com.utmstack.modules-config",
		"module":         moduleName,
		"url":            url,
		"internalKey":    internalKey,
		"internalKeyLen": len(internalKey),
	})

	response, status, err := utils.DoReq[ConfigurationSection](
		url,
		nil,
		"GET",
		map[string]string{"Utm-Internal-Key": internalKey},
		true,
	)
	// Remove after testing, before release to production
	catcher.Info("fetchModuleConfig: response received", map[string]any{
		"process":    "plugin_com.utmstack.modules-config",
		"module":     moduleName,
		"status":     status,
		"err":        fmt.Sprintf("%v", err),
		"groupCount": len(response.ModuleGroups),
		"moduleName": response.ModuleName,
	})

	if err != nil || status != http.StatusOK {
		return nil, status, err
	}
	// Remove after testing, before release to production
	for _, g := range response.ModuleGroups {
		for _, cnf := range g.ModuleGroupConfigurations {
			catcher.Info("fetchModuleConfig: incoming field (pre-decrypt)", map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"module":       moduleName,
				"groupId":      g.Id,
				"confKey":      cnf.ConfKey,
				"confDataType": cnf.ConfDataType,
				"valueLen":     len(cnf.ConfValue),
				"confValue":    cnf.ConfValue,
			})
		}
	}

	if err := s.runDecrypter(&response); err != nil {
		return nil, status, catcher.Error("failed to decrypt module config", err, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		})
	}
	// Remove after testing, before release to production
	for _, g := range response.ModuleGroups {
		for _, cnf := range g.ModuleGroupConfigurations {
			catcher.Info("fetchModuleConfig: field (post-decrypt)", map[string]any{
				"process":      "plugin_com.utmstack.modules-config",
				"module":       moduleName,
				"groupId":      g.Id,
				"confKey":      cnf.ConfKey,
				"confDataType": cnf.ConfDataType,
				"valueLen":     len(cnf.ConfValue),
				"confValue":    cnf.ConfValue,
			})
		}
	}

	return &response, status, nil
}

func (s *ConfigServer) updateCache(pluginType PluginType, config *ConfigurationSection) []*pluginConnection {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[pluginType] = config
	return append([]*pluginConnection{}, s.plugins[pluginType]...)
}

func (s *ConfigServer) clearFailedModule(pluginType PluginType, moduleName string) {
	s.failedMu.Lock()
	defer s.failedMu.Unlock()

	if _, exists := s.failedModules[pluginType]; exists {
		delete(s.failedModules, pluginType)
		catcher.Info(
			fmt.Sprintf("Module %s removed from retry list after successful config update", moduleName), map[string]any{
				"process":    "plugin_com.utmstack.modules-config",
				"pluginType": pluginType,
			},
		)
	}
}

func (s *ConfigServer) notifyConnectedPlugins(connections []*pluginConnection, config *ConfigurationSection, moduleName string) {
	if len(connections) == 0 {
		catcher.Info("No active connections for plugin type", map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		},
		)
		return
	}

	for _, conn := range connections {
		if err := conn.stream.Send(&BiDirectionalMessage{
			Payload: &BiDirectionalMessage_Config{Config: config},
		}); err != nil {
			catcher.Error("failed to send config update", err, map[string]any{
				"process": "plugin_com.utmstack.modules-config",
				"module":  moduleName,
			},
			)
		}
	}
}

func (s *ConfigServer) trackFailure(pluginType PluginType, moduleName string, err error) {
	s.failedMu.Lock()
	defer s.failedMu.Unlock()

	errMsg := "unknown error"
	if err != nil {
		errMsg = err.Error()
	}

	if existing, exists := s.failedModules[pluginType]; exists {
		existing.retries++
		existing.lastAttempt = time.Now()
		existing.lastError = errMsg
	} else {
		s.failedModules[pluginType] = &failedModule{
			moduleName:  moduleName,
			pluginType:  pluginType,
			lastAttempt: time.Now(),
			retries:     1,
			lastError:   errMsg,
		}
	}
}

func (s *ConfigServer) syncModuleWithRetry(moduleName string, pluginType PluginType, backend string, internalKey string) error {
	const maxRetries = 5
	baseDelay := 2 * time.Second

	var lastErr error
	var lastStatus int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		config, status, err := s.fetchModuleConfig(backend, moduleName, internalKey)
		lastStatus = status
		lastErr = err

		if err == nil && status == http.StatusOK {
			connections := s.updateCache(pluginType, config)
			s.clearFailedModule(pluginType, moduleName)
			s.notifyConnectedPlugins(connections, config, moduleName)
			return nil
		}

		if attempt < maxRetries {
			delay := time.Duration(1<<attempt) * baseDelay
			time.Sleep(delay)
		}
	}

	s.trackFailure(pluginType, moduleName, lastErr)

	return catcher.Error("failed to sync module after max retries, will retry periodically", lastErr, map[string]any{
		"process":     "plugin_com.utmstack.modules-config",
		"module":      moduleName,
		"pluginType":  pluginType,
		"retries":     maxRetries + 1,
		"status_code": lastStatus,
	},
	)
}

var AllModules = map[string]PluginType{
	"AWS_IAM_USER": PluginType_AWS_IAM_USER,
	"AZURE":        PluginType_AZURE,
	"BITDEFENDER":  PluginType_BITDEFENDER,
	"GCP":          PluginType_GCP,
	"O365":         PluginType_O365,
	"SOC_AI":       PluginType_SOC_AI,
	"SOPHOS":       PluginType_SOPHOS,
	"CROWDSTRIKE":  PluginType_CROWDSTRIKE,
}

func (s *ConfigServer) SyncConfigs(backend string, internalKey string) {
	g := errgroup.Group{}
	g.SetLimit(4)

	for name, t := range AllModules {
		g.Go(func() error {
			return s.syncModuleWithRetry(name, t, backend, internalKey)
		})
	}

	if err := g.Wait(); err != nil {
		catcher.Error("module config sync failed", err, map[string]any{"process": "plugin_com.utmstack.modules-config"})
	}
}

func (s *ConfigServer) StartPeriodicRetry(backend string, internalKey string) {
	const retryInterval = 5 * time.Minute

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.failedMu.RLock()
		toRetry := make([]*failedModule, 0, len(s.failedModules))
		for _, fm := range s.failedModules {
			toRetry = append(toRetry, fm)
		}
		s.failedMu.RUnlock()

		if len(toRetry) == 0 {
			continue
		}

		catcher.Info(fmt.Sprintf("Retrying %d failed module(s)", len(toRetry)), map[string]any{"process": "plugin_com.utmstack.modules-config"})

		for _, fm := range toRetry {
			err := s.syncModuleWithRetry(fm.moduleName, fm.pluginType, backend, internalKey)

			if err != nil {
				s.failedMu.RLock()
				currentRetries := 0
				if existing, ok := s.failedModules[fm.pluginType]; ok {
					currentRetries = existing.retries
				}
				s.failedMu.RUnlock()

				catcher.Error(fmt.Sprintf("Module sync retry failed (attempt %d)", currentRetries), err, map[string]any{
					"process":    "plugin_com.utmstack.modules-config",
					"module":     fm.moduleName,
					"pluginType": fm.pluginType,
					"retries":    currentRetries,
				},
				)
			}
		}
	}
}
