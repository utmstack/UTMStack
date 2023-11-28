package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/handlers"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	autService := logservice.NewLogAuthService()
	go autService.SyncAuth()
	authInterceptor := middleware.NewLogAuthInterceptor(autService)

	cert, key, err := loadCerts()
	if err != nil {
		log.Fatal(err)
	}

	logOutputService := logservice.NewLogOutputService()
	go logOutputService.SyncOutputs()

	go startHTTPServer(authInterceptor, logOutputService, cert, key)
	// go startTCPServer(authInterceptor, logOutputService, cert, key)
	go startGRPCServer(authInterceptor, logOutputService)

	// Block the main thread until an interrupt is received
	select {}
}

func startHTTPServer(interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService, cert string, key string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/v1/log", interceptor.HTTPAuthInterceptor(), handlers.HttpLog(logOutputService))
	router.POST("/v1/logs", interceptor.HTTPAuthInterceptor(), handlers.HttpBulkLog(logOutputService))
	router.POST("/v1/github-webhook", interceptor.HTTPGitHubAuthInterceptor(), handlers.HttpGitHubHandler(logOutputService))
	router.GET("/v1/ping", handlers.HttpPing)
	err := router.RunTLS(":8080", cert, key)
	log.Println("Starting HTTP server on 0.0.0.0:8080")
	if err != nil {
		return
	}
}

func startTCPServer(interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService, certFile string, keyFile string) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	tcpTLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	l, err := tls.Listen("tcp", ":2056", tcpTLSConfig)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()

	log.Println("Starting TCP server on 0.0.0.0:8081")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}
		go handlers.HandleRequest(conn, interceptor, logOutputService)
	}
}

func loadCerts() (string, string, error) {
	certsLocation := os.Getenv(config.UTMCertsLocationEnv)
	certPath := filepath.Join(certsLocation, config.UTMCertFileName)
	keyPath := filepath.Join(certsLocation, config.UTMCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}
	return certPath, keyPath, nil
}

func startGRPCServer(interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService) {

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := &logservice.Grpc{
		OutputService: logOutputService,
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor.GrpcRecoverInterceptor),
		grpc.ChainUnaryInterceptor(interceptor.GrpcAuthInterceptor))
	logservice.RegisterLogServiceServer(grpcServer, s)

	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	// Start the gRPC server
	log.Println("Starting gRPC server on 0.0.0.0:50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
