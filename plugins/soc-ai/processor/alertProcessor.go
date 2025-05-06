package processor

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func (p *Processor) processAlertsInfo() {
	for alert := range p.AlertInfoQueue {
		correlation, err := elastic.FindRelatedAlerts(alert)
		if err != nil {
			_ = catcher.Error("error finding related alerts: %v", err, nil)
		}

		details := schema.ConvertFromAlertToAlertDB(alert)

		if correlation != nil && len(correlation.RelatedAlerts) > 0 {
			correlationContext := elastic.BuildCorrelationContext(correlation)
			details.Description = details.Description + "\n\n" + correlationContext
		}

		p.GPTQueue <- cleanAlerts(&details)
	}
}
