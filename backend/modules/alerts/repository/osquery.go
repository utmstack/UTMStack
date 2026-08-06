package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// ---------------------------------------------------------------------------
// Raw OpenSearch query-map builders (relocated from the former pkg/opensearch
// so this module owns its DSL helpers and the go-sdk `os` client is the only
// OpenSearch client).
// ---------------------------------------------------------------------------

func termQuery(field string, value any) map[string]any {
	return map[string]any{"term": map[string]any{field: value}}
}

func termsQuery(field string, values []string) map[string]any {
	iValues := make([]any, len(values))
	for i, v := range values {
		iValues[i] = v
	}
	return map[string]any{"terms": map[string]any{field: iValues}}
}

// Script holds a Painless script source and its parameters. Always populate
// Params — never embed user input directly in Source.
type Script struct {
	Source string
	Params map[string]any
}

// Render produces the script block expected by update_by_query.
func (s Script) Render() map[string]any {
	p := s.Params
	if p == nil {
		p = map[string]any{}
	}
	return map[string]any{"source": s.Source, "lang": "painless", "params": p}
}

// ---------------------------------------------------------------------------
// go-sdk `os` call wrappers preserving the former pkg client method shapes.
// ---------------------------------------------------------------------------

func scopeTenantQuery(ctx context.Context, query map[string]any) map[string]any {
	tid := authz.TenantIDFromContext(ctx)
	if tid == "" {
		return query
	}
	tenantFilter := termQuery("tenantId", tid)
	if query == nil {
		return tenantFilter
	}
	return map[string]any{
		"bool": map[string]any{
			"must":   []map[string]any{query},
			"filter": []map[string]any{tenantFilter},
		},
	}
}

func osUpdateByQuery(ctx context.Context, index string, query map[string]any, script Script) error {
	body := map[string]any{
		"query":  scopeTenantQuery(ctx, query),
		"script": script.Render(),
	}
	const maxAttempts = 5
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		_, err := osdk.UpdateByQuery(ctx, []string{index}, body)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isVersionConflict(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(60*(attempt+1)) * time.Millisecond):
		}
	}
	return lastErr
}

func isVersionConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "version_conflict") || strings.Contains(msg, "version conflict")
}

func osSearchSources(ctx context.Context, index string, query map[string]any, size int) ([]json.RawMessage, error) {
	query = scopeTenantQuery(ctx, query)
	body := map[string]any{"size": size}
	if query != nil {
		body["query"] = query
	}
	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, err
	}
	docs := make([]json.RawMessage, 0, len(res.Hits.Hits))
	for _, h := range res.Hits.Hits {
		b, err := json.Marshal(h.Source)
		if err != nil {
			return nil, err
		}
		docs = append(docs, b)
	}
	return docs, nil
}

// osSearchPage runs a paged + sorted search and returns the raw source docs
// alongside the matching total count.
func osSearchPage(ctx context.Context, index string, query map[string]any, from, size int, sortBy, sortOrder string) ([]json.RawMessage, int64, error) {
	query = scopeTenantQuery(ctx, query)
	body := map[string]any{
		"from":             from,
		"size":             size,
		"track_total_hits": true,
	}
	if query != nil {
		body["query"] = query
	}
	if sortBy != "" {
		body["sort"] = []map[string]any{
			{sortBy: map[string]any{"order": sortOrder}},
		}
	}
	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, 0, err
	}
	docs := make([]json.RawMessage, 0, len(res.Hits.Hits))
	for _, h := range res.Hits.Hits {
		b, err := json.Marshal(h.Source)
		if err != nil {
			return nil, 0, err
		}
		docs = append(docs, b)
	}
	return docs, res.Hits.Total.Value, nil
}

func osCount(ctx context.Context, index string, query map[string]any) (int64, error) {
	query = scopeTenantQuery(ctx, query)
	body := map[string]any{"size": 0, "track_total_hits": true}
	if query != nil {
		body["query"] = query
	}
	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return 0, err
	}
	return res.Hits.Total.Value, nil
}

// osAggregate runs an aggregation-only search and returns the raw
// {"aggregations": {...}} envelope for the caller to decode.
//
// Like every other wrapper here it scopes the query to the acting tenant first.
// Aggregations are the easiest place to leak across tenants without noticing:
// the response carries no documents, so nothing in it looks like another
// tenant's data — only counts and terms derived from it.
func osAggregate(ctx context.Context, index string, query map[string]any, aggs map[string]any, timeout string) (json.RawMessage, error) {
	body := map[string]any{
		"size": 0,
		"aggs": aggs,
	}
	if timeout != "" {
		body["timeout"] = timeout
	}
	if q := scopeTenantQuery(ctx, query); q != nil {
		body["query"] = q
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"aggregations": res.Aggregations})
}
