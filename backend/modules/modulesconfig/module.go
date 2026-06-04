package modulesconfig

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/handler"
)

type Module struct {
	moduleUsecase connectors.ModuleUsecase
	groupUsecase  connectors.GroupUsecase
	configUsecase connectors.ConfigUsecase
	factory       connectors.ModuleFactory

	moduleHandler *handler.ModuleHandler
	groupHandler  *handler.GroupHandler
	configHandler *handler.ConfigurationHandler
}

func NewModule(
	moduleUC connectors.ModuleUsecase,
	groupUC connectors.GroupUsecase,
	configUC connectors.ConfigUsecase,
	factory connectors.ModuleFactory,
) *Module {
	return &Module{
		moduleUsecase: moduleUC,
		groupUsecase:  groupUC,
		configUsecase: configUC,
		factory:       factory,
		moduleHandler: handler.NewModuleHandler(moduleUC),
		groupHandler:  handler.NewGroupHandler(groupUC),
		configHandler: handler.NewConfigurationHandler(configUC),
	}
}

// Factory exposes the module-kind registry so other bootstrap code (per-kind
// init in modulekinds/) can register implementations.
func (m *Module) Factory() connectors.ModuleFactory { return m.factory }

// ModuleUsecase is exposed for cross-module use (e.g. collectors may need to
// query module activation state).
func (m *Module) ModuleUsecase() connectors.ModuleUsecase { return m.moduleUsecase }
