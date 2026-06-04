package repository

import (
	"context"
	"fmt"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type osGatewayRepo struct {
	mapper *osdk.FieldMapper
}

func NewOSGatewayRepository() connectors.GatewayRepository {
	return &osGatewayRepo{mapper: osdk.NewFieldMapper()}
}

func (r *osGatewayRepo) PropertyValues(ctx context.Context, index, field string) ([]string, error) {
	body := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"values": map[string]any{
				"terms": map[string]any{
					"field": field,
					"size":  10000,
					"order": map[string]any{"_count": "desc"},
				},
			},
		},
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, err
	}

	buckets := termsBuckets(res.Aggregations, "values")
	out := make([]string, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, fmt.Sprintf("%v", b["key"]))
	}
	return out, nil
}

func (r *osGatewayRepo) PropertyValuesWithCount(
	ctx context.Context,
	index, field string,
	filters map[string]any,
	top int,
	orderByCount, sortAsc bool,
) (map[string]int64, error) {
	dir := "desc"
	if sortAsc {
		dir = "asc"
	}
	orderKey := "_key"
	if orderByCount {
		orderKey = "_count"
	}
	if top <= 0 {
		top = 10000
	}

	body := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"values": map[string]any{
				"terms": map[string]any{
					"field": field,
					"size":  top,
					"order": map[string]any{orderKey: dir},
				},
			},
		},
	}
	if len(filters) > 0 {
		body["query"] = filters
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, err
	}

	buckets := termsBuckets(res.Aggregations, "values")
	out := make(map[string]int64, len(buckets))
	for _, b := range buckets {
		key := fmt.Sprintf("%v", b["key"])
		if c, ok := b["doc_count"].(float64); ok {
			out[key] = int64(c)
		}
	}
	return out, nil
}

func (r *osGatewayRepo) IndexProperties(ctx context.Context, pattern string) ([]dto.IndexPropertyType, error) {
	merged, err := r.mapper.GetMergedMapping(ctx, pattern)
	if err != nil {
		return nil, err
	}

	out := make([]dto.IndexPropertyType, 0, len(merged.Fields))
	for name, info := range merged.Fields {
		out = append(out, dto.IndexPropertyType{Name: name, Type: info.Type})
		// Emit sub-fields (e.g. <field>.keyword) so callers see them too.
		for sub, subType := range info.Fields {
			out = append(out, dto.IndexPropertyType{Name: name + "." + sub, Type: subType})
		}
	}
	return out, nil
}

func (r *osGatewayRepo) Indices(ctx context.Context, pattern string) ([]dto.IndexInfo, error) {
	cats, err := osdk.ListIndices(ctx, pattern)
	if err != nil {
		return nil, err
	}

	out := make([]dto.IndexInfo, 0, len(cats))
	for _, c := range cats {
		out = append(out, dto.IndexInfo{
			Name:      c.Index,
			Status:    c.Status,
			Health:    c.Health,
			DocsCount: c.DocsCount,
			StoreSize: c.StoreSize,
		})
	}
	return out, nil
}

func (r *osGatewayRepo) DeleteIndices(ctx context.Context, indices []string) error {
	for _, idx := range indices {
		if err := osdk.DeleteIndex(ctx, idx); err != nil {
			return err
		}
	}
	return nil
}

func (r *osGatewayRepo) Search(
	ctx context.Context,
	index string,
	query map[string]any,
	from, size int,
	sortField, sortOrder string,
) ([]map[string]any, int64, error) {
	body := map[string]any{
		"from": from,
		"size": size,
	}
	if len(query) > 0 {
		body["query"] = query
	}
	if sortField != "" {
		order := sortOrder
		if order == "" {
			order = "desc"
		}
		body["sort"] = []map[string]any{{sortField: map[string]any{"order": order}}}
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, 0, err
	}

	hits := make([]map[string]any, len(res.Hits.Hits))
	for i, h := range res.Hits.Hits {
		hits[i] = h.Source
	}
	return hits, res.Hits.Total.Value, nil
}

func (r *osGatewayRepo) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	body := map[string]any{"size": 0}
	if len(query) > 0 {
		body["query"] = query
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return 0, err
	}
	return res.Hits.Total.Value, nil
}

func (r *osGatewayRepo) CountExists(ctx context.Context, index string, query map[string]any) (bool, error) {
	total, err := r.Count(ctx, index, query)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *osGatewayRepo) SearchSQL(ctx context.Context, sql string) ([]map[string]any, error) {
	return osdk.SearchSQL(ctx, sql)
}

func (r *osGatewayRepo) CountSQL(ctx context.Context, wrappedSQL string) (int64, error) {
	return osdk.CountSQL(ctx, wrappedSQL)
}

func (r *osGatewayRepo) ClusterHealth(ctx context.Context) (*dto.ClusterHealth, error) {
	h, err := osdk.GetClusterHealth(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.ClusterHealth{
		ClusterName:       h.ClusterName,
		Status:            h.Status,
		NumberOfNodes:     h.NumberOfNodes,
		NumberOfDataNodes: h.NumberOfDataNodes,
		ActiveShards:      h.ActiveShards,
	}, nil
}

// termsBuckets extracts the buckets array of a named terms aggregation from a
// raw aggregations map.
func termsBuckets(aggs map[string]any, name string) []map[string]any {
	agg, ok := aggs[name].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := agg["buckets"].([]any)
	if !ok {
		return nil
	}
	buckets := make([]map[string]any, 0, len(raw))
	for _, b := range raw {
		if m, ok := b.(map[string]any); ok {
			buckets = append(buckets, m)
		}
	}
	return buckets
}
