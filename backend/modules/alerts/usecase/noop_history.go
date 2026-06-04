package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
)

type noopHistoryRecorder struct{}

func (n *noopHistoryRecorder) Record(_ context.Context, _ []connectors.HistoryEntry) error {
	return nil
}

func NewNoopHistoryRecorder() connectors.HistoryRecorder {
	return &noopHistoryRecorder{}
}
