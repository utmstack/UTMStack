package appconfig

import (
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/handler"
	"github.com/utmstack/utmstack/backend/modules/appconfig/repository"
	"github.com/utmstack/utmstack/backend/modules/appconfig/usecase"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"gorm.io/gorm"
)

type mailerSetter interface {
	SetMailer(mail_connectors.MailService)
}

type Module struct {
	usecase         connectors.Usecase
	store           connectors.Store
	handler         *handler.Handler
	branding        connectors.BrandingUsecase
	brandingHandler *handler.BrandingHandler
}

func NewModule(db *gorm.DB, cipher *secret.Cipher, uploadDir string) *Module {
	repo := repository.NewRepository(db)
	brandingSvc := usecase.NewBranding(repo)
	svc := usecase.New(repo, cipher, brandingSvc)
	return &Module{
		usecase:         svc,
		store:           svc,
		handler:         handler.NewHandler(svc),
		branding:        brandingSvc,
		brandingHandler: handler.NewBrandingHandler(brandingSvc, uploadDir),
	}
}

type entitlementSetter interface {
	SetEntitlement(func() bool)
}

func (m *Module) SetWhiteLabelEntitlement(fn func() bool) {
	if s, ok := m.branding.(entitlementSetter); ok {
		s.SetEntitlement(fn)
	}
}

func (m *Module) Handler() *handler.Handler                 { return m.handler }
func (m *Module) BrandingHandler() *handler.BrandingHandler { return m.brandingHandler }
func (m *Module) Branding() connectors.BrandingUsecase      { return m.branding }
func (m *Module) Store() connectors.Store                   { return m.store }
func (m *Module) Usecase() connectors.Usecase               { return m.usecase }

func (m *Module) SetMailer(mailer mail_connectors.MailService) {
	if s, ok := m.usecase.(mailerSetter); ok {
		s.SetMailer(mailer)
	}
}
