package main

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
)

func startHTTPServer(middlewares *Middlewares, cert string, key string) {
	// Retry logic for starting HTTP server
	maxRetries := 3
	retryDelay := 2 * time.Second

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.POST("/v1/log", middlewares.HttpAuth(), Log)
	router.POST("/v1/github-webhook", middlewares.GitHubAuth(), GitHub)
	router.GET("/v1/ping", Ping)
	router.GET("/v1/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	var loadedCert tls.Certificate
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		loadedCert, err = tls.LoadX509KeyPair(cert, key)

		if err == nil {
			break
		}

		_ = catcher.Error("failed to read the certificate files, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("could not read the certificate files, all retries failed", err, nil)
			return
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{loadedCert},
		MinVersion:   tls.VersionTLS13,
	}

	server := &http.Server{
		Addr:      ":8080",
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		_ = catcher.Error("could not start http server", err, nil)
	}

}

func Log(c *gin.Context) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := catcher.Error("failed to read request body", err, nil)
		e.GinError(c)
		return
	}

	body := buf.String()

	var l = new(plugins.Log)

	err = utils.ToObject(&body, l)
	if err != nil {
		e := catcher.Error("failed to parse log", err, nil)
		e.GinError(c)
		return
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

	c.JSON(http.StatusOK, plugins.Ack{LastId: l.Id})
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "ok"})
}

func GitHub(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := catcher.Error("failed to read request body", err, nil)
		e.GinError(c)
		return
	}

	var l = new(plugins.Log)

	l.Raw = buf.String()

	lastId := uuid.New().String()

	l.Id = lastId
	l.DataType = "github"
	l.DataSource = "github"
	l.TenantId = defaultTenant
	l.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)

	localLogsChannel <- l

	c.JSON(http.StatusOK, plugins.Ack{LastId: l.Id})
}

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
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			return catcher.Error("could not read the certificate files, all retries failed", err, nil)
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
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			return catcher.Error("all retries failed when listening to grpc", err, nil)
		}
	}

	// Serve with error handling
	if err := server.Serve(listener); err != nil {
		return catcher.Error("failed to serve grpc", err, nil)
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
				"lastId": l.Id,
			})
		}
	}
}
