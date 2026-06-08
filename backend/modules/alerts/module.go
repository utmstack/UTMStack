package alerts

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/handler"
	"github.com/utmstack/utmstack/backend/modules/alerts/repository"
	"github.com/utmstack/utmstack/backend/modules/alerts/usecase"
	"gorm.io/gorm"
)

type Module struct {
	alertHandler        *handler.AlertHandler
	alertUsecase        connectors.AlertUsecase
	alertTagHandler     *handler.AlertTagHandler
	alertTagRuleHandler *handler.AlertTagRuleHandler
	adversaryHandler    *handler.AdversaryHandler
	scheduler           *usecase.Scheduler
	schedulerEnabled    bool
}

func NewModule(
	db *gorm.DB,
	schedulerEnabled bool,
) *Module {
	alertRepo := repository.NewOSAlertRepository()

	historyRecorder := repository.NewHistoryRecorder()

	alertUC := usecase.NewAlertUsecase(alertRepo, historyRecorder)
	alertH := handler.NewAlertHandler(alertUC)

	alertTagRepo := repository.NewAlertTagRepository(db)
	alertTagUC := usecase.NewAlertTagUsecase(alertTagRepo)
	alertTagH := handler.NewAlertTagHandler(alertTagUC)

	alertTagRuleRepo := repository.NewAlertTagRuleRepository(db)
	alertTagRuleUC := usecase.NewAlertTagRuleUsecase(alertTagRuleRepo, alertTagRepo)
	alertTagRuleH := handler.NewAlertTagRuleHandler(alertTagRuleUC)

	sched := usecase.NewScheduler(alertRepo)

	adversaryUC := usecase.NewAdversaryUsecase()
	adversaryH := handler.NewAdversaryHandler(adversaryUC)

	return &Module{
		alertHandler:        alertH,
		alertUsecase:        alertUC,
		alertTagHandler:     alertTagH,
		alertTagRuleHandler: alertTagRuleH,
		adversaryHandler:    adversaryH,
		scheduler:           sched,
		schedulerEnabled:    schedulerEnabled,
	}
}

func (m *Module) Start(ctx context.Context) {
	if !m.schedulerEnabled {
		catcher.Info("alerts scheduler: disabled (ALERTS_SCHEDULER_ENABLED=false)", nil)
		return
	}
	catcher.Info("alerts scheduler: enabled — launching goroutine", nil)
	go m.scheduler.Start(ctx)
}

func (m *Module) GetAlertHandler() *handler.AlertHandler { return m.alertHandler }

func (m *Module) GetAlertUsecase() connectors.AlertUsecase { return m.alertUsecase }

func (m *Module) GetAlertTagHandler() *handler.AlertTagHandler { return m.alertTagHandler }

func (m *Module) GetAlertTagRuleHandler() *handler.AlertTagRuleHandler {
	return m.alertTagRuleHandler
}

func (m *Module) GetAdversaryHandler() *handler.AdversaryHandler { return m.adversaryHandler }

func (m *Module) IsSchedulerEnabled() bool { return m.schedulerEnabled }
