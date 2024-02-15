package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/auth"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/migration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// Handle the panic here
			fmt.Println("Panic occurred:", r)
		}
	}()
	config.InitDb()
	migration.MigrateDatabase()

	// Create a new server instance
	s := &pb.Grpc{
		AgentStreamMap:     make(map[string]*pb.StreamAgent),
		CollectorStreamMap: make(map[string]*pb.StreamCollector),
		ResultChannel:      make(map[string]chan *pb.CommandResult),
	}
	pb.CacheAgent = make(map[uint]string)
	pb.CacheCollector = make(map[uint]string)

	err := s.LoadAgentCacheFromDatabase()
	err = s.LoadCollectorsCacheFromDatabase()

	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	// Create a listener on a specific address and port
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a gRPC server with the authInterceptor
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(recoverInterceptor),
		grpc.ChainUnaryInterceptor(auth.ConnectionKeyInterceptor),
		grpc.ChainUnaryInterceptor(auth.StreamAuthInterceptor),
		grpc.StreamInterceptor(auth.ProcessCommandInterceptor))
	pb.RegisterAgentServiceServer(grpcServer, s)
	pb.RegisterPanelServiceServer(grpcServer, s)
	pb.RegisterAgentGroupServiceServer(grpcServer, s)

	pb.RegisterCollectorServiceServer(grpcServer, s)
	pb.RegisterPingServiceServer(grpcServer, s)
	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the health status to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.InitPingSync()
	// Start the gRPC server
	log.Println("Starting gRPC server on 0.0.0.0:50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
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
			log.Printf("Panic occurred: %v", r)
			err = status.Errorf(codes.Internal, "Internal server error")
		}
	}()

	// Call the gRPC handler
	return handler(ctx, req)
}
