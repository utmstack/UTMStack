package initializer

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/association"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/extractor"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/mapper"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/scheduler"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/service"
)

func (a *App) buildProcessingPipeline() error {
	catcher.Info("building processing pipeline", nil)

	fieldExtractor := extractor.NewFieldExtractor()
	entityMapper := mapper.NewEntityMapper()
	associationBuilder := association.NewAssociationBuilder()

	entityBuilder := service.NewEntityBuilder(entityMapper)

	alertProcessor := service.NewAlertProcessor(
		a.clients.OpenSearch,
		fieldExtractor,
		entityBuilder,
	)

	incidentProcessor := service.NewIncidentProcessor(
		a.clients,
		alertProcessor,
		associationBuilder,
	)

	a.scheduler = scheduler.NewIngestionScheduler(
		a.config,
		a.clients,
		incidentProcessor,
	)

	catcher.Info("processing pipeline built successfully", nil)
	return nil
}
