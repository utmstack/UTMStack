package incidents

import (
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/handler"
	"github.com/utmstack/utmstack/backend/modules/incidents/repository"
	"github.com/utmstack/utmstack/backend/modules/incidents/usecase"
	"gorm.io/gorm"
)

type Module struct {
	incidentHandler      *handler.IncidentHandler
	incidentAlertHandler *handler.IncidentAlertHandler
	incidentNoteHandler  *handler.IncidentNoteHandler
	historyHandler       *handler.IncidentHistoryHandler
}

func NewModule(
	db *gorm.DB,
	mailer connectors.IncidentMailer,
	alertsGateway connectors.AlertsGateway,
	iamGateway connectors.IAMGateway,
	audit audit_connectors.Logger,
) *Module {
	if mailer == nil {
		mailer = connectors.NewNoopMailer()
	}
	if alertsGateway == nil {
		alertsGateway = connectors.NewNoopAlertsGateway()
	}
	if iamGateway == nil {
		iamGateway = connectors.NewNoopIAMGateway()
	}

	incidentRepo := repository.NewIncidentRepository(db)
	alertRepo := repository.NewIncidentAlertRepository(db)
	noteRepo := repository.NewIncidentNoteRepository(db)
	historyRepo := repository.NewIncidentHistoryRepository(db)

	incidentUC := usecase.NewIncidentUsecase(incidentRepo, alertRepo, historyRepo, mailer, alertsGateway, iamGateway, audit)
	incidentAlertUC := usecase.NewIncidentAlertUsecase(alertRepo, historyRepo)
	incidentNoteUC := usecase.NewIncidentNoteUsecase(noteRepo, historyRepo)
	historyUC := usecase.NewIncidentHistoryUsecase(historyRepo)

	return &Module{
		incidentHandler:      handler.NewIncidentHandler(incidentUC),
		incidentAlertHandler: handler.NewIncidentAlertHandler(incidentAlertUC),
		incidentNoteHandler:  handler.NewIncidentNoteHandler(incidentNoteUC),
		historyHandler:       handler.NewIncidentHistoryHandler(historyUC),
	}
}

func (m *Module) GetIncidentHandler() *handler.IncidentHandler { return m.incidentHandler }
func (m *Module) GetIncidentAlertHandler() *handler.IncidentAlertHandler {
	return m.incidentAlertHandler
}
func (m *Module) GetIncidentNoteHandler() *handler.IncidentNoteHandler       { return m.incidentNoteHandler }
func (m *Module) GetIncidentHistoryHandler() *handler.IncidentHistoryHandler { return m.historyHandler }
