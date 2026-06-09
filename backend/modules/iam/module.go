package iam

import (
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/handler"
)

type Module struct {
	authHandler   *handler.AuthHandler
	userHandler   *handler.UserHandler
	roleHandler   *handler.RoleHandler
	tfaHandler    *handler.TfaHandler
	apiKeyHandler *handler.APIKeyHandler
	idpHandler    *handler.IdentityProviderHandler
	samlHandler   *handler.SAMLHandler

	authUsecase   connectors.AuthUsecase
	userUsecase   connectors.UserUsecase
	roleUsecase   connectors.RoleUsecase
	tfaUsecase    connectors.TfaUsecase
	apiKeyUsecase connectors.APIKeyUsecase
	idpUsecase    connectors.IdentityProviderUsecase
	samlUsecase   connectors.SAMLUsecase
}

func NewModule(
	authUsecase connectors.AuthUsecase,
	userUsecase connectors.UserUsecase,
	roleUsecase connectors.RoleUsecase,
	tfaUsecase connectors.TfaUsecase,
	apiKeyUsecase connectors.APIKeyUsecase,
	idpUsecase connectors.IdentityProviderUsecase,
	samlUsecase connectors.SAMLUsecase,
	uploadDir string,
) *Module {
	return &Module{
		authHandler:   handler.NewAuthHandler(authUsecase, uploadDir),
		userHandler:   handler.NewUserHandler(userUsecase),
		roleHandler:   handler.NewRoleHandler(roleUsecase),
		tfaHandler:    handler.NewTfaHandler(tfaUsecase),
		apiKeyHandler: handler.NewAPIKeyHandler(apiKeyUsecase),
		idpHandler:    handler.NewIdentityProviderHandler(idpUsecase),
		samlHandler:   handler.NewSAMLHandler(samlUsecase),
		authUsecase:   authUsecase,
		userUsecase:   userUsecase,
		roleUsecase:   roleUsecase,
		tfaUsecase:    tfaUsecase,
		apiKeyUsecase: apiKeyUsecase,
		idpUsecase:    idpUsecase,
		samlUsecase:   samlUsecase,
	}
}

func (m *Module) GetAuthHandler() *handler.AuthHandler            { return m.authHandler }
func (m *Module) GetUserHandler() *handler.UserHandler            { return m.userHandler }
func (m *Module) GetRoleHandler() *handler.RoleHandler            { return m.roleHandler }
func (m *Module) GetTfaHandler() *handler.TfaHandler              { return m.tfaHandler }
func (m *Module) GetAPIKeyHandler() *handler.APIKeyHandler        { return m.apiKeyHandler }
func (m *Module) GetIDPHandler() *handler.IdentityProviderHandler { return m.idpHandler }
func (m *Module) GetSAMLHandler() *handler.SAMLHandler            { return m.samlHandler }
func (m *Module) GetAuthUsecase() connectors.AuthUsecase          { return m.authUsecase }
func (m *Module) GetTfaUsecase() connectors.TfaUsecase            { return m.tfaUsecase }
func (m *Module) GetAPIKeyUsecase() connectors.APIKeyUsecase      { return m.apiKeyUsecase }
func (m *Module) GetUserUsecase() connectors.UserUsecase          { return m.userUsecase }
func (m *Module) GetRoleUsecase() connectors.RoleUsecase          { return m.roleUsecase }
func (m *Module) GetIdentityProviderUsecase() connectors.IdentityProviderUsecase {
	return m.idpUsecase
}
func (m *Module) GetSAMLUsecase() connectors.SAMLUsecase { return m.samlUsecase }
