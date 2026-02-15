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
	// 1. Obtener index patterns activos
	patterns, err := e.backend.GetActiveIndexPatterns(ctx)
	if err != nil {
		return models.ControlEvaluation{}, fmt.Errorf("failed to get index patterns: %w", err)
	}

	results := make([]models.QueryEvaluation, 0)
	applicable := make([]models.QueryEvaluation, 0) // solo queries con indexPattern activo

	// 2. Evaluar cada QueryConfig
	for _, q := range cfg.QueriesConfigs {
		catcher.Info("Evaluating query", map[string]any{
			"query_id": q.ID,
		})

		// 2.1 Si el index pattern NO está activo → NOT_APPLICABLE
		if !patternExists(int(q.IndexPatternID), patterns) {
			reason := "Index pattern not active"
			qr := models.QueryEvaluation{
				QueryConfigID: q.ID,
				QueryName:     q.QueryName,
				Hits:          0,
				Status:        models.QueryStatusNotApplicable,
				ErrorMessage:  &reason,
			}
			results = append(results, qr)
			continue
		}

		// 2.2 Evaluar query normalmente
		qr := e.evaluateQuery(ctx, q)
		results = append(results, qr)

		// 2.3 Solo las queries aplicables participan en la estrategia ALL/ANY
		if qr.Status != models.QueryStatusNotApplicable {
			applicable = append(applicable, qr)
		}
	}

	// 3. Combinar resultados según la estrategia del control
	finalStatus := computeControlStatus(cfg.ControlStrategy, applicable)

	// 4. Construir evaluación final
	return models.ControlEvaluation{
		ControlConfigID:  cfg.ID,
		ControlName:      cfg.ControlName,
		Status:           finalStatus,
		QueryEvaluations: results,
	}, nil
}

func (e *Evaluator) evaluateQuery(ctx context.Context, q models.QueryConfig) models.QueryEvaluation {
	// Ejecutar la query SQL real contra OpenSearch
	res, err := e.backend.ExecuteSQLQuery(ctx, q.SQLQuery)
	if err != nil {
		msg := fmt.Sprintf("query execution failed: %v", err)
		return models.QueryEvaluation{
			QueryConfigID: q.ID,
			QueryName:     q.QueryName,
			Hits:          0,
			Status:        models.QueryStatusError,
			ErrorMessage:  &msg,
			Evidence:      nil,
		}
	}

	// Limitar evidencia (por ejemplo, a 50 filas)
	const evidenceLimit = 50
	rawRows := res.Rows
	if len(rawRows) > evidenceLimit {
		rawRows = rawRows[:evidenceLimit]
	}

	// Convertir [][]any → []map[string]any para OpenSearch
	evidence := make([]map[string]any, 0, len(rawRows))
	for _, row := range rawRows {
		rowMap := map[string]any{}
		for i, col := range row {
			rowMap[fmt.Sprintf("col_%d", i)] = col
		}
		evidence = append(evidence, rowMap)
	}

	// Evaluar la regla con los hits obtenidos
	status, errMsg := evaluateQueryRule(q, res.Count)

	return models.QueryEvaluation{
		QueryConfigID: q.ID,
		QueryName:     q.QueryName,
		Hits:          int64(res.Count),
		Status:        status,
		ErrorMessage:  errMsg,
		Evidence:      evidence,
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

	switch strategy {

	case models.StrategyAll:
		// ALL → todas deben ser COMPLIANT
		for _, r := range results {
			if r.Status != models.QueryStatusCompliant {
				return models.ControlStatusNonCompliant
			}
		}
		return models.ControlStatusCompliant

	case models.StrategyAny:
		// ANY → basta con que una sea COMPLIANT
		for _, r := range results {
			if r.Status == models.QueryStatusCompliant {
				return models.ControlStatusCompliant
			}
		}
		return models.ControlStatusNonCompliant

	default:
		// fallback seguro
		return models.ControlStatusNonCompliant
	}
}

func patternExists(pattern int, active []models.IndexPattern) bool {
	for _, p := range active {
		if int(p.ID) == pattern && p.Active {
			return true
		}
	}
	return false
}
