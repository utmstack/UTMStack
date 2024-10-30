package main

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/api/option"
)

const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

type PluginConfig struct {
	InternalKey string `yaml:"internalKey"`
	Backend     string `yaml:"backend"`
}

type GroupModule struct {
	GroupName      string
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	CTX            context.Context
	Cancel         context.CancelFunc
}

func (g *GroupModule) PullLogs() {
	for {
		select {
		case <-g.CTX.Done():
			go_sdk.Logger().LogF(100, "stopping logs puller for %s", g.GroupName)
			return
		default:
			go_sdk.Logger().LogF(100, "pulling logs for %s", g.GroupName)
			client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
			if err != nil {
				go_sdk.Logger().ErrorF("failed to create client: %v", err)
				return
			}
			defer client.Close()

			sub := client.Subscription(g.SubscriptionID)
			go_sdk.Logger().LogF(100, "subscribing to %s", g.SubscriptionID)

			err = sub.Receive(g.CTX, func(ctx context.Context, msg *pubsub.Message) {
				go_sdk.Logger().LogF(100, "received message: %v", string(msg.Data))
				logsQueue <- &go_sdk.Log{
					Id:         uuid.NewString(),
					TenantId:   DefaultTenant,
					DataType:   "google",
					DataSource: g.GroupName,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        string(msg.Data),
				}

				msg.Ack()
			})
			if err != nil {
				go_sdk.Logger().ErrorF("failed to receive messages: %v", err)
				continue
			}
		}
	}
}

func (g *GroupModule) VerifyCredentials() bool {
	client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
	if err != nil {
		go_sdk.Logger().ErrorF("failed to create client at verify credentials: %v", err)
		return false
	}
	defer client.Close()

	sub := client.Subscription(g.SubscriptionID)
	exist, err := sub.Exists(g.CTX)
	if err != nil {
		go_sdk.Logger().ErrorF("failed to verify subscription: %v", err)
		return false
	}
	if !exist {
		go_sdk.Logger().ErrorF("subscription does not exist")
		return false
	}
	return true
}

func GetModuleConfig(newConf types.ModuleGroup) GroupModule {
	ctx, cancel := context.WithCancel(context.Background())
	return GroupModule{
		GroupName:      newConf.GroupName,
		JsonKey:        newConf.Configurations[0].ConfValue,
		ProjectID:      newConf.Configurations[1].ConfValue,
		SubscriptionID: newConf.Configurations[2].ConfValue,
		CTX:            ctx,
		Cancel:         cancel,
	}
}

func compareConfigs(saveConf GroupModule, newConf types.ModuleGroup) (isNecessaryVerify bool) {
	if saveConf.JsonKey != newConf.Configurations[0].ConfValue ||
		saveConf.ProjectID != newConf.Configurations[1].ConfValue ||
		saveConf.SubscriptionID != newConf.Configurations[2].ConfValue {
		return true
	}
	return false
}
