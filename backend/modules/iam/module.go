package iam

import (
	"context"
	"github.com/utmstack/utmstack/backend/modules/iam/usecase"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/handler"
)

type Module struct {
	sessionPurger     *usecase.SessionPurger
	authHandler       *handler.AuthHandler
	userHandler       *handler.UserHandler
	roleHandler       *handler.RoleHandler
	tfaHandler        *handler.TfaHandler
	apiKeyHandler     *handler.APIKeyHandler
	idpHandler        *handler.IdentityProviderHandler
	bulkIDPHandler    *handler.BulkIDPHandler
	federationHandler *handler.FederationHandler

	authUsecase   connectors.AuthUsecase
	userUsecase   connectors.UserUsecase
	roleUsecase   connectors.RoleUsecase
	tfaUsecase    connectors.TfaUsecase
	apiKeyUsecase connectors.APIKeyUsecase
	idpUsecase    connectors.IdentityProviderUsecase
	federationUC  connectors.FederationUsecase
}

func NewModule(
	authUsecase connectors.AuthUsecase,
	userUsecase connectors.UserUsecase,
	roleUsecase connectors.RoleUsecase,
	tfaUsecase connectors.TfaUsecase,
	apiKeyUsecase connectors.APIKeyUsecase,
	idpUsecase connectors.IdentityProviderUsecase,
	federationUC connectors.FederationUsecase,
	uploadDir string,
	tenantLister func(context.Context) ([]string, error),
) *Module {
	return &Module{
		authHandler:       handler.NewAuthHandler(authUsecase, uploadDir),
		userHandler:       handler.NewUserHandler(userUsecase),
		roleHandler:       handler.NewRoleHandler(roleUsecase),
		tfaHandler:        handler.NewTfaHandler(tfaUsecase),
		apiKeyHandler:     handler.NewAPIKeyHandler(apiKeyUsecase),
		idpHandler:        handler.NewIdentityProviderHandler(idpUsecase),
		bulkIDPHandler:    handler.NewBulkIDPHandler(idpUsecase, tenantLister),
		federationHandler: handler.NewFederationHandler(federationUC),
		authUsecase:       authUsecase,
		userUsecase:       userUsecase,
		roleUsecase:       roleUsecase,
		tfaUsecase:        tfaUsecase,
		apiKeyUsecase:     apiKeyUsecase,
		idpUsecase:        idpUsecase,
		federationUC:      federationUC,
	}
}

func (m *Module) GetAuthHandler() *handler.AuthHandler             { return m.authHandler }
func (m *Module) GetUserHandler() *handler.UserHandler             { return m.userHandler }
func (m *Module) GetRoleHandler() *handler.RoleHandler             { return m.roleHandler }
func (m *Module) GetTfaHandler() *handler.TfaHandler               { return m.tfaHandler }
func (m *Module) GetAPIKeyHandler() *handler.APIKeyHandler         { return m.apiKeyHandler }
func (m *Module) GetIDPHandler() *handler.IdentityProviderHandler     { return m.idpHandler }
func (m *Module) GetBulkIDPHandler() *handler.BulkIDPHandler          { return m.bulkIDPHandler }
func (m *Module) GetFederationHandler() *handler.FederationHandler { return m.federationHandler }
func (m *Module) GetAuthUsecase() connectors.AuthUsecase           { return m.authUsecase }
func (m *Module) GetTfaUsecase() connectors.TfaUsecase             { return m.tfaUsecase }
func (m *Module) GetAPIKeyUsecase() connectors.APIKeyUsecase       { return m.apiKeyUsecase }
func (m *Module) GetUserUsecase() connectors.UserUsecase           { return m.userUsecase }
func (m *Module) GetRoleUsecase() connectors.RoleUsecase           { return m.roleUsecase }
func (m *Module) GetIdentityProviderUsecase() connectors.IdentityProviderUsecase {
	return m.idpUsecase
}
func (m *Module) GetFederationUsecase() connectors.FederationUsecase { return m.federationUC }

func (m *Module) SetSessionPurger(p *usecase.SessionPurger) { m.sessionPurger = p }

func (m *Module) Start(ctx context.Context) {
	if m.sessionPurger != nil {
		go m.sessionPurger.Start(ctx)
	}
}
