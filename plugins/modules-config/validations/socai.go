package validations

import (
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateSOCAIConfig(config *config.ModuleGroup) error {
	var apiKey, provider string

	if config == nil {
		return catcher.Error("SOC_AI configuration is nil", nil, nil)
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
		return catcher.Error("API Key is required in SOC_AI configuration", nil, nil)
	}
	if provider == "" {
		return catcher.Error("Provider is required in SOC_AI configuration", nil, nil)
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
			return catcher.Error("SOC_AI connection timed out", err, map[string]any{
				"status_code": status,
				"response":    response,
			})
		}
		if status == http.StatusUnauthorized {
			return catcher.Error("SOC_AI API Key is invalid", err, map[string]any{
				"status_code": status,
				"response":    response,
			})
		}
		catcher.Info("Error validating SOC_AI connection", map[string]any{
			"error":       err,
			"status_code": status,
			"response":    response,
		})
		return catcher.Error("SOC_AI API Key is invalid", err, map[string]any{
			"status_code": status,
			"response":    response,
		})
	}

	return nil
}
