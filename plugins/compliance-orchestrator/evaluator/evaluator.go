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

func (e *Evaluator) Evaluate(ctx context.Context, cfg models.ReportConfig) (models.Evaluation, error) {
	// 1. Obtener index patterns activos
	_, err := e.backend.GetActiveIndexPatterns(ctx)
	if err != nil {
		return models.Evaluation{}, fmt.Errorf("failed to get index patterns: %w", err)
	}

	// 2. Evaluar cada QuerySpec
	/*var results []models.QueryResult*/
	for _, q := range cfg.Queries {
		/*qr := e.evaluateQuery(ctx, q, patterns)
		results = append(results, qr)*/
		catcher.Info("Evaluating query", map[string]any{
			"query_id": q.ID,
		})
	}

	/*final := combineResults(cfg, results)

	return final, nil*/

	return models.Evaluation{}, nil
}

func (e *Evaluator) evaluateQuery(ctx context.Context, q models.QuerySpec, patterns []models.IndexPattern) models.QueryResult {

	/*if !patternExists(q.IndexPatternID, patterns) {
		return models.QueryResult{
			QueryID: int(q.ID),
			Status:  models.StatusNotApplicable,
			Reason:  "Index pattern not active",
		}
	}*/

	return models.QueryResult{
		QueryID: int(q.ID),
		Status:  models.StatusCompliant,
		Reason:  "Query executed successfully (placeholder)",
	}
}

func patternExists(pattern int, active []models.IndexPattern) bool {
	for _, p := range active {
		if p.ID == pattern && p.Active {
			return true
		}
	}
	return false
}

func combineResults(cfg models.ReportConfig, results []models.QueryResult) models.Evaluation {
	final := models.Evaluation{
		ReportID: int(cfg.ID),
		Results:  results,
	}

	// Estrategia simple: si alguna query es NON_COMPLIANT → NON_COMPLIANT
	for _, r := range results {
		if r.Status == models.StatusNonCompliant {
			final.Status = models.StatusNonCompliant
			return final
		}
	}

	final.Status = models.StatusCompliant
	return final
}
