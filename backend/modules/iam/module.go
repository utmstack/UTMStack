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

	authUsecase   connectors.AuthUsecase
	userUsecase   connectors.UserUsecase
	roleUsecase   connectors.RoleUsecase
	tfaUsecase    connectors.TfaUsecase
	apiKeyUsecase connectors.APIKeyUsecase
}

func NewModule(
	authUsecase connectors.AuthUsecase,
	userUsecase connectors.UserUsecase,
	roleUsecase connectors.RoleUsecase,
	tfaUsecase connectors.TfaUsecase,
	apiKeyUsecase connectors.APIKeyUsecase,
	uploadDir string,
) *Module {
	return &Module{
		authHandler:   handler.NewAuthHandler(authUsecase, uploadDir),
		userHandler:   handler.NewUserHandler(userUsecase),
		roleHandler:   handler.NewRoleHandler(roleUsecase),
		tfaHandler:    handler.NewTfaHandler(tfaUsecase),
		apiKeyHandler: handler.NewAPIKeyHandler(apiKeyUsecase),
		authUsecase:   authUsecase,
		userUsecase:   userUsecase,
		roleUsecase:   roleUsecase,
		tfaUsecase:    tfaUsecase,
		apiKeyUsecase: apiKeyUsecase,
	}
}

func (m *Module) GetAuthHandler() *handler.AuthHandler       { return m.authHandler }
func (m *Module) GetUserHandler() *handler.UserHandler       { return m.userHandler }
func (m *Module) GetRoleHandler() *handler.RoleHandler       { return m.roleHandler }
func (m *Module) GetTfaHandler() *handler.TfaHandler         { return m.tfaHandler }
func (m *Module) GetAPIKeyHandler() *handler.APIKeyHandler   { return m.apiKeyHandler }
func (m *Module) GetAuthUsecase() connectors.AuthUsecase     { return m.authUsecase }
func (m *Module) GetTfaUsecase() connectors.TfaUsecase       { return m.tfaUsecase }
func (m *Module) GetAPIKeyUsecase() connectors.APIKeyUsecase { return m.apiKeyUsecase }
