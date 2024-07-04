package main

import (
	"context"
	"net"

	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	pb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/auth"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/handlers"
	"github.com/utmstack/UTMStack/agent-manager/migration"
	"github.com/utmstack/UTMStack/agent-manager/updates"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func main() {
	utils.InitLogger()

	go func() {
		// http://localhost:6060/debug/pprof/
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	defer func() {
		if r := recover(); r != nil {
			utils.ALogger.ErrorF("Panic occurred: %v", r)
		}
	}()

	config.InitDb()
	migration.MigrateDatabase()

	s, err := pb.InitGrpc()
	if err != nil {
		utils.ALogger.Fatal("failed to inititialize gRPC: %v", err)
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		utils.ALogger.Fatal("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(recoverInterceptor),
		grpc.ChainUnaryInterceptor(auth.UnaryInterceptor),
		grpc.StreamInterceptor(auth.StreamInterceptor))

	pb.RegisterAgentServiceServer(grpcServer, s)
	pb.RegisterPanelServiceServer(grpcServer, s)
	pb.RegisterAgentGroupServiceServer(grpcServer, s)

	pb.RegisterCollectorServiceServer(grpcServer, s)
	pb.RegisterPanelCollectorServiceServer(grpcServer, s)

	s.ProcessPendingConfigs()

	pb.RegisterPingServiceServer(grpcServer, s)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.InitPingSync()

	updates.ManageUpdates()
	go StartHttpServer()

	utils.ALogger.Info("Starting gRPC server on 0.0.0.0:50051")
	if err := grpcServer.Serve(lis); err != nil {
		utils.ALogger.Fatal("failed to serve: %v", err)
	}
}

func recoverInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.ALogger.ErrorF("panic occurred: %v", r)
			err = status.Errorf(codes.Internal, "Internal server error")
		}
	}()

	return handler(ctx, req)
}

func StartHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/dependencies/agent", auth.HTTPAuthInterceptor(), handlers.HandleAgentUpdates)
	router.GET("/dependencies/collector", auth.HTTPAuthInterceptor(), handlers.HandleCollectorUpdates)

	utils.ALogger.Info("Starting HTTP server on port 8080")
	err := router.Run(":8080")
	if err != nil {
		utils.ALogger.ErrorF("error starting HTTP server: %v", err)
		return
	}
}
