package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

type LogRecord struct {
	ID        string
	Raw       string
	EventTime string
	Source    string
}

type source struct {
	name string
	poll func(p *GraphProcessor, group *ModuleGroup, payload *cursorPayload, windowStart, windowEnd time.Time) ([]LogRecord, time.Time, error)
}

// listSource builds a source that pages an OData list endpoint bounded by a
// $filter on one of its time fields.
func listSource(name, path, timeField, top string, idFn func(item map[string]any) (string, string)) source {
	return source{
		name: name,
		poll: func(p *GraphProcessor, group *ModuleGroup, payload *cursorPayload, windowStart, windowEnd time.Time) ([]LogRecord, time.Time, error) {
			if payload.sourceUnavailable(name) {
				return nil, windowEnd, nil
			}
			start, end := sourceWindow(payload.sourcePosition(name), windowStart, windowEnd)

			// url.Values.Encode: a hand-built URL with raw spaces is a malformed
			// request line that Graph rejects with a generic 400.
			q := path + "?" + url.Values{
				"$filter": {fmt.Sprintf("(%s ge %s and %s lt %s)", timeField, odata(start), timeField, odata(end))},
				"$top":    {top},
			}.Encode()

			items, err := p.getJSONPages(q)
			if err != nil {
				if isPermanent(err) {
					payload.markSourceUnavailable(name)
					catcher.Warn("graph source permanently unavailable", map[string]any{
						"process": processName, "source": name, "error": err.Error(),
					})
					return nil, windowEnd, nil
				}
				return nil, windowEnd, fmt.Errorf("%s: %w", name, err)
			}

			records := make([]LogRecord, 0, len(items))
			for _, item := range items {
				objectID, lastUpdate := idFn(item)
				raw, err := json.Marshal(item)
				if err != nil {
					continue
				}
				records = append(records, LogRecord{
					ID:        graphIdentity(p.UtmTenantId, group.Key(), name, objectID, lastUpdate),
					Raw:       string(raw),
					EventTime: stringsToTime(item[timeField]),
					Source:    name,
				})
			}
			return records, end, nil
		},
	}
}

func stringsToTime(v any) string {
	s, _ := v.(string)
	return s
}

func idAndTime(item map[string]any) (string, string) {
	id, _ := item["id"].(string)
	last, _ := item["lastUpdateDateTime"].(string)
	return id, last
}

func idOnly(item map[string]any) (string, string) {
	id, _ := item["id"].(string)
	return id, ""
}

func huntingPoll(p *GraphProcessor, group *ModuleGroup, payload *cursorPayload, windowStart, windowEnd time.Time) ([]LogRecord, time.Time, error) {
	name := "hunting"
	if payload.sourceUnavailable(name) {
		return nil, windowEnd, nil
	}
	start, end := sourceWindow(payload.sourcePosition(name), windowStart, windowEnd)

	table := "EmailEvents"
	// KQL literal timestamps must be single-quoted.
	q := table + " | where Timestamp >= datetime('" + start.UTC().Format("2006-01-02T15:04:05") +
		"') and Timestamp < datetime('" + end.UTC().Format("2006-01-02T15:04:05") + "') | limit 5000"
	body, _ := json.Marshal(map[string]any{"Query": q})

	respBody, err := p.doGraphRequest("POST", graphAPIVersion+"/security/runHuntingQuery", body)
	if err != nil {
		if isPermanent(err) {
			payload.markSourceUnavailable(name)
			catcher.Warn("graph source permanently unavailable", map[string]any{
				"process": processName, "source": name, "error": err.Error(),
			})
			return nil, windowEnd, nil
		}
		return nil, windowEnd, fmt.Errorf("%s: %w", name, err)
	}

	// runHuntingQuery returns {"schema":[...], "results":[{...row...}, ...]};
	// results is the row array, not a nested table object.
	var resp struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, windowEnd, fmt.Errorf("%s: response decode: %w", name, err)
	}

	records := make([]LogRecord, 0, len(resp.Results))
	for _, row := range resp.Results {
		raw, err := json.Marshal(row)
		if err != nil {
			continue
		}
		records = append(records, LogRecord{
			ID:        huntingIdentity(p.UtmTenantId, group.Key(), table, row),
			Raw:       string(raw),
			EventTime: stringsToTime(row["Timestamp"]),
			Source:    name,
		})
	}
	return records, end, nil
}

// collect pulls every source for the group, advancing each source's position
// independently so one failing source does not stall the others.
func collect(payload *cursorPayload, windowStart, windowEnd time.Time, group *ModuleGroup) (int, *cursorPayload, error) {
	p := newGraphProcessor(group)

	if _, err := p.getToken(); err != nil {
		_ = catcher.Error("graph collect: auth failed", err, map[string]any{
			"process": processName, "group": group.Key(),
		})
		return 0, payload, err
	}

	next := &cursorPayload{WindowEnd: windowEnd, Sources: map[string]time.Time{}}
	for k, v := range payload.Sources {
		next.Sources[k] = v
	}
	if payload.UnavailableSources != nil {
		next.UnavailableSources = make(map[string]time.Time, len(payload.UnavailableSources))
		for k, v := range payload.UnavailableSources {
			next.UnavailableSources[k] = v
		}
	}

	sources := []source{
		listSource("alerts", graphAPIVersion+"/security/alerts_v2", "lastUpdateDateTime", strconv.Itoa(topParam), idAndTime),
		// incidents caps $top at 50; other list sources accept 500.
		listSource("incidents", graphAPIVersion+"/security/incidents", "lastUpdateDateTime", "50", idAndTime),
		listSource("directory-audit", graphAPIVersion+"/auditLogs/directoryAudits", "activityDateTime", strconv.Itoa(topParam), idOnly),
		listSource("sign-ins", graphAPIVersion+"/auditLogs/signIns", "createdDateTime", "1000", idOnly),
		listSource("risk-detections", graphAPIVersion+"/identityProtection/riskDetections", "lastUpdatedDateTime", "500", idAndTime),
		{name: "hunting", poll: huntingPoll},
		{name: "purview-audit", poll: purviewPoll},
	}

	var firstErr error
	ingested := 0

	for _, s := range sources {
		records, newPos, err := s.poll(p, group, next, windowStart, windowEnd)
		if err != nil {
			_ = catcher.Error("graph collect: source failed", err, map[string]any{
				"process": processName, "group": group.Key(), "source": s.name,
			})
			// Never park a source past where it already is: on repeated
			// failure the position stays put and re-covers the still-unread
			// part of the window, so consecutive failed ticks cannot skip
			// data.
			park := windowStart
			if old := next.sourcePosition(s.name); old.After(park) {
				park = old
			}
			next.setSourcePosition(s.name, park)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		next.setSourcePosition(s.name, newPos)
		ingested += len(records)
		for _, r := range records {
			enqueueRecord(p, group, r)
		}
	}

	return ingested, next, firstErr
}

func enqueueRecord(p *GraphProcessor, group *ModuleGroup, r LogRecord) {
	utmTenantId := p.UtmTenantId
	if utmTenantId == "" {
		utmTenantId = DefaultTenant
	}
	ts := r.EventTime
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339Nano)
	}
	_ = plugins.EnqueueLog(&plugins.Log{
		Id:         r.ID,
		TenantId:   utmTenantId,
		DataType:   "o365-graph",
		DataSource: group.GroupName,
		Timestamp:  ts,
		Raw:        r.Raw,
	}, pluginName)
}
