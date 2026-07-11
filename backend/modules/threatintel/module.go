package threatintel

import (
	"context"

	"github.com/utmstack/utmstack/backend/pkg/instanceconfig"
)

type Module struct {
	updatesDir string
}

func NewModule(updatesDir string) *Module {
	instanceconfig.Init(updatesDir)
	return &Module{updatesDir: updatesDir}
}

func (m *Module) Start(ctx context.Context) {}
