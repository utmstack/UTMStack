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
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
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
			return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusAccepted {
			catcher.Info("entity ingested successfully", map[string]any{
				"entity_type": entityType,
				"status_code": resp.StatusCode,
				"response":    string(body),
			})
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return fmt.Errorf("client error %d: %s", resp.StatusCode, string(body))
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

		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("max retries exceeded")
}

func (c *ThreadWindsClient) IngestBatch(ctx context.Context, entityBatch []*entities.Entity) error {
	successCount := 0
	errorCount := 0

	for i, entity := range entityBatch {
		select {
		case <-ctx.Done():
			return fmt.Errorf("batch ingestion cancelled: %w (processed %d/%d)", ctx.Err(), successCount, len(entityBatch))
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
		return fmt.Errorf("batch completed with %d errors out of %d entities", errorCount, len(entityBatch))
	}

	return nil
}
