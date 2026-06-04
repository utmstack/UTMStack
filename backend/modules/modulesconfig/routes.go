package modulesconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// RegisterRoutes wires the modules-config endpoints under the /api/v1 group.
// Permissions: every read is gated by "modules.read"; every write (activation,
// group CRUD, config update) is gated by "modules.write". The internal-key
// gated decrypted endpoint stays behind userAuth too — the handler verifies
// the "internal" context flag the auth middleware sets.
func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	modules := api.Group("/utm-modules", userAuth)
	modules.GET("", middleware.RequirePermission("modules.read"), m.moduleHandler.List)
	modules.GET("/moduleDetails", middleware.RequirePermission("modules.read"), m.moduleHandler.ModuleDetails)
	modules.GET("/module-details-decrypted", m.moduleHandler.ModuleDetailsDecrypted)
	modules.GET("/checkRequirements", middleware.RequirePermission("modules.read"), m.moduleHandler.CheckRequirements)
	modules.GET("/moduleCategories", middleware.RequirePermission("modules.read"), m.moduleHandler.Categories)
	modules.GET("/is-active", middleware.RequirePermission("modules.read"), m.moduleHandler.IsActive)
	modules.GET("/:id", middleware.RequirePermission("modules.read"), m.moduleHandler.Get)
	modules.PUT("/activateDeactivate", middleware.RequirePermission("modules.write"), m.moduleHandler.ActivateDeactivate)

	groups := api.Group("/utm-configuration-groups", userAuth)
	groups.GET("/module-groups", middleware.RequirePermission("modules.read"), m.groupHandler.ListByModuleID)
	groups.GET("/:id", middleware.RequirePermission("modules.read"), m.groupHandler.Get)
	groups.POST("", middleware.RequirePermission("modules.write"), m.groupHandler.Create)
	groups.PUT("", middleware.RequirePermission("modules.write"), m.groupHandler.Update)
	groups.DELETE("/:id", middleware.RequirePermission("modules.write"), m.groupHandler.Delete)

	configs := api.Group("/module-group-configurations", userAuth)
	configs.GET("/by-group-id", middleware.RequirePermission("modules.read"), m.configHandler.ListByGroupID)
	configs.GET("/by-group-and-key", middleware.RequirePermission("modules.read"), m.configHandler.GetByGroupAndKey)
	configs.PUT("/update", middleware.RequirePermission("modules.write"), m.configHandler.Update)
}
