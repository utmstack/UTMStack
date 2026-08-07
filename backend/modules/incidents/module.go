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

	incidentUsecase      connectors.IncidentUsecase
	incidentAlertUsecase connectors.IncidentAlertUsecase
	incidentNoteUsecase  connectors.IncidentNoteUsecase
	historyUsecase       connectors.IncidentHistoryUsecase
}

func NewModule(
	db *gorm.DB,
	mailer connectors.IncidentMailer,
	alertsGateway connectors.AlertsGateway,
	audit audit_connectors.Logger,
) *Module {
	incidentRepo := repository.NewIncidentRepository(db)
	alertRepo := repository.NewIncidentAlertRepository(db)
	noteRepo := repository.NewIncidentNoteRepository(db)
	historyRepo := repository.NewIncidentHistoryRepository(db)

	incidentUC := usecase.NewIncidentUsecase(incidentRepo, alertRepo, historyRepo, mailer, alertsGateway, audit)
	incidentAlertUC := usecase.NewIncidentAlertUsecase(alertRepo, historyRepo, incidentRepo)
	incidentNoteUC := usecase.NewIncidentNoteUsecase(noteRepo, historyRepo)
	historyUC := usecase.NewIncidentHistoryUsecase(historyRepo)

	return &Module{
		incidentHandler:      handler.NewIncidentHandler(incidentUC),
		incidentAlertHandler: handler.NewIncidentAlertHandler(incidentAlertUC),
		incidentNoteHandler:  handler.NewIncidentNoteHandler(incidentNoteUC),
		historyHandler:       handler.NewIncidentHistoryHandler(historyUC),
		incidentUsecase:      incidentUC,
		incidentAlertUsecase: incidentAlertUC,
		incidentNoteUsecase:  incidentNoteUC,
		historyUsecase:       historyUC,
	}
}

func (m *Module) GetIncidentHandler() *handler.IncidentHandler { return m.incidentHandler }
func (m *Module) GetIncidentAlertHandler() *handler.IncidentAlertHandler {
	return m.incidentAlertHandler
}
func (m *Module) GetIncidentNoteHandler() *handler.IncidentNoteHandler       { return m.incidentNoteHandler }
func (m *Module) GetIncidentHistoryHandler() *handler.IncidentHistoryHandler { return m.historyHandler }

func (m *Module) GetIncidentUsecase() connectors.IncidentUsecase { return m.incidentUsecase }
func (m *Module) GetIncidentAlertUsecase() connectors.IncidentAlertUsecase {
	return m.incidentAlertUsecase
}
func (m *Module) GetIncidentNoteUsecase() connectors.IncidentNoteUsecase {
	return m.incidentNoteUsecase
}
func (m *Module) GetIncidentHistoryUsecase() connectors.IncidentHistoryUsecase {
	return m.historyUsecase
}
