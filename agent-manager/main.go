package main

import (
	"context"
	"crypto/tls"
	"net"

	_ "net/http/pprof"

	pb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/auth"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/migration"
	"github.com/utmstack/UTMStack/agent-manager/updates"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func main() {
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

	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		h.Fatal("failed to load server certificates: %v", err)
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

	// Start the gRPC server
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		h.Fatal("Failed to listen: %v", err)
	}

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
