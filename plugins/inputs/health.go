package main

import (
	context "context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func CheckAgentManagerHealth() {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		helpers.Logger().ErrorF("failed to get the plugin config: %v", e)
		os.Exit(1)
	}

	serverAddress := pCfg.AgentManager
	if serverAddress == "" {
		helpers.Logger().ErrorF("failed to get the SERVER_ADDRESS ")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		helpers.Logger().ErrorF("failed to connect to gRPC server: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	healthClient := grpc_health_v1.NewHealthClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		cancel()

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func CheckBackendHealth() {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		helpers.Logger().ErrorF("failed to get the plugin config: %v", e)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/healthcheck", pCfg.Backend)

	for {
		resp, err := http.Get(url)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode == 200 {
			return
		}

		time.Sleep(5 * time.Second)
	}
}
