package collector

import (
	"context"
	"strings"
	"time"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	pb "github.com/utmstack/UTMStack/as400/agent"
	"github.com/utmstack/UTMStack/as400/config"
	"github.com/utmstack/UTMStack/as400/conn"
	"github.com/utmstack/UTMStack/as400/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConfigStreamManager struct {
	currentConfig  *AS400CollectorConfig
	onConfigChange func(*AS400CollectorConfig)
}

type AS400CollectorConfig struct {
	Servers []AS400ServerConfig `json:"servers"`
}

type AS400ServerConfig struct {
	Tenant   string `json:"tenant"`
	Hostname string `json:"hostname"`
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

func NewConfigStreamManager(onConfigChange func(*AS400CollectorConfig)) *ConfigStreamManager {
	return &ConfigStreamManager{
		onConfigChange: onConfigChange,
	}
}

func (csm *ConfigStreamManager) Start(cnf *config.Config, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			connection, err := conn.GetAgentManagerConnection(cnf)
			if err != nil {
				utils.Logger.ErrorF("error connecting to backend: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			stream, err := pb.NewCollectorServiceClient(connection).CollectorStream(ctx)
			if err != nil {
				utils.Logger.ErrorF("error opening stream: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			utils.Logger.Info("Config stream connected")
			csm.handleStream(stream)
			time.Sleep(5 * time.Second)
		}
	}
}

func (csm *ConfigStreamManager) handleStream(stream pb.CollectorService_CollectorStreamClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				return
			}
			if st, ok := status.FromError(err); ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
				return
			}
			utils.Logger.ErrorF("stream error: %v", err)
			return
		}

		if protoConfig := msg.GetConfig(); protoConfig != nil {
			csm.processConfiguration(stream, protoConfig)
		}
	}
}

func (csm *ConfigStreamManager) processConfiguration(stream pb.CollectorService_CollectorStreamClient, protoConfig *pb.CollectorConfig) {
	requestID := protoConfig.GetRequestId()

	as400Config, err := csm.protoToConfig(protoConfig)
	if err != nil {
		utils.Logger.ErrorF("invalid config: %v", err)
		csm.sendAcknowledgment(stream, requestID, false)
		return
	}

	if err := csm.validateConfig(as400Config); err != nil {
		utils.Logger.ErrorF("validation failed: %v", err)
		csm.sendAcknowledgment(stream, requestID, false)
		return
	}

	if csm.currentConfig != nil && csm.configEquals(csm.currentConfig, as400Config) {
		csm.sendAcknowledgment(stream, requestID, true)
		return
	}

	csm.currentConfig = as400Config
	utils.Logger.Info("Config received: %d servers", len(as400Config.Servers))

	if csm.onConfigChange != nil {
		csm.onConfigChange(as400Config)
	}

	csm.sendAcknowledgment(stream, requestID, true)
}

func (csm *ConfigStreamManager) protoToConfig(protoConfig *pb.CollectorConfig) (*AS400CollectorConfig, error) {
	config := &AS400CollectorConfig{
		Servers: make([]AS400ServerConfig, 0),
	}

	for _, group := range protoConfig.GetGroups() {
		server := AS400ServerConfig{
			Tenant: group.GetGroupName(),
		}

		utils.Logger.Info("Processing group: %s", server.Tenant)

		confs := group.GetConfigurations()
		utils.Logger.Info("  Configurations count: %d", len(confs))

		for _, conf := range confs {
			key := conf.GetConfKey()
			value := conf.GetConfValue()

			switch key {
			case "hostname", "collector.as400.hostname":
				server.Hostname = value
			case "userId", "collector.as400.user":
				server.UserId = value
			case "password", "collector.as400.password":
				server.Password = value
			default:
				utils.Logger.Info("  WARNING: Unknown config key '%s' (ignored)", key)
			}
		}

		config.Servers = append(config.Servers, server)
	}

	return config, nil
}

func (csm *ConfigStreamManager) validateConfig(config *AS400CollectorConfig) error {
	// Empty config is valid - means no servers to collect from
	for i, s := range config.Servers {
		if s.Tenant == "" || s.Hostname == "" || s.UserId == "" || s.Password == "" {
			return utils.Logger.ErrorF("server %d (%s): missing required fields", i, s.Tenant)
		}
	}
	return nil
}

func (csm *ConfigStreamManager) configEquals(a, b *AS400CollectorConfig) bool {
	if len(a.Servers) != len(b.Servers) {
		return false
	}

	for i := range a.Servers {
		if a.Servers[i].Tenant != b.Servers[i].Tenant ||
			a.Servers[i].Hostname != b.Servers[i].Hostname ||
			a.Servers[i].UserId != b.Servers[i].UserId ||
			a.Servers[i].Password != b.Servers[i].Password {
			return false
		}
	}

	return true
}

func (csm *ConfigStreamManager) sendAcknowledgment(stream pb.CollectorService_CollectorStreamClient, requestId string, accepted bool) {
	acceptedStr := "false"
	if accepted {
		acceptedStr = "true"
	}

	ack := &pb.CollectorMessages{
		StreamMessage: &pb.CollectorMessages_Result{
			Result: &pb.ConfigKnowledge{
				Accepted:  acceptedStr,
				RequestId: requestId,
			},
		},
	}

	if err := stream.Send(ack); err != nil {
		utils.Logger.ErrorF("ack send failed: %v", err)
	}
}

func EncryptPasswords(cfg *AS400CollectorConfig) error {
	for i := range cfg.Servers {
		encPassword, err := aesCrypt.AESEncrypt(cfg.Servers[i].Password, []byte(config.REPLACE_KEY))
		if err != nil {
			return err
		}
		cfg.Servers[i].Password = encPassword
	}

	return nil
}
