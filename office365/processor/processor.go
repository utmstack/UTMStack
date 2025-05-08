package processor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
		return err
	}

	o.Credentials = result

	return nil
}

func (o *OfficeProcessor) StartSubscriptions() error {
	for _, subscription := range o.Subscriptions {
		utils.Logger.Info("starting subscription: %s...", subscription)
		url := configuration.GetStartSubscriptionLink(o.TenantId) + "?contentType=" + subscription
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
		}

		resp, status, e := utils.DoReq[StartSubscriptionResponse](url, []byte("{}"), http.MethodPost, headers)
		if e != nil || status != http.StatusOK {
			if strings.Contains(e.Error(), "subscription is already enabled") {
				// utils.Logger.Info("subscription is already enabled") // Debug
				return nil
			}
			return e
		}

		respJson, err := json.Marshal(resp)
		if err != nil || status != http.StatusOK {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}

		utils.Logger.Info("starting subscription response: %v", respJson)
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

	respBody, status, err := utils.DoReq[ContentDetailsResponse](url, nil, http.MethodGet, headers)
	if err != nil || status != http.StatusOK {
		return ContentDetailsResponse{}, err
	}

	return respBody, nil
}

func (o *OfficeProcessor) GetLogs(startTime string, endTime string, group types.ModuleGroup) {
	for _, subscription := range o.Subscriptions {
		contentList, err := o.GetContentList(subscription, startTime, endTime, group)
		if err != nil {
			utils.Logger.ErrorF("error getting content list: %v", err) // Debug
			continue
		}
		logsCounter := 0
		if len(contentList) > 0 {
			for _, log := range contentList {
				details, err := o.GetContentDetails(log.ContentUri)
				if err != nil {
					utils.Logger.ErrorF("error getting content details: %v", err) // Debug
					continue
				}
				if len(details) > 0 {
					logsCounter += len(details)
					cleanLogs := ETLProcess(details, group)
					err := SendToLogstash(cleanLogs)
					if err != nil {
						utils.Logger.ErrorF("error sending logs to logstash: %v", err) // Debug
						continue
					}
				}
			}
		}

		utils.Logger.Info("found: %d new logs in %s for group %s", logsCounter, subscription, group.GroupName)
	}
}
