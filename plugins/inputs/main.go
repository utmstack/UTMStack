package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var localLogsChannel = make(chan *plugins.Log)

type PluginConfig struct {
	ServerName   string `yaml:"server_name"`
	InternalKey  string `yaml:"internal_key"`
	AgentManager string `yaml:"agent_manager"`
	Backend      string `yaml:"backend"`
	Logstash     string `yaml:"logstash"`
	CertsFolder  string `yaml:"certs_folder"`
}

func main() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(helpers.GetCfg().Env.Workdir, "sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		helpers.Logger().ErrorF("failed to connect to engine server: %v", err)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		helpers.Logger().ErrorF("failed to create input client: %v", err)
		os.Exit(1)
	}

	notifyClient, err := client.Notify(context.Background())
	if err != nil {
		helpers.Logger().ErrorF("failed to create notify client: %v", err)
		os.Exit(1)
	}

	autService := NewLogAuthService()
	go autService.SyncAuth()

	middlewares := NewMiddlewares(autService)

	cert, key, err := loadCerts()
	if err != nil {
		helpers.Logger().ErrorF("failed to load certificates: %v", err)
		os.Exit(1)
	}

	go startHTTPServer(middlewares, cert, key)
	go startGRPCServer(middlewares, cert, key)

	cpu := runtime.NumCPU()

	for i := 0; i < cpu; i++ {
		go sendLog(inputClient, notifyClient)
	}
}

func startHTTPServer(middlewares *Middlewares, cert string, key string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/v1/log", middlewares.HttpAuth(), Log)
	router.POST("/v1/github-webhook", middlewares.GitHubAuth(), GitHub)
	router.GET("/v1/ping", Ping)
	router.GET("/v1/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	err := router.RunTLS(":8080", cert, key)
	helpers.Logger().Info("starting HTTP server on 0.0.0.0:8080...")
	if err != nil {
		helpers.Logger().ErrorF("failed to start HTTP server: %v", err)
		return
	}
}

func loadCerts() (string, string, error) {
	cnf, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	certPath := filepath.Join(cnf.CertsFolder, utmCertFileName)
	keyPath := filepath.Join(cnf.CertsFolder, utmCertFileKey)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("certificate file does not exist: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("key file does not exist: %s", keyPath)
	}
	return certPath, keyPath, nil
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
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

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

func sendLog(inputClient plugins.Engine_InputClient, notifyClient plugins.Engine_NotifyClient) {
	for {
		l := <-localLogsChannel

		err := inputClient.Send(l)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
			notify(notifyClient, "sending_failure", err.Error())
			continue
		}

		// TODO: implement a logic to resend failed logs
		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			notify(notifyClient, "ack_failure", err.Error())
			continue
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
		notify(notifyClient, "sending_success", "")
	}
}

func notify(notifyClient plugins.Engine_NotifyClient, topic string, cause string) {
	msg := &plugins.Message{
		Id:        uuid.NewString(),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Topic:     topic,
	}

	if topic != "sending_success" {
		if cause != "" {
			msg.Message = cause
		} else {
			msg.Message = "unknown cause"
		}
	}

	err := notifyClient.Send(msg)
	if err != nil {
		helpers.Logger().ErrorF("failed to send notification: %v", err)
		return
	}

	// TODO: implement a logic to resend failed notifications
	ack, err := notifyClient.Recv()
	if err != nil {
		helpers.Logger().ErrorF("failed to receive notification ack: %v", err)
		return
	}

	helpers.Logger().LogF(100, "received notification ack: %v", ack)
}
