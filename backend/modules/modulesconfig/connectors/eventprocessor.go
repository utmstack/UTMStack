package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/pkg/eventprocessor"
)

// EventProcessorClient is the seam the usecase uses to talk to
// event-processor-manager. The real implementation lives in pkg/eventprocessor;
// tests substitute a fake.
type EventProcessorClient interface {
	UpdateModule(ctx context.Context, nameShort string, module eventprocessor.ModulePayload) error
	ValidateModule(ctx context.Context, nameShort string, module eventprocessor.ModulePayload) (*eventprocessor.ValidationResult, error)
}
