package main

import (
	"context"
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
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

type CloudEnvironment string

const (
	GRANTTYPE                                  = "client_credentials"
	endPointLogin                              = "/oauth2/v2.0/token"
	endPointStartSubscription                  = "/activity/feed/subscriptions/start"
	endPointContent                            = "/activity/feed/subscriptions/content"
	DefaultTenant                              = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	apiVersion                                 = "api/v1.0/"
	connectionTimeout         time.Duration    = 30 * time.Second
	wait                      time.Duration    = 1 * time.Second
	CloudCommercial           CloudEnvironment = "Commercial"
	CloudGCC                  CloudEnvironment = "GCC"
	CloudGCCHigh              CloudEnvironment = "GCCHigh"
	CloudDoD                  CloudEnvironment = "DoD"
)

type CloudConfig struct {
	LoginAuthority     string
	ManagementEndpoint string
	Scope              string
}

func GetCloudConfig(env CloudEnvironment) CloudConfig {
	configs := map[CloudEnvironment]CloudConfig{
		CloudCommercial: {
			LoginAuthority:     "https://login.microsoftonline.com/",
			ManagementEndpoint: "https://manage.office.com/",
			Scope:              "https://manage.office.com/.default",
		},
		CloudGCC: {
			LoginAuthority:     "https://login.microsoftonline.com/",
			ManagementEndpoint: "https://manage-gcc.office.com/",
			Scope:              "https://manage-gcc.office.com/.default",
		},
		CloudGCCHigh: {
			LoginAuthority:     "https://login.microsoftonline.us/",
			ManagementEndpoint: "https://manage.office365.us/",
			Scope:              "https://manage.office365.us/.default",
		},
		CloudDoD: {
			LoginAuthority:     "https://login.microsoftonline.us/",
			ManagementEndpoint: "https://manage.protection.apps.mil/",
			Scope:              "https://manage.protection.apps.mil/.default",
		},
	}

	config, exists := configs[env]
	if !exists {
		return configs[CloudCommercial]
	}
	return config
}

var (
	activeGroupsMu sync.RWMutex
	activeGroups   = make(map[string]*ModuleGroup)
)

// coordinationRetryDelay paces retries while coordination is unreachable.
const coordinationRetryDelay = 5 * time.Second

func main() {
	mode := plugins.GetCfg(processName).Env.Mode
	if mode != "worker" {
		return
	}

	go StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel(pluginName)
	}

	ctx := context.Background()

	// Blocks until NATS coordination is up. Fails closed: never ingests without
	// coordination.
	setup, err := coordination.SetupWithRetry(ctx, coordination.DialQueuePath(processName, queuePlugin), coordinationRetryDelay, func(err error) {
		_ = catcher.Error("coordination setup failed, waiting to retry", err, map[string]any{
			"process": processName,
		})
	})
	if err != nil {
		// Only reachable on ctx cancellation, which production never does; the
		// branch exists so tests can kill the retry loop.
		_ = catcher.Error("coordination setup cancelled before it could succeed", err, map[string]any{"process": processName})
		return
	}
	defer setup.Close()
	logCoordinationReady()

	go runQueueConsumer(ctx, setup.Consumer, setup.Cursors, getEncryptionKey)

	watchConfigAndPull(setup.Scheduler, setup.Publisher, uuid.New().String())
}

func syncActiveGroups(newConfig *ConfigurationSection) {
	activeGroupsMu.Lock()
	defer activeGroupsMu.Unlock()

	if newConfig == nil || !newConfig.ModuleActive {
		catcher.Info("Module deactivated, clearing all groups", map[string]any{
			"process": processName,
		})
		activeGroups = make(map[string]*ModuleGroup)
		return
	}

	newGroups := make(map[string]*ModuleGroup)
	for _, grp := range newConfig.ModuleGroups {
		newGroups[grp.Key()] = grp
	}

	for key := range activeGroups {
		if _, exists := newGroups[key]; !exists {
			catcher.Info("Group removed from configuration", map[string]any{
				"group":   key,
				"process": processName,
			})
		}
	}

	for key := range newGroups {
		if _, exists := activeGroups[key]; !exists {
			catcher.Info("New group added to configuration", map[string]any{
				"group":   key,
				"process": processName,
			})
		}
	}

	activeGroups = newGroups
}

func getActiveGroups() []*ModuleGroup {
	activeGroupsMu.RLock()
	defer activeGroupsMu.RUnlock()

	groups := make([]*ModuleGroup, 0, len(activeGroups))
	for _, grp := range activeGroups {
		groups = append(groups, grp)
	}
	return groups
}

// The local startTime below only seeds published jobs; authoritative per-group
// progress lives in the persisted cursor, since another worker may handle a tick.
func watchConfigAndPull(scheduler coordination.Store, publisher coordination.JobPublisher, holder string) {
	time.Sleep(3 * time.Second)

	initialConfig := GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		syncActiveGroups(initialConfig)
	}

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for {
		select {
		case newConfig := <-GetConfigUpdateChannel():
			catcher.Info("Received config update, syncing groups", map[string]any{
				"moduleActive": newConfig != nil && newConfig.ModuleActive,
				"process":      processName,
			})
			syncActiveGroups(newConfig)

		case <-ticker.C:
			endTime := time.Now().UTC()

			groups := getActiveGroups()
			if len(groups) == 0 {
				catcher.Info("No active groups, skipping pull", map[string]any{
					"process": processName,
				})
				startTime = endTime.Add(1 * time.Nanosecond)
				continue
			}

			checkConfiguredEnvironments(groups)

			publishTickJobs(context.Background(), scheduler, publisher, holder, groups, startTime, endTime)

			startTime = endTime.Add(1 * time.Nanosecond)
		}
	}
}

func checkConfiguredEnvironments(groups []*ModuleGroup) {
	uniqueAuthorities := make(map[string]CloudEnvironment)

	for _, group := range groups {
		env := getGroupEnvironment(group)
		cloudConfig := GetCloudConfig(env)
		uniqueAuthorities[cloudConfig.LoginAuthority] = env
	}

	for authority, env := range uniqueAuthorities {
		if err := ConnectionChecker(authority); err != nil {
			_ = catcher.Error("External connection failure detected", err, map[string]any{
				"process":     processName,
				"environment": env,
				"authority":   authority,
			})
		}
	}
}

func getGroupEnvironment(group *ModuleGroup) CloudEnvironment {
	for _, cnf := range group.ModuleGroupConfigurations {
		if cnf.ConfKey == "office365_cloud_environment" && cnf.ConfValue != "" {
			return CloudEnvironment(cnf.ConfValue)
		}
	}
	return CloudCommercial
}

// pull fetches one window of audit logs for group and enqueues them, returning
// the record count. Callers must not advance the group's cursor when the error
// is non-nil: that would skip past un-ingested data.
func pull(startTime time.Time, endTime time.Time, group *ModuleGroup) (int, error) {
	agent := GetOfficeProcessor(group)

	if err := agent.GetAuth(); err != nil {
		_ = catcher.Error("error getting auth", err, map[string]any{"process": processName})
		return 0, err
	}

	if err := agent.StartSubscriptions(); err != nil {
		_ = catcher.Error("error starting subscriptions", err, map[string]any{"process": processName})
		return 0, err
	}

	utmTenantId := agent.UtmTenantId
	if utmTenantId == "" {
		utmTenantId = DefaultTenant
	}

	records := agent.GetLogs(startTime, endTime, group)
	for _, record := range records {
		plugins.EnqueueLog(&plugins.Log{
			Id:         record.ID,
			TenantId:   utmTenantId,
			DataType:   "o365",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        record.Raw,
		}, pluginName)
	}
	return len(records), nil
}

type OfficeProcessor struct {
	Credentials MicrosoftLoginResponse
	// TenantId is the customer's Microsoft Azure AD tenant, used to build API
	// URLs. It is NOT UTMStack's platform tenant; that is UtmTenantId.
	TenantId         string
	UtmTenantId      string
	ClientId         string
	ClientSecret     string
	Subscriptions    []string
	CloudEnvironment CloudEnvironment
	CloudConfig      CloudConfig
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

func GetOfficeProcessor(group *ModuleGroup) OfficeProcessor {
	offProc := OfficeProcessor{
		CloudEnvironment: CloudCommercial,
		UtmTenantId:      group.UtmTenantId,
	}

	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "office365_client_id":
			offProc.ClientId = cnf.ConfValue
		case "office365_client_secret":
			offProc.ClientSecret = cnf.ConfValue
		case "office365_tenant_id":
			offProc.TenantId = cnf.ConfValue
		case "office365_cloud_environment":
			if cnf.ConfValue != "" {
				offProc.CloudEnvironment = CloudEnvironment(cnf.ConfValue)
			}
		}
	}

	offProc.CloudConfig = GetCloudConfig(offProc.CloudEnvironment)

	offProc.Subscriptions = []string{
		"Audit.AzureActiveDirectory",
		"Audit.Exchange",
		"Audit.General",
		"DLP.All",
		"Audit.SharePoint",
	}

	return offProc
}

func (o *OfficeProcessor) GetAuth() error {
	requestUrl := fmt.Sprintf("%s%s%s", o.CloudConfig.LoginAuthority, o.TenantId, endPointLogin)

	data := url.Values{}
	data.Set("grant_type", GRANTTYPE)
	data.Set("client_id", o.ClientId)
	data.Set("client_secret", o.ClientSecret)
	data.Set("scope", o.CloudConfig.Scope)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	dataBytes := []byte(data.Encode())

	maxRetries := 3
	retryDelay := 2 * time.Second

	var result MicrosoftLoginResponse
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		result, _, err = utils.DoReq[MicrosoftLoginResponse](requestUrl, dataBytes, http.MethodPost, headers, false)
		if err == nil {
			o.Credentials = result
			return nil
		}

		_ = catcher.Error("error getting authentication, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return catcher.Error("all retries failed when getting authentication", err, map[string]any{"process": processName})
}

func (o *OfficeProcessor) StartSubscriptions() error {
	for _, subscription := range o.Subscriptions {
		link := fmt.Sprintf("%s%s%s%s?contentType=%s",
			o.CloudConfig.ManagementEndpoint,
			apiVersion,
			o.TenantId,
			endPointStartSubscription,
			subscription)
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
		}

		maxRetries := 3
		retryDelay := 2 * time.Second

		var err error

		for retry := 0; retry < maxRetries; retry++ {
			_, _, err = utils.DoReq[StartSubscriptionResponse](link, []byte("{}"), http.MethodPost, headers, false)
			if err == nil {
				break
			}

			// Microsoft reports an already-enabled subscription as an error.
			if strings.Contains(err.Error(), "subscription is already enabled") {
				return nil
			}

			_ = catcher.Error("error starting subscription, retrying", err, map[string]any{
				"process":      processName,
				"retry":        retry + 1,
				"maxRetries":   maxRetries,
				"subscription": subscription,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
			}
		}

		if err != nil {
			return catcher.Error("all retries failed when starting subscription", err, map[string]any{
				"process":      processName,
				"subscription": subscription,
			})
		}
	}

	return nil
}

func (o *OfficeProcessor) GetContentList(subscription string, startTime time.Time, endTime time.Time) ([]ContentList, error) {
	link := fmt.Sprintf("%s%s%s%s?startTime=%s&endTime=%s&contentType=%s",
		o.CloudConfig.ManagementEndpoint,
		apiVersion,
		o.TenantId,
		endPointContent,
		startTime.UTC().Format("2006-01-02T15:04:05"),
		endTime.UTC().Format("2006-01-02T15:04:05"),
		subscription)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	maxRetries := 3
	retryDelay := 2 * time.Second

	var respBody []ContentList
	var status int
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		respBody, status, err = utils.DoReq[[]ContentList](link, nil, http.MethodGet, headers, false)
		if err == nil && status == http.StatusOK {
			return respBody, nil
		}

		_ = catcher.Error("error getting content list, retrying", err, map[string]any{
			"process":      processName,
			"retry":        retry + 1,
			"maxRetries":   maxRetries,
			"subscription": subscription,
			"status":       status,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return []ContentList{}, catcher.Error("all retries failed when getting content list", err, map[string]any{
		"process":      processName,
		"subscription": subscription,
		"status":       status,
	})
}

func (o *OfficeProcessor) GetContentDetails(url string) (ContentDetailsResponse, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.Credentials.TokenType, o.Credentials.AccessToken),
	}

	maxRetries := 3
	retryDelay := 2 * time.Second

	var respBody ContentDetailsResponse
	var status int
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		respBody, status, err = utils.DoReq[ContentDetailsResponse](url, nil, http.MethodGet, headers, false)
		if err == nil {
			return respBody, nil
		}

		_ = catcher.Error("error getting content details, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"url":        url,
			"status":     status,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return ContentDetailsResponse{}, catcher.Error("all retries failed when getting content details", err, map[string]any{
		"process": processName,
		"url":     url,
		"status":  status,
	})
}

// GetLogs returns one LogRecord per audit detail. The Management Activity API
// is two-step: list content blobs per subscription, then fetch each content URI.
func (o *OfficeProcessor) GetLogs(startTime, endTime time.Time, group *ModuleGroup) []LogRecord {
	records := make([]LogRecord, 0, 10)
	groupKey := group.Key()
	for _, subscription := range o.Subscriptions {
		contentList, err := o.GetContentList(subscription, startTime, endTime)
		if err != nil {
			_ = catcher.Error("error getting content list", err, map[string]any{"process": processName})
			continue
		}

		if len(contentList) > 0 {
			for _, log := range contentList {
				details, err := o.GetContentDetails(log.ContentUri)
				if err != nil {
					_ = catcher.Error("error getting content details", err, map[string]any{"process": processName})
					continue
				}
				if len(details) > 0 {
					for _, detail := range details {
						rawDetail, err := json.Marshal(detail)
						if err != nil {
							_ = catcher.Error("error marshalling content details", err, map[string]any{"process": processName})
							continue
						}
						recordID, _ := detail["Id"].(string)
						records = append(records, LogRecord{
							ID:  eventIdentity(o.UtmTenantId, groupKey, log.ContentUri, recordID),
							Raw: string(rawDetail),
						})
					}
				}
			}
		}
	}
	return records
}

func ConnectionChecker(url string) error {
	checkConn := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()

		if err := checkConnection(url, ctx); err != nil {
			return fmt.Errorf("connection failed")
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string, ctx context.Context) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_ = catcher.Error("error closing response body: %v", err, map[string]any{"process": processName})
		}
	}()

	return nil
}

func infiniteRetryIfXError(f func() error, exception string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exception) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, map[string]any{"process": processName})
				xErrorWasLogged = true
			}
			time.Sleep(wait)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
