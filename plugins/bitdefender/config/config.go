package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	sync "sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/bitdefender/utils"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	DefaultTenant      string = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	EndpointPush       string = "/v1.0/jsonrpc/push"
	BitdefenderGZPort  string = "8000"
	UrlCheckConnection string = "https://cloud.gravityzone.bitdefender.com"

	reconnectDelay = 5 * time.Second
	maxMessageSize = 1024 * 1024 * 1024
)

var (
	cnf *ConfigurationSection
	mu  sync.Mutex

	internalKey       string
	modulesConfigHost string

	configsSent = make(map[string]BDGZModuleConfig)
)

func GetConfig() *ConfigurationSection {
	mu.Lock()
	defer mu.Unlock()
	if cnf == nil {
		return &ConfigurationSection{}
	}
	return cnf
}

func StartConfigurationSystem() {
	for {
		if err := utils.ConnectionChecker(UrlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, nil)
		}
		pluginConfig := plugins.PluginCfg("com.utmstack", false)
		if !pluginConfig.Exists() {
			_ = catcher.Error("plugin configuration not found", nil, nil)
			time.Sleep(reconnectDelay)
			continue
		}
		internalKey = pluginConfig.Get("internalKey").String()
		modulesConfigHost = pluginConfig.Get("modulesConfig").String()

		if internalKey == "" || modulesConfigHost == "" {
			fmt.Println("Internal key or Modules Config Host is not set, skipping UTMStack plugin execution")
			time.Sleep(reconnectDelay)
			continue
		}
		break
	}

	for {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", internalKey)
		conn, err := grpc.NewClient(
			modulesConfigHost,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
		)

		if err != nil {
			catcher.Error("Failed to connect to server", err, nil)
			cancel()
			time.Sleep(reconnectDelay)
			continue
		}

		state := conn.GetState()
		if state == connectivity.Shutdown || state == connectivity.TransientFailure {
			catcher.Error("Connection is in shutdown or transient failure state", nil, nil)
			cancel()
			time.Sleep(reconnectDelay)
			continue
		}

		client := NewConfigServiceClient(conn)
		stream, err := client.StreamConfig(ctx)
		if err != nil {
			catcher.Error("Failed to create stream", err, nil)
			conn.Close()
			cancel()
			time.Sleep(reconnectDelay)
			continue
		}

		err = stream.Send(&BiDirectionalMessage{
			Payload: &BiDirectionalMessage_PluginInit{
				PluginInit: &PluginInit{Type: PluginType_BITDEFENDER},
			},
		})
		if err != nil {
			catcher.Error("Failed to send PluginInit", err, nil)
			conn.Close()
			cancel()
			time.Sleep(reconnectDelay)
			continue
		}

		for {
			in, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					catcher.Info("Stream closed by server, reconnecting...", nil)
					conn.Close()
					cancel()
					time.Sleep(reconnectDelay)
					break
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					catcher.Error("Stream error: "+st.Message(), err, nil)
					conn.Close()
					cancel()
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
				cnf = message.Config
				go processConfigurations(cnf)
			}
		}
	}
}

func processConfigurations(config *ConfigurationSection) {
	for _, group := range config.ModuleGroups {
		newConfig := GetBDGZModuleConfig(group)
		if isNeededUpdate(configsSent, newConfig, group.GroupName) {
			if newConfig.ConnectionKey == "" || newConfig.AccessUrl == "" || newConfig.MasterIp == "" || len(newConfig.CompaniesIDs) == 0 {
				_ = catcher.Error("Invalid configuration for group", nil, map[string]any{
					"group": group.GroupName,
				})
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						_ = catcher.Error("recovered from panic in API operations", nil, map[string]any{
							"panic": r,
							"group": group.GroupName,
						})
					}
				}()

				if err := apiPush(newConfig, "sendConf"); err != nil {
					_ = catcher.Error("error sending configuration", err, map[string]any{
						"group": group.GroupName,
					})
					return
				}
				time.Sleep(15 * time.Second)
				if err := apiPush(newConfig, "getConf"); err != nil {
					_ = catcher.Error("error getting configuration", err, map[string]any{
						"group": group.GroupName,
					})
					return
				}
				if err := apiPush(newConfig, "sendTest"); err != nil {
					_ = catcher.Error("error sending test event", err, map[string]any{
						"group": group.GroupName,
					})
					return
				}

				configsSent[group.GroupName] = newConfig
			}()
		}
	}
}

func apiPush(config BDGZModuleConfig, operation string) error {
	operationFunc := map[string]func(BDGZModuleConfig) (*http.Response, error){
		"sendConf": sendPushEventSettings,
		"getConf":  getPushEventSettings,
		"sendTest": sendTestPushEvent,
	}

	fn, ok := operationFunc[operation]
	if !ok {
		return catcher.Error("wrong operation", nil, nil)
	}

	for i := 0; i < 5; i++ {
		response, err := fn(config)
		if err != nil {
			_ = catcher.Error(fmt.Sprintf("%v", err), err, nil)
			time.Sleep(1 * time.Minute)
			continue
		}

		func() { _ = response.Body.Close() }()

		return nil
	}

	return catcher.Error("error sending configuration after 5 retries", nil, nil)
}

func sendPushEventSettings(config BDGZModuleConfig) (*http.Response, error) {
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, catcher.Error("error when marshaling the request body to send the configuration", err, nil)
	}
	return sendRequest(body, config)
}

func getPushEventSettings(config BDGZModuleConfig) (*http.Response, error) {
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, catcher.Error("error when marshaling the request body to get the configuration", err, nil)
	}
	return sendRequest(body, config)
}

func sendTestPushEvent(config BDGZModuleConfig) (*http.Response, error) {
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, catcher.Error("error when marshaling the request body to send the test event", err, nil)
	}
	return sendRequest(body, config)
}

func isNeededUpdate(savedConfigs map[string]BDGZModuleConfig, newConf BDGZModuleConfig, groupName string) bool {
	cnf, ok := savedConfigs[groupName]
	if !ok {
		return true
	}

	return isDifferent(cnf.CompaniesIDs, newConf.CompaniesIDs) ||
		cnf.ConnectionKey != newConf.ConnectionKey ||
		cnf.AccessUrl != newConf.AccessUrl ||
		cnf.MasterIp != newConf.MasterIp
}

func isDifferent(a1 []string, a2 []string) bool {
	m := make(map[string]bool)
	for _, v := range a1 {
		m[v] = true
	}
	for _, v := range a2 {
		if !m[v] {
			return true
		}
		delete(m, v)
	}
	return len(m) > 0
}
