package usecase

import (
	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/mitre"
)

func phase1Intrinsic(a *alertdomain.UtmAlert) domain.PhaseResult {
	impactPts, impactRaw := impactComponent(a)
	tacticPts, matchedTactic := mitre.TacticScore(a.Category)
	techID := mitre.ExtractTechniqueID(a.Technique, a.Name, a.Description, a.Category)
	techPts := mitre.TechniqueScore(techID)

	mult, ok := mitre.SeverityMultiplier[a.Severity]
	if !ok {
		mult = mitre.DefaultSeverityMultiplier
	}

	raw := float64(impactPts+tacticPts+techPts) * mult
	final := clampScore(raw)

	return domain.PhaseResult{
		Score: final,
		Breakdown: map[string]any{
			"impactScoreRaw":     impactRaw,
			"impactPoints":       impactPts,
			"category":           a.Category,
			"matchedTactic":      matchedTactic,
			"tacticPoints":       tacticPts,
			"techniqueId":        techID,
			"techniquePoints":    techPts,
			"severity":           a.Severity,
			"severityMultiplier": mult,
			"rawScore":           round1(raw),
		},
	}
}

func impactComponent(a *alertdomain.UtmAlert) (points int, raw int) {
	raw = a.ImpactScore
	if raw == 0 && a.Impact != nil {
		raw = a.Impact.Score
	}
	if raw <= 0 {
		switch a.Severity {
		case 4:
			return 36, 0
		case 3:
			return 28, 0
		case 2:
			return 16, 0
		case 1:
			return 8, 0
		default:
			return 12, 0
		}
	}
	if raw > 100 {
		raw = 100
	}
	pts := int(float64(raw)*0.40 + 0.5)
	if pts > 40 {
		pts = 40
	}
	return pts, a.ImpactScore
}
