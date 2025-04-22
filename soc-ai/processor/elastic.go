package processor

import (
	"fmt"
	"strings"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
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

		if gptConfig.ChangeAlertStatus && alert.GPTClassification == "possible false positive" {
			err = elastic.ChangeAlertStatus(alert.AlertID, configurations.API_ALERT_COMPLETED_STATUS_CODE, alert.GPTClassification+" - "+alert.GPTReasoning)
			if err != nil {
				utils.Logger.ErrorF("error while changing alert status in elastic: %v", err)
				continue
			}
			utils.Logger.Info("alert %s status changed to COMPLETED in Panel", alert.AlertID)
		}

		if gptConfig.AutomaticIncidentCreation && alert.GPTClassification == "possible incident" {
			incidentsDetails, err := elastic.GetIncidentsByPattern("Incident in " + alert.DataSource)
			if err != nil {
				utils.Logger.ErrorF("error while getting incidents by pattern: %v", err)
				continue
			}

			incidentExists := false
			if len(incidentsDetails) != 0 {
				for _, incident := range incidentsDetails {
					if strings.HasSuffix(incident.IncidentName, "Incident in "+alert.DataSource) {
						incidentExists = true
						err = elastic.AddAlertToIncident(incident.ID, alert)
						if err != nil {
							utils.Logger.ErrorF("error while adding alert to incident: %v", err)
							continue
						}
					}
				}
			}

			if !incidentExists {
				err = elastic.CreateNewIncident(alert)
				if err != nil {
					utils.Logger.ErrorF("error while creating incident: %v", err)
					continue
				}
			}
			utils.Logger.Info("alert %s added to incident in Panel", alert.AlertID)
		}

		utils.Logger.Info("alert %s processed correctly", alert.AlertID)

	}
}
