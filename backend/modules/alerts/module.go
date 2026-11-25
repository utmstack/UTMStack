package alerts

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/handler"
	"github.com/utmstack/utmstack/backend/modules/alerts/repository"
	"github.com/utmstack/utmstack/backend/modules/alerts/usecase"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"gorm.io/gorm"
)

type Module struct {
	alertHandler        *handler.AlertHandler
	alertUsecase        connectors.AlertUsecase
	alertTagHandler     *handler.AlertTagHandler
	alertTagUsecase     connectors.AlertTagUsecase
	alertTagRuleHandler *handler.AlertTagRuleHandler
	alertTagRuleUsecase connectors.AlertTagRuleUsecase
	adversaryHandler    *handler.AdversaryHandler
	adversaryUsecase    connectors.AdversaryUsecase

	events     *eventstore.Store
	floodGuard *usecase.FloodGuard
}

func NewModule(db *gorm.DB, events *eventstore.Store, mailer connectors.AlertMailer) *Module {
	alertRepo := repository.NewCHAlertRepository(events, db)

	alertUC := usecase.NewAlertUsecase(alertRepo, mailer)
	alertH := handler.NewAlertHandler(alertUC)

	alertTagRepo := repository.NewAlertTagRepository(db)
	alertTagUC := usecase.NewAlertTagUsecase(alertTagRepo)
	alertTagH := handler.NewAlertTagHandler(alertTagUC)

	alertTagRuleRepo := repository.NewAlertTagRuleRepository(db)
	alertTagRuleUC := usecase.NewAlertTagRuleUsecase(alertTagRuleRepo, alertTagRepo)
	alertTagRuleH := handler.NewAlertTagRuleHandler(alertTagRuleUC)

	adversaryUC := usecase.NewAdversaryUsecase(repository.NewAdversaryRepository(events))
	adversaryH := handler.NewAdversaryHandler(adversaryUC)

	return &Module{
		alertHandler:        alertH,
		alertUsecase:        alertUC,
		alertTagHandler:     alertTagH,
		alertTagUsecase:     alertTagUC,
		alertTagRuleHandler: alertTagRuleH,
		alertTagRuleUsecase: alertTagRuleUC,
		adversaryHandler:    adversaryH,
		adversaryUsecase:    adversaryUC,
		events:              events,
	}
}

func (m *Module) GetAlertHandler() *handler.AlertHandler { return m.alertHandler }

func (m *Module) GetAlertUsecase() connectors.AlertUsecase { return m.alertUsecase }

// AddRaisedHook registers something to run for each new alert.
func (m *Module) AddRaisedHook(h connectors.RaisedHook) { m.alertUsecase.AddRaisedHook(h) }

func (m *Module) SetCorrelationResolver(r connectors.CorrelationResolver) {
	m.alertUsecase.SetCorrelationResolver(r)
}

func (m *Module) GetAlertTagHandler() *handler.AlertTagHandler { return m.alertTagHandler }

func (m *Module) GetAlertTagUsecase() connectors.AlertTagUsecase { return m.alertTagUsecase }

func (m *Module) GetAlertTagRuleHandler() *handler.AlertTagRuleHandler {
	return m.alertTagRuleHandler
}

func (m *Module) GetAlertTagRuleUsecase() connectors.AlertTagRuleUsecase {
	return m.alertTagRuleUsecase
}

func (m *Module) GetAdversaryHandler() *handler.AdversaryHandler { return m.adversaryHandler }

func (m *Module) GetAdversaryUsecase() connectors.AdversaryUsecase { return m.adversaryUsecase }

// EnableFloodGuard sets up the guard that disables a rule that floods a tenant's alerts. It needs what
// other modules own, the rules and the notification bell, so the caller passes them in.
func (m *Module) EnableFloodGuard(rules connectors.FloodRuleSwitch, notifier connectors.FloodNotifier, config func() usecase.FloodConfig) {
	m.floodGuard = usecase.NewFloodGuard(repository.NewFloodRepository(m.events), rules, notifier, config)
}

// Start begins the module's periodic work.
func (m *Module) Start(ctx context.Context) {
	if m.floodGuard != nil {
		go m.floodGuard.Start(ctx)
	}
}
