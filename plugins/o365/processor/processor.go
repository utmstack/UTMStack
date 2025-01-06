package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/o365/configuration"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var LogQueue = make(chan *plugins.Log)

type OfficeProcessor struct {
	Credentials   MicrosoftLoginResponse
	TenantId      string
	ClientId      string
	ClientSecret  string
	Subscriptions []string
}

func GetOfficeProcessor(group types.ModuleGroup) OfficeProcessor {
	offProc := OfficeProcessor{}
	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "Client ID":
			offProc.ClientId = cnf.ConfValue
		case "Client Secret":
			offProc.ClientSecret = cnf.ConfValue
		case "Tenant ID":
			offProc.TenantId = cnf.ConfValue
		}
	}

	offProc.Subscriptions = append(offProc.Subscriptions, []string{
		"Audit.AzureActiveDirectory",
		"Audit.Exchange",
		"Audit.General",
		"DLP.All",
		"Audit.SharePoint"}...)

	return offProc
}

func (o *OfficeProcessor) GetAuth() error {
	requestUrl := configuration.GetMicrosoftLoginLink(o.TenantId)

	data := url.Values{}
	data.Set("grant_type", configuration.GRANTTYPE)
	data.Set("client_id", o.ClientId)
	data.Set("client_secret", o.ClientSecret)
	data.Set("scope", configuration.SCOPE)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	dataBytes := []byte(data.Encode())

	result, status, err := utils.DoReq[MicrosoftLoginResponse](requestUrl, dataBytes, http.MethodPost, headers)
	if err != nil || status != http.StatusOK {
		return err
	}

	o.Credentials = result

	return nil
}

func (o *OfficeProcessor) StartSubscriptions() error {
	for _, subscription := range o.Subscriptions {
		link := configuration.GetStartSubscriptionLink(o.TenantId) + "?contentType=" + subscription
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
		}

		_, _, err := utils.DoReq[StartSubscriptionResponse](link, []byte("{}"), http.MethodPost, headers)
		if err != nil {
			if strings.Contains(err.Error(), "subscription is already enabled") {
				return nil
			}
			return err
		}
	}

	return nil
}

func (o *OfficeProcessor) GetContentList(subscription string, startTime string, endTime string) ([]ContentList, error) {
	link := configuration.GetContentLink(o.TenantId) + fmt.Sprintf("?startTime=%s&endTime=%s&contentType=%s", startTime, endTime, subscription)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	respBody, status, err := utils.DoReq[[]ContentList](link, nil, http.MethodGet, headers)
	if err != nil || status != http.StatusOK {
		return []ContentList{}, err
	}

	return respBody, nil

}

func (o *OfficeProcessor) GetContentDetails(url string) (ContentDetailsResponse, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	respBody, _, err := utils.DoReq[ContentDetailsResponse](url, nil, http.MethodGet, headers)
	if err != nil {
		return ContentDetailsResponse{}, err
	}

	return respBody, nil
}

func (o *OfficeProcessor) GetLogs(startTime string, endTime string) []string {
	logs := make([]string, 0, 10)
	for _, subscription := range o.Subscriptions {
		contentList, err := o.GetContentList(subscription, startTime, endTime)
		if err != nil {
			_ = catcher.Error("error getting content list", err, map[string]any{})
			continue
		}

		if len(contentList) > 0 {
			for _, log := range contentList {
				details, err := o.GetContentDetails(log.ContentUri)
				if err != nil {
					_ = catcher.Error("error getting content details", err, map[string]any{})
					continue
				}
				if len(details) > 0 {
					for _, detail := range details {
						rawDetail, err := json.Marshal(detail)
						if err != nil {
							_ = catcher.Error("error marshalling content details", err, map[string]any{})
							continue
						}
						logs = append(logs, string(rawDetail))
					}
				}
			}
		}
	}
	return logs
}

func ProcessLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", path.Join(plugins.GetCfg().Env.Workdir,
		"sockets", "engine_server.sock")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = catcher.Error("failed to connect to engine server", err, map[string]any{})
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		_ = catcher.Error("failed to create input client", err, map[string]any{})
		os.Exit(1)
	}

	go func() {
		for {
			log := <-LogQueue

			err := inputClient.Send(log)
			if err != nil {
				_ = catcher.Error("failed to send log", err, map[string]any{})
				os.Exit(1)
			}
		}
	}()

	for {
		_, err := inputClient.Recv()
		if err != nil {
			_ = catcher.Error("failed to receive ack", err, map[string]any{})
			os.Exit(1)
		}
	}
}
