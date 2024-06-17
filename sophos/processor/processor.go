package processor

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/sophos/utils"
	"github.com/utmstack/config-client-go/types"
)

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

func (p *SophosCentralProcessor) GetLogs(group types.ModuleGroup, fromTime int) ([]TransformedLog, *logger.Error) {
	baseURL := p.ApiUrl + "/siem/v1/events"

	u, parseerr := url.Parse(baseURL)
	if parseerr != nil {
		return nil, utils.Logger.ErrorF("error parsing URL params: %v", parseerr)
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
		return nil, err
	}

	logs := ETLProcess(response, group)
	return logs, nil
}
