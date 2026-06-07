package datasources

import (
	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/handler"
)

type Module struct {
	datasourceHandler *handler.DatasourceHandler
	assetGroupHandler *handler.AssetGroupHandler
}

func NewModule(dsUC connectors.DatasourceUsecase, groupUC connectors.AssetGroupUsecase) *Module {
	return &Module{
		datasourceHandler: handler.NewDatasourceHandler(dsUC),
		assetGroupHandler: handler.NewAssetGroupHandler(groupUC),
	}
}

func (m *Module) GetDatasourceHandler() *handler.DatasourceHandler { return m.datasourceHandler }
func (m *Module) GetAssetGroupHandler() *handler.AssetGroupHandler { return m.assetGroupHandler }
