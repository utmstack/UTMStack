package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	purviewPollInterval = 5 * time.Second
	purviewPollMaxWait  = 5 * time.Minute
)

// purview-audit is an async job API available only in the global cloud.
func purviewPoll(p *GraphProcessor, group *ModuleGroup, payload *cursorPayload, windowStart, windowEnd time.Time) ([]LogRecord, time.Time, error) {
	name := "purview-audit"
	pos := payload.sourcePosition(name)
	if payload.sourceUnavailable(name) {
		return nil, pos, nil
	}
	if p.CloudEnvironment != CloudCommercial && p.CloudEnvironment != CloudGCC {
		return nil, windowEnd, nil
	}

	start, end := sourceWindow(payload.sourcePosition(name), windowStart, windowEnd)

	createBody, _ := json.Marshal(map[string]any{
		"filterStartDateTime": odata(start),
		"filterEndDateTime":   odata(end),
	})

	respBody, err := p.doGraphRequest(http.MethodPost, graphAPIVersion+"/security/auditLog/queries", createBody)
	if err != nil {
		if isPermanent(err) || isThrottled(err) {
			// The window is still unread: keep the position so it is
			// re-covered once the backoff expires.
			payload.markSourceUnavailable(name)
			catcher.Warn("graph source unavailable", map[string]any{
				"process": processName, "source": name, "error": err.Error(),
			})
			return nil, pos, nil
		}
		return nil, pos, fmt.Errorf("%s: %w", name, err)
	}

	var query struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &query); err != nil {
		return nil, pos, fmt.Errorf("%s: query decode: %w", name, err)
	}

	deadline := time.Now().Add(purviewPollMaxWait)
	for query.Status == "notStarted" || query.Status == "running" {
		if time.Now().After(deadline) {
			return nil, pos, fmt.Errorf("%s: query still running after %s", name, purviewPollMaxWait)
		}
		time.Sleep(purviewPollInterval)
		statusBody, err := p.doGraphRequest(http.MethodGet, graphAPIVersion+"/security/auditLog/queries/"+query.Id, nil)
		if err != nil {
			return nil, pos, fmt.Errorf("%s: poll: %w", name, err)
		}
		_ = json.Unmarshal(statusBody, &query)
	}
	if query.Status != "succeeded" {
		return nil, pos, fmt.Errorf("%s: query status %s", name, query.Status)
	}

	items, err := p.getJSONPages(graphAPIVersion + "/security/auditLog/queries/" + query.Id + "/records")
	if err != nil {
		return nil, pos, fmt.Errorf("%s: records: %w", name, err)
	}

	records := make([]LogRecord, 0, len(items))
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			continue
		}
		activity, _ := item["ActivityDisplayName"].(string)
		objectID, _ := item["Id"].(string)
		if objectID == "" {
			objectID, _ = item["RecordId"].(string)
		}
		records = append(records, LogRecord{
			ID:        graphIdentity(p.UtmTenantId, group.Key(), name, objectID, activity),
			Raw:       string(raw),
			EventTime: stringsToTime(item["ActivityDateTime"]),
			Source:    name,
		})
	}
	return records, end, nil
}
