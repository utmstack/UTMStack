package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type alertScoreInput struct {
	AlertID string `json:"alert_id" jsonschema:"Alert id (the alert's own id field) to score"`
}

func registerAlertScoring(m *Module) {
	if m.deps.AlertScoring == nil {
		return
	}
	uc := m.deps.AlertScoring.GetScoringUsecase()

	Add(m, &mcp.Tool{
		Name:  "alerts.score",
		Title: "Score an alert (deterministic 4-phase triage)",
		Description: "Run UTMStack's deterministic 4-phase scoring engine on one alert and return a reproducible 0–100 verdict with a recommended decision (COMPLETE / IN_REVIEW / INCIDENT) and an explainable per-phase breakdown. " +
			"Phases: intrinsic risk (MITRE tactic/technique + impact), baseline deviation (rarity + per-asset novelty + how prior identical alerts were resolved), asset criticality (data-source type, asset CIA sensitivity, network exposure, agent reachability), and attack-chain (related alerts, kill-chain progression, temporal clustering). " +
			"No LLM is involved, so the same alert always yields the same score. Call this first to anchor your triage, then use the read tools to confirm or refute the verdict.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "alerts.read"},
		func(ctx context.Context, _ *authz.Actor, in alertScoreInput) (any, error) {
			if in.AlertID == "" {
				return nil, fmt.Errorf("alert_id is required")
			}
			return uc.ScoreAlert(ctx, in.AlertID)
		})
}
