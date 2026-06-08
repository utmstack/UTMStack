package handler

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/utmstack/UTMStack/plugins/soar/internal/cache"
	"github.com/utmstack/UTMStack/plugins/soar/internal/client"
	"github.com/utmstack/UTMStack/plugins/soar/internal/eval"
)

var marshaler = protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: true}

// New creates a Correlate handler bound to a backend client. The returned
// closure matches the plugins SDK's correlation signature.
func New(c *client.BackendClient) func(context.Context, *plugins.Alert) (*emptypb.Empty, error) {
	return func(ctx context.Context, alert *plugins.Alert) (*emptypb.Empty, error) {
		defer func() {
			if r := recover(); r != nil {
				_ = catcher.Error("recovered from panic in Correlate", nil, map[string]any{
					"panic":   r,
					"alert":   alert.GetName(),
					"process": "plugin_com.utmstack.soar",
				})
			}
		}()

		alertJSONBytes, err := marshaler.Marshal(alert)
		if err != nil {
			_ = catcher.Error("failed to marshal alert", err, map[string]any{"alert": alert.GetId()})
			return &emptypb.Empty{}, nil
		}
		alertJSON := string(alertJSONBytes)

		for _, rule := range cache.ActiveRules() {
			evaluateRuleForAlert(ctx, c, rule, alert, alertJSON, true)
			if rule.DefaultAgent != "" {
				evaluateRuleForAlert(ctx, c, rule, alert, alertJSON, false)
			}
		}
		return &emptypb.Empty{}, nil
	}
}

func evaluateRuleForAlert(parent context.Context, c *client.BackendClient, rule cache.CachedRule, alert *plugins.Alert, alertJSON string, isAgent bool) {
	if isAgent && len(rule.AgentNames) == 0 {
		return
	}

	conditions, err := eval.BuildRuleConditions(rule, isAgent)
	if err != nil {
		_ = catcher.Error("failed to build rule conditions", err, map[string]any{"rule": rule.RelPath})
		return
	}
	if !eval.Matches(alertJSON, conditions) {
		return
	}

	targetAgent := eval.ResolveAgent(alertJSON, rule, isAgent)
	if targetAgent == "" {
		return
	}

	// One execution row per command — each gets its own status/retry counter
	// in the dispatcher.
	for _, raw := range rule.Commands {
		command := eval.BuildCommand(raw, alertJSON)
		ctx, cancel := context.WithTimeout(parent, 10*time.Second)
		err := c.CreateExecution(ctx, client.CreateExecutionRequest{
			RulePath: rule.RelPath,
			AlertID:  alert.GetId(),
			Command:  command,
			Agent:    targetAgent,
		})
		cancel()
		if err != nil {
			_ = catcher.Error("failed to create execution", err, map[string]any{
				"rule":  rule.RelPath,
				"alert": alert.GetId(),
			})
		}
	}
}
