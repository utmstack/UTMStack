package processor

import (
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
)

func (p *Processor) processAlertsInfo() {
	for alert := range p.AlertInfoQueue {
		catcher.Info("Processing alert info for ID", map[string]any{"alert": alert.AlertID})

		alertInfo, err := elastic.GetAlertsInfo(alert.AlertID)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while getting alert %s info: %v", alert.AlertID, err), alert.AlertID)
			continue
		}
		catcher.Info("Alert info retrieved successfully for ID", map[string]any{"alert": alert.AlertID})

		correlation, err := elastic.FindRelatedAlerts(alertInfo)
		if err != nil {
			catcher.Error("error finding related alerts", err, nil)
		}

		details := schema.ConvertFromAlertToAlertDB(alertInfo)

		if correlation != nil && len(correlation.RelatedAlerts) > 0 {
			correlationContext := elastic.BuildCorrelationContext(correlation)
			details.Description = details.Description + "\n\n" + correlationContext
		}

		p.GPTQueue <- cleanAlerts(&details)
	}
}
