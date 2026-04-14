package evaluator

import (
	"context"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/client"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/models"
)

type Evaluator struct {
	backend *client.BackendClient
}

func NewEvaluator(backend *client.BackendClient) *Evaluator {
	return &Evaluator{backend: backend}
}

func (e *Evaluator) Evaluate(ctx context.Context, cfg models.ControlConfig) (models.ControlEvaluation, error) {

	patterns, err := e.backend.GetActiveIndexPatterns(ctx)
	if err != nil {
		return models.ControlEvaluation{}, fmt.Errorf("failed to get index patterns: %w", err)
	}

	results := make([]models.QueryEvaluation, 0)
	applicable := make([]models.QueryEvaluation, 0)

	for _, q := range cfg.QueriesConfigs {
		catcher.Info("Evaluating query", map[string]any{
			"query_id": q.ID,
		})

		if !patternExists(int(q.IndexPatternID), patterns) {
			reason := "Index pattern not active"
			qr := models.QueryEvaluation{
				QueryConfigID:  q.ID,
				QueryName:      q.QueryName,
				EvaluationRule: q.EvaluationRule,
				RuleValue:      q.RuleValue,
				Hits:           0,
				Status:         models.QueryStatusNotApplicable,
				ErrorMessage:   &reason,
			}
			results = append(results, qr)
			continue
		}

		qr := e.evaluateQuery(ctx, q)
		results = append(results, qr)

		if qr.Status != models.QueryStatusNotApplicable {
			applicable = append(applicable, qr)
		}
	}

	finalStatus := computeControlStatus(cfg.ControlStrategy, applicable)

	return models.ControlEvaluation{
		ControlConfigID:  cfg.ID,
		ControlName:      cfg.ControlName,
		Status:           finalStatus,
		QueryEvaluations: results,
	}, nil
}

func (e *Evaluator) evaluateQuery(ctx context.Context, q models.QueryConfig) models.QueryEvaluation {
	res, err := e.backend.ExecuteSQLQuery(ctx, q.SQLQuery)
	if err != nil {
		msg := fmt.Sprintf("query execution failed: %v", err)
		return models.QueryEvaluation{
			QueryConfigID:  q.ID,
			QueryName:      q.QueryName,
			EvaluationRule: q.EvaluationRule,
			RuleValue:      q.RuleValue,
			Hits:           0,
			Status:         models.QueryStatusError,
			ErrorMessage:   &msg,
			Evidence:       nil,
		}
	}

	const evidenceLimit = 10
	rawRows := res.Rows
	if len(rawRows) > evidenceLimit {
		rawRows = rawRows[:evidenceLimit]
	}

	evidence := make([]map[string]any, 0, len(rawRows))
	for _, row := range rawRows {
		rowMap := map[string]any{}
		for i, col := range row {
			rowMap[fmt.Sprintf("col_%d", i)] = col
		}
		evidence = append(evidence, rowMap)
	}

	status, errMsg := evaluateQueryRule(q, res.Count)

	return models.QueryEvaluation{
		QueryConfigID:  q.ID,
		QueryName:      q.QueryName,
		EvaluationRule: q.EvaluationRule,
		RuleValue:      q.RuleValue,
		Hits:           int64(res.Count),
		Status:         status,
		ErrorMessage:   errMsg,
		Evidence:       evidence,
	}
}

func evaluateQueryRule(q models.QueryConfig, hits int) (models.QueryEvaluationStatus, *string) {
	switch q.EvaluationRule {
	case models.NoHitsAllowed:
		if hits == 0 {
			return models.QueryStatusCompliant, nil
		}
		return models.QueryStatusNonCompliant, nil

	case models.MinHitsRequired:
		if q.RuleValue == nil {
			msg := "ruleValue is required for MIN_HITS_REQUIRED"
			return models.QueryStatusError, &msg
		}
		if hits >= *q.RuleValue {
			return models.QueryStatusCompliant, nil
		}
		return models.QueryStatusNonCompliant, nil

	case models.ThresholdMax:
		if q.RuleValue == nil {
			msg := "ruleValue is required for THRESHOLD_MAX"
			return models.QueryStatusError, &msg
		}
		if hits <= *q.RuleValue {
			return models.QueryStatusCompliant, nil
		}
		return models.QueryStatusNonCompliant, nil

	default:
		msg := "unknown evaluation rule"
		return models.QueryStatusError, &msg
	}
}

func computeControlStatus(strategy models.ComplianceStrategy, results []models.QueryEvaluation) models.ControlEvaluationStatus {

	hasCompliant := false
	hasNonCompliant := false
	hasError := false
	allNotApplicable := true

	applicableCount := 0
	errorCount := 0

	for _, r := range results {
		switch r.Status {
		case models.QueryStatusCompliant:
			hasCompliant = true
			allNotApplicable = false
			applicableCount++

		case models.QueryStatusNonCompliant:
			hasNonCompliant = true
			allNotApplicable = false
			applicableCount++

		case models.QueryStatusError:
			hasError = true
			allNotApplicable = false
			applicableCount++
			errorCount++

		case models.QueryStatusNotApplicable:
		}
	}

	if allNotApplicable {
		return models.ControlStatusNotApplicable
	}

	if applicableCount > 0 && errorCount == applicableCount {
		return models.ControlStatusNotEvaluated
	}

	switch strategy {
	case models.StrategyAny:
		if hasCompliant {
			return models.ControlStatusCompliant
		}

		if hasError {
			return models.ControlStatusNonCompliant
		}

		return models.ControlStatusNonCompliant

	case models.StrategyAll:

		if hasNonCompliant {
			return models.ControlStatusNonCompliant
		}

		if hasError {
			return models.ControlStatusNonCompliant
		}

		if hasCompliant {
			return models.ControlStatusCompliant
		}
	}

	return models.ControlStatusNonCompliant
}

func patternExists(pattern int, active []models.IndexPattern) bool {
	for _, p := range active {
		if int(p.ID) == pattern && p.Active {
			return true
		}
	}
	return false
}
