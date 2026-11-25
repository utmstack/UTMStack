// Package threadwinds sends entities to ThreatWinds through the Customers Manager's proxy. The
// instance never holds ThreatWinds credentials: it presents its own id and key to the proxy, which
// forwards the request with the credentials it keeps for the instance.
package threadwinds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/entities"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
)

const (
	ingestPath        = "/api/ingest/v1/entity"
	maxIngestAttempts = 3
)

// Pauses, as variables so tests do not have to wait.
var (
	entitySleep   = 100 * time.Millisecond
	initialBackof = time.Second
)

// InstanceFunc returns where and as whom to call the proxy, or nil when the instance has not
// registered with the Customers Manager.
type InstanceFunc func() *instanceconfig.InstanceConfig

type Client struct {
	instance   InstanceFunc
	httpClient *http.Client
}

func New(instance InstanceFunc) *Client {
	return &Client{instance: instance, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// errNotRegistered is what an instance with no Customers Manager credentials gets: nothing can be
// sent, and trying the next entity would fail the same way.
var errNotRegistered = errors.New("the instance has no Customers Manager credentials")

// IngestBatch sends the entities one by one. It stops at the first answer that says the instance is
// not provisioned, since every other entity would get the same one.
func (c *Client) IngestBatch(ctx context.Context, batch []*entities.Entity) models.IngestOutcome {
	outcome := models.IngestOutcome{}

	cfg := c.instance()
	if cfg == nil || cfg.Server == "" || cfg.InstanceID == "" || cfg.InstanceKey == "" {
		outcome.ProvisioningBlocked = true
		_ = catcher.Error("threat intelligence contribution cannot run", errNotRegistered, nil)

		return outcome
	}

	for i, entity := range batch {
		select {
		case <-ctx.Done():
			return outcome
		default:
		}

		payload, err := json.Marshal(entity)
		if err != nil {
			outcome.Failures = append(outcome.Failures, models.EntityIngestResult{
				Entity: entity,
				Err:    catcher.Error("failed to marshal entity", err, map[string]any{"entity_type": entity.Type}),
			})

			continue
		}

		status, blocked, err := c.ingestOne(ctx, cfg, payload, entity.Type)
		if err != nil {
			outcome.Failures = append(outcome.Failures, models.EntityIngestResult{Entity: entity, Err: err, Status: status})
			_ = catcher.Error("failed to ingest entity", err, map[string]any{
				"entity_type": entity.Type, "batch_index": i, "status": status,
			})

			if blocked {
				outcome.ProvisioningBlocked = true

				break
			}

			continue
		}

		outcome.OKCount++
		outcome.Sent = append(outcome.Sent, entity)

		if i < len(batch)-1 {
			time.Sleep(entitySleep)
		}
	}

	return outcome
}

func (c *Client) ingestOne(ctx context.Context, cfg *instanceconfig.InstanceConfig, payload []byte, entityType string) (int, bool, error) {
	url := cfg.Server + "/proxy" + ingestPath
	backoff := initialBackof

	var lastErr error

	for attempt := 1; attempt <= maxIngestAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return 0, false, catcher.Error("failed to create request", err, map[string]any{"entity_type": entityType})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("id", cfg.InstanceID)
		req.Header.Set("key", cfg.InstanceKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = catcher.Error("http request failed", err, map[string]any{"attempt": attempt, "entity_type": entityType})
			if attempt < maxIngestAttempts {
				time.Sleep(backoff)
				backoff *= 2

				continue
			}

			return 0, false, lastErr
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		_ = resp.Body.Close()

		if readErr != nil {
			lastErr = catcher.Error("failed to read response body", readErr, map[string]any{"entity_type": entityType})
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
				"status": resp.StatusCode, "entity_type": entityType, "response": string(body),
			})

			return resp.StatusCode, true, fmt.Errorf("proxy provisioning error: status %d", resp.StatusCode)

		case resp.StatusCode == http.StatusNotFound:
			_ = catcher.Error("proxy path not allowed", nil, map[string]any{
				"status": resp.StatusCode, "entity_type": entityType, "response": string(body),
			})

			return resp.StatusCode, false, fmt.Errorf("proxy path not allowed: status %d", resp.StatusCode)

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("retryable proxy error: status %d, response %s", resp.StatusCode, string(body))
			_ = catcher.Error("retryable proxy error", lastErr, map[string]any{
				"attempt": attempt, "status": resp.StatusCode, "entity_type": entityType,
			})

			if attempt < maxIngestAttempts {
				time.Sleep(backoff)
				backoff *= 2

				continue
			}

			return resp.StatusCode, false, lastErr

		default:
			_ = catcher.Error("unexpected status from proxy", nil, map[string]any{
				"status": resp.StatusCode, "entity_type": entityType, "response": string(body),
			})

			return resp.StatusCode, false, fmt.Errorf("unexpected status from proxy: %d", resp.StatusCode)
		}
	}

	return 0, false, lastErr
}
