package main

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type integration struct {
	plugins.UnimplementedIntegrationServer
}

func startGRPCServer(middlewares *Middlewares, cert string, key string) error {
	// Retry logic for creating credentials
	maxRetries := 3
	retryDelay := 2 * time.Second
	var loadedCert tls.Certificate
	var err error

	for retry := 0; retry < maxRetries; retry++ {

		loadedCert, err = tls.LoadX509KeyPair(cert, key)
		if err == nil {
			break
		}

		_ = catcher.Error("failed to read the certificate files, retrying", err, map[string]any{
			"process":    "plugin_com.utmstack.inputs",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			return catcher.Error("could not read the certificate files, all retries failed", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
		}
	}

	transportCredentials := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{loadedCert},
		MinVersion:   tls.VersionTLS13,
	})

	server := grpc.NewServer(
		grpc.Creds(transportCredentials),
		grpc.ChainUnaryInterceptor(middlewares.GrpcAuth),
		grpc.ChainStreamInterceptor(middlewares.GrpcStreamAuth),
	)

	integrationInstance := new(integration)

	plugins.RegisterIntegrationServer(server, integrationInstance)
	healthServer := health.NewServer()
	grpcHealth.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpcHealth.HealthCheckResponse_SERVING)

	// Retry logic for listening to gRPC
	retryDelay = 2 * time.Second
	var listener net.Listener

	for retry := 0; retry < maxRetries; retry++ {
		listener, err = net.Listen("tcp", "0.0.0.0:50051")
		if err == nil {
			break
		}

		_ = catcher.Error("failed to listen to grpc, retrying", err, map[string]any{
			"process":    "plugin_com.utmstack.inputs",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			return catcher.Error("all retries failed when listening to grpc", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
		}
	}

	// Serve with error handling
	if err := server.Serve(listener); err != nil {
		return catcher.Error("failed to serve grpc", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
	}

	return nil
}

func (i *integration) ProcessLog(srv plugins.Integration_ProcessLogServer) error {
	for {
		l, err := srv.Recv()
		if err != nil {
			return err
		}

		if l.Id == "" {
			lastId := uuid.New().String()
			l.Id = lastId
		}

		if l.TenantId == "" {
			l.TenantId = defaultTenant
		}

		if l.DataType == "" {
			l.DataType = "generic"
		}

		if l.DataSource == "" {
			l.DataSource = "unknown"
		}

		if l.Timestamp == "" {
			l.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
		}

		localLogsChannel <- l

		if err := srv.Send(&plugins.Ack{LastId: l.Id}); err != nil {
			return catcher.Error("failed to send ack", err, map[string]any{
				"process": "plugin_com.utmstack.inputs",
				"lastId":  l.Id,
			})
		}
	}
}
