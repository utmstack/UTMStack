package repository

import (
	"context"
	"fmt"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type osAnalyzerRepository struct{}

func NewAnalyzerRepository() connectors.AnalyzerRepository {
	return &osAnalyzerRepository{}
}

// TopValues runs a terms aggregation on `field` (+ a value_count for the total)
// over the index, applying the given filters.
func (r *osAnalyzerRepository) TopValues(ctx context.Context, index, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error) {
	if top <= 0 {
		top = 10
	}
	body := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"values": map[string]any{
				"terms": map[string]any{
					"field": field,
					"size":  top,
					"order": map[string]any{"_count": "desc"},
				},
			},
			"total": map[string]any{
				"value_count": map[string]any{"field": field},
			},
		},
	}
	if q := common_models.FiltersToQuery(filters); len(q) > 0 {
		body["query"] = q
	}

	res, err := osdk.RawSearch(ctx, []string{index}, body)
	if err != nil {
		return nil, err
	}

	total := aggValue(res.Aggregations, "total")
	out := &dto.TopValuesResponse{Total: int64(total), Top: []dto.TopValue{}}
	for _, b := range buckets(res.Aggregations, "values") {
		count := docCount(b)
		percent := 0.0
		if total > 0 {
			percent = float64(count) / total * 100
		}
		out.Top = append(out.Top, dto.TopValue{
			Value:   fmt.Sprintf("%v", b["key"]),
			Count:   count,
			Percent: percent,
		})
	}
	return out, nil
}

// ChartView runs a date_histogram (when Interval is set) or a terms aggregation
// over Field, returning parallel category/value arrays.
func (r *osAnalyzerRepository) ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error) {
	var chart map[string]any
	if req.Interval != "" {
		chart = map[string]any{
			"date_histogram": map[string]any{
				"field":             req.Field,
				"calendar_interval": req.Interval,
			},
		}
	} else {
		top := req.Top
		if top <= 0 {
			top = 10
		}
		chart = map[string]any{
			"terms": map[string]any{
				"field": req.Field,
				"size":  top,
				"order": map[string]any{"_count": "desc"},
			},
		}
	}
	body := map[string]any{
		"size": 0,
		"aggs": map[string]any{"chart": chart},
	}
	if q := common_models.FiltersToQuery(req.Filters); len(q) > 0 {
		body["query"] = q
	}

	res, err := osdk.RawSearch(ctx, []string{req.IndexPattern}, body)
	if err != nil {
		return nil, err
	}

	out := &dto.ChartViewResponse{Categories: []string{}, Values: []int64{}}
	for _, b := range buckets(res.Aggregations, "chart") {
		label := fmt.Sprintf("%v", b["key"])
		if s, ok := b["key_as_string"].(string); ok && s != "" {
			label = s
		}
		out.Categories = append(out.Categories, label)
		out.Values = append(out.Values, docCount(b))
	}
	return out, nil
}

// --- aggregation parse helpers ---

func buckets(aggs map[string]any, name string) []map[string]any {
	agg, ok := aggs[name].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := agg["buckets"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, b := range raw {
		if m, ok := b.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func docCount(b map[string]any) int64 {
	if c, ok := b["doc_count"].(float64); ok {
		return int64(c)
	}
	return 0
}

func aggValue(aggs map[string]any, name string) float64 {
	agg, ok := aggs[name].(map[string]any)
	if !ok {
		return 0
	}
	if v, ok := agg["value"].(float64); ok {
		return v
	}
	return 0
}
