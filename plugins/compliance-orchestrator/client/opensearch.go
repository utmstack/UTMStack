package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
)

func ConnectOpenSearch() error {
	osUrl := plugins.PluginCfg("org.opensearch", false).Get("opensearch").String()

	err := sdkos.Connect([]string{osUrl}, "", "")
	if err != nil {
		return catcher.Error("failed to connect to OpenSearch", err, map[string]any{
			"url": osUrl,
		})
	}

	catcher.Info("Connected to OpenSearch", map[string]any{
		"url": osUrl,
	})

	return nil
}

type SQLResponse struct {
	Schema   []any   `json:"schema"`
	DataRows [][]any `json:"datarows"`
	Total    int     `json:"total"`
}

func (b *BackendClient) ExecuteSQLQuery(ctx context.Context, sql string) (SQLResult, error) {
	baseURL := plugins.PluginCfg("org.opensearch", false).Get("opensearch").String()
	sqlEndpoint := fmt.Sprintf("%s/_plugins/_sql", baseURL)

	body := map[string]string{
		"query": sql,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return SQLResult{}, fmt.Errorf("failed to marshal SQL body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sqlEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return SQLResult{}, fmt.Errorf("failed to create SQL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return SQLResult{}, fmt.Errorf("SQL request failed: %w", err)
	}
	defer resp.Body.Close()

	var result SQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SQLResult{}, fmt.Errorf("failed to decode SQL response: %w", err)
	}

	// Convertir a tu tipo interno
	return SQLResult{
		Rows:  result.DataRows,
		Count: result.Total,
	}, nil
}
