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
)

const (
	authURL            = "https://id.sophos.com/api/v2/oauth2/token"
	whoamiURL          = "https://api.central.sophos.com/whoami/v1"
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://id.sophos.com"
	wait               = 3 * time.Second
)

var (
	nextKeys   = make(map[int]string)
	nextKeysMu sync.RWMutex

	activeGroupsMu sync.RWMutex
	activeGroups   = make(map[int32]*ModuleGroup)
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.sophos").Env.Mode
	if mode != "manager" {
		return
	}

	go StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go plugins.SendLogsFromChannel("com.utmstack.sophos")
	}

	watchConfigAndPull()
}

func syncActiveGroups(newConfig *ConfigurationSection) {
	activeGroupsMu.Lock()
	defer activeGroupsMu.Unlock()

	if newConfig == nil || !newConfig.ModuleActive {
		catcher.Info("Module deactivated, clearing all groups", map[string]any{
			"process": "plugin_com.utmstack.sophos",
		})
		activeGroups = make(map[int32]*ModuleGroup)
		return
	}

	newGroups := make(map[int32]*ModuleGroup)
	for _, grp := range newConfig.ModuleGroups {
		newGroups[grp.Id] = grp
	}

	for id := range activeGroups {
		if _, exists := newGroups[id]; !exists {
			catcher.Info("Group removed from configuration", map[string]any{
				"process": "plugin_com.utmstack.sophos",
			})
		}
	}

	for id := range newGroups {
		if _, exists := activeGroups[id]; !exists {
			catcher.Info("New group added to configuration", map[string]any{
				"process": "plugin_com.utmstack.sophos",
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

func watchConfigAndPull() {
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
				"process":      "plugin_com.utmstack.sophos",
			})
			syncActiveGroups(newConfig)

		case <-ticker.C:
			endTime := time.Now().UTC()

			if err := connectionChecker(urlCheckConnection); err != nil {
				_ = catcher.Error("External connection failure detected", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
				continue
			}

			groups := getActiveGroups()
			if len(groups) == 0 {
				catcher.Info("No active groups, skipping pull", map[string]any{
					"process": "plugin_com.utmstack.sophos",
				})
				startTime = endTime.Add(1 * time.Nanosecond)
				continue
			}

			var wg sync.WaitGroup
			wg.Add(len(groups))
			for _, grp := range groups {
				go func(group *ModuleGroup) {
					defer wg.Done()
					var invalid bool
					for _, c := range group.ModuleGroupConfigurations {
						if strings.TrimSpace(c.ConfValue) == "" {
							invalid = true
							break
						}
					}
					if !invalid {
						pull(startTime, group)
					}
				}(grp)
			}
			wg.Wait()

			startTime = endTime.Add(1 * time.Nanosecond)
		}
	}
}

func pull(startTime time.Time, group *ModuleGroup) {
	nextKeysMu.RLock()
	prevKey := nextKeys[int(group.Id)]
	nextKeysMu.RUnlock()

	agent := getSophosCentralProcessor(group)
	logs, newNextKey, err := agent.getLogs(startTime.Unix(), prevKey)
	if err != nil {
		_ = catcher.Error("error getting logs", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
		return
	}

	tenantId := group.TenantId
	if tenantId == "" {
		tenantId = defaultTenant
	}

	if len(logs) > 0 {
		for _, log := range logs {
			plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.New().String(),
				TenantId:   tenantId,
				DataType:   "sophos-central",
				DataSource: group.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        log,
			}, "com.utmstack.sophos")
		}
	}

	nextKeysMu.Lock()
	nextKeys[int(group.Id)] = newNextKey
	nextKeysMu.Unlock()
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

	// Retry logic for getting access token
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
			"process":    "plugin_com.utmstack.sophos",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		return "", catcher.Error("all retries failed when getting access token", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return "", catcher.Error("access_token not found in response after all retries", nil, map[string]any{
			"process":  "plugin_com.utmstack.sophos",
			"response": response,
		})
	}

	expiresIn, ok := response["expires_in"].(float64)
	if !ok {
		return "", catcher.Error("expires_in not found in response after all retries", nil, map[string]any{
			"process":  "plugin_com.utmstack.sophos",
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

	// Retry logic for getting tenant info
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
			"process":    "plugin_com.utmstack.sophos",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		return catcher.Error("all retries failed when getting tenant info", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
	}

	if response.ID == "" {
		return catcher.Error("tenant ID not found in whoami response after all retries", nil, map[string]any{
			"process":  "plugin_com.utmstack.sophos",
			"response": response,
		})
	}
	p.TenantID = response.ID

	if response.ApiHosts.DataRegion == "" {
		return catcher.Error("dataRegion not found in whoami response after all retries", nil, map[string]any{
			"process":  "plugin_com.utmstack.sophos",
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

func (p *SophosCentralProcessor) getLogs(fromTime int64, nextKey string) ([]string, string, error) {
	// Retry logic for getting access token
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
			"process":    "plugin_com.utmstack.sophos",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		return nil, "", catcher.Error("all retries failed when getting access token", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
	}

	if p.TenantID == "" || p.DataRegion == "" {
		// Retry logic for getting tenant info
		for retry := 0; retry < maxRetries; retry++ {
			err = p.getTenantInfo(accessToken)
			if err == nil {
				break
			}

			_ = catcher.Error("error getting tenant info, retrying", err, map[string]any{
				"process":    "plugin_com.utmstack.sophos",
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				// Increase delay for next retry
				retryDelay *= 2
			}
		}

		if err != nil {
			return nil, "", catcher.Error("all retries failed when getting tenant info", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
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

		// Retry logic for getting logs
		var response EventAggregate
		for retry := 0; retry < maxRetries; retry++ {
			response, _, err = utils.DoReq[EventAggregate](u.String(), nil, http.MethodGet, headers, false)
			if err == nil {
				break
			}

			_ = catcher.Error("error getting logs, retrying", err, map[string]any{
				"process":    "plugin_com.utmstack.sophos",
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				// Increase delay for next retry
				retryDelay *= 2
			}
		}

		if err != nil {
			return nil, "", catcher.Error("all retries failed when getting logs", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
		}

		for _, item := range response.Items {
			jsonItem, err := json.Marshal(item)
			if err != nil {
				_ = catcher.Error("error marshalling content details", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
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

func (p *SophosCentralProcessor) buildURL(fromTime int64, nextKey string) (*url.URL, error) {
	baseURL := p.DataRegion + "/siem/v1/events"
	u, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return nil, catcher.Error("error parsing url", parseErr, map[string]any{
			"process": "sophos-plugin",
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

func connectionChecker(url string) error {
	checkConn := func() error {
		if err := checkConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
			_ = catcher.Error("error closing response body", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
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
				_ = catcher.Error("An error occurred, will keep retrying indefinitely...", err, map[string]any{"process": "plugin_com.utmstack.sophos"})
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
