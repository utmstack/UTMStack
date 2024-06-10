package types

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/installer/utils"
)

type LogstashConfig struct {
	Host     string `yaml:"http.host"`
	Size     int    `yaml:"pipeline.batch.size"`
	ECS      string `yaml:"pipeline.ecs_compatibility"`
	Workers  int    `yaml:"pipeline.workers"`
	MaxBytes string `yaml:"queue.max_bytes"`
	Type     string `yaml:"queue.type"`
}

func CreateLogstashConfig(logstasConfigPath string) error {
	config := LogstashConfig{
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
