package iam

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, module *Module, userAuth gin.HandlerFunc, enterprise gin.HandlerFunc) {
	auth := module.GetAuthHandler()
	users := module.GetUserHandler()
	roles := module.GetRoleHandler()
	tfa := module.GetTfaHandler()
	apiKeys := module.GetAPIKeyHandler()

	authGroup := api.Group("/auth")
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/refresh", auth.Refresh)
	authGroup.POST("/tfa/verify-code", tfa.VerifyLoginCode)

	authGroup.POST("/logout", auth.Logout)
	authGroup.POST("/reset-password/init", auth.RequestPasswordReset)
	authGroup.POST("/reset-password/finish", auth.FinishPasswordReset)
	authGroup.GET("/me", userAuth, auth.Me)
	authGroup.PUT("/me", userAuth, auth.UpdateMe)
	authGroup.GET("/authenticate", userAuth, auth.Authenticate)
	authGroup.POST("/me/avatar", userAuth, auth.UploadAvatar)
	authGroup.DELETE("/me/avatar", userAuth, auth.RemoveAvatar)

	authGroup.POST("/change-password", userAuth, auth.ChangePassword)
	authGroup.GET("/sessions", userAuth, auth.ListSessions)
	authGroup.DELETE("/sessions", userAuth, auth.RevokeOtherSessions)
	authGroup.DELETE("/sessions/:id", userAuth, auth.RevokeSession)

	userGroup := api.Group("/users", userAuth)
	userGroup.GET("", middleware.RequirePermission("users.read"), users.List)
	userGroup.GET("/:id", middleware.RequirePermission("users.read"), users.Get)
	userGroup.POST("", middleware.RequirePermission("users.write"), users.Create)
	userGroup.PUT("/:id", middleware.RequirePermission("users.write"), users.Update)
	userGroup.DELETE("/:id", middleware.RequirePermission("users.delete"), users.Deactivate)
	userGroup.PUT("/:id/roles", middleware.RequirePermission("users.write"), users.AssignRoles)

	roleGroup := api.Group("/roles", userAuth)
	roleGroup.GET("", middleware.RequirePermission("roles.read"), roles.List)
	roleGroup.GET("/:name", middleware.RequirePermission("roles.read"), roles.Get)

	tfaGroup := api.Group("/tfa", userAuth)
	tfaGroup.POST("/init", tfa.InitEnrollment)
	tfaGroup.POST("/verify", tfa.VerifyEnrollment)
	tfaGroup.POST("/complete", tfa.CompleteEnrollment)
	tfaGroup.GET("/refresh", tfa.RefreshChallenge)

	api.POST("/enrollment/tfa", userAuth, tfa.UnifiedEnrollment)

	keyGroup := api.Group("/api-keys", userAuth)
	keyGroup.POST("", enterprise, apiKeys.Create)
	keyGroup.GET("", apiKeys.List)
	keyGroup.GET("/:id", apiKeys.Get)
	keyGroup.PUT("/:id", apiKeys.Update)
	keyGroup.DELETE("/:id", apiKeys.Delete)
	keyGroup.POST("/:id/generate", enterprise, apiKeys.Generate)
	// Internal: the inputs gateway validates a pusher's API key here.
	keyGroup.POST("/authenticate", middleware.RequireInternal(), apiKeys.Authenticate)

	idp := module.GetIDPHandler()
	idpGroup := api.Group("/identity-providers", userAuth)
	idpGroup.POST("", enterprise, middleware.RequirePermission("idp.write"), idp.Create)
	idpGroup.PUT("", enterprise, middleware.RequirePermission("idp.write"), idp.Update)
	idpGroup.GET("", middleware.RequirePermission("idp.read"), idp.List)
	idpGroup.GET("/:id", middleware.RequirePermission("idp.read"), idp.GetByID)
	idpGroup.DELETE("/:id", middleware.RequirePermission("idp.write"), idp.Delete)

	api.GET("/idp-providers", enterprise, idp.PublicList)
	saml := module.GetSAMLHandler()
	ssoGroup := api.Group("/sso/saml/:name", enterprise)
	ssoGroup.GET("/login", saml.Initiate)
	ssoGroup.POST("/acs", saml.ACS)
}
