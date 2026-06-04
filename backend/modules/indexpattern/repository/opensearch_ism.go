package repository

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/dto"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

type ISMClient struct {
	base string // e.g. "https://os:9200"
	user string
	pass string
	hc   *http.Client
}

func NewISMClient(cfg ospkg.Config) *ISMClient {
	scheme := "http"
	if cfg.UseTLS {
		scheme = "https"
	}
	base := fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, cfg.Port)

	transport := &http.Transport{}
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &ISMClient{
		base: base,
		user: cfg.User,
		pass: cfg.Password,
		hc:   &http.Client{Transport: transport},
	}
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

func (c *ISMClient) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("ism: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.base + "/" + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("ism: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("ism: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("ism: read body: %w", err)
	}
	return data, resp.StatusCode, nil
}

// ---------------------------------------------------------------------------
// ISM policy operations
// ---------------------------------------------------------------------------

func (c *ISMClient) GetPolicy(ctx context.Context) (*domain.IndexPolicy, error) {
	path := fmt.Sprintf("_plugins/_ism/policies/%s", domain.PolicyID)
	data, status, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ism: GetPolicy: status %d: %s", status, string(data))
	}

	var p domain.IndexPolicy
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("ism: GetPolicy: decode: %w", err)
	}
	return &p, nil
}

func (c *ISMClient) CreatePolicy(ctx context.Context, policy domain.IndexPolicy) error {
	path := fmt.Sprintf("_plugins/_ism/policies/%s", domain.PolicyID)
	body := map[string]any{"policy": policy.Policy}
	data, status, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("ism: CreatePolicy: status %d: %s", status, string(data))
	}
	return nil
}

func (c *ISMClient) UpdatePolicy(ctx context.Context, policy domain.IndexPolicy, seqNo, primaryTerm int64) error {
	path := fmt.Sprintf("_plugins/_ism/policies/%s?if_seq_no=%d&if_primary_term=%d",
		domain.PolicyID, seqNo, primaryTerm)
	body := map[string]any{"policy": policy.Policy}
	data, status, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("ism: UpdatePolicy: status %d: %s", status, string(data))
	}
	return nil
}

func (c *ISMClient) AddPolicyToIndex(ctx context.Context, indexPattern string) error {
	path := fmt.Sprintf("_plugins/_ism/add/%s", indexPattern)
	body := domain.AddPolicyRequest{PolicyID: domain.PolicyID}
	data, status, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("ism: AddPolicyToIndex(%s): status %d: %s", indexPattern, status, string(data))
	}
	return nil
}

func (c *ISMClient) ChangePolicyForIndex(
	ctx context.Context,
	pattern string,
	req domain.UpdateManagedIndexPolicyConfiguration,
) (*domain.UpdateManagedIndexPolicyResponse, error) {
	path := fmt.Sprintf("_plugins/_ism/change_policy/%s", pattern)
	data, status, err := c.do(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ism: ChangePolicyForIndex(%s): status %d: %s", pattern, status, string(data))
	}

	var resp domain.UpdateManagedIndexPolicyResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ism: ChangePolicyForIndex: decode: %w", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// Snapshot repository
// ---------------------------------------------------------------------------

func (c *ISMClient) GetSnapshotRepo(ctx context.Context) (bool, error) {
	path := fmt.Sprintf("_snapshot/%s", domain.SnapshotRepoName)
	data, status, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if status < 200 || status >= 300 {
		return false, fmt.Errorf("ism: GetSnapshotRepo: status %d: %s", status, string(data))
	}
	return true, nil
}

func (c *ISMClient) RegisterSnapshotRepo(ctx context.Context) error {
	path := fmt.Sprintf("_snapshot/%s", domain.SnapshotRepoName)
	body := domain.RegisterRepositoryRequest{
		Type: "fs",
		Settings: domain.RepoSettings{
			Location: domain.SnapshotRepoPath,
		},
	}
	data, status, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("ism: RegisterSnapshotRepo: status %d: %s", status, string(data))
	}
	return nil
}

// ---------------------------------------------------------------------------
// ISM explain (is index removable?)
// ---------------------------------------------------------------------------

func (c *ISMClient) IsIndexRemovable(ctx context.Context, indexName string) (bool, error) {
	path := fmt.Sprintf("_plugins/_ism/explain/%s", indexName)
	data, status, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, nil //nolint:nilerr — safe default on transport error
	}
	if status < 200 || status >= 300 {
		return false, nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, nil
	}

	indexRaw, ok := raw[indexName]
	if !ok {
		return false, nil
	}

	state := extractCurrentState(indexRaw, indexName)
	return state == domain.StateDelete || state == domain.StateSafeDelete, nil
}

func extractCurrentState(indexRaw json.RawMessage, _ string) string {
	var info map[string]json.RawMessage
	if err := json.Unmarshal(indexRaw, &info); err != nil {
		return ""
	}

	if stateRaw, ok := info["state"]; ok {
		var stateObj map[string]json.RawMessage
		if err := json.Unmarshal(stateRaw, &stateObj); err == nil {
			if nameRaw, ok := stateObj["name"]; ok {
				var name string
				if err := json.Unmarshal(nameRaw, &name); err == nil && name != "" {
					return name
				}
			}
		}
	}

	legacyPaths := []string{
		"index.plugins.index_state_management.current_state",
		"index.opendistro.index_state_management.current_state",
		"opendistro.index_state_management.current_state",
	}
	for _, p := range legacyPaths {
		if raw, ok := info[p]; ok {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil && s != "" {
				return s
			}
		}
	}
	return ""
}

func (c *ISMClient) GetIndexProperties(ctx context.Context, pattern string) ([]dto.IndexField, error) {
	path := fmt.Sprintf("%s/_mapping", pattern)
	data, status, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return []dto.IndexField{}, nil
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ism: GetIndexProperties(%s): status %d: %s", pattern, status, string(data))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("ism: GetIndexProperties: decode: %w", err)
	}

	seen := make(map[string]struct{})
	var fields []dto.IndexField

	for _, indexRaw := range raw {
		var indexMapping struct {
			Mappings struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"mappings"`
		}
		if err := json.Unmarshal(indexRaw, &indexMapping); err != nil {
			continue
		}
		flattenProperties(indexMapping.Mappings.Properties, "", seen, &fields)
	}

	return fields, nil
}

func flattenProperties(props map[string]json.RawMessage, prefix string, seen map[string]struct{}, out *[]dto.IndexField) {
	for name, raw := range props {
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		var node struct {
			Type       string                     `json:"type"`
			Properties map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(raw, &node); err != nil {
			continue
		}

		if node.Properties != nil {
			flattenProperties(node.Properties, fullName, seen, out)
			continue
		}

		if _, dup := seen[fullName]; dup {
			continue
		}
		seen[fullName] = struct{}{}
		*out = append(*out, dto.IndexField{Name: fullName, Type: node.Type})
	}
}
