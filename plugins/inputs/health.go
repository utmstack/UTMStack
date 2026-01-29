package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

const healthMaxMessageSize = 5 * 1024 * 1024 // 5MB

func CheckAgentManagerHealth() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)

	for {
		pConfig := plugins.PluginCfg("com.utmstack")
		agentManager := pConfig.Get("agentManager").String()
		internalKey := pConfig.Get("internalKey").String()

		if agentManager == "" {
			_ = catcher.Error("Could not connect to the Agent Manager. This is a common occurrence during the startup process and typically resolves on its own after a short while.", fmt.Errorf("configuration is empty"), map[string]any{"process": "plugin_com.utmstack.inputs"})
			time.Sleep(5 * time.Second)
			continue
		}

		conn, err := grpc.NewClient(agentManager, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(healthMaxMessageSize)))
		if err != nil {
			_ = catcher.Error("Could not connect to the Agent Manager. This is a common occurrence during the startup process and typically resolves on its own after a short while.", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
			time.Sleep(5 * time.Second)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())

		ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", internalKey)

		client := grpc_health_v1.NewHealthClient(conn)

		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: ""})
		if err != nil {
			cancel()
			_ = conn.Close()
			_ = catcher.Error("Could not connect to the Agent Manager. This is a common occurrence during the startup process and typically resolves on its own after a short while.", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			cancel()
			_ = conn.Close()
			break
		}
	}
}
