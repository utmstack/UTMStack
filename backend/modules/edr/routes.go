package edr

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// RegisterRoutes wires /api/v1/edr. Reads need edr.read; mutations edr.write.
// Both are seeded by migration 000004_edr_permissions.
func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	h := m.Handler()

	g := api.Group("/edr", userAuth)

	g.GET("/overview", middleware.RequirePermission("edr.read"), h.Overview)
	g.GET("/endpoints", middleware.RequirePermission("edr.read"), h.Endpoints)
	g.GET("/endpoints/:agentId", middleware.RequirePermission("edr.read"), h.Endpoint)
	g.POST("/endpoints/:agentId/actions", middleware.RequirePermission("edr.write"), h.Action)

	g.GET("/policies", middleware.RequirePermission("edr.read"), h.ListPolicies)
	g.POST("/policies", middleware.RequirePermission("edr.write"), h.CreatePolicy)
	g.GET("/policies/:id", middleware.RequirePermission("edr.read"), h.GetPolicy)
	g.PUT("/policies/:id", middleware.RequirePermission("edr.write"), h.UpdatePolicy)
	g.DELETE("/policies/:id", middleware.RequirePermission("edr.write"), h.DeletePolicy)
	g.PUT("/policies/:id/assignments", middleware.RequirePermission("edr.write"), h.AssignPolicy)

	g.GET("/quarantine", middleware.RequirePermission("edr.read"), h.Quarantine)
	g.POST("/quarantine/:itemId/restore", middleware.RequirePermission("edr.write"), h.QuarantineRestore)
	g.POST("/quarantine/:itemId/purge", middleware.RequirePermission("edr.write"), h.QuarantinePurge)

	g.GET("/blocklist/status", middleware.RequirePermission("edr.read"), h.BlocklistStatus)
	g.GET("/blocklist/allowlist", middleware.RequirePermission("edr.read"), h.Allowlist)
	g.POST("/blocklist/allowlist", middleware.RequirePermission("edr.write"), h.AllowAdd)
	g.DELETE("/blocklist/allowlist", middleware.RequirePermission("edr.write"), h.AllowRemove)
}
