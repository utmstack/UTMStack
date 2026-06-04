package repository

import (
	"context"
	"encoding/json"

	osdk "github.com/threatwinds/go-sdk/os"
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

func matchQuery(field string, value any) map[string]any {
	return map[string]any{"match": map[string]any{field: value}}
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

func osUpdateByQuery(ctx context.Context, index string, query map[string]any, script Script) error {
	body := map[string]any{
		"query":  query,
		"script": script.Render(),
	}
	_, err := osdk.UpdateByQuery(ctx, []string{index}, body)
	return err
}

func osSearchSources(ctx context.Context, index string, query map[string]any, size int) ([]json.RawMessage, error) {
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

func osCount(ctx context.Context, index string, query map[string]any) (int64, error) {
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
