package billing

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/billing/usecase"
)

var (
	PublicKey   string
	EncryptSalt string
)

type Module struct {
	version *usecase.VersionUsecase
	license *usecase.LicenseService
}

func NewModule(updatesDir string) *Module {
	return &Module{
		version: usecase.NewVersionUsecase(updatesDir),
		license: usecase.NewLicenseService(updatesDir, PublicKey, EncryptSalt),
	}
}

func (m *Module) Start(ctx context.Context) {
	m.license.Start(ctx)
}

func (m *Module) Version() *usecase.VersionUsecase { return m.version }

func (m *Module) License() *usecase.LicenseService { return m.license }
