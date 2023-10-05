package processor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/utils"
	"github.com/utmstack/config-client-go/types"
)

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
		return fmt.Errorf("failed to get authentication: %v", err)
	}

	o.Credentials = result
	return nil
}

func (o *OfficeProcessor) StartSubscriptions(h *holmes.Logger) error {
	for _, subscription := range o.Subscriptions {
		h.Info("Starting subscription: %s...", subscription)
		url := configuration.GetStartSubscriptionLink(o.TenantId) + "?contentType=" + subscription
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
		}

		resp, status, err := utils.DoReq[StartSubscriptionResponse](url, []byte("{}"), http.MethodPost, headers)
		if err != nil || status != http.StatusOK {
			if status == http.StatusBadRequest && resp.Error.Code == "AF20024" {
				continue
			}
			h.Error("failed to start subscription: %v", err)
			continue
		}
		respJson, err := json.Marshal(resp)
		if err != nil || status != http.StatusOK {
			h.Error("failed to unmarshal response: %v", err)
			continue
		}

		h.Info("Starting subscription response: %v", respJson)
	}
	return nil
}

func (o *OfficeProcessor) GetContentList(subscription string, startTime string, endTime string, group types.ModuleGroup) ([]ContentList, error) {
	url := configuration.GetContentLink(o.TenantId) + fmt.Sprintf("?startTime=%s&endTime=%s&contentType=%s", startTime, endTime, subscription)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	respBody, status, err := utils.DoReq[[]ContentList](url, nil, http.MethodGet, headers)
	if err != nil && status != http.StatusOK {
		return []ContentList{}, err
	}

	return respBody, nil

}

func (o *OfficeProcessor) GetContentDetails(url string) (ContentDetailsResponse, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	respBody, status, err := utils.DoReq[ContentDetailsResponse](url, nil, http.MethodGet, headers)
	if err != nil || status != http.StatusOK {
		return ContentDetailsResponse{}, err
	}

	return respBody, nil
}

func (o *OfficeProcessor) GetLogs(startTime string, endTime string, group types.ModuleGroup, h *holmes.Logger) error {
	for _, subscription := range o.Subscriptions {
		contentList, err := o.GetContentList(subscription, startTime, endTime, group)
		if err != nil {
			h.Error("error getting content list: %v", err)
			continue
		}
		logsCounter := 0
		if len(contentList) > 0 {
			for _, log := range contentList {
				details, err := o.GetContentDetails(log.ContentUri)
				if err != nil {
					h.Error("error getting content details: %v", err)
					continue
				}
				if len(details) > 0 {
					logsCounter += len(details)
					cleanLogs := ETLProcess(details, group)
					err = SendToCorrelation(cleanLogs, h)
					if err != nil {
						h.Error("error sending log to correlation engine: %v", err)
						continue
					}
				}
			}
		}
		h.Info("Found: %d new logs in %s. for group %s", logsCounter, subscription, group.GroupName)
	}
	return nil
}
