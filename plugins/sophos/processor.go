package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/utmstack/config-client-go/types"
)

const authURL string = "https://id.sophos.com/api/v2/oauth2/token"
const whoamiURL = "https://api.central.sophos.com/whoami/v1"

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

func (p *SophosCentralProcessor) getAccessToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)
	data.Set("scope", "token")

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	response, _, err := utils.DoReq[map[string]any](authURL, []byte(data.Encode()), http.MethodPost, headers)
	if err != nil {
		return "", catcher.Error("error making auth request", err, map[string]any{})
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return "", catcher.Error("access_token not found in response", nil, map[string]any{
			"response": response,
		})
	}

	expiresIn, ok := response["expires_in"].(float64)
	if !ok {
		return "", catcher.Error("expires_in not found in response", nil, map[string]any{
			"response": response,
		})
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

func (p *SophosCentralProcessor) getTenantInfo(accessToken string) error {
	headers := map[string]string{
		"accept":        "application/json",
		"Authorization": "Bearer " + accessToken,
	}

	response, _, err := utils.DoReq[WhoamiResponse](whoamiURL, nil, http.MethodGet, headers)
	if err != nil {
		return catcher.Error("error making whoami request", err, map[string]any{})
	}

	if response.ID == "" {
		return catcher.Error("tenant ID not found in whoami response", nil, map[string]any{
			"response": response,
		})
	}
	p.TenantID = response.ID

	if response.ApiHosts.DataRegion == "" {
		return catcher.Error("dataRegion not found in whoami response", nil, map[string]any{
			"response": response,
		})
	}
	p.DataRegion = response.ApiHosts.DataRegion

	return nil
}

func (p *SophosCentralProcessor) getValidAccessToken() (string, error) {
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

func (p *SophosCentralProcessor) getLogs(fromTime int, nextKey string) ([]string, string, error) {
	accessToken, err := p.getValidAccessToken()
	if err != nil {
		return nil, "", fmt.Errorf("error getting access token: %v", err)
	}

	if p.TenantID == "" || p.DataRegion == "" {
		if err := p.getTenantInfo(accessToken); err != nil {
			return nil, "", fmt.Errorf("error getting tenant information: %v", err)
		}
	}

	logs := make([]string, 0, 1000)

	for {
		u, err := p.buildURL(fromTime, nextKey)
		if err != nil {
			return nil, "", err
		}

		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + accessToken,
			"X-Tenant-ID":   p.TenantID,
		}

		response, _, err := utils.DoReq[EventAggregate](u.String(), nil, http.MethodGet, headers)
		if err != nil {
			return nil, "", fmt.Errorf("error getting logs: %v", err)
		}

		for _, item := range response.Items {
			jsonItem, err := json.Marshal(item)
			if err != nil {
				_ = catcher.Error("error marshalling content details", err, map[string]any{})
				continue
			}
			logs = append(logs, string(jsonItem))
		}

		if response.Pages.NextKey == "" {
			break
		}
		nextKey = response.Pages.NextKey
	}

	return logs, nextKey, nil
}

func (p *SophosCentralProcessor) buildURL(fromTime int, nextKey string) (*url.URL, error) {
	baseURL := p.DataRegion + "/siem/v1/events"
	u, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return nil, catcher.Error("error parsing url", parseErr, map[string]any{
			"url": baseURL,
		})
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
