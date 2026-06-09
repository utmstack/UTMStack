package alerts

import (
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
}

func NewModule(db *gorm.DB) *Module {
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

	adversaryUC := usecase.NewAdversaryUsecase()
	adversaryH := handler.NewAdversaryHandler(adversaryUC)

	return &Module{
		alertHandler:        alertH,
		alertUsecase:        alertUC,
		alertTagHandler:     alertTagH,
		alertTagRuleHandler: alertTagRuleH,
		adversaryHandler:    adversaryH,
	}
}

func (m *Module) GetAlertHandler() *handler.AlertHandler { return m.alertHandler }

func (m *Module) GetAlertUsecase() connectors.AlertUsecase { return m.alertUsecase }

func (m *Module) GetAlertTagHandler() *handler.AlertTagHandler { return m.alertTagHandler }

func (m *Module) GetAlertTagRuleHandler() *handler.AlertTagRuleHandler {
	return m.alertTagRuleHandler
}

func (m *Module) GetAdversaryHandler() *handler.AdversaryHandler { return m.adversaryHandler }
