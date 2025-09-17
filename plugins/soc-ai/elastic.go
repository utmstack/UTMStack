package main

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func processAlertToElastic(alert *schema.AlertFields) error {
	resp := elastic.ConvertFromAlertDBToGPTResponse(alert)
	resp.Status = "Completed"

	err := elastic.ElasticQuery(config.SOC_AI_INDEX, resp, "update")
	if err != nil {
		if strings.Contains(err.Error(), "index_not_found_exception") || strings.Contains(err.Error(), "no such index") {

			if createErr := elastic.CreateIndexIfNotExist(config.SOC_AI_INDEX); createErr != nil {
				return fmt.Errorf("error updating alert in elastic: %v (failed to create index: %v)", err, createErr)
			}

			if retryErr := elastic.ElasticQuery(config.SOC_AI_INDEX, resp, "update"); retryErr != nil {
				return fmt.Errorf("error updating alert in elastic after index creation: %v", retryErr)
			}
		} else {
			return fmt.Errorf("error updating alert in elastic: %v", err)
		}
	}

	if config.GetConfig().ChangeAlertStatus {
		err = elastic.ChangeAlertStatus(alert.ID, config.API_ALERT_COMPLETED_STATUS_CODE, alert.DataSource, alert.GPTClassification+" - "+alert.GPTReasoning)
		if err != nil {
			_ = catcher.Error("error while changing alert status in elastic: %v", err, nil)
		}
	}

	if config.GetConfig().AutomaticIncidentCreation && alert.GPTClassification == "possible incident" {
		incidentsDetails, err := elastic.GetIncidentsByPattern("Incident in " + alert.DataSource)
		if err != nil {
			_ = catcher.Error("error while getting incidents by pattern: %v", err, nil)
		}

		incidentExists := false
		if len(incidentsDetails) != 0 {
			for _, incident := range incidentsDetails {
				if strings.HasSuffix(incident.IncidentName, "Incident in "+alert.DataSource) {
					incidentExists = true
					err = elastic.AddAlertToIncident(incident.ID, alert)
					if err != nil {
						_ = catcher.Error("error while adding alert to incident: %v", err, nil)
					}
				}
			}
		}

		if !incidentExists {
			err = elastic.CreateNewIncident(alert)
			if err != nil {
				_ = catcher.Error("error while creating incident: %v", err, nil)
			}
		}
	}
	return nil
}
