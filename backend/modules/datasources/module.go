package datasources

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/handler"
	"github.com/utmstack/utmstack/backend/modules/datasources/usecase"
)

type Module struct {
	datasourceHandler *handler.DatasourceHandler
	assetGroupHandler *handler.AssetGroupHandler
	reconciler        *usecase.StatsReconciler
}

func NewModule(dsUC connectors.DatasourceUsecase, groupUC connectors.AssetGroupUsecase, reconciler *usecase.StatsReconciler) *Module {
	return &Module{
		datasourceHandler: handler.NewDatasourceHandler(dsUC),
		assetGroupHandler: handler.NewAssetGroupHandler(groupUC),
		reconciler:        reconciler,
	}
}

func (m *Module) Start(ctx context.Context) {
	if m.reconciler != nil {
		go m.reconciler.Start(ctx)
	}
}

func (m *Module) GetDatasourceHandler() *handler.DatasourceHandler { return m.datasourceHandler }
func (m *Module) GetAssetGroupHandler() *handler.AssetGroupHandler { return m.assetGroupHandler }
