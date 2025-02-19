package correlation

import (
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/rules"
	"github.com/utmstack/UTMStack/correlation/utils"
)

func processResponse(logs []string, rule rules.Rule, save []utils.SavedField, tmpLogs *[20][]map[string]string,
	steps, step, minCount int) {
	if len(logs) >= func() int {
		switch minCount {
		case 0:
			return 1
		default:
			return minCount
		}
	}() {
		for _, l := range logs {
			fields := utils.ExtractDetails(save, l)

			// Alert in the last step or save data to the next iteration
			if steps-1 == step {
				// Use content of AlertName as Name if exists
				var alertName string
				if fields["AlertName"] != "" {
					alertName = fields["AlertName"]
				} else {
					alertName = rule.Name
				}
				// Use content of AlertCategory as Category if exists
				var alertCategory string
				if fields["AlertCategory"] != "" {
					alertCategory = fields["AlertCategory"]
				} else {
					alertCategory = rule.Category
				}

				// Run alert function
				Alert(
					alertName,
					rule.Severity,
					rule.Description,
					rule.Solution,
					alertCategory,
					rule.Tactic,
					rule.Reference,
					gjson.Get(l, "dataType").String(),
					gjson.Get(l, "dataSource").String(),
					fields,
				)
			} else {
				tmpLogs[step] = append(tmpLogs[step], fields)
			}
		}
	}
}
