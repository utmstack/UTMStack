package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/o365/config"
)

const (
	loginUrl                  = "https://login.microsoftonline.com/"
	GRANTTYPE                 = "client_credentials"
	SCOPE                     = "https://manage.office.com/.default"
	endPointLogin             = "/oauth2/v2.0/token"
	endPointStartSubscription = "/activity/feed/subscriptions/start"
	endPointContent           = "/activity/feed/subscriptions/content"
	BASEURL                   = "https://manage.office.com/api/v1.0/"
	DefaultTenant             = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

func GetMicrosoftLoginLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", loginUrl, tenant, endPointLogin)
}

func GetStartSubscriptionLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", BASEURL, tenant, endPointStartSubscription)
}

func GetContentLink(tenant string) string {
	return fmt.Sprintf("%s%s%s", BASEURL, tenant, endPointContent)
}

func GetTenantId() string {
	return DefaultTenant
}

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		return
	}

	go config.StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel()
	}

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for range ticker.C {
		endTime := time.Now().UTC()

		if err := ConnectionChecker(loginUrl); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, nil)
		}

		moduleConfig := config.GetConfig()
		if moduleConfig != nil && moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ModuleGroups))

			for _, grp := range moduleConfig.ModuleGroups {
				go func(group *config.ModuleGroup) {
					defer wg.Done()
					var invalid bool
					for _, c := range group.ModuleGroupConfigurations {
						if strings.TrimSpace(c.ConfValue) == "" {
							invalid = true
							break
						}
					}

					if !invalid {
						pull(startTime, endTime, group)
					}
				}(grp)
			}

			wg.Wait()
		}

		startTime = endTime.Add(1 * time.Nanosecond)
	}
}

func pull(startTime time.Time, endTime time.Time, group *config.ModuleGroup) {
	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		_ = catcher.Error("error getting auth", err, nil)
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		_ = catcher.Error("error starting subscriptions", err, nil)
		return
	}

	logs := agent.GetLogs(startTime, endTime)
	for _, log := range logs {
		plugins.EnqueueLog(&plugins.Log{
			Id:         uuid.New().String(),
			TenantId:   GetTenantId(),
			DataType:   "o365",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		})
	}
}

type OfficeProcessor struct {
	Credentials   MicrosoftLoginResponse
	TenantId      string
	ClientId      string
	ClientSecret  string
	Subscriptions []string
}

type MicrosoftLoginResponse struct {
	TokenType   string `json:"token_type,omitempty"`
	Expires     int    `json:"expires_in,omitempty"`
	ExtExpires  int    `json:"ext_expires_in,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type StartSubscriptionResponse struct {
	ContentType string `json:"contentType,omitempty"`
	Status      string `json:"status,omitempty"`
	WebHook     any    `json:"webhook,omitempty"`
	Error       struct {
		Message string `json:"message,omitempty"`
		Code    string `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

type ContentList struct {
	ContentUri        string `json:"contentUri,omitempty"`
	ContentId         string `json:"contentId,omitempty"`
	ContentType       string `json:"contentType,omitempty"`
	ContentCreated    string `json:"contentCreated,omitempty"`
	ContentExpiration string `json:"contentExpiration,omitempty"`
}

type ContentDetailsResponse []map[string]any

func GetOfficeProcessor(group *config.ModuleGroup) OfficeProcessor {
	offProc := OfficeProcessor{}
	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "office365_client_id":
			offProc.ClientId = cnf.ConfValue
		case "office365_client_secret":
			offProc.ClientSecret = cnf.ConfValue
		case "office365_tenant_id":
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
	requestUrl := GetMicrosoftLoginLink(o.TenantId)

	data := url.Values{}
	data.Set("grant_type", GRANTTYPE)
	data.Set("client_id", o.ClientId)
	data.Set("client_secret", o.ClientSecret)
	data.Set("scope", SCOPE)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	dataBytes := []byte(data.Encode())

	// Retry logic for authentication
	maxRetries := 3
	retryDelay := 2 * time.Second

	var result MicrosoftLoginResponse
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		result, _, err = utils.DoReq[MicrosoftLoginResponse](requestUrl, dataBytes, http.MethodPost, headers)
		if err == nil {
			o.Credentials = result
			return nil
		}

		_ = catcher.Error("error getting authentication, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	return catcher.Error("all retries failed when getting authentication", err, nil)
}

func (o *OfficeProcessor) StartSubscriptions() error {
	for _, subscription := range o.Subscriptions {
		link := GetStartSubscriptionLink(o.TenantId) + "?contentType=" + subscription
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
		}

		// Retry logic for starting subscriptions
		maxRetries := 3
		retryDelay := 2 * time.Second

		var err error

		for retry := 0; retry < maxRetries; retry++ {
			_, _, err = utils.DoReq[StartSubscriptionResponse](link, []byte("{}"), http.MethodPost, headers)
			if err == nil {
				break
			}

			// If the subscription is already enabled, that's not an error
			if strings.Contains(err.Error(), "subscription is already enabled") {
				return nil
			}

			_ = catcher.Error("error starting subscription, retrying", err, map[string]any{
				"retry":        retry + 1,
				"maxRetries":   maxRetries,
				"subscription": subscription,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				// Increase delay for next retry
				retryDelay *= 2
			}
		}

		if err != nil {
			return catcher.Error("all retries failed when starting subscription", err, map[string]any{
				"subscription": subscription,
			})
		}
	}

	return nil
}

func (o *OfficeProcessor) GetContentList(subscription string, startTime time.Time, endTime time.Time) ([]ContentList, error) {
	link := GetContentLink(o.TenantId) + fmt.Sprintf("?startTime=%s&endTime=%s&contentType=%s",
		startTime.UTC().Format("2006-01-02T15:04:05"),
		endTime.UTC().Format("2006-01-02T15:04:05"),
		subscription)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	// Retry logic for getting content list
	maxRetries := 3
	retryDelay := 2 * time.Second

	var respBody []ContentList
	var status int
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		respBody, status, err = utils.DoReq[[]ContentList](link, nil, http.MethodGet, headers)
		if err == nil && status == http.StatusOK {
			return respBody, nil
		}

		_ = catcher.Error("error getting content list, retrying", err, map[string]any{
			"retry":        retry + 1,
			"maxRetries":   maxRetries,
			"subscription": subscription,
			"status":       status,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	return []ContentList{}, catcher.Error("all retries failed when getting content list", err, map[string]any{
		"subscription": subscription,
		"status":       status,
	})
}

func (o *OfficeProcessor) GetContentDetails(url string) (ContentDetailsResponse, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	// Retry logic for getting content details
	maxRetries := 3
	retryDelay := 2 * time.Second

	var respBody ContentDetailsResponse
	var status int
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		respBody, status, err = utils.DoReq[ContentDetailsResponse](url, nil, http.MethodGet, headers)
		if err == nil {
			return respBody, nil
		}

		_ = catcher.Error("error getting content details, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"url":        url,
			"status":     status,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	return ContentDetailsResponse{}, catcher.Error("all retries failed when getting content details", err, map[string]any{
		"url":    url,
		"status": status,
	})
}

func (o *OfficeProcessor) GetLogs(startTime, endTime time.Time) []string {
	logs := make([]string, 0, 10)
	for _, subscription := range o.Subscriptions {
		contentList, err := o.GetContentList(subscription, startTime, endTime)
		if err != nil {
			_ = catcher.Error("error getting content list", err, nil)
			continue
		}

		if len(contentList) > 0 {
			for _, log := range contentList {
				details, err := o.GetContentDetails(log.ContentUri)
				if err != nil {
					_ = catcher.Error("error getting content details", err, nil)
					continue
				}
				if len(details) > 0 {
					for _, detail := range details {
						rawDetail, err := json.Marshal(detail)
						if err != nil {
							_ = catcher.Error("error marshalling content details", err, nil)
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
