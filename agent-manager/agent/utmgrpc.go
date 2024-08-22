package agent

import (
	"net"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func InitGrpcServer() {
	err := InitAgentService()
	if err != nil {
		utils.ALogger.Fatal("failed to init agent service: %v", err)
	}

	err = InitCollectorService()
	if err != nil {
		utils.ALogger.Fatal("failed to init collector service: %v", err)
	}

	InitLastSeenService()
	InitModulesService()

	StartGrpcServer()
}

func StartGrpcServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		utils.ALogger.Fatal("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile(config.CertPath, config.CertKeyPath)
	if err != nil {
		utils.ALogger.Fatal("failed to load TLS credentials: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(UnaryInterceptor),
		grpc.StreamInterceptor(StreamInterceptor))

	RegisterAgentServiceServer(grpcServer, AgentServ)
	RegisterPanelServiceServer(grpcServer, AgentServ)
	RegisterCollectorServiceServer(grpcServer, CollectorServ)
	RegisterPanelCollectorServiceServer(grpcServer, CollectorServ)
	RegisterPingServiceServer(grpcServer, LastSeenServ)
	RegisterModuleConfigServiceServer(grpcServer, ModulesServ)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	utils.ALogger.Info("Starting gRPC server on 0.0.0.0:50051")
	if err := grpcServer.Serve(listener); err != nil {
		utils.ALogger.Fatal("failed to serve: %v", err)
	}
}
