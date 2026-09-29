package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// Only the changed high/medium predicates use these unchanged history clauses.
// A local OpenSearch mock checks their identity binding and count threshold.
func TestSuricataSeverityHistory(t *testing.T) {
	if os.Getenv("UTM_SURICATA_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestSuricataSeverityHistory$")
		command.Env = append(os.Environ(), "UTM_SURICATA_HISTORY_CHILD=1")
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("isolated history test: %v\n%s", err, out)
		}
		return
	}
	var historyCount uint64
	var expectedIP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_, _ = w.Write([]byte(`{"v11-log-suricata-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"origin":{"properties":{"ip":{"type":"ip"}}},"target":{"properties":{"ip":{"type":"ip"}}}}}}}`))
			return
		}
		if r.URL.Path != "/v11-log-suricata-*/_search" {
			t.Errorf("unexpected history path %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		query := string(body)
		if !strings.Contains(query, expectedIP) || !strings.Contains(query, "@timestamp") {
			t.Errorf("history query lacks the rule IP identity or time scope: %s", query)
		}
		hits := make([]map[string]any, 0, historyCount)
		for i := uint64(0); i < historyCount; i++ {
			hits = append(hits, map[string]any{"_id": fmt.Sprint(i), "_index": "v11-log-suricata-test", "_source": map[string]any{}})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{
			"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits,
		}})
	}))
	defer server.Close()
	if err := sdkos.Connect([]string{server.URL}, "", ""); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		severity string
		ip       string
	}{
		{"high_severity_suricata_alerts_were_detected", "critical", "192.0.2.10"},
		{"medium_severity_suricata_alerts_were_detected", "warning", "203.0.113.20"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/suricata", tc.name+".yml"))
			if err != nil {
				t.Fatal(err)
			}
			rule := new(plugins.Rule)
			if err := protojson.Unmarshal(b, rule); err != nil {
				t.Fatal(err)
			}
			rule.Normalize()
			if len(rule.Correlation) != 1 {
				t.Fatalf("expected one history clause, got %d", len(rule.Correlation))
			}
			stamp := time.Now().UTC().Format(time.RFC3339Nano)
			event, err := json.Marshal(map[string]any{
				"@timestamp": stamp, "dataType": "suricata", "dataSource": "synthetic-suricata",
				"origin":   map[string]any{"ip": "192.0.2.10"},
				"target":   map[string]any{"ip": "203.0.113.20"},
				"severity": tc.severity, "log": map[string]any{"eventType": "alert"},
			})
			if err != nil {
				t.Fatal(err)
			}
			current := string(event)
			expectedIP = tc.ip
			search := rule.Correlation[0]
			historyCount = search.Count - 1
			matched, _, err := search.Execute(&current)
			if err != nil || matched {
				t.Fatalf("below threshold: matched=%v error=%v", matched, err)
			}
			historyCount = search.Count
			matched, _, err = search.Execute(&current)
			if err != nil || !matched {
				t.Fatalf("at threshold: matched=%v error=%v", matched, err)
			}
		})
	}
}
