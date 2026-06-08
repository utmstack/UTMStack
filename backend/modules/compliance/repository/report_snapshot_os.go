package repository

import (
	"context"
	"encoding/json"
	"fmt"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

const reportIndex = "v11-compliance-report"

type osReportStore struct{}

// NewReportStore persists/reads compliance report snapshots in OpenSearch.
func NewReportStore() connectors.ReportStore { return &osReportStore{} }

func (r *osReportStore) Save(ctx context.Context, snap *domain.ReportSnapshot) error {
	data, status, err := sdkos.DoRequest(ctx, "PUT", "/"+reportIndex+"/_doc/"+snap.ID, snap)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("save report snapshot returned %d: %s", status, string(data))
	}
	return nil
}

func (r *osReportStore) List(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error) {
	if limit <= 0 {
		limit = 50
	}
	query := map[string]any{"match_all": map[string]any{}}
	if frameworkKey != "" {
		query = map[string]any{"term": map[string]any{"frameworkKey.keyword": frameworkKey}}
	}
	body := map[string]any{
		"size":    limit,
		"_source": []string{"id", "frameworkKey", "frameworkName", "@timestamp", "score"},
		"query":   query,
		"sort":    []map[string]any{{"@timestamp": map[string]any{"order": "desc"}}},
	}
	data, status, err := sdkos.DoRequest(ctx, "POST", "/"+reportIndex+"*/_search", body)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return []domain.ReportSnapshotMeta{}, nil // index not created yet
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("list report snapshots returned %d", status)
	}
	var resp struct {
		Hits struct {
			Hits []struct {
				Source domain.ReportSnapshotMeta `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	out := make([]domain.ReportSnapshotMeta, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		out = append(out, h.Source)
	}
	return out, nil
}

func (r *osReportStore) Get(ctx context.Context, id string) (*domain.ReportSnapshot, error) {
	data, status, err := sdkos.DoRequest(ctx, "GET", "/"+reportIndex+"/_doc/"+id, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, domain.ErrReportNotFound
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("get report snapshot returned %d", status)
	}
	var resp struct {
		Found  bool                  `json:"found"`
		Source domain.ReportSnapshot `json:"_source"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if !resp.Found {
		return nil, domain.ErrReportNotFound
	}
	return &resp.Source, nil
}
