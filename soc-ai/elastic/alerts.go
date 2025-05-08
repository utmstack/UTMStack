package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

func GetAlertsInfo(id string) (schema.Alert, error) {
	result, err := ElasticSearch(configurations.ALERT_INDEX_PATTERN, "id", id)
	if err != nil {
		return schema.Alert{}, fmt.Errorf("error while getting alert %s info: %v", id, err)
	}

	var alertDetails schema.AlertDetails
	err = json.Unmarshal(result, &alertDetails)
	if err != nil {
		return schema.Alert{}, fmt.Errorf("error decoding response: %v", err)
	}

	if len(alertDetails) == 0 {
		return schema.Alert{}, fmt.Errorf("no alert found for id")
	}

	if len(alertDetails[0].Logs) > 0 {
		var logs []string
		if len(alertDetails[0].Logs) > 3 {
			logs = alertDetails[0].Logs[:3]
		} else {
			logs = alertDetails[0].Logs
		}

		for i, log := range logs {
			resp, err := ElasticSearch(configurations.LOGS_INDEX_PATTERN, "id", log)
			if err != nil {
				continue
			}
			alertDetails[0].Logs[i] = string(resp)
		}
	}
	return alertDetails[0], nil
}

func ChangeAlertStatus(id string, status int, observations string) error {
	url := configurations.GetPanelServiceName() + configurations.API_ALERT_STATUS_ENDPOINT
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": configurations.GetInternalKey(),
	}

	body := schema.ChangeAlertStatus{AlertIDs: []string{id}, Status: status, StatusObservation: observations}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling body: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, configurations.HTTP_TIMEOUT)
	if err != nil || statusCode != http.StatusOK {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	return nil
}

type AlertCounts struct {
	Incidents     int
	FalsePositive int
	Standard      int
	Unclassified  int
}

type MatchTypeCounts struct {
	SourceIP        AlertCounts
	DestinationIP   AlertCounts
	SourceUser      AlertCounts
	DestinationUser AlertCounts
}

type AlertCorrelation struct {
	CurrentAlert  schema.Alert
	RelatedAlerts []schema.Alert
	Counts        MatchTypeCounts
}

func GetRelatedAlerts(alertName string) ([]schema.Alert, error) {
	result, err := ElasticSearch(configurations.ALERT_INDEX_PATTERN, "name", alertName)
	if err != nil {
		return nil, fmt.Errorf("error getting historical alerts: %v", err)
	}

	var alerts []schema.Alert
	if err := json.Unmarshal(result, &alerts); err != nil {
		return nil, fmt.Errorf("error unmarshalling alerts: %v", err)
	}

	return alerts, nil
}

func FindRelatedAlerts(current schema.Alert) (*AlertCorrelation, error) {
	alerts, err := GetRelatedAlerts(current.Name)
	if err != nil {
		return nil, err
	}

	corr := &AlertCorrelation{CurrentAlert: current}
	for _, hist := range alerts {
		if hist.ID == current.ID {
			continue
		}
		if related, matches := isAlertRelated(current, hist); related {
			classif := getAlertClassification(hist)
			for _, m := range matches {
				incrementCount(&corr.Counts, m, classif)
			}
			corr.RelatedAlerts = append(corr.RelatedAlerts, hist)
		}
	}
	return corr, nil
}

func isAlertRelated(current, historical schema.Alert) (bool, []string) {
	if current.ID == historical.ID || current.Name != historical.Name {
		return false, nil
	}

	var matches []string

	if current.Source.IP != "" && current.Source.IP == historical.Source.IP {
		matches = append(matches, "SourceIP")
	}
	if current.Destination.IP != "" && current.Destination.IP == historical.Destination.IP {
		matches = append(matches, "DestinationIP")
	}
	if current.Source.User != "" && current.Source.User == historical.Source.User {
		matches = append(matches, "SourceUser")
	}
	if current.Destination.User != "" && current.Destination.User == historical.Destination.User {
		matches = append(matches, "DestinationUser")
	}

	sort.Strings(matches)
	return len(matches) > 0, matches
}

func getAlertClassification(alert schema.Alert) string {
	if len(alert.Tags) == 0 {
		return "Unclassified alert"
	}
	switch strings.ToLower(alert.Tags[0]) {
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

func incrementCount(cnts *MatchTypeCounts, matchType, classif string) {
	var ac *AlertCounts

	switch matchType {
	case "SourceIP":
		ac = &cnts.SourceIP
	case "DestinationIP":
		ac = &cnts.DestinationIP
	case "SourceUser":
		ac = &cnts.SourceUser
	case "DestinationUser":
		ac = &cnts.DestinationUser
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

func BuildCorrelationContext(corr *AlertCorrelation) string {
	if corr == nil || len(corr.RelatedAlerts) == 0 {
		return "No related alerts exist"
	}
	// Group alerts by matches and classifications
	// Example: "SourceIP+DestinationIP" -> { "Possible incident": 2, "False positive": 1 }
	groups := make(map[string]map[string]int)
	for _, alert := range corr.RelatedAlerts {
		if rel, mts := isAlertRelated(corr.CurrentAlert, alert); rel {
			key := strings.Join(mts, "+")
			if _, ok := groups[key]; !ok {
				groups[key] = make(map[string]int)
			}
			classif := getAlertClassification(alert)
			groups[key][classif]++
		}
	}
	// Ordered summary
	var sb strings.Builder
	total := len(corr.RelatedAlerts)
	sb.WriteString("\nHistorical Context: ")
	sb.WriteString(fmt.Sprintf("In the past, there are %d alerts with the same name", total))

	// Ordered keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sub := groups[k]
		// Count total for this group
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
