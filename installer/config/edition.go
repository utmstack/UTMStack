package config

import "github.com/utmstack/UTMStack/installer/utils"

type Edition struct {
	Edition string `yaml:"edition"`
}

func GetEdition() (*Edition, error) {
	edition := &Edition{}
	err := utils.ReadYAML(EditionFile, edition)
	if err != nil {
		return nil, err
	}
	return edition, nil
}

func SaveEdition(edition *Edition) error {
	return utils.WriteYAML(EditionFile, edition)
}
