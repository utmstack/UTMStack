package usecase

import (
	"strings"

	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
)

func phase2Baseline(alertNameCount int64, sameAssetCount int64, pastAlerts []alertdomain.UtmAlert) domain.PhaseResult {
	rarity := rarityScore(alertNameCount)
	novelty := assetNoveltyScore(sameAssetCount)
	modifier, detail := historicalResolutionModifier(pastAlerts)

	final := clampScore(float64(rarity + novelty + modifier))

	return domain.PhaseResult{
		Score: final,
		Breakdown: map[string]any{
			"alertCount":         alertNameCount,
			"rarityPoints":       rarity,
			"sameAssetCount":     sameAssetCount,
			"noveltyPoints":      novelty,
			"resolutionModifier": modifier,
			"resolutionDetail":   detail,
			"pastAlertsAnalyzed": len(pastAlerts),
		},
	}
}

func rarityScore(count int64) int {
	switch {
	case count == 0:
		return 40
	case count <= 2:
		return 35
	case count <= 5:
		return 28
	case count <= 15:
		return 20
	case count <= 50:
		return 12
	case count <= 200:
		return 6
	default:
		return 2
	}
}

func assetNoveltyScore(count int64) int {
	switch {
	case count == 0:
		return 30
	case count <= 2:
		return 22
	case count <= 5:
		return 15
	case count <= 15:
		return 8
	default:
		return 3
	}
}

func historicalResolutionModifier(pastAlerts []alertdomain.UtmAlert) (int, string) {
	if len(pastAlerts) == 0 {
		return 10, "no_history"
	}

	var falsePositive, completedNoIncident, partOfIncident int
	for _, a := range pastAlerts {
		isFP := false
		for _, t := range a.Tags {
			if strings.Contains(strings.ToLower(t), "false positive") {
				isFP = true
				break
			}
		}
		switch {
		case isFP:
			falsePositive++
		case a.IsIncident:
			partOfIncident++
		case a.Status == int(alertdomain.AlertStatusCompleted):
			completedNoIncident++
		}
	}

	totalResolved := falsePositive + completedNoIncident + partOfIncident
	if totalResolved == 0 {
		return 10, "no_resolved_history"
	}

	fpRatio := float64(falsePositive) / float64(totalResolved)
	incidentRatio := float64(partOfIncident) / float64(totalResolved)

	switch {
	case fpRatio >= 0.7:
		return -25, "mostly_false_positives"
	case fpRatio >= 0.4:
		return -15, "significant_false_positives"
	case incidentRatio >= 0.5:
		return 25, "mostly_incidents"
	case incidentRatio >= 0.2:
		return 15, "some_incidents"
	case completedNoIncident > falsePositive:
		return -5, "mostly_benign_resolved"
	default:
		return 0, "mixed_history"
	}
}
