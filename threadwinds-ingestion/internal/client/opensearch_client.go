package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

type OpenSearchClient struct {
	client *opensearch.Client
}

func NewOpenSearchClient(cfg *config.TWConfig) (*OpenSearchClient, error) {
	osConfig := opensearch.Config{
		Addresses: []string{
			fmt.Sprintf("http://%s:%s", cfg.OpenSearchHost, cfg.OpenSearchPort),
		},
	}

	client, err := opensearch.NewClient(osConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create opensearch client: %w", err)
	}

	info, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("opensearch connection failed: %w", err)
	}
	defer info.Body.Close()

	catcher.Info("opensearch client connected successfully", map[string]any{
		"host": cfg.OpenSearchHost,
		"port": cfg.OpenSearchPort,
	})

	return &OpenSearchClient{client: client}, nil
}

func (c *OpenSearchClient) GetAlertByID(ctx context.Context, alertID string) (*models.Alert, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"id.keyword": alertID,
			},
		},
	}

	return c.searchSingleAlert(ctx, "alert-*", query)
}

func (c *OpenSearchClient) searchSingleAlert(ctx context.Context, index string, query map[string]any) (*models.Alert, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error %d: %s", res.StatusCode, string(body))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source models.Alert `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Hits.Hits) == 0 {
		return nil, fmt.Errorf("alert not found")
	}

	return &result.Hits.Hits[0].Source, nil
}
