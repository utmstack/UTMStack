package threatintel

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/threatintel/internal"
)

type Module struct {
	updatesDir string
}

func NewModule(updatesDir string) *Module {
	m := &Module{updatesDir: updatesDir}

	if _, err := internal.LoadInstanceConfig(updatesDir); err != nil {
		catcher.Warn("threatintel: failed to load instance config", map[string]any{"error": err.Error()})
	}

	return m
}

func (m *Module) Start(ctx context.Context) {}
