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

type ConfigServer struct {
	UnimplementedConfigServiceServer

	mu      sync.RWMutex
	plugins map[PluginType][]*pluginConnection
	cache   map[PluginType]*ConfigurationSection
}

func GetConfigServer() *ConfigServer {
	configOnce.Do(func() {
		configServer = &ConfigServer{
			plugins: make(map[PluginType][]*pluginConnection),
			cache:   make(map[PluginType]*ConfigurationSection),
		}
	})
	return configServer
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
	pluginType := PluginType_UNKNOWN

	switch moduleName {
	case "AWS_IAM_USER":
		pluginType = PluginType_AWS_IAM_USER
	case "AZURE":
		pluginType = PluginType_AZURE
	case "BITDEFENDER":
		pluginType = PluginType_BITDEFENDER
	case "GCP":
		pluginType = PluginType_GCP
	case "O365":
		pluginType = PluginType_O365
	case "SOC_AI":
		pluginType = PluginType_SOC_AI
	case "SOPHOS":
		pluginType = PluginType_SOPHOS
	case "CROWDSTRIKE":
		pluginType = PluginType_CROWDSTRIKE
	default:
		catcher.Error("unknown module name", nil, map[string]any{"process": "plugin_com.utmstack.modules-config", "module": moduleName})
		return
	}

	s.mu.Lock()
	s.cache[pluginType] = section
	connectedPlugins := append([]*pluginConnection{}, s.plugins[pluginType]...)
	s.mu.Unlock()

	if len(connectedPlugins) == 0 {
		catcher.Info(fmt.Sprintf("No active connections for plugin type: %s", pluginType), map[string]any{"process": "plugin_com.utmstack.modules-config"})
		return
	}

	for _, conn := range connectedPlugins {
		err := conn.stream.Send(&BiDirectionalMessage{
			Payload: &BiDirectionalMessage_Config{
				Config: section,
			},
		})
		if err != nil {
			catcher.Error("error sending configuration update", err, map[string]any{"process": "plugin_com.utmstack.modules-config"})
			continue
		}
	}
}

func (s *ConfigServer) syncModuleWithRetry(
	moduleName string,
	pluginType PluginType,
	backend string,
	internalKey string,
) error {
	const maxRetries = 5
	baseDelay := 2 * time.Second

	url := fmt.Sprintf("%s/api/utm-modules/module-details-decrypted?nameShort=%s&serverId=1", backend, moduleName)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, status, err := utils.DoReq[ConfigurationSection](
			url,
			nil,
			"GET",
			map[string]string{"Utm-Internal-Key": internalKey},
			true,
		)

		if err == nil && status == http.StatusOK {
			s.mu.Lock()
			s.cache[pluginType] = &response
			connectedPlugins := append([]*pluginConnection{}, s.plugins[pluginType]...)
			s.mu.Unlock()

			if len(connectedPlugins) > 0 {
				for _, conn := range connectedPlugins {
					if err := conn.stream.Send(&BiDirectionalMessage{
						Payload: &BiDirectionalMessage_Config{Config: &response},
					}); err != nil {
						catcher.Error(
							"failed to send late-arrival config",
							err,
							map[string]any{
								"process": "plugin_com.utmstack.modules-config",
								"module":  moduleName,
							},
						)
					}
				}
			}

			return nil
		}

		if attempt < maxRetries {
			delay := time.Duration(1<<attempt) * baseDelay
			time.Sleep(delay)
		}
	}

	return catcher.Error("failed to sync module after max retries", nil, map[string]any{
		"process": "plugin_com.utmstack.modules-config",
		"module":  moduleName,
		"retries": maxRetries + 1,
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
		catcher.Error("module config sync failed", err, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
		})
	}
}
