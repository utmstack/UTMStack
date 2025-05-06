package processor

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func (p *Processor) processAlertToElastic() {
	for alert := range p.ElasticQueue {
		gptConfig := configurations.GetPluginConfig()

		resp := schema.ConvertFromAlertDBToGPTResponse(alert)
		resp.Status = "Completed"
		query, err := schema.ConvertGPTResponseToUpdateQuery(resp)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error converting gpt response to update query: %v", err), alert.ID)
			continue
		}
		err = elastic.ElasticQuery(configurations.SOC_AI_INDEX, query, "update")
		if err != nil {
			p.RegisterError(fmt.Sprintf("error indexing gpt response in elastic: %v", err), alert.ID)
			continue
		}

		if gptConfig.ChangeAlertStatus && alert.GPTClassification == "possible false positive" {
			err = elastic.ChangeAlertStatus(alert.ID, configurations.API_ALERT_COMPLETED_STATUS_CODE, alert.GPTClassification+" - "+alert.GPTReasoning)
			if err != nil {
				_ = catcher.Error("error while changing alert status in elastic: %v", err, nil)
				continue
			}
		}

		if gptConfig.AutomaticIncidentCreation && alert.GPTClassification == "possible incident" {
			incidentsDetails, err := elastic.GetIncidentsByPattern("Incident in " + alert.DataSource)
			if err != nil {
				_ = catcher.Error("error while getting incidents by pattern: %v", err, nil)
				continue
			}

			incidentExists := false
			if len(incidentsDetails) != 0 {
				for _, incident := range incidentsDetails {
					if strings.HasSuffix(incident.IncidentName, "Incident in "+alert.DataSource) {
						incidentExists = true
						err = elastic.AddAlertToIncident(incident.ID, alert)
						if err != nil {
							_ = catcher.Error("error while adding alert to incident: %v", err, nil)
							continue
						}
					}
				}
			}

			if !incidentExists {
				err = elastic.CreateNewIncident(alert)
				if err != nil {
					_ = catcher.Error("error while creating incident: %v", err, nil)
					continue
				}
			}
		}

	}
}
