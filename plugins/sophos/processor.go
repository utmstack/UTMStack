package main

import (
	"encoding/json"
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
	"net/http"
	"net/url"

	"github.com/utmstack/config-client-go/types"
)

type SophosCentralProcessor struct {
	XApiKey       string
	Authorization string
	ApiUrl        string
}

func getSophosCentralProcessor(group types.ModuleGroup) SophosCentralProcessor {
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
	HasMore    bool             `json:"has_more"`
	Items      []map[string]any `json:"items"`
	NextCursor string           `json:"next_cursor"`
}

func (p *SophosCentralProcessor) getLogs(fromTime int) ([]string, error) {
	baseURL := p.ApiUrl + "/siem/v1/events"

	u, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return nil, catcher.Error("error parsing url", parseErr, map[string]any{
			"url": baseURL,
		})
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

	logs := make([]string, 0, 1000)
	for _, item := range response.Items {
		jsonItem, err := json.Marshal(item)
		if err != nil {
			_ = catcher.Error("error marshalling content details", err, map[string]any{})
			continue
		}
		logs = append(logs, string(jsonItem))
	}

	return logs, nil
}
