package processor

import (
	"fmt"

	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
)

func (p *Processor) processAlertsInfo() {
	for alert := range p.AlertInfoQueue {
		alertInfo, err := elastic.GetAlertsInfo(alert.AlertID)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while getting alert %s info: %v", alert.AlertID, err), alert.AlertID)
			continue
		}

		details := schema.ConvertFromAlertToAlertDB(alertInfo)
		p.GPTQueue <- cleanAlerts(&details)
	}
}
