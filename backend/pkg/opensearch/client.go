// Package opensearch provides a thin wrapper around the official opensearch-go/v4
// client. Build a single Client singleton in modules.go and inject it into any
// module that needs it.
package opensearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// Config holds the connection settings read from environment variables.
type Config struct {
	Host               string
	Port               int
	User               string
	Password           string
	UseTLS             bool
	InsecureSkipVerify bool
}

// Client wraps the opensearch-go v4 client and exposes the three helpers needed
// by the alerts module: UpdateByQuery, Count, and Search.
type Client struct {
	inner *opensearchapi.Client
	// raw HTTP fields — used for endpoints not covered by the typed SDK.
	base string
	user string
	pass string
	hc   *http.Client
}

// IndexInfo holds summary information about an OpenSearch index, as returned
// by the _cat/indices API.
type IndexInfo struct {
	Name      string `json:"index"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	DocsCount string `json:"docs.count"`
	StoreSize string `json:"store.size"`
}

// ClusterHealth holds the key fields from the _cluster/health response.
type ClusterHealth struct {
	ClusterName       string `json:"cluster_name"`
	Status            string `json:"status"`
	NumberOfNodes     int    `json:"number_of_nodes"`
	NumberOfDataNodes int    `json:"number_of_data_nodes"`
	ActiveShards      int    `json:"active_shards"`
}

// IndexPropertyType represents a flattened index mapping field name + type.
type IndexPropertyType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// NewClient builds an OpenSearch client from the supplied Config.
// Returns an error if the underlying client cannot be constructed.
func NewClient(cfg Config) (*Client, error) {
	scheme := "http"
	if cfg.UseTLS {
		scheme = "https"
	}
	addr := fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, cfg.Port)

	transport := &http.Transport{}
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	inner, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{addr},
			Username:  cfg.User,
			Password:  cfg.Password,
			Transport: transport,
		},
	})
	if err != nil {
		return nil, WrapOS(err, "NewClient")
	}
	return &Client{
		inner: inner,
		base:  addr,
		user:  cfg.User,
		pass:  cfg.Password,
		hc:    &http.Client{Transport: transport},
	}, nil
}

// rawDo executes a raw HTTP request against the OpenSearch cluster.
// It handles auth, content-type, and response body reading.
func (c *Client) rawDo(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, WrapOS(fmt.Errorf("marshal body: %w", err), "rawDo")
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.base + "/" + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, WrapOS(fmt.Errorf("new request: %w", err), "rawDo")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, 0, WrapOS(fmt.Errorf("%s %s: %w", method, path, err), "rawDo")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, WrapOS(fmt.Errorf("read body: %w", err), "rawDo")
	}
	return data, resp.StatusCode, nil
}

// UpdateByQuery executes an update_by_query request against index using the
// provided query filter and Painless script. The script MUST be built with
// a params map — no raw user input inside Source.
func (c *Client) UpdateByQuery(ctx context.Context, index string, query map[string]any, script Script) error {
	body := map[string]any{
		"query":  query,
		"script": script.Render(),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return WrapOS(err, "UpdateByQuery.marshal")
	}

	_, err = c.inner.UpdateByQuery(ctx, opensearchapi.UpdateByQueryReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "UpdateByQuery")
	}
	return nil
}

// Count returns the number of documents in index that match query.
func (c *Client) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	bodyBytes, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		return 0, WrapOS(err, "Count.marshal")
	}

	resp, err := c.inner.Indices.Count(ctx, &opensearchapi.IndicesCountReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return 0, WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "Count")
	}

	return int64(resp.Count), nil
}

// SearchWithSort returns up to size raw JSON document sources from index matching query,
// sorted by sortField in sortOrder ("asc" or "desc").
// Used by the logstash pipeline scheduler to retrieve the latest statistics document.
func (c *Client) SearchWithSort(ctx context.Context, index string, query map[string]any, sortField, sortOrder string, size int) ([]json.RawMessage, error) {
	bodyBytes, err := json.Marshal(map[string]any{
		"query": query,
		"sort":  []map[string]any{{sortField: map[string]any{"order": sortOrder}}},
		"size":  size,
	})
	if err != nil {
		return nil, WrapOS(err, "SearchWithSort.marshal")
	}

	resp, err := c.inner.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "SearchWithSort")
	}

	docs := make([]json.RawMessage, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		docs = append(docs, h.Source)
	}
	return docs, nil
}

// Search returns up to size raw JSON document sources from index matching query.
func (c *Client) Search(ctx context.Context, index string, query map[string]any, size int) ([]json.RawMessage, error) {
	bodyBytes, err := json.Marshal(map[string]any{
		"query": query,
		"size":  size,
	})
	if err != nil {
		return nil, WrapOS(err, "Search.marshal")
	}

	resp, err := c.inner.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "Search")
	}

	docs := make([]json.RawMessage, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		docs = append(docs, h.Source)
	}
	return docs, nil
}

// GetFieldValues runs a terms aggregation on field within index, optionally
// filtered by filters (pass nil for match_all). Returns a map of value→docCount.
func (c *Client) GetFieldValues(
	ctx context.Context,
	index, field string,
	filters map[string]any,
	size int,
	sortByCount, sortAsc bool,
) (map[string]int64, error) {
	order := "desc"
	if sortAsc {
		order = "asc"
	}
	orderKey := "_count"
	if !sortByCount {
		orderKey = "_key"
	}

	query := map[string]any{"match_all": map[string]any{}}
	if filters != nil {
		query = filters
	}

	body := map[string]any{
		"size": 0,
		"query": query,
		"aggs": map[string]any{
			"field_values": map[string]any{
				"terms": map[string]any{
					"field": field,
					"size":  size,
					"order": map[string]any{
						orderKey: order,
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, WrapOS(err, "GetFieldValues.marshal")
	}

	resp, err := c.inner.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "GetFieldValues")
	}

	// Aggregations is json.RawMessage — parse it directly.
	var aggs struct {
		FieldValues struct {
			Buckets []struct {
				Key      string  `json:"key"`
				DocCount float64 `json:"doc_count"`
			} `json:"buckets"`
		} `json:"field_values"`
	}
	if err := json.Unmarshal(resp.Aggregations, &aggs); err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "GetFieldValues")
	}

	result := make(map[string]int64, len(aggs.FieldValues.Buckets))
	for _, b := range aggs.FieldValues.Buckets {
		result[b.Key] = int64(b.DocCount)
	}
	return result, nil
}

// RawSearch executes a raw _search POST against the given index pattern and returns
// the response body bytes and HTTP status. The caller is responsible for parsing
// the aggregations from the returned JSON.
func (c *Client) RawSearch(ctx context.Context, index string, body any) ([]byte, int, error) {
	path := index + "/_search"
	data, status, err := c.rawDo(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, 0, WrapOS(err, "RawSearch")
	}
	return data, status, nil
}

// IndexExists checks whether at least one index matching the given pattern exists.
// It uses the _cat/indices API and treats a 404 as "no match" (returns false, nil).
func (c *Client) IndexExists(ctx context.Context, pattern string) (bool, error) {
	path := "_cat/indices/" + pattern + "?format=json&h=index"
	data, status, err := c.rawDo(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, WrapOS(err, "IndexExists")
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if status < 200 || status >= 300 {
		return false, WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "IndexExists")
	}
	// An empty JSON array means no indices matched.
	var indices []json.RawMessage
	if err := json.Unmarshal(data, &indices); err != nil {
		return false, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "IndexExists")
	}
	return len(indices) > 0, nil
}

// GetIndices returns index info from _cat/indices matching pattern.
// If pattern is empty, all indices are returned.
func (c *Client) GetIndices(ctx context.Context, pattern string) ([]IndexInfo, error) {
	path := "_cat/indices?format=json&h=index,status,health,docs.count,store.size"
	if pattern != "" {
		path = "_cat/indices/" + pattern + "?format=json&h=index,status,health,docs.count,store.size"
	}

	data, status, err := c.rawDo(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, WrapOS(err, "GetIndices")
	}
	if status == http.StatusNotFound {
		return []IndexInfo{}, nil
	}
	if status < 200 || status >= 300 {
		return nil, WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "GetIndices")
	}

	var indices []IndexInfo
	if err := json.Unmarshal(data, &indices); err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "GetIndices")
	}
	return indices, nil
}

// DeleteIndices deletes one or more indices by name.
func (c *Client) DeleteIndices(ctx context.Context, indices []string) error {
	path := strings.Join(indices, ",")
	data, status, err := c.rawDo(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return WrapOS(err, "DeleteIndices")
	}
	if status < 200 || status >= 300 {
		return WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "DeleteIndices")
	}
	return nil
}

// GetClusterHealth returns the cluster health from _cluster/health.
func (c *Client) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	data, status, err := c.rawDo(ctx, http.MethodGet, "_cluster/health", nil)
	if err != nil {
		return nil, WrapOS(err, "GetClusterHealth")
	}
	if status < 200 || status >= 300 {
		return nil, WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "GetClusterHealth")
	}

	var ch ClusterHealth
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "GetClusterHealth")
	}
	return &ch, nil
}

// SearchSQL executes a SQL query against the OpenSearch SQL plugin and returns
// flattened rows as []map[string]any. The JDBC-style response (schema+datarows)
// is parsed and converted.
func (c *Client) SearchSQL(ctx context.Context, sql string) ([]map[string]any, error) {
	body := map[string]any{"query": sql}
	data, status, err := c.rawDo(ctx, http.MethodPost, "_plugins/_sql", body)
	if err != nil {
		return nil, WrapOS(err, "SearchSQL")
	}
	if status < 200 || status >= 300 {
		return nil, WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "SearchSQL")
	}

	return parseSQLResponse(data)
}

// CountSQL executes a wrapped COUNT(*) SQL query and returns the count as int64.
func (c *Client) CountSQL(ctx context.Context, wrappedSQL string) (int64, error) {
	rows, err := c.SearchSQL(ctx, wrappedSQL)
	if err != nil {
		return 0, WrapOS(err, "CountSQL")
	}
	if len(rows) == 0 {
		return 0, nil
	}
	// The wrapped SQL is: SELECT COUNT(*) FROM (...) AS total_count
	// OpenSearch returns the column as "COUNT(*)" with a double value.
	for _, v := range rows[0] {
		switch n := v.(type) {
		case float64:
			return int64(n), nil
		case int64:
			return n, nil
		}
	}
	return 0, nil
}

// SearchWithTotal executes a paginated search and returns hits as []map[string]any
// plus the total hit count. Supports optional sort (pass empty strings for default _score desc).
func (c *Client) SearchWithTotal(
	ctx context.Context,
	index string,
	query map[string]any,
	from, size int,
	sortField, sortOrder string,
) ([]map[string]any, int64, error) {
	body := map[string]any{
		"query": query,
		"from":  from,
		"size":  size,
	}
	if sortField != "" {
		order := sortOrder
		if order == "" {
			order = "desc"
		}
		body["sort"] = []map[string]any{
			{sortField: map[string]any{"order": order}},
		}
	} else {
		// Default: sort by relevance score descending (matches Java behavior).
		body["sort"] = []map[string]any{
			{"_score": map[string]any{"order": "desc"}},
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, 0, WrapOS(err, "SearchWithTotal.marshal")
	}

	resp, err := c.inner.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, 0, WrapOS(fmt.Errorf("%w: %v", ErrOSRequest, err), "SearchWithTotal")
	}

	total := int64(resp.Hits.Total.Value)
	hits := make([]map[string]any, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		var doc map[string]any
		if err := json.Unmarshal(h.Source, &doc); err != nil {
			continue
		}
		hits = append(hits, doc)
	}
	return hits, total, nil
}

// GetIndexProperties returns a deduplicated flat list of field name+type for
// the given index pattern using the _mapping API.
func (c *Client) GetIndexProperties(ctx context.Context, pattern string) ([]IndexPropertyType, error) {
	path := fmt.Sprintf("%s/_mapping", pattern)
	data, status, err := c.rawDo(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, WrapOS(err, "GetIndexProperties")
	}
	if status == http.StatusNotFound {
		return []IndexPropertyType{}, nil
	}
	if status < 200 || status >= 300 {
		return nil, WrapOS(fmt.Errorf("%w: status %d: %s", ErrOSRequest, status, string(data)), "GetIndexProperties")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "GetIndexProperties")
	}

	seen := make(map[string]struct{})
	var props []IndexPropertyType

	for _, indexRaw := range raw {
		var indexMapping struct {
			Mappings struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"mappings"`
		}
		if err := json.Unmarshal(indexRaw, &indexMapping); err != nil {
			continue
		}
		flattenIndexProperties(indexMapping.Mappings.Properties, "", seen, &props)
	}
	return props, nil
}

// ---------------------------------------------------------------------------
// private helpers
// ---------------------------------------------------------------------------

// parseSQLResponse converts an OpenSearch SQL JDBC-style response (schema + datarows)
// into a []map[string]any where each map's keys are column names.
func parseSQLResponse(data []byte) ([]map[string]any, error) {
	var raw struct {
		Schema []struct {
			Name string `json:"name"`
		} `json:"schema"`
		Datarows [][]any `json:"datarows"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, WrapOS(fmt.Errorf("%w: %v", ErrOSDecode, err), "parseSQLResponse")
	}

	result := make([]map[string]any, 0, len(raw.Datarows))
	for _, row := range raw.Datarows {
		m := make(map[string]any, len(raw.Schema))
		for i, col := range raw.Schema {
			if i < len(row) {
				m[col.Name] = row[i]
			}
		}
		result = append(result, m)
	}
	return result, nil
}

// flattenIndexProperties recursively flattens OpenSearch mapping properties.
func flattenIndexProperties(
	props map[string]json.RawMessage,
	prefix string,
	seen map[string]struct{},
	out *[]IndexPropertyType,
) {
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
			flattenIndexProperties(node.Properties, fullName, seen, out)
			continue
		}

		if _, dup := seen[fullName]; dup {
			continue
		}
		seen[fullName] = struct{}{}
		*out = append(*out, IndexPropertyType{Name: fullName, Type: node.Type})
	}
}
