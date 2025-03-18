package processor

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/sophos/configuration"
	"github.com/utmstack/UTMStack/sophos/utils"
	"github.com/utmstack/config-client-go/types"
)

type SophosCentralProcessor struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	DataRegion   string
	AccessToken  string
	ExpiresAt    time.Time
}

func getSophosCentralProcessor(group types.ModuleGroup) SophosCentralProcessor {
	sophosProcessor := SophosCentralProcessor{}

	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "Client Id":
			sophosProcessor.ClientID = cnf.ConfValue
		case "Client Secret":
			sophosProcessor.ClientSecret = cnf.ConfValue
		}
	}
	return sophosProcessor
}

func (p *SophosCentralProcessor) getAccessToken() (string, *logger.Error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)
	data.Set("scope", "token")

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	response, _, err := utils.DoReq[map[string]any](configuration.AUTHURL, []byte(data.Encode()), http.MethodPost, headers)
	if err != nil {
		return "", utils.Logger.ErrorF("error making auth request: %v", err)
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return "", utils.Logger.ErrorF("access_token not found in response")
	}

	expiresIn, ok := response["expires_in"].(float64)
	if !ok {
		return "", utils.Logger.ErrorF("expires_in not found in response")
	}

	p.AccessToken = accessToken
	p.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return accessToken, nil
}

type WhoamiResponse struct {
	ID       string   `json:"id"`
	ApiHosts ApiHosts `json:"apiHosts"`
}
type ApiHosts struct {
	Global     string `json:"global"`
	DataRegion string `json:"dataRegion"`
}

func (p *SophosCentralProcessor) getTenantInfo(accessToken string) *logger.Error {
	headers := map[string]string{
		"accept":        "application/json",
		"Authorization": "Bearer " + accessToken,
	}

	response, _, err := utils.DoReq[WhoamiResponse](configuration.WHOAMIURL, nil, http.MethodGet, headers)
	if err != nil {
		return utils.Logger.ErrorF("error making whoami request: %v", err)
	}

	if response.ID == "" {
		return utils.Logger.ErrorF("tenant ID not found in whoami response")
	}
	p.TenantID = response.ID

	if response.ApiHosts.DataRegion == "" {
		return utils.Logger.ErrorF("dataRegion not found in whoami response")
	}
	p.DataRegion = response.ApiHosts.DataRegion

	return nil
}

func (p *SophosCentralProcessor) getValidAccessToken() (string, *logger.Error) {
	if p.AccessToken != "" && time.Now().Before(p.ExpiresAt) {
		return p.AccessToken, nil
	}
	return p.getAccessToken()
}

type EventAggregate struct {
	Pages Pages            `json:"pages"`
	Items []map[string]any `json:"items"`
}

type Pages struct {
	FromKey string `json:"fromKey"`
	NextKey string `json:"nextKey"`
	Size    int64  `json:"size"`
	MaxSize int64  `json:"maxSize"`
}

func (p *SophosCentralProcessor) getLogs(fromTime int, nextKey string, group types.ModuleGroup) ([]TransformedLog, string, *logger.Error) {
	accessToken, err := p.getValidAccessToken()
	if err != nil {
		return nil, "", utils.Logger.ErrorF("error getting access token: %v", err)
	}

	if p.TenantID == "" || p.DataRegion == "" {
		if err := p.getTenantInfo(accessToken); err != nil {
			return nil, "", utils.Logger.ErrorF("error getting tenant information: %v", err)
		}
	}

	var aggregatedEvents EventAggregate
	aggregatedEvents.Items = make([]map[string]any, 0)
	currentNextKey := nextKey

	for {
		u, err := p.buildURL(fromTime, currentNextKey)
		if err != nil {
			return nil, "", utils.Logger.ErrorF("error building URL: %v", err)
		}

		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + accessToken,
			"X-Tenant-ID":   p.TenantID,
		}

		response, _, err := utils.DoReq[EventAggregate](u.String(), nil, http.MethodGet, headers)
		if err != nil {
			return nil, "", err
		}

		aggregatedEvents.Items = append(aggregatedEvents.Items, response.Items...)

		if response.Pages.NextKey == "" {
			break
		}
		currentNextKey = response.Pages.NextKey
	}

	transformedLogs := ETLProcess(aggregatedEvents, group)

	return transformedLogs, currentNextKey, nil
}

func (p *SophosCentralProcessor) buildURL(fromTime int, nextKey string) (*url.URL, *logger.Error) {
	baseURL := p.DataRegion + "/siem/v1/events"
	u, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return nil, utils.Logger.ErrorF("error parsing url: %v", parseErr)
	}

	params := url.Values{}
	if nextKey != "" {
		params.Set("pageFromKey", nextKey)
	} else {
		params.Set("from_date", fmt.Sprintf("%d", fromTime))
	}

	u.RawQuery = params.Encode()
	return u, nil
}
