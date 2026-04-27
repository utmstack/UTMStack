package validations

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/plugins/modules-config/config"
	"github.com/utmstack/UTMStack/plugins/modules-config/validations/socai"
)




func ValidateSOCAIConfig(cfg *config.ModuleGroup) error {
	if cfg == nil {
		return fmt.Errorf("SOC AI configuration is not provided")
	}
	socai_config := parseSOCAIConfig(cfg)
	verificator := socai.NewSocaiVerification(socai_config)
	return verificator.Validate()
}

func parseSOCAIConfig(cfg *config.ModuleGroup) socai.SOCAIConfig {
	socai := socai.SOCAIConfig{
		AuthType:      "none",
		Model:         "",
		CustomHeaders: make(map[string]string),
	}

	for _, cnf := range cfg.ModuleGroupConfigurations {
		cnf.ConfValue=strings.Trim(cnf.ConfValue," ")
		switch cnf.ConfKey {
		case "utmstack.socai.autoAnalyze":
			socai.AutoAnalyze = cnf.ConfValue == "true"
		case "utmstack.socai.incidentCreation":
			socai.IncidentCreation = cnf.ConfValue == "true"
		case "utmstack.socai.changeAlertStatus":
			socai.ChangeAlertStatus = cnf.ConfValue == "true"
		case "utmstack.socai.provider":
			socai.Provider = cnf.ConfValue
		case "utmstack.socai.url":
			socai.URL = cnf.ConfValue
		case "utmstack.socai.model":
			socai.Model = cnf.ConfValue
		case "utmstack.socai.authType":
			if cnf.ConfValue != "" {
				socai.AuthType = cnf.ConfValue
			}
		case "utmstack.socai.maxTokens":
			socai.MaxTokens = cnf.ConfValue
		case "utmstack.socai.customHeaders":
			if cnf.ConfValue != "" {
				if err := json.Unmarshal([]byte(cnf.ConfValue), &socai.CustomHeaders); err != nil {
					fmt.Printf("Warning: Failed to parse customHeaders JSON: %v\n", err)
				}
			}
		}
	}

	return socai
}



