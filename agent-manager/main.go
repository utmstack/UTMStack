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
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func main() {
	go func() {
		// http://localhost:6060/debug/pprof/
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	h := util.GetLogger()

	defer func() {
		if r := recover(); r != nil {
			// Handle the panic here
			h.ErrorF("Panic occurred: %v", r)
		}
	}()

	config.InitDb()
	migration.MigrateDatabase(h)

	s, err := pb.InitGrpc()

	if err != nil {
		h.Fatal("Failed to inititialize gRPC: %v", err)
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		h.Fatal("Failed to listen: %v", err)
	}

	// Create a gRPC server with the authInterceptor.
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(recoverInterceptor),
		grpc.ChainUnaryInterceptor(auth.UnaryInterceptor),
		grpc.StreamInterceptor(auth.StreamInterceptor))

	pb.RegisterAgentServiceServer(grpcServer, s)
	pb.RegisterPanelServiceServer(grpcServer, s)
	pb.RegisterAgentGroupServiceServer(grpcServer, s)

	pb.RegisterCollectorServiceServer(grpcServer, s)
	pb.RegisterPanelCollectorServiceServer(grpcServer, s)

	s.ProcessPendingConfigs()

	agentUpdater := updates.NewAgentUpdater()
	as400Updater := updates.NewAs400Updater()
	collectorInstallerUpdater := updates.NewCollectorUpdater()
	go agentUpdater.UpdateDependencies()
	go as400Updater.UpdateDependencies()
	go collectorInstallerUpdater.UpdateDependencies()

	pb.RegisterPingServiceServer(grpcServer, s)

	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the health status to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.InitPingSync()

	go StartHttpServer()

	// Start the gRPC server
	h.Info("Starting gRPC server on 0.0.0.0:50051")
	if err := grpcServer.Serve(lis); err != nil {
		h.Fatal("Failed to serve: %v", err)
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
			// Handle the panic here
			h := util.GetLogger()
			h.ErrorF("Panic occurred: %v", r)
			err = status.Errorf(codes.Internal, "Internal server error")
		}
	}()

	// Call the gRPC handler
	return handler(ctx, req)
}

func StartHttpServer() {
	h := util.GetLogger()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/dependencies/agent", auth.HTTPAuthInterceptor(), handlers.HandleAgentUpdates)
	router.GET("/dependencies/collector", auth.HTTPAuthInterceptor(), handlers.HandleCollectorUpdates)

	h.Info("Starting HTTP server on port 8080")
	err := router.Run(":8080")
	if err != nil {
		h.ErrorF("Error starting HTTP server: %v", err)
		return
	}
}
