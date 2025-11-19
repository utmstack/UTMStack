package processor

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
)

func (p *Processor) processAlertToElastic() {
	for alert := range p.ElasticQueue {
		gptConfig := configurations.GetGPTConfig()

		resp := schema.ConvertFromAlertDBToGPTResponse(alert)
		resp.Status = "Completed"
		query, err := schema.ConvertGPTResponseToUpdateQuery(resp)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error converting gpt response to update query: %v", err), alert.AlertID)
			continue
		}
		err = elastic.ElasticQuery(configurations.SOC_AI_INDEX, query, "update")
		if err != nil {
			p.RegisterError(fmt.Sprintf("error indexing gpt response in elastic: %v", err), alert.AlertID)
			continue
		}

		if gptConfig.ChangeAlertStatus {
			err = elastic.ChangeAlertStatus(alert.AlertID, configurations.API_ALERT_COMPLETED_STATUS_CODE, alert.GPTClassification+" - "+alert.GPTReasoning)
			if err != nil {
				catcher.Error("error while changing alert status in elastic", err, nil)
				continue
			}
			catcher.Info("alert status changed to COMPLETED in Panel", map[string]any{"alert": alert.AlertID})
		}

		if gptConfig.AutomaticIncidentCreation && alert.GPTClassification == "possible incident" {
			incidentsDetails, err := elastic.GetIncidentsByPattern("Incident in " + alert.DataSource)
			if err != nil {
				catcher.Error("error while getting incidents by pattern", err, nil)
				continue
			}

			incidentExists := false
			if len(incidentsDetails) != 0 {
				for _, incident := range incidentsDetails {
					if strings.HasSuffix(incident.IncidentName, "Incident in "+alert.DataSource) {
						incidentExists = true
						err = elastic.AddAlertToIncident(incident.ID, alert)
						if err != nil {
							catcher.Error("error while adding alert to incident", err, nil)
							continue
						}
					}
				}
			}

			if !incidentExists {
				err = elastic.CreateNewIncident(alert)
				if err != nil {
					catcher.Error("error while creating incident", err, nil)
					continue
				}
			}
			catcher.Info("alert added to incident in Panel", map[string]any{"alert": alert.AlertID})
		}

		catcher.Info("alert processed correctly", map[string]any{"alert": alert.AlertID})

	}
}
