package processor

import (
	"fmt"

	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

func (p *Processor) processAlertsInfo() {
	for alert := range p.AlertInfoQueue {
		utils.Logger.Info("Processing alert info for ID: %s", alert.AlertID)

		alertInfo, err := elastic.GetAlertsInfo(alert.AlertID)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while getting alert %s info: %v", alert.AlertID, err), alert.AlertID)
			continue
		}
		utils.Logger.Info("Alert info retrieved successfully for ID: %s", alert.AlertID)

		correlation, err := elastic.FindRelatedAlerts(alertInfo)
		if err != nil {
			utils.Logger.ErrorF("error finding related alerts: %v", err)
		}

		details := schema.ConvertFromAlertToAlertDB(alertInfo)

		if correlation != nil && len(correlation.RelatedAlerts) > 0 {
			correlationContext := elastic.BuildCorrelationContext(correlation)
			details.Description = details.Description + "\n\n" + correlationContext
		}

		p.GPTQueue <- cleanAlerts(&details)
	}
}
