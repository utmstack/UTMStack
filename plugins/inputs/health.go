package main

import (
	"context"
	"crypto/tls"
	"github.com/threatwinds/go-sdk/plugins"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func CheckAgentManagerHealth() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)

	for {
		pConfig := plugins.PluginCfg("com.utmstack", false)
		agentManager := pConfig.Get("agentManager").String()
		internalKey := pConfig.Get("internalKey").String()

		if agentManager == "" {
			panic("agentManager config is empty")
		}

		conn, err := grpc.NewClient(agentManager, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
		if err != nil {
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
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			cancel()
			_ = conn.Close()
			break
		}

		cancel()
		_ = conn.Close()
		time.Sleep(5 * time.Second)
	}
}
