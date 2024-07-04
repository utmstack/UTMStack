package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

	go receiveAcks(inputClient)

	autService := NewLogAuthService()
	go autService.SyncAuth()

	middlewares := NewMiddlewares(autService)

	// Load certificates
	cert, key, err := loadCerts()
	if err != nil {
		helpers.Logger().ErrorF("failed to load certificates: %v", err)
		os.Exit(1)
	}

	go startHTTPServer(middlewares, cert, key)
	go startGRPCServer(middlewares)

	for {
		l := <-localLogsChannel

		err := inputClient.Send(l)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
		}
	}
}

func startHTTPServer(middlewares *Middlewares, cert string, key string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/v1/log", middlewares.HttpAuth(), Log)
	router.POST("/v1/github-webhook", middlewares.GitHubAuth(), GitHub)
	router.GET("/v1/ping", Ping)
	err := router.RunTLS(":8080", cert, key)
	helpers.Logger().Info("starting HTTP server on 0.0.0.0:8080")
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

func startGRPCServer(middlewares *Middlewares) {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(middlewares.GrpcAuth))

	integrationInstance := new(integration)

	plugins.RegisterIntegrationServer(server, integrationInstance)

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

func receiveAcks(inputClient plugins.Engine_InputClient) {
	for {
		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
	}
}
