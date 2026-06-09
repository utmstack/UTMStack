package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

const pluginID = "com.utmstack.soar"

var marshaler = protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: false}

var (
	backendURL  string
	internalKey string
	httpClient  = &http.Client{Timeout: 30 * time.Second}
)

func main() {
	flowsDir := filepath.Join(plugins.WorkDir, "soar")
	loadFlows(flowsDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchFlows(ctx, flowsDir)

	cfg := plugins.PluginCfg("com.utmstack")
	backendURL = cfg.Get("backend").String()
	if !strings.HasPrefix(backendURL, "http://") && !strings.HasPrefix(backendURL, "https://") {
		backendURL = "http://" + backendURL
	}
	internalKey = cfg.Get("internalKey").String()

	if err := plugins.InitCorrelationPlugin(pluginID, correlate); err != nil {
		_ = catcher.Error(pluginID, err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func correlate(ctx context.Context, alert *plugins.Alert) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate", nil, map[string]any{
				"panic": r, "alert": alert.GetName(), "process": "plugin_" + pluginID,
			})
		}
	}()

	alertJSON, err := marshaler.Marshal(alert)
	if err != nil {
		_ = catcher.Error("failed to marshal alert", err, map[string]any{"alert": alert.GetId()})
		return &emptypb.Empty{}, nil
	}

	for _, f := range activeFlows() {
		if !matches(string(alertJSON), f.Conditions) {
			continue
		}
		rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := reportMatch(rctx, f.RelPath, alertJSON); err != nil {
			_ = catcher.Error("failed to report flow match", err, map[string]any{"rule": f.RelPath})
		}
		cancel()
	}
	return &emptypb.Empty{}, nil
}

func reportMatch(ctx context.Context, rulePath string, alert []byte) error {
	payload, err := json.Marshal(struct {
		RulePath string          `json:"rulePath"`
		Alert    json.RawMessage `json:"alert"`
	}{rulePath, alert})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendURL+"/api/v1/soar/rule-executions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Key", internalKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("backend returned error status", nil, map[string]any{"status": resp.StatusCode, "body": string(body)})
	}
	return nil
}
