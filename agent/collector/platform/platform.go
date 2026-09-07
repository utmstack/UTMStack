package platform

import (
	"context"

	"github.com/threatwinds/go-sdk/plugins"
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
