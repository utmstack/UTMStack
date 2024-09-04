package main

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func CheckAgentManagerHealth() {
	pCfg, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		go_sdk.Logger().ErrorF("failed to get the PluginConfig: %v", e)
		os.Exit(1)
	}

	serverAddress := pCfg.AgentManager
	if serverAddress == "" {
		go_sdk.Logger().ErrorF("failed to get the SERVER_ADDRESS ")
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)

	for {
		conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", pCfg.InternalKey)

		client := grpc_health_v1.NewHealthClient(conn)

		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: ""})
		if err != nil {
			cancel()
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			cancel()
			conn.Close()
			break
		}

		cancel()
		conn.Close()
		time.Sleep(5 * time.Second)
	}
}
