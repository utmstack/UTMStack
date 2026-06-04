package repository

import (
	"context"
	"encoding/json"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

const statisticsIndexPattern = "v11-statistics-*"

type osStatisticsRepository struct {
	client *ospkg.Client
}

func NewStatisticsRepository(client *ospkg.Client) connectors.StatisticsRepository {
	return &osStatisticsRepository{client: client}
}

func (r *osStatisticsRepository) GetLatestStatistic(ctx context.Context, dataType string) (*dto.StatisticDocument, error) {
	query := map[string]any{
		"match": map[string]any{
			"dataType": dataType,
		},
	}

	hits, err := r.client.SearchWithSort(ctx, statisticsIndexPattern, query, "@timestamp", "desc", 1)
	if err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return nil, nil
	}

	var doc dto.StatisticDocument
	if err := json.Unmarshal(hits[0], &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
