package config

import (
	"context"
	"fmt"
	"log"
	"strings"
	sync "sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	reconnectDelay = 5 * time.Second
	maxMessageSize = 1024 * 1024 * 1024
)

var (
	config      Config
	configMutex sync.RWMutex
	configOnce  sync.Once
)

type Config struct {
	Backend                   string
	InternalKey               string
	Opensearch                string
	ModulesConfigHost         string
	APIKey                    string
	ChangeAlertStatus         bool
	AutomaticIncidentCreation bool
	Provider                  string
	Model                     string
	Url                       string
	ModuleActive              bool
}

func GetConfig() *Config {
	configOnce.Do(func() {
		config = Config{}
	})
	return &config
}

func StartConfigurationSystem() {
	GetConfig()

	for {
		pluginConfig := plugins.PluginCfg("com.utmstack", false)
		if !pluginConfig.Exists() {
			_ = catcher.Error("plugin configuration not found", nil, nil)
			time.Sleep(reconnectDelay)
			continue
		}

		configMutex.Lock()
		config.Backend = pluginConfig.Get("backend").String()
		config.InternalKey = pluginConfig.Get("internalKey").String()
		config.Opensearch = pluginConfig.Get("opensearch").String()
		config.ModulesConfigHost = pluginConfig.Get("modulesConfig").String()
		configMutex.Unlock()

		if config.Backend == "" || config.InternalKey == "" || config.Opensearch == "" || config.ModulesConfigHost == "" {
			fmt.Println("Backend, Internal key, Opensearch or Modules Config Host is not set, skipping UTMStack plugin execution")
			time.Sleep(reconnectDelay)
			continue
		}
		break
	}

	for {
		connCtx, connCancel := context.WithCancel(context.Background())
		connCtx = metadata.AppendToOutgoingContext(connCtx, "internal-key", config.InternalKey)
		conn, err := grpc.NewClient(
			config.ModulesConfigHost,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
		)

		if err != nil {
			catcher.Error("Failed to connect to server", err, nil)
			connCancel()
			time.Sleep(reconnectDelay)
			continue
		}

		state := conn.GetState()
		if state == connectivity.Shutdown || state == connectivity.TransientFailure {
			catcher.Error("Connection is in shutdown or transient failure state", nil, nil)
			conn.Close()
			connCancel()
			time.Sleep(reconnectDelay)
			continue
		}

		client := NewConfigServiceClient(conn)
		stream, err := client.StreamConfig(connCtx)
		if err != nil {
			catcher.Error("Failed to create stream", err, nil)
			conn.Close()
			connCancel()
			time.Sleep(reconnectDelay)
			continue
		}

		err = stream.Send(&BiDirectionalMessage{
			Payload: &BiDirectionalMessage_PluginInit{
				PluginInit: &PluginInit{Type: PluginType_SOC_AI},
			},
		})
		if err != nil {
			catcher.Error("Failed to send PluginInit", err, nil)
			conn.Close()
			connCancel()
			time.Sleep(reconnectDelay)
			continue
		}

		for {
			in, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					catcher.Info("Stream closed by server, reconnecting...", nil)
					conn.Close()
					connCancel()
					time.Sleep(reconnectDelay)
					break
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					catcher.Error("Stream error: "+st.Message(), err, nil)
					conn.Close()
					connCancel()
					time.Sleep(reconnectDelay)
					break
				} else {
					catcher.Error("Stream receive error", err, nil)
					time.Sleep(reconnectDelay)
					continue
				}
			}

			switch message := in.Payload.(type) {
			case *BiDirectionalMessage_Config:
				log.Printf("Received configuration update: %v", message.Config)
				updateConfigFromGRPC(message.Config)
			}
		}

		conn.Close()
		connCancel()
		time.Sleep(reconnectDelay)
	}
}

func updateConfigFromGRPC(grpcConf *ConfigurationSection) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if grpcConf == nil {
		utils.Logger.LogF(100, "Received nil configuration from gRPC")
		return
	}

	config.ModuleActive = grpcConf.ModuleActive

	if len(grpcConf.ModuleGroups) == 0 {
		return
	}

	model, customModel, customURL := "", "", ""
	for _, c := range grpcConf.ModuleGroups[0].ModuleGroupConfigurations {
		switch c.ConfKey {
		case "utmstack.socai.incidentCreation":
			config.AutomaticIncidentCreation = c.ConfValue == "true"
		case "utmstack.socai.changeAlertStatus":
			config.ChangeAlertStatus = c.ConfValue == "true"
		case "utmstack.socai.provider":
			config.Provider = c.ConfValue
		case "utmstack.socai.key":
			config.APIKey = c.ConfValue
		case "utmstack.socai.model":
			model = c.ConfValue
		case "utmstack.socai.custom.model":
			customModel = c.ConfValue
		case "utmstack.socai.custom.url":
			customURL = c.ConfValue
		default:
			utils.Logger.LogF(100, "Unknown configuration key: %s", c.ConfKey)
		}
	}

	if config.Provider == "openai" {
		config.Url = GPT_API_ENDPOINT
		config.Model = model
	} else {
		config.Url = customURL
		config.Model = customModel
	}
}
