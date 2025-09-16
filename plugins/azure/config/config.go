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
	cnf *ConfigurationSection
	mu  sync.Mutex

	internalKey       string
	modulesConfigHost string
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
				PluginInit: &PluginInit{Type: PluginType_AZURE},
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
			}
		}
	}
}
