package configuration

import (
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/runner/utils"
)

type Environment struct {
	Branch string `yaml:"branch"`
}

func ReadEnv() (*Environment, error) {
	var env Environment
	path, err := utils.GetMyPath()
	if err != nil {
		return nil, err
	}

	path = filepath.Join(path, "env.yml")

	if _, err = os.Stat(path); os.IsNotExist(err) {
		return &Environment{Branch: "release"}, nil
	} else {
		err = utils.ReadYAML(path, &env)
		if err != nil {
			return nil, err
		}
	}
	return &env, nil
}
