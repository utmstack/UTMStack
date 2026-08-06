package iam

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, module *Module, userAuth gin.HandlerFunc, enterprise, enterpriseLicense gin.HandlerFunc) {
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
	userGroup.POST("/:id/tfa/disable", middleware.RequirePermission("users.write"), users.ResetTfa)

	roleGroup := api.Group("/roles", userAuth)
	roleGroup.GET("", middleware.RequirePermission("roles.read"), roles.List)
	roleGroup.GET("/permissions", middleware.RequirePermission("roles.read"), roles.ListPermissions)
	roleGroup.GET("/:id", middleware.RequirePermission("roles.read"), roles.Get)
	roleGroup.POST("", middleware.RequirePermission("roles.write"), roles.Create)
	roleGroup.PUT("/:id", middleware.RequirePermission("roles.write"), roles.Update)
	roleGroup.DELETE("/:id", middleware.RequirePermission("roles.write"), roles.Delete)

	tfaGroup := api.Group("/tfa", userAuth)
	tfaGroup.POST("/enroll", tfa.Enroll)
	tfaGroup.POST("/disable", tfa.Disable)

	keyGroup := api.Group("/api-keys", userAuth)
	keyGroup.POST("", enterprise, apiKeys.Create)
	keyGroup.GET("", apiKeys.List)
	keyGroup.GET("/:id", apiKeys.Get)
	keyGroup.PUT("/:id", apiKeys.Update)
	keyGroup.DELETE("/:id", apiKeys.Delete)
	keyGroup.POST("/:id/generate", enterprise, apiKeys.Generate)
	keyGroup.POST("/authenticate", middleware.RequireInternal(), apiKeys.Authenticate)

	idp := module.GetIDPHandler()
	idpGroup := api.Group("/identity-providers", userAuth)
	idpGroup.POST("", enterprise, middleware.RequirePermission("idp.write"), idp.Create)
	idpGroup.PUT("", enterprise, middleware.RequirePermission("idp.write"), idp.Update)
	idpGroup.GET("", middleware.RequirePermission("idp.read"), idp.List)
	idpGroup.GET("/:id", middleware.RequirePermission("idp.read"), idp.GetByID)
	idpGroup.DELETE("/:id", middleware.RequirePermission("idp.write"), idp.Delete)
	idpGroup.GET("/:id/group-mappings", middleware.RequirePermission("idp.read"), idp.ListMappings)

	api.GET("/idp-providers", idp.PublicList)
	sso := module.GetFederationHandler()
	ssoGroup := api.Group("/sso/:name", enterpriseLicense)
	ssoGroup.GET("/login", sso.Start)
	ssoGroup.POST("/acs", sso.ACS)
	ssoGroup.GET("/callback", sso.Callback)
}
