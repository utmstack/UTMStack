package correlation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func GetCorrelationContext(alert schema.AlertFields) (string, error) {
	relatedAlerts, err := findRelatedAlerts(alert)
	if err != nil {
		return "", fmt.Errorf("error finding related alerts: %v", err)
	}

	if len(relatedAlerts.RelatedAlerts) > 0 {
		correlationContext := buildCorrelationContext(relatedAlerts)
		return fmt.Sprintf(" %s", correlationContext), nil
	}

	return "", nil
}

func findRelatedAlerts(current schema.AlertFields) (schema.AlertCorrelation, error) {
	correlation := schema.AlertCorrelation{CurrentAlert: current}

	result, err := elastic.ElasticSearch(config.ALERT_INDEX_PATTERN, "name", current.Name)
	if err != nil {
		return correlation, fmt.Errorf("error getting historical alerts: %v", err)
	}

	var alerts []schema.AlertFields
	if err := json.Unmarshal(result, &alerts); err != nil {
		return correlation, fmt.Errorf("error unmarshalling alerts: %v", err)
	}

	for _, hist := range alerts {
		if hist.ID == current.ID {
			continue
		}
		if related, matches := isAlertRelated(current, hist); related {
			classif := getAlertClassification(hist.Tags)
			for _, m := range matches {
				incrementCount(correlation.Counts, m, classif)
			}
			correlation.RelatedAlerts = append(correlation.RelatedAlerts, hist)
		}
	}
	return correlation, nil
}

func isAlertRelated(current, historical schema.AlertFields) (bool, []string) {
	if current.ID == historical.ID || current.Name != historical.Name {
		return false, nil
	}

	var matches []string

	if current.Adversary != nil && current.Adversary.Ip != "" && historical.Adversary != nil && current.Adversary.Ip == historical.Adversary.Ip {
		matches = append(matches, "AdversaryIP")
	}
	if current.Target != nil && current.Target.Ip != "" && historical.Target != nil && current.Target.Ip == historical.Target.Ip {
		matches = append(matches, "TargetIP")
	}
	if current.Adversary != nil && current.Adversary.User != "" && historical.Adversary != nil && current.Adversary.User == historical.Adversary.User {
		matches = append(matches, "AdversaryUser")
	}
	if current.Target != nil && current.Target.User != "" && historical.Adversary != nil && current.Target.User == historical.Adversary.User {
		matches = append(matches, "AdversaryUser")
	}

	sort.Strings(matches)
	return len(matches) > 0, matches
}

func getAlertClassification(tags []string) string {
	if len(tags) == 0 {
		return "Unclassified alert"
	}
	switch strings.ToLower(tags[0]) {
	case "possible incident":
		return "Possible incident"
	case "false positive":
		return "False positive"
	case "standard alert":
		return "Standard alert"
	default:
		return "Unclassified alert"
	}
}

func incrementCount(cnts schema.MatchTypeCounts, matchType, classif string) {
	var ac schema.AlertCounts

	switch matchType {
	case "SourceIP":
		ac = cnts.OriginIP
	case "DestinationIP":
		ac = cnts.TargetIP
	case "SourceUser":
		ac = cnts.OriginUser
	case "DestinationUser":
		ac = cnts.TargetUser
	}
	switch classif {
	case "Possible incident":
		ac.Incidents++
	case "False positive":
		ac.FalsePositive++
	case "Standard Alert":
		ac.Standard++
	default:
		ac.Unclassified++
	}
}

func buildCorrelationContext(corr schema.AlertCorrelation) string {
	if len(corr.RelatedAlerts) == 0 {
		return "No related alerts exist"
	}

	groups := make(map[string]map[string]int)
	for _, alert := range corr.RelatedAlerts {
		if rel, mts := isAlertRelated(corr.CurrentAlert, alert); rel {
			key := strings.Join(mts, "+")
			if _, ok := groups[key]; !ok {
				groups[key] = make(map[string]int)
			}
			classif := getAlertClassification(alert.Tags)
			groups[key][classif]++
		}
	}

	var sb strings.Builder
	total := len(corr.RelatedAlerts)
	sb.WriteString("\nHistorical Context: ")
	sb.WriteString(fmt.Sprintf("In the past, there are %d alerts with the same name", total))

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sub := groups[k]

		n := 0
		for _, v := range sub {
			n += v
		}
		sb.WriteString(fmt.Sprintf("\n- %d match the same %s", n, translateMatchTypes(strings.Split(k, "+"))))
		if n > 0 {
			sb.WriteString(" and of these " + formatClassifications(sub))
		}
	}
	return sb.String()
}

func translateMatchTypes(types []string) string {
	sort.Strings(types)
	var out []string

	for _, t := range types {
		switch t {
		case "SourceIP":
			out = append(out, "Source IP")
		case "DestinationIP":
			out = append(out, "Destination IP")
		case "SourceUser":
			out = append(out, "Source User")
		case "DestinationUser":
			out = append(out, "Destination User")
		}
	}
	return strings.Join(out, " and ")
}

func formatClassifications(m map[string]int) string {
	parts := make([]string, 0, len(m))
	for classif, cnt := range m {
		parts = append(parts, fmt.Sprintf("%d were classified as %s", cnt, classif))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
