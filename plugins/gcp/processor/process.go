package processor

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/UTMStack/plugins/gcp/utils"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	maxMessageSize = 1024 * 1024 * 1024
)

func PullLogs(queue chan schema.LogEntry, stopChan chan struct{}, tenant string, cnf schema.ModuleConfig) {
	uuid := uuid.New()
	filePath := "./" + cnf.ProjectID + "-" + uuid.String() + ".json"
	file, err := os.Create(filePath)
	if err != nil {
		utils.Logger.ErrorF("Error creating config file: %v", err)
		return
	}
	_, err = file.WriteString(cnf.JsonKey)
	if err != nil {
		utils.Logger.ErrorF("Error writing to config file: %v", err)
		return
	}
	defer file.Close()

	for {
		select {
		case <-stopChan:
			err = os.RemoveAll(filePath)
			if err != nil {
				utils.Logger.ErrorF("Error removing config file: %v", err)
			}
			return
		default:
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, cnf.ProjectID, option.WithCredentialsFile(filePath))
			if err != nil {
				utils.Logger.ErrorF("Failed to create client: %v", err)
				return
			}

			sub := client.Subscription(cnf.SubscriptionID)

			messageData := ""
			err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				messageData = string(msg.Data)
				if err != nil {
					utils.Logger.ErrorF("Failed to unmarshal message: %v", err)
					msg.Nack()
					return
				}
				msg.Ack()
			})
			if err != nil {
				utils.Logger.ErrorF("Failed to receive messages: %v", err)
				continue
			}
			if messageData != "" {
				queue <- schema.LogEntry{
					Log:    messageData,
					Tenant: tenant,
				}
			}
		}
	}
}

func ProcessLogs(queue chan schema.LogEntry, cnf schema.PluginConfig) {
	tlsCredentials, err := utils.LoadTLSCredentials(filepath.Join(cnf.CertsFolder, "ca.crt"))
	if err != nil {
		utils.Logger.Fatal("error loading TLS credentials: %v", err)
	}

	conn, err := grpc.NewClient(cnf.LogInput, grpc.WithTransportCredentials(tlsCredentials), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	if err != nil {
		utils.Logger.Fatal("error connecting to server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "internal-key", cnf.InternalKey)

	client := plugins.NewIntegrationClient(conn)
	plClient, err := client.ProcessLog(ctx)
	if err != nil {
		utils.Logger.Fatal("error creating client: %v", err)
	}

	for log := range queue {
		utils.Logger.Info("Received log: %v", log)
		sendLog(plClient, log.Log, log.Tenant)
	}
}

func sendLog(plClient plugins.Integration_ProcessLogClient, newLog string, tenant string) {
	err := plClient.Send(&plugins.Log{
		DataType:   "google",
		DataSource: tenant,
		Raw:        newLog,
	})
	if err != nil {
		if strings.Contains(err.Error(), "EOF") {
			return
		}
		st, ok := status.FromError(err)
		if !ok || (st.Code() != codes.Unavailable && st.Code() != codes.Canceled) {
			utils.Logger.ErrorF("failed to send log: %v :log: %s", err, newLog)
		}
	}
}
