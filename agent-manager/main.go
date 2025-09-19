package main

import (
	"context"
	"crypto/tls"
	"net"
	"os"

	_ "net/http/pprof"

	"github.com/threatwinds/go-sdk/catcher"
	pb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/auth"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/migration"
	"github.com/utmstack/UTMStack/agent-manager/updates"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func main() {
	catcher.Info("Starting UTMStack Agent Manager", map[string]any{})

	defer func() {
		if r := recover(); r != nil {
			// Handle the panic here
			catcher.Error("Panic occurred", nil, map[string]any{"message": r})
		}
	}()

	catcher.Info("Initializing database...", map[string]any{})
	config.InitDb()
	migration.MigrateDatabase()
	catcher.Info("[OK] Database initialized", map[string]any{})

	s, err := pb.InitGrpc()
	if err != nil {
		catcher.Error("Failed to initialize gRPC", err, map[string]any{})
		os.Exit(1)
	}

	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		catcher.Error("failed to load server certificates", err, map[string]any{})
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}

	creds := credentials.NewTLS(tlsConfig)

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(recoverInterceptor),
		grpc.ChainUnaryInterceptor(auth.UnaryInterceptor),
		grpc.StreamInterceptor(auth.StreamInterceptor),
	)

	pb.RegisterAgentServiceServer(grpcServer, s)
	pb.RegisterPanelServiceServer(grpcServer, s)
	pb.RegisterAgentGroupServiceServer(grpcServer, s)

	pb.RegisterCollectorServiceServer(grpcServer, s)
	pb.RegisterPanelCollectorServiceServer(grpcServer, s)
	s.ProcessPendingConfigs()

	pb.RegisterPingServiceServer(grpcServer, s)

	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the health status to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.InitPingSync()
	updates.InitUpdatesManager()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		catcher.Error("Failed to listen", err, map[string]any{})
		os.Exit(1)
	}

	catcher.Info("Starting gRPC server on 0.0.0.0:50051", map[string]any{})
	if err := grpcServer.Serve(lis); err != nil {
		catcher.Error("Failed to serve", err, map[string]any{})
		os.Exit(1)
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
			catcher.Error("Panic occurred", nil, map[string]any{"message": r})
			err = status.Errorf(codes.Internal, "Internal server error")
		}
	}()

	return handler(ctx, req)
}
