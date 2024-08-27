package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/sophos/utils"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LogQueue = make(chan *plugins.Log)

type SophosCentralProcessor struct {
	XApiKey       string
	Authorization string
	ApiUrl        string
}

func GetSophosCentralProcessor(group types.ModuleGroup) SophosCentralProcessor {
	sophosProcessor := SophosCentralProcessor{}

	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "X-API-KEY":
			sophosProcessor.XApiKey = cnf.ConfValue
		case "Authorization":
			sophosProcessor.Authorization = cnf.ConfValue
		case "API Url":
			sophosProcessor.ApiUrl = cnf.ConfValue
		}
	}
	return sophosProcessor
}

type EventAggregate struct {
	HasMore    bool                     `json:"has_more"`
	Items      []map[string]interface{} `json:"items"`
	NextCursor string                   `json:"next_cursor"`
}

func (p *SophosCentralProcessor) GetLogs(group types.ModuleGroup, fromTime int) ([]string, error) {
	baseURL := p.ApiUrl + "/siem/v1/events"

	u, parseerr := url.Parse(baseURL)
	if parseerr != nil {
		return nil, fmt.Errorf("error parsing URL params: %v", parseerr)
	}

	params := url.Values{}
	params.Add("limit", "1000")
	params.Add("from_date", fmt.Sprintf("%d", fromTime))

	u.RawQuery = params.Encode()

	headers := map[string]string{
		"accept":        "application/json",
		"Authorization": p.Authorization,
		"x-api-key":     p.XApiKey,
	}

	response, _, err := utils.DoReq[EventAggregate](u.String(), nil, http.MethodGet, headers)
	if err != nil {
		return nil, fmt.Errorf("error getting logs: %v", err)
	}

	logs := []string{}
	for _, item := range response.Items {
		jsonItem, err := json.Marshal(item)
		if err != nil {
			helpers.Logger().ErrorF("error marshalling item: %v", err)
			continue
		}
		logs = append(logs, string(jsonItem))
	}

	return logs, nil
}

func ProcessLogs() {
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

	for {
		log := <-LogQueue
		helpers.Logger().LogF(100, "sending log: %v", log)
		err := inputClient.Send(log)
		if err != nil {
			helpers.Logger().ErrorF("failed to send log: %v", err)
		} else {
			helpers.Logger().LogF(100, "successfully sent log to processing engine: %v", log)
		}

		ack, err := inputClient.Recv()
		if err != nil {
			helpers.Logger().ErrorF("failed to receive ack: %v", err)
			os.Exit(1)
		}

		helpers.Logger().LogF(100, "received ack: %v", ack)
	}
}
