package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

const (
	authURL       = "https://id.sophos.com/api/v2/oauth2/token"
	whoamiURL     = "https://api.central.sophos.com/whoami/v1"
	defaultTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

// tickInterval is how often jobs are published, and the width of a
// first-activation window.
const tickInterval = 5 * time.Minute

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

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go plugins.SendLogsFromChannel(pluginName)
	}

	ctx := context.Background()

	// Retry forever and fail closed. Never turn this into a
	// timeout-then-degrade path: ingesting without coordination means every
	// replica pulls every group with no durable position.
	coord, err := coordination.SetupWithRetry(ctx, coordination.DialQueuePath(processName, queuePlugin), coordinationRetryDelay,
		func(err error) {
			_ = catcher.Error("coordination setup failed, retrying", err, map[string]any{"process": processName})
		})
	if err != nil {
		return
	}
	defer coord.Close()
	logCoordinationReady()

	// Runs on every replica regardless of the election below: the election
	// only decides who publishes, not who ingests.
	go runQueueConsumer(ctx, coord.Consumer, coord.Cursors, getEncryptionKey)

	watchConfigAndPublish(coord.Scheduler, coord.Publisher, uuid.New().String())
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

// No connectivity pre-check here on purpose: blocking on id.sophos.com would
// freeze the config-update branch. Failed pulls advance no cursor and ack no
// job, so the window is simply redelivered.
func watchConfigAndPublish(scheduler coordination.Store, publisher coordination.JobPublisher, holder string) {
	time.Sleep(3 * time.Second)

	initialConfig := GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		syncActiveGroups(initialConfig)
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case newConfig := <-GetConfigUpdateChannel():
			catcher.Info("Received config update, syncing groups", map[string]any{
				"moduleActive": newConfig != nil && newConfig.ModuleActive,
				"process":      processName,
			})
			syncActiveGroups(newConfig)

		case <-ticker.C:
			groups := getActiveGroups()
			if len(groups) == 0 {
				catcher.Info("No active groups, publishing nothing", map[string]any{
					"process": processName,
				})
				continue
			}

			publishTickJobs(context.Background(), scheduler, publisher, holder, groups, time.Now().UTC())
		}
	}
}

// ingestion isolates pull's side effects so the cursor rules stay testable.
type ingestion struct {
	fetch func(fromTime int64, nextKey string, groupKey string) ([]LogRecord, string, error)
	emit  func(log *plugins.Log)
}

func liveIngestion(group *ModuleGroup) ingestion {
	agent := getSophosCentralProcessor(group)
	return ingestion{
		fetch: func(fromTime int64, nextKey string, groupKey string) ([]LogRecord, string, error) {
			return agent.getLogs(fromTime, nextKey, groupKey)
		},
		emit: func(log *plugins.Log) {
			_ = plugins.EnqueueLog(log, pluginName)
		},
	}
}

func pull(group *ModuleGroup, floor int64, pageKey string, in ingestion) (string, int, error) {
	records, nextKey, err := in.fetch(floor, pageKey, group.Key())
	if err != nil {
		return "", 0, catcher.Error("error getting logs", err, map[string]any{
			"process": processName,
			"group":   group.Key(),
		})
	}

	tenantId := group.TenantId
	if tenantId == "" {
		tenantId = defaultTenant
	}

	for _, record := range records {
		in.emit(&plugins.Log{
			Id:         record.ID,
			TenantId:   tenantId,
			DataType:   "sophos-central",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        record.Raw,
		})
	}

	return nextKey, len(records), nil
}

type SophosCentralProcessor struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	DataRegion   string
	AccessToken  string
	ExpiresAt    time.Time
}

func getSophosCentralProcessor(group *ModuleGroup) SophosCentralProcessor {
	sophosProcessor := SophosCentralProcessor{}

	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "sophos_client_id":
			sophosProcessor.ClientID = cnf.ConfValue
		case "sophos_x_api_key":
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

	maxRetries := 3
	retryDelay := 2 * time.Second

	var response map[string]any
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		response, _, err = utils.DoReq[map[string]any](authURL, []byte(data.Encode()), http.MethodPost, headers, false)
		if err == nil {
			accessToken, ok := response["access_token"].(string)
			if ok && accessToken != "" {
				expiresIn, ok := response["expires_in"].(float64)
				if ok {
					p.AccessToken = accessToken
					p.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
					return accessToken, nil
				}
			}
		}

		_ = catcher.Error("error getting access token, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if err != nil {
		return "", catcher.Error("all retries failed when getting access token", err, map[string]any{"process": processName})
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return "", catcher.Error("access_token not found in response after all retries", nil, map[string]any{
			"process":  processName,
			"response": response,
		})
	}

	expiresIn, ok := response["expires_in"].(float64)
	if !ok {
		return "", catcher.Error("expires_in not found in response after all retries", nil, map[string]any{
			"process":  processName,
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

	maxRetries := 3
	retryDelay := 2 * time.Second

	var response WhoamiResponse
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		response, _, err = utils.DoReq[WhoamiResponse](whoamiURL, nil, http.MethodGet, headers, false)
		if err == nil {
			if response.ID != "" && response.ApiHosts.DataRegion != "" {
				p.TenantID = response.ID
				p.DataRegion = response.ApiHosts.DataRegion
				return nil
			}
		}

		_ = catcher.Error("error getting tenant info, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if err != nil {
		return catcher.Error("all retries failed when getting tenant info", err, map[string]any{"process": processName})
	}

	if response.ID == "" {
		return catcher.Error("tenant ID not found in whoami response after all retries", nil, map[string]any{
			"process":  processName,
			"response": response,
		})
	}
	p.TenantID = response.ID

	if response.ApiHosts.DataRegion == "" {
		return catcher.Error("dataRegion not found in whoami response after all retries", nil, map[string]any{
			"process":  processName,
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

func (p *SophosCentralProcessor) getLogs(fromTime int64, nextKey string, groupKey string) ([]LogRecord, string, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	var accessToken string
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		accessToken, err = p.getValidAccessToken()
		if err == nil {
			break
		}

		_ = catcher.Error("error getting access token, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if err != nil {
		return nil, "", catcher.Error("all retries failed when getting access token", err, map[string]any{"process": processName})
	}

	if p.TenantID == "" || p.DataRegion == "" {
		for retry := 0; retry < maxRetries; retry++ {
			err = p.getTenantInfo(accessToken)
			if err == nil {
				break
			}

			_ = catcher.Error("error getting tenant info, retrying", err, map[string]any{
				"process":    processName,
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
			}
		}

		if err != nil {
			return nil, "", catcher.Error("all retries failed when getting tenant info", err, map[string]any{"process": processName})
		}
	}

	records := make([]LogRecord, 0, 1000)

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

		var response EventAggregate
		for retry := 0; retry < maxRetries; retry++ {
			response, _, err = utils.DoReq[EventAggregate](u.String(), nil, http.MethodGet, headers, false)
			if err == nil {
				break
			}

			_ = catcher.Error("error getting logs, retrying", err, map[string]any{
				"process":    processName,
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
			}
		}

		if err != nil {
			return nil, "", catcher.Error("all retries failed when getting logs", err, map[string]any{"process": processName})
		}

		for _, item := range response.Items {
			record, err := newLogRecord(groupKey, item)
			if err != nil {
				_ = catcher.Error("error marshalling content details", err, map[string]any{"process": processName})
				continue
			}
			records = append(records, record)
		}

		if response.Pages.NextKey == "" {
			break
		}
		nextKey = response.Pages.NextKey
	}

	return records, nextKey, nil
}

func (p *SophosCentralProcessor) buildURL(fromTime int64, nextKey string) (*url.URL, error) {
	baseURL := p.DataRegion + "/siem/v1/events"
	u, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return nil, catcher.Error("error parsing url", parseErr, map[string]any{
			"process": processName,
			"url":     baseURL,
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
