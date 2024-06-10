package main

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

func CreateLogstashConfig(logstasConfigPath string) error {
	config := types.LogstashConfig{
		Host:     "0.0.0.0",
		Size:     125,
		ECS:      "disabled",
		Workers:  4,
		MaxBytes: "1gb",
		Type:     "persisted",
	}

	err := utils.CreatePathIfNotExist(logstasConfigPath)
	if err != nil {
		return fmt.Errorf("error creating logstash config path: %v", err)
	}

	err = utils.WriteYAML(filepath.Join(logstasConfigPath, "logstash.yml"), &config)
	if err != nil {
		return fmt.Errorf("error writing logstash config: %v", err)
	}

	return nil
}
