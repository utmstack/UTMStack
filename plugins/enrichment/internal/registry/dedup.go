package registry

import "sync"

var (
	missingDatasetSeen sync.Map
	missingColumnSeen  sync.Map
)

func MarkMissingDataset(datasetID string) (firstTime bool) {
	_, loaded := missingDatasetSeen.LoadOrStore(datasetID, struct{}{})
	return !loaded
}

func MarkMissingColumn(datasetID, column string) (firstTime bool) {
	key := datasetID + "|" + column
	_, loaded := missingColumnSeen.LoadOrStore(key, struct{}{})
	return !loaded
}
