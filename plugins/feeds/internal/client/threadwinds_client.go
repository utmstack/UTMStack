package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/entities"
)

const (
	ingestPath        = "/api/ingest/v1/entity"
	maxIngestAttempts = 3
	entitySleep       = 100 * time.Millisecond
)

type ThreadWindsClient struct {
	baseURL     string
	instanceID  string
	instanceKey string
	httpClient  *http.Client
}

func NewThreadWindsClient(ic *CustomersManagerClient) *ThreadWindsClient {
	return &ThreadWindsClient{
		baseURL:     ic.Server + "/proxy",
		instanceID:  ic.InstanceID,
		instanceKey: ic.InstanceKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type EntityIngestResult struct {
	Entity *entities.Entity
	Err    error
	Status int
}

type IngestOutcome struct {
	OKCount             int
	Failures            []EntityIngestResult
	ProvisioningBlocked bool
}

func (c *ThreadWindsClient) IngestBatch(ctx context.Context, batch []*entities.Entity) IngestOutcome {
	outcome := IngestOutcome{}

	for i, entity := range batch {
		select {
		case <-ctx.Done():
			return outcome
		default:
		}

		payload, err := json.Marshal(entity)
		if err != nil {
			outcome.Failures = append(outcome.Failures, EntityIngestResult{
				Entity: entity,
				Err:    catcher.Error("failed to marshal entity", err, map[string]any{"entity_type": entity.Type}),
			})
			continue
		}

		status, provisioningBlocked, err := c.ingestOne(ctx, payload, entity.Type)
		if err != nil {
			outcome.Failures = append(outcome.Failures, EntityIngestResult{
				Entity: entity,
				Err:    err,
				Status: status,
			})
			_ = catcher.Error("failed to ingest entity", err, map[string]any{
				"entity_type": entity.Type,
				"batch_index": i,
				"status":      status,
			})
			if provisioningBlocked {
				outcome.ProvisioningBlocked = true
				break
			}
			continue
		}

		outcome.OKCount++

		if i < len(batch)-1 {
			time.Sleep(entitySleep)
		}
	}

	return outcome
}

func (c *ThreadWindsClient) ingestOne(ctx context.Context, payload []byte, entityType string) (int, bool, error) {
	url := c.baseURL + ingestPath
	backoff := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxIngestAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return 0, false, catcher.Error("failed to create request", err, map[string]any{
				"entity_type": entityType,
			})
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("id", c.instanceID)
		req.Header.Set("key", c.instanceKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = catcher.Error("http request failed", err, map[string]any{
				"attempt":     attempt,
				"entity_type": entityType,
			})
			if attempt < maxIngestAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return 0, false, lastErr
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = catcher.Error("failed to read response body", readErr, map[string]any{
				"entity_type": entityType,
			})
			if attempt < maxIngestAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return resp.StatusCode, false, lastErr
		}

		switch {
		case resp.StatusCode == http.StatusAccepted:
			return resp.StatusCode, false, nil

		case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusUnprocessableEntity:
			_ = catcher.Error("proxy provisioning error, not retrying", nil, map[string]any{
				"status":      resp.StatusCode,
				"entity_type": entityType,
				"response":    string(body),
			})
			return resp.StatusCode, true, fmt.Errorf("proxy provisioning error: status %d", resp.StatusCode)

		case resp.StatusCode == http.StatusNotFound:
			_ = catcher.Error("proxy path not allowed", nil, map[string]any{
				"status":      resp.StatusCode,
				"entity_type": entityType,
				"response":    string(body),
			})
			return resp.StatusCode, false, fmt.Errorf("proxy path not allowed: status %d", resp.StatusCode)

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("retryable proxy error: status %d, response %s", resp.StatusCode, string(body))
			_ = catcher.Error("retryable proxy error", lastErr, map[string]any{
				"attempt":     attempt,
				"status":      resp.StatusCode,
				"entity_type": entityType,
			})
			if attempt < maxIngestAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return resp.StatusCode, false, lastErr

		default:
			_ = catcher.Error("unexpected status from proxy", nil, map[string]any{
				"status":      resp.StatusCode,
				"entity_type": entityType,
				"response":    string(body),
			})
			return resp.StatusCode, false, fmt.Errorf("unexpected status from proxy: %d", resp.StatusCode)
		}
	}

	return 0, false, lastErr
}
