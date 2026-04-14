package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
)

func ConnectOpenSearch() error {
	osCfg := LoadOpenSearchConfig()

	err := sdkos.Connect([]string{osCfg.URL}, osCfg.User, osCfg.Pass)
	if err != nil {
		return catcher.Error("failed to connect to OpenSearch", err, map[string]any{
			"url":  osCfg.URL,
			"user": osCfg.User,
		})
	}

	catcher.Info("Connected to OpenSearch", map[string]any{
		"url": osCfg.URL,
	})
	return nil
}

func (b *BackendClient) ExecuteSQLQuery(ctx context.Context, sql string) (SQLResult, error) {
	osCfg := LoadOpenSearchConfig()
	client := NewOpenSearchHTTPClient()

	sqlEndpoint := fmt.Sprintf("%s/_plugins/_sql", osCfg.URL)

	body := map[string]string{"query": sql}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return SQLResult{}, fmt.Errorf("failed to marshal SQL body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sqlEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return SQLResult{}, fmt.Errorf("failed to create SQL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(osCfg.User, osCfg.Pass)

	resp, err := client.Do(req)
	if err != nil {
		return SQLResult{}, fmt.Errorf("SQL request failed: %w", err)
	}
	defer resp.Body.Close()

	var result SQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SQLResult{}, fmt.Errorf("failed to decode SQL response: %w", err)
	}

	return SQLResult{
		Rows:  result.DataRows,
		Count: result.Total,
	}, nil
}
