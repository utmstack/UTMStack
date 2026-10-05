package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

const (
	graphAPIVersion        = "v1.0/"
	endPointRiskDetections = "identityProtection/riskDetections"
	riskDetectionsPageSize = "500"
	riskDetectionsMaxPages = 200
)

type riskDetectionEnvelope struct {
	Value    []json.RawMessage `json:"value"`
	NextLink string            `json:"@odata.nextLink"`
}

type riskDetection struct {
	Id                  string `json:"id"`
	LastUpdatedDateTime string `json:"lastUpdatedDateTime"`
}

func (o *OfficeProcessor) GetGraphAuth() error {
	requestUrl := fmt.Sprintf("%s%s%s", o.CloudConfig.LoginAuthority, o.TenantId, endPointLogin)

	data := url.Values{}
	data.Set("grant_type", GRANTTYPE)
	data.Set("client_id", o.ClientId)
	data.Set("client_secret", o.ClientSecret)
	data.Set("scope", o.CloudConfig.GraphScope)

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
			o.GraphCredentials = result
			return nil
		}

		_ = catcher.Error("error getting graph authentication, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return catcher.Error("all retries failed when getting graph authentication", err, map[string]any{"process": processName})
}

func riskDetectionsURL(graphEndpoint string, since time.Time) string {
	q := url.Values{}
	q.Set("$filter", fmt.Sprintf("lastUpdatedDateTime ge %s", since.UTC().Format(time.RFC3339)))
	q.Set("$orderby", "lastUpdatedDateTime asc")
	q.Set("$top", riskDetectionsPageSize)

	query := strings.ReplaceAll(q.Encode(), "+", "%20")

	return fmt.Sprintf("%s%s%s?%s", graphEndpoint, graphAPIVersion, endPointRiskDetections, query)
}

var forbiddenReported sync.Map // groupKey -> struct{}

var errRiskDetectionsForbidden = errors.New("risk detections forbidden")

func (o *OfficeProcessor) getRiskDetectionsPage(link, groupKey string) (riskDetectionEnvelope, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", o.GraphCredentials.TokenType, o.GraphCredentials.AccessToken),
	}

	maxRetries := 3
	retryDelay := 2 * time.Second

	var respBody riskDetectionEnvelope
	var status int
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		respBody, status, err = utils.DoReq[riskDetectionEnvelope](link, nil, http.MethodGet, headers, false)
		if err == nil && status == http.StatusOK {
			forbiddenReported.Delete(groupKey)
			return respBody, nil
		}

		if status == http.StatusForbidden {
			if _, alreadyReported := forbiddenReported.LoadOrStore(groupKey, struct{}{}); !alreadyReported {
				_ = catcher.Error("risk detections forbidden, skipping this group until permission and licensing are in place", err, map[string]any{
					"process": processName,
					"group":   groupKey,
					"status":  status,
				})
			}
			return riskDetectionEnvelope{}, errRiskDetectionsForbidden
		}

		_ = catcher.Error("error getting risk detections, retrying", err, map[string]any{
			"process":    processName,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"status":     status,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return riskDetectionEnvelope{}, catcher.Error("all retries failed when getting risk detections", err, map[string]any{
		"process": processName,
		"status":  status,
	})
}

func pullRiskDetections(since time.Time, group *ModuleGroup) (int, time.Time) {
	agent := GetOfficeProcessor(group)

	if err := agent.GetGraphAuth(); err != nil {
		return 0, since
	}

	utmTenantId := agent.UtmTenantId
	if utmTenantId == "" {
		utmTenantId = DefaultTenant
	}

	groupKey := group.Key()
	link := riskDetectionsURL(agent.CloudConfig.GraphEndpoint, since)

	count := 0
	watermark := since
	staleReported := false

	for page := 0; page < riskDetectionsMaxPages && link != ""; page++ {
		envelope, err := agent.getRiskDetectionsPage(link, groupKey)
		if err != nil {
			// Already logged one level down. Hand back the original watermark so
			// the next tick re-requests this window rather than skipping it.
			return count, since
		}

		for _, raw := range envelope.Value {
			var detection riskDetection
			if err := json.Unmarshal(raw, &detection); err != nil {
				_ = catcher.Error("error parsing risk detection", err, map[string]any{
					"process": processName,
					"group":   groupKey,
				})
				continue
			}

			lastUpdated, err := time.Parse(time.RFC3339, detection.LastUpdatedDateTime)
			if err != nil {
				_ = catcher.Error("error parsing risk detection timestamp", err, map[string]any{
					"process":   processName,
					"group":     groupKey,
					"detection": detection.Id,
				})
			} else {
				if lastUpdated.Before(since) && !staleReported {
					_ = catcher.Error("graph returned risk detections older than the requested watermark, $filter on lastUpdatedDateTime is being ignored", nil, map[string]any{
						"process": processName,
						"group":   groupKey,
						"since":   since.UTC().Format(time.RFC3339Nano),
					})
					staleReported = true
				}
				if lastUpdated.After(watermark) {
					watermark = lastUpdated
				}
			}

			_ = plugins.EnqueueLog(&plugins.Log{
				Id:         eventIdentity(utmTenantId, groupKey, "riskDetections", detection.Id),
				TenantId:   utmTenantId,
				DataType:   "o365",
				DataSource: group.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(raw),
			}, pluginName)
			count++
		}

		link = envelope.NextLink
	}

	return count, watermark
}
