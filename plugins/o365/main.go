package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const delayCheck = 300

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
	if mode != "worker" {
		os.Exit(0)
	}

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel()
	}

	st := time.Now().Add(-delayCheck * time.Second)

	for {
		pConfig := plugins.PluginCfg("com.utmstack", false)
		backend := pConfig.Get("backend").String()
		internalKey := pConfig.Get("internalKey").String()

		client := utmconf.NewUTMClient(internalKey, backend)

		et := st.Add(299 * time.Second)
		startTime := st.UTC().Format("2006-01-02T15:04:05")
		endTime := et.UTC().Format("2006-01-02T15:04:05")

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot obtain module configuration", err, nil)
			}

			time.Sleep(time.Second * delayCheck)
			st = time.Now().Add(-delayCheck * time.Second)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							skip = true
							break
						}
					}

					if !skip {
						PullLogs(startTime, endTime, group)
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)
		st = time.Now().Add(-delayCheck * time.Second)
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

	result, _, err := utils.DoReq[MicrosoftLoginResponse](requestUrl, dataBytes, http.MethodPost, headers)
	if err != nil {
		return err
	}

	o.Credentials = result

	return nil
}

func (o *OfficeProcessor) StartSubscriptions() error {
	for _, subscription := range o.Subscriptions {
		link := GetStartSubscriptionLink(o.TenantId) + "?contentType=" + subscription
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
	link := GetContentLink(o.TenantId) + fmt.Sprintf("?startTime=%s&endTime=%s&contentType=%s", startTime, endTime, subscription)

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

func PullLogs(startTime string, endTime string, group types.ModuleGroup) {
	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		_ = catcher.Error("error getting auth", err, map[string]any{})
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		_ = catcher.Error("error starting subscriptions", err, map[string]any{})
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
