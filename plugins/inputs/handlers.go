package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
)

func startHTTPServer(middlewares *Middlewares, cert string, key string) {
	go_sdk.Logger().Info("starting HTTP server on 0.0.0.0:8080...")

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.POST("/v1/log", middlewares.HttpAuth(), Log)
	router.POST("/v1/github-webhook", middlewares.GitHubAuth(), GitHub)
	router.GET("/v1/ping", Ping)
	router.GET("/v1/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	err := router.RunTLS(":8080", cert, key)
	if err != nil {
		go_sdk.Logger().ErrorF("failed to start HTTP server: %v", err)
		return
	}
}

func Log(c *gin.Context) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := go_sdk.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	body := buf.String()

	var l = new(go_sdk.Log)

	e := go_sdk.ToObject(&body, l)
	if e != nil {
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

	c.JSON(http.StatusOK, go_sdk.Ack{LastId: l.Id})
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "ok"})
}

func GitHub(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := go_sdk.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	var l = new(go_sdk.Log)

	l.Raw = buf.String()

	lastId := uuid.New().String()

	l.Id = lastId
	l.DataType = "github"
	l.DataSource = "github"
	l.TenantId = defaultTenant
	l.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)

	localLogsChannel <- l

	c.JSON(http.StatusOK, go_sdk.Ack{LastId: l.Id})
}

type integration struct {
	go_sdk.UnimplementedIntegrationServer
}

func startGRPCServer(middlewares *Middlewares, cert string, key string) {
	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		go_sdk.Logger().Fatal("failed to load TLS credentials: %v", err)
	}

	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(middlewares.GrpcAuth),
		grpc.ChainStreamInterceptor(middlewares.GrpcStreamAuth),
	)

	integrationInstance := new(integration)

	go_sdk.RegisterIntegrationServer(server, integrationInstance)
	healthServer := health.NewServer()
	grpcHealth.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpcHealth.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		go_sdk.Logger().ErrorF("failed to listen grpc server: %v", err)
		os.Exit(1)
	}

	go_sdk.Logger().Info("starting gRPC server on 0.0.0.0:50051")
	if err := server.Serve(listener); err != nil {
		go_sdk.Logger().Fatal("failed to serve grpc: %v", err)
		os.Exit(1)
	}
}

func (i *integration) ProcessLog(srv go_sdk.Integration_ProcessLogServer) error {
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

		if err := srv.Send(&go_sdk.Ack{LastId: l.Id}); err != nil {
			go_sdk.Logger().ErrorF("failed to send ack: %v", err)
			return err
		}
	}
}
