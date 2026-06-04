package connectors

import "context"

type DataInputStatusReader interface {
	FindDataSourcesToConfigure(ctx context.Context, excludeType string) ([]string, error)
}

// AssetSyncer operates on utm_network_scan (module #33, not yet ported).
// TODO(module-33): replace with real implementation when module #33 is ported.
type AssetSyncer interface {
	DeleteAllAssetsByDataType(ctx context.Context, dataTypes []string) error
	SynchronizeSourcesToAssets(ctx context.Context) error
}

type noopAssetSyncer struct{}

func NewNoopAssetSyncer() AssetSyncer {
	return &noopAssetSyncer{}
}

// TODO(module-33): implement when utm_network_scan is ported.
func (n *noopAssetSyncer) DeleteAllAssetsByDataType(_ context.Context, _ []string) error {
	return nil
}

// TODO(module-33): implement when utm_network_scan is ported.
func (n *noopAssetSyncer) SynchronizeSourcesToAssets(_ context.Context) error {
	return nil
}
