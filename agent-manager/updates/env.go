package updates

import (
	"os"
	"path"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

type Environment struct {
	Branch string `yaml:"branch"`
}

func readEnv() (string, error) {
	var env Environment
	configPath := path.Join(config.VOLPATH, "config.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "release", nil
	} else {
		err = utils.ReadYAML(configPath, &env)
		if err != nil {
			return "", err
		}
	}

	switch env.Branch {
	case "v10-dev":
		return "dev", nil
	case "v10-qa":
		return "qa", nil
	case "v10-rc":
		return "rc", nil
	default:
		return "release", nil
	}
}
