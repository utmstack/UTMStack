package validations

import (
	"fmt"

	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateModuleConfig(moduleName string, config *config.ModuleGroup) error {
	switch moduleName {
	case "AWS_IAM_USER":
		if err := ValidateAwsConfig(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "AZURE":
		if err := ValidateAzureConfig(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "BITDEFENDER":
		if err := ValidateBdgzConfig(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "GCP":
		if err := ValidateGcpConfig(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "O365":
		if err := ValidateO365Config(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "SOC_AI":
		if err := ValidateSOCAIConfig(config); err != nil {
			return fmt.Errorf("%v", err)
		}
	case "SOPHOS":
		// if err := ValidateSophosConfig(config); err != nil {
		// 	return fmt.Errorf("%v", err)
		// }
	default:
		return fmt.Errorf("unsupported module: %s", moduleName)
	}

	return nil
}
