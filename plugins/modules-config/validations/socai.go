package validations

import (
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateSOCAIConfig(config *config.ModuleGroup) error {
	var apiKey, provider string

	if config == nil {
		return fmt.Errorf("SOC_AI configuration is nil")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch {
		case cnf.ConfKey == "utmstack.socai.key":
			apiKey = cnf.ConfValue
		case cnf.ConfKey == "utmstack.socai.provider":
			provider = cnf.ConfValue
		}
	}

	if apiKey == "" {
		return fmt.Errorf("API Key is required in SOC_AI configuration")
	}
	if provider == "" {
		return fmt.Errorf("Provider is required in SOC_AI configuration")
	} else if provider != "openai" {
		return nil
	}

	url := "https://api.openai.com/v1/chat/completions"
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
		"Content-Type":  "application/json",
	}

	response, status, err := utils.DoReq[map[string]any](url, nil, "GET", headers)
	if err != nil || status != http.StatusOK {
		if status == http.StatusRequestTimeout {
			return fmt.Errorf("SOC_AI connection timed out")
		}
		if status == http.StatusUnauthorized {
			return fmt.Errorf("SOC_AI API Key is invalid")
		}
		fmt.Printf("Error validating SOC_AI connection: %v, status code: %d, response: %v\n", err, status, response)
		return fmt.Errorf("SOC_AI API Key is invalid")
	}

	return nil
}
