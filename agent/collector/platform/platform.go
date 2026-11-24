package platform

import (
	"context"

	"github.com/utmstack/go-sdk/v2/plugins"
)

// CollectorConfig holds paths for log collection.
type CollectorConfig struct {
	LogsPath    string
	LogFileName string
}

type Collector interface {
	Name() string
	Start(ctx context.Context, enqueue func(*plugins.Log) error)
	Stop()
}
