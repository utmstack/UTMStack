package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/entities"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
)

type ThreadWindsClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	mu         sync.RWMutex
}

func NewThreadWindsClient(cfg *config.TWConfig) *ThreadWindsClient {
	return &ThreadWindsClient{
		baseURL: cfg.ThreadWindsURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ThreadWindsClient) UpdateCredentials(apiKey, apiSecret string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.apiKey = apiKey
	c.apiSecret = apiSecret

	catcher.Info("ThreadWinds credentials updated", nil)
}

func (c *ThreadWindsClient) ingestEntity(ctx context.Context, entity *entities.Entity) error {
	url := fmt.Sprintf("%s/api/ingest/v1/entity", c.baseURL)

	payload, err := json.Marshal(entity)
	if err != nil {
		return catcher.Error("failed to marshal entity", err, map[string]any{
			"entity_type": entity.Type,
		})
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return catcher.Error("failed to create request", err, map[string]any{
			"entity_type": entity.Type,
		})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("api-secret", c.apiSecret)

	return c.executeWithRetry(req, entity.Type)
}

func (c *ThreadWindsClient) executeWithRetry(req *http.Request, entityType string) error {
	maxRetries := 3
	backoff := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			catcher.Error("http request failed", err, map[string]any{
				"attempt":     attempt,
				"entity_type": entityType,
			})
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return catcher.Error("failed after max attempts", err, map[string]any{
				"max_retries": maxRetries,
				"entity_type": entityType,
			})
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return catcher.Error("failed to read response body", err, map[string]any{
				"entity_type": entityType,
			})
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return catcher.Error("client error from ThreadWinds", nil, map[string]any{
				"status": resp.StatusCode,
			})
		}

		if resp.StatusCode >= 500 && attempt < maxRetries {
			catcher.Error("server error, retrying", fmt.Errorf("server error %d", resp.StatusCode), map[string]any{
				"attempt":     attempt,
				"entity_type": entityType,
				"response":    string(body),
			})
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return catcher.Error("unexpected status code from ThreadWinds", nil, map[string]any{
			"status": resp.StatusCode,
		})
	}

	return catcher.Error("max retries exceeded", nil, map[string]any{
		"entity_type": entityType,
	})
}

func (c *ThreadWindsClient) IngestBatch(ctx context.Context, entityBatch []*entities.Entity) error {
	successCount := 0
	errorCount := 0

	for i, entity := range entityBatch {
		select {
		case <-ctx.Done():
			return catcher.Error("batch ingestion cancelled", ctx.Err(), map[string]any{
				"processed": successCount,
			})
		default:
		}

		err := c.ingestEntity(ctx, entity)
		if err != nil {
			errorCount++
			catcher.Error("failed to ingest entity", err, map[string]any{
				"entity_type":   entity.Type,
				"batch_index":   i,
				"success_count": successCount,
				"error_count":   errorCount,
			})
			continue
		}
		successCount++

		if i < len(entityBatch)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if errorCount > 0 {
		return catcher.Error("batch completed with errors", nil, map[string]any{
			"error_count":   errorCount,
			"success_count": successCount,
			"total_count":   len(entityBatch),
		})
	}

	return nil
}
