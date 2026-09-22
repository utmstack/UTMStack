package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/models"
)

const internalKeyHeader = "X-Internal-Key"

const (
	pageSize = 200
	maxPages = 20
)

type BackendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewBackendClient(cfg *config.TWConfig) *BackendClient {
	return &BackendClient{
		baseURL:     cfg.BackendURL,
		internalKey: cfg.InternalKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getJSON performs an internal-authenticated GET and decodes the body into out.
func (c *BackendClient) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return catcher.Error("failed to create request", err, nil)
	}
	req.Header.Set(internalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("request failed", err, nil)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("unexpected status from backend", nil, map[string]any{
			"url": url, "status": resp.StatusCode, "body": string(body),
		})
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return catcher.Error("failed to decode response", err, nil)
	}
	return nil
}

func fetchPaged[T any](c *BackendClient, ctx context.Context, urlBase string, out *[]T) error {
	for page := 1; page <= maxPages; page++ {
		u, err := url.Parse(urlBase)
		if err != nil {
			return catcher.Error("failed to parse url", err, map[string]any{"url": urlBase})
		}
		q := u.Query()
		q.Set("page", strconv.Itoa(page))
		q.Set("size", strconv.Itoa(pageSize))
		u.RawQuery = q.Encode()

		var batch []T
		if err := c.getJSON(ctx, u.String(), &batch); err != nil {
			return err
		}

		*out = append(*out, batch...)

		if len(batch) < pageSize {
			return nil
		}
	}

	catcher.Warn("pagination cap reached, results may be incomplete", map[string]any{
		"url":       urlBase,
		"max_pages": maxPages,
		"page_size": pageSize,
	})
	return nil
}

func (c *BackendClient) GetRecentIncidents(ctx context.Context) ([]*models.Incident, error) {
	var all []*models.Incident
	for _, status := range []string{"Open", "In review"} {
		q := url.Values{}
		q.Set("incidentStatus", status)
		q.Set("sort", "incidentCreatedDate,desc")
		urlBase := c.baseURL + "/api/v1/incidents?" + q.Encode()

		var batch []*models.Incident
		if err := fetchPaged(c, ctx, urlBase, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
	}
	return all, nil
}

func (c *BackendClient) GetIncidentAlerts(ctx context.Context, incidentID string) ([]*models.IncidentAlert, error) {
	q := url.Values{}
	q.Set("incidentId", incidentID)
	urlBase := c.baseURL + "/api/v1/incident-alerts?" + q.Encode()

	var alerts []*models.IncidentAlert
	if err := fetchPaged(c, ctx, urlBase, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}
