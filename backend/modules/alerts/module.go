package alerts

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/handler"
	"github.com/utmstack/utmstack/backend/modules/alerts/repository"
	"github.com/utmstack/utmstack/backend/modules/alerts/usecase"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
	"gorm.io/gorm"
)

type Module struct {
	alertHandler        *handler.AlertHandler
	alertUsecase        connectors.AlertUsecase
	alertTagHandler     *handler.AlertTagHandler
	alertTagUsecase     connectors.AlertTagUsecase
	alertTagRuleHandler *handler.AlertTagRuleHandler
	alertTagRuleUsecase connectors.AlertTagRuleUsecase
	scheduler           *usecase.Scheduler
	schedulerEnabled    bool
}

func NewModule(
	db *gorm.DB,
	osClient *ospkg.Client,
	audit audit_connectors.Logger,
	schedulerEnabled bool,
) *Module {
	alertRepo := repository.NewOSAlertRepository(osClient)

	historyRecorder := repository.NewHistoryRecorder(osClient)

	alertUC := usecase.NewAlertUsecase(alertRepo, historyRecorder, audit)
	alertH := handler.NewAlertHandler(alertUC)

	alertTagRepo := repository.NewAlertTagRepository(db)
	alertTagUC := usecase.NewAlertTagUsecase(alertTagRepo)
	alertTagH := handler.NewAlertTagHandler(alertTagUC)

	alertTagRuleRepo := repository.NewAlertTagRuleRepository(db)
	alertTagRuleUC := usecase.NewAlertTagRuleUsecase(alertTagRuleRepo, alertTagRepo)
	alertTagRuleH := handler.NewAlertTagRuleHandler(alertTagRuleUC)

	sched := usecase.NewScheduler(alertRepo, alertTagRuleRepo, alertTagRepo, historyRecorder)

	return &Module{
		alertHandler:        alertH,
		alertUsecase:        alertUC,
		alertTagHandler:     alertTagH,
		alertTagUsecase:     alertTagUC,
		alertTagRuleHandler: alertTagRuleH,
		alertTagRuleUsecase: alertTagRuleUC,
		scheduler:           sched,
		schedulerEnabled:    schedulerEnabled,
	}
}

func (m *Module) Start(ctx context.Context) {
	if !m.schedulerEnabled {
		logger.Info("alerts scheduler: disabled (ALERTS_SCHEDULER_ENABLED=false)")
		return
	}
	logger.Info("alerts scheduler: enabled — launching goroutine")
	go m.scheduler.Start(ctx)
}

func (m *Module) GetAlertHandler() *handler.AlertHandler { return m.alertHandler }

func (m *Module) GetAlertUsecase() connectors.AlertUsecase { return m.alertUsecase }

func (m *Module) GetAlertTagHandler() *handler.AlertTagHandler { return m.alertTagHandler }

func (m *Module) GetAlertTagUsecase() connectors.AlertTagUsecase { return m.alertTagUsecase }

func (m *Module) GetAlertTagRuleHandler() *handler.AlertTagRuleHandler {
	return m.alertTagRuleHandler
}

func (m *Module) GetAlertTagRuleUsecase() connectors.AlertTagRuleUsecase {
	return m.alertTagRuleUsecase
}

func (m *Module) IsSchedulerEnabled() bool { return m.schedulerEnabled }
