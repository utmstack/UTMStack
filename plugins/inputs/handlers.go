package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func startHTTPServer(middlewares *Middlewares, cert string, key string) {
	helpers.Logger().Info("starting HTTP server on 0.0.0.0:8080...")

	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()
	router.POST("/v1/log", middlewares.HttpAuth(), Log)
	router.POST("/v1/github-webhook", middlewares.GitHubAuth(), GitHub)
	router.GET("/v1/ping", Ping)
	router.GET("/v1/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	
	err := router.RunTLS(":8080", cert, key)
	if err != nil {
		helpers.Logger().ErrorF("failed to start HTTP server: %v", err)
		return
	}
}

func Log(c *gin.Context) {
	buf := new(bytes.Buffer)
	
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := helpers.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	var l = new(plugins.Log)
	
	err = protojson.Unmarshal(buf.Bytes(), l)
	if err != nil {
		e := helpers.Logger().ErrorF(err.Error())
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
		e := helpers.Logger().ErrorF(err.Error())
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

func startGRPCServer(middlewares *Middlewares, cert string, key string) {
	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		helpers.Logger().Fatal("failed to load TLS credentials: %v", err)
	}

	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(middlewares.GrpcAuth),
		grpc.ChainStreamInterceptor(middlewares.GrpcStreamAuth),
	)

	integrationInstance := new(integration)

	plugins.RegisterIntegrationServer(server, integrationInstance)
	healthServer := health.NewServer()
	grpcHealth.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpcHealth.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		helpers.Logger().ErrorF("failed to listen grpc server: %v", err)
		os.Exit(1)
	}

	helpers.Logger().Info("starting gRPC server on 0.0.0.0:50051")
	if err := server.Serve(listener); err != nil {
		helpers.Logger().Fatal("failed to serve grpc: %v", err)
		os.Exit(1)
	}
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
			helpers.Logger().ErrorF("failed to send ack: %v", err)
			return err
		}
	}
}
