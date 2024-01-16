package configuration

import (
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/updater/utils"
)

type Config struct {
	Server             string `yaml:"server"`
	AgentID            uint   `yaml:"agent-id"`
	AgentKey           string `yaml:"agent-key"`
	SkipCertValidation bool   `yaml:"insecure"`
}

type Environment struct {
	Branch string `yaml:"branch"`
}

// ReadEnv reads the environment file
// If the file does not exist, it returns a default environment with release branch
// If the file exists, it returns the environment
func ReadEnv() (string, error) {
	var env Environment
	path, err := utils.GetMyPath()
	if err != nil {
		return "", err
	}

	path = filepath.Join(path, "env.yml")

	if _, err = os.Stat(path); os.IsNotExist(err) {
		return "release", nil
	} else {
		err = utils.ReadYAML(path, &env)
		if err != nil {
			return "", err
		}
	}
	return env.Branch, nil
}
