package repository

import (
	"context"
	"encoding/json"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

const statisticsIndexPattern = "v11-statistics-*"

type osStatisticsRepository struct{}

func NewStatisticsRepository() connectors.StatisticsRepository {
	return &osStatisticsRepository{}
}

func (r *osStatisticsRepository) GetLatestStatistic(ctx context.Context, dataType string) (*dto.StatisticDocument, error) {
	body := map[string]any{
		"size": 1,
		"query": map[string]any{
			"match": map[string]any{"dataType": dataType},
		},
		"sort": []map[string]any{
			{"@timestamp": map[string]any{"order": "desc"}},
		},
	}

	res, err := osdk.RawSearch(ctx, []string{statisticsIndexPattern}, body)
	if err != nil {
		return nil, err
	}
	if len(res.Hits.Hits) == 0 {
		return nil, nil
	}

	raw, err := json.Marshal(res.Hits.Hits[0].Source)
	if err != nil {
		return nil, err
	}
	var doc dto.StatisticDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
