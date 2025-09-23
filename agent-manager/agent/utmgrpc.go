package agent

import (
	"crypto/tls"
	"net"
	"os"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/config"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func InitGrpcServer() {
	err := InitAgentService()
	if err != nil {
		catcher.Error("failed to init agent service", err, nil)
		os.Exit(1)
	}

	go InitCollectorService()
	InitLastSeenService()

	StartGrpcServer()
}

func StartGrpcServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		catcher.Error("failed to listen", err, nil)
		os.Exit(1)
	}

	loadedCert, err := tls.LoadX509KeyPair(config.CertPath, config.CertKeyPath)
	if err != nil {
		catcher.Error("failed to load TLS credentials: %v", err, nil)
		os.Exit(1)
	}

	transportCredentials := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{loadedCert},
		MinVersion:   tls.VersionTLS13,
	})

	grpcServer := grpc.NewServer(
		grpc.Creds(transportCredentials),
		grpc.ChainUnaryInterceptor(UnaryInterceptor),
		grpc.StreamInterceptor(StreamInterceptor))

	RegisterAgentServiceServer(grpcServer, AgentServ)
	RegisterPanelServiceServer(grpcServer, AgentServ)
	RegisterCollectorServiceServer(grpcServer, CollectorServ)
	RegisterPanelCollectorServiceServer(grpcServer, CollectorServ)
	RegisterPingServiceServer(grpcServer, LastSeenServ)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	catcher.Info("Starting gRPC server on 0.0.0.0:50051", nil)
	if err := grpcServer.Serve(listener); err != nil {
		catcher.Error("failed to serve", err, nil)
		os.Exit(1)
	}
}
