package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/log-auth-proxy/handlers"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	catcher.Info("Starting Log Auth Proxy...", nil)
	autService := logservice.NewLogAuthService()
	go autService.SyncAuth()
	authInterceptor := middleware.NewLogAuthInterceptor(autService)

	logOutputService := logservice.NewLogOutputService()
	go logOutputService.SyncOutputs()

	go startHTTPServer(authInterceptor, logOutputService)
	go startGRPCServer(authInterceptor, logOutputService)

	select {}
}

func startHTTPServer(interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/v1/log", interceptor.HTTPAuthInterceptor(), handlers.HttpLog(logOutputService))
	router.POST("/v1/logs", interceptor.HTTPAuthInterceptor(), handlers.HttpBulkLog(logOutputService))
	router.POST("/v1/github-webhook", interceptor.HTTPGitHubAuthInterceptor(), handlers.HttpGitHubHandler(logOutputService))
	router.GET("/v1/ping", handlers.HttpPing)

	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		catcher.Error("failed to load server certificates", err, nil)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}

	server := &http.Server{
		Addr:      ":8080",
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	catcher.Info("Starting HTTP server on 0.0.0.0:8080", nil)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		catcher.Error("Failed to start HTTP server", err, nil)
		os.Exit(1)
	}
}

func startGRPCServer(interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService) {
	cert, err := tls.LoadX509KeyPair("/cert/utm.crt", "/cert/utm.key")
	if err != nil {
		catcher.Error("failed to load server certificates", err, nil)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}

	creds := credentials.NewTLS(tlsConfig)

	s := &logservice.Grpc{
		OutputService: logOutputService,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptor.GrpcRecoverInterceptor),
		grpc.ChainUnaryInterceptor(interceptor.GrpcAuthInterceptor),
	)

	logservice.RegisterLogServiceServer(grpcServer, s)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		catcher.Error("failed to listen grpc server", err, nil)
		os.Exit(1)
	}

	catcher.Info("Starting gRPC server on 0.0.0.0:50051", nil)
	if err := grpcServer.Serve(lis); err != nil {
		catcher.Error("Failed to serve grpc", err, nil)
		os.Exit(1)
	}
}
