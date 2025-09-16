package config

import (
	"fmt"
	"io"
	"net/http"
	sync "sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
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
		catcher.Error("module group not found", fmt.Errorf("module: %s", moduleName), nil)
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
			catcher.Info(fmt.Sprintf("Plugin (%s) connected", pluginType), nil)

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
			catcher.Error("unexpected message type", fmt.Errorf("received: %T", payload), nil)
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
	default:
		_ = catcher.Error("unknown module name", fmt.Errorf("module: %s", moduleName), nil)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[pluginType] = section

	if len(s.plugins[pluginType]) == 0 {
		catcher.Info(fmt.Sprintf("No active connections for plugin type: %s", pluginType), nil)
		return
	}

	for _, conn := range s.plugins[pluginType] {
		err := conn.stream.Send(&BiDirectionalMessage{
			Payload: &BiDirectionalMessage_Config{
				Config: section,
			},
		})
		if err != nil {
			_ = catcher.Error("error sending configuration update", err, nil)
			continue
		}
	}
}

func (s *ConfigServer) SyncConfigs(backend string, internalKey string) {
	var AllModules = map[string]PluginType{
		"AWS_IAM_USER": PluginType_AWS_IAM_USER,
		"AZURE":        PluginType_AZURE,
		"BITDEFENDER":  PluginType_BITDEFENDER,
		"GCP":          PluginType_GCP,
		"O365":         PluginType_O365,
		"SOC_AI":       PluginType_SOC_AI,
		"SOPHOS":       PluginType_SOPHOS,
	}

	for name, t := range AllModules {
		url := fmt.Sprintf("%s/api/utm-modules/module-details-decrypted?nameShort=%s&serverId=1", backend, name)

		for {
			response, status, err := utils.DoReq[ConfigurationSection](url, nil, "GET", map[string]string{"Utm-Internal-Key": internalKey})
			if err == nil && status == http.StatusOK {
				s.mu.Lock()
				s.cache[t] = &response
				s.mu.Unlock()
				break
			}

			fmt.Printf("Error fetching configuration for %s: %v, status code: %d. Retrying...\n", name, err, status)
			time.Sleep(5 * time.Second)
		}
	}
}
