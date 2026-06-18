package usecase

import (
	"sort"

	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/mitre"
)

func phase4Chain(currentID, currentName string, related []alertdomain.UtmAlert) domain.PhaseResult {
	others := make([]alertdomain.UtmAlert, 0, len(related))
	for _, a := range related {
		if a.ID != currentID {
			others = append(others, a)
		}
	}

	volumePts, uniqueNames := volumeScore(others, currentName)
	chainPts, stageLabels, chainLen := killChainScore(related)
	temporalPts, spanHours := temporalClusteringScore(related)
	incidentPts, hasIncident := incidentModifier(others)

	final := clampScore(float64(volumePts + chainPts + temporalPts + incidentPts))

	return domain.PhaseResult{
		Score: final,
		Breakdown: map[string]any{
			"relatedAlertCount":   len(others),
			"uniqueAlertNames":    uniqueNames,
			"volumePoints":        volumePts,
			"killChainStages":     stageLabels,
			"killChainLength":     chainLen,
			"chainPoints":         chainPts,
			"timeSpanHours":       spanHours,
			"temporalPoints":      temporalPts,
			"hasExistingIncident": hasIncident,
			"incidentPoints":      incidentPts,
		},
	}
}

func volumeScore(others []alertdomain.UtmAlert, currentName string) (int, []string) {
	seen := map[string]bool{}
	for _, a := range others {
		if a.Name != "" && a.Name != currentName {
			seen[a.Name] = true
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	count := len(names)
	switch {
	case count == 0:
		return 0, names
	case count == 1:
		return 8, names
	case count <= 3:
		return 15, names
	case count <= 6:
		return 20, names
	default:
		return 25, names
	}
}

func killChainScore(all []alertdomain.UtmAlert) (int, []string, int) {
	stageSet := map[int]bool{}
	for _, a := range all {
		for _, s := range mitre.Stages(a.Category) {
			stageSet[s] = true
		}
	}
	if len(stageSet) == 0 {
		return 0, []string{}, 0
	}

	stages := make([]int, 0, len(stageSet))
	for s := range stageSet {
		stages = append(stages, s)
	}
	sort.Ints(stages)

	chainLen := longestChainWithGap(stages, 1)

	var score int
	switch {
	case chainLen >= 5:
		score = 40
	case chainLen >= 4:
		score = 32
	case chainLen >= 3:
		score = 24
	case chainLen >= 2:
		score = 15
	default:
		score = 0
	}

	// Resolve stage numbers back to representative tactic names.
	stageToTactic := map[int]string{}
	for tactic, num := range mitre.KillChainOrder {
		if stageSet[num] {
			if _, exists := stageToTactic[num]; !exists {
				stageToTactic[num] = tactic
			}
		}
	}
	labels := make([]string, 0, len(stages))
	for _, s := range stages {
		labels = append(labels, stageToTactic[s])
	}
	return score, labels, chainLen
}

func longestChainWithGap(sorted []int, maxGap int) int {
	if len(sorted) == 0 {
		return 0
	}
	best, current := 1, 1
	for i := 1; i < len(sorted); i++ {
		if sorted[i]-sorted[i-1] <= maxGap+1 {
			current++
			if current > best {
				best = current
			}
		} else {
			current = 1
		}
	}
	return best
}

func temporalClusteringScore(all []alertdomain.UtmAlert) (int, float64) {
	var times []int64 // unix seconds
	for _, a := range all {
		if t, ok := parseAlertTime(a.Timestamp); ok {
			times = append(times, t.Unix())
		}
	}
	if len(times) < 2 {
		return 0, 0
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	spanHours := float64(times[len(times)-1]-times[0]) / 3600.0

	switch {
	case spanHours <= 1:
		return 20, round1(spanHours)
	case spanHours <= 6:
		return 16, round1(spanHours)
	case spanHours <= 24:
		return 12, round1(spanHours)
	case spanHours <= 72:
		return 6, round1(spanHours)
	default:
		return 2, round1(spanHours)
	}
}

func incidentModifier(others []alertdomain.UtmAlert) (int, bool) {
	for _, a := range others {
		if a.IsIncident {
			return 15, true
		}
	}
	return 0, false
}
