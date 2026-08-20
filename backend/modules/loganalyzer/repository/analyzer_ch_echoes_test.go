package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/threatwinds/go-sdk/store"
)

type echoFakeReader struct {
	docs    []json.RawMessage
	buckets []store.Bucket
}

func (f *echoFakeReader) DescribeFields(_ context.Context, _ store.Scope) ([]store.Field, error) {
	return nil, nil
}
func (f *echoFakeReader) FetchPage(_ context.Context, _ store.Scope, _ []store.Filter, _ store.Page) ([]json.RawMessage, int64, error) {
	return f.docs, int64(len(f.docs)), nil
}
func (f *echoFakeReader) TopValues(_ context.Context, _ store.Scope, _ string, _ []store.Filter, _ int) ([]store.Bucket, error) {
	return f.buckets, nil
}
func (f *echoFakeReader) Timeline(_ context.Context, _ store.Scope, _ []store.Filter, _ store.Interval) ([]store.Point, error) {
	return nil, nil
}
func (f *echoFakeReader) Count(_ context.Context, _ store.Scope, _ []store.Filter) (int64, error) {
	return 0, nil
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEnrichEchoCounts(t *testing.T) {
	docs := []json.RawMessage{
		mustJSON(t, map[string]any{"id": "a", "parentId": ""}),
		mustJSON(t, map[string]any{"id": "b", "parentId": ""}),
		mustJSON(t, map[string]any{"id": "c", "parentId": ""}),
	}
	buckets := []store.Bucket{
		{Key: "a", Count: 5},
		{Key: "c", Count: 2},
	}

	repo := &chAnalyzerRepository{store: &echoFakeReader{docs: docs, buckets: buckets}}
	result := repo.enrichEchoCounts(context.Background(), store.Scope{}, docs)

	for _, raw := range result {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		id := m["id"].(string)
		echoes, hasEchoes := m["echoes"]

		switch id {
		case "a":
			if !hasEchoes {
				t.Error("doc a: want echoes field")
			} else if int64(echoes.(float64)) != 5 {
				t.Errorf("doc a: want echoes=5, got %v", echoes)
			}
		case "b":
			if hasEchoes {
				t.Errorf("doc b: want no echoes field, got %v", echoes)
			}
		case "c":
			if !hasEchoes {
				t.Error("doc c: want echoes field")
			} else if int64(echoes.(float64)) != 2 {
				t.Errorf("doc c: want echoes=2, got %v", echoes)
			}
		}
	}
}
