package types

type LogstashConfig struct {
	Host     string `yaml:"http.host"`
	Size     int    `yaml:"pipeline.batch.size"`
	ECS      string `yaml:"pipeline.ecs_compatibility"`
	Workers  int    `yaml:"pipeline.workers"`
	MaxBytes string `yaml:"queue.max_bytes"`
	Type     string `yaml:"queue.type"`
}
