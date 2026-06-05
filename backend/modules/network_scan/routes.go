package network_scan

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	ns := m.GetNetworkScanHandler()
	gr := m.GetAssetGroupHandler()
	tp := m.GetAssetTypesHandler()
	pt := m.GetPortsHandler()
	pr := m.GetProbeHandler()

	nsGroup := api.Group("/network-scans", userAuth)
	nsGroup.POST("/custom", middleware.RequirePermission("assets.write"), ns.SaveCustom)
	nsGroup.PUT("", middleware.RequirePermission("assets.write"), ns.Update)
	nsGroup.PUT("/type", middleware.RequirePermission("assets.write"), ns.UpdateType)
	nsGroup.PUT("/group", middleware.RequirePermission("assets.write"), ns.UpdateGroup)
	nsGroup.GET("", middleware.RequirePermission("assets.read"), ns.ListByCriteria)
	nsGroup.GET("/count", middleware.RequirePermission("assets.read"), ns.CountByCriteria)
	nsGroup.GET("/search", middleware.RequirePermission("assets.read"), ns.SearchByFilters)
	nsGroup.GET("/property-values", middleware.RequirePermission("assets.read"), ns.SearchPropertyValues)
	nsGroup.GET("/count-new", middleware.RequirePermission("assets.read"), ns.CountNewAssets)
	nsGroup.GET("/can-run-command", middleware.RequirePermission("assets.read"), ns.CanRunCommand)
	nsGroup.GET("/agent-os-platform", middleware.RequirePermission("assets.read"), ns.AgentOSPlatform)
	nsGroup.GET("/report", middleware.RequirePermission("assets.read"), ns.Report)
	nsGroup.DELETE("/custom/:id", middleware.RequirePermission("assets.write"), ns.DeleteCustom)
	nsGroup.GET("/:id", middleware.RequirePermission("assets.read"), ns.Get)

	// Probe sub-routes — separate so the gin tree splits cleanly between /:id and /probe.
	probeGroup := api.Group("/network-scans/probe", userAuth)
	probeGroup.POST("/scan", middleware.RequirePermission("assets.write"), pr.Scan)
	probeGroup.GET("/ping", middleware.RequirePermission("assets.read"), pr.Ping)
	probeGroup.GET("/check-interface", middleware.RequirePermission("assets.read"), pr.CheckInterface)

	portsGroup := api.Group("/open-ports", userAuth)
	portsGroup.POST("", middleware.RequirePermission("assets.write"), pt.Create)
	portsGroup.PUT("", middleware.RequirePermission("assets.write"), pt.Update)
	portsGroup.GET("", middleware.RequirePermission("assets.read"), pt.List)
	portsGroup.GET("/count", middleware.RequirePermission("assets.read"), pt.Count)
	portsGroup.GET("/:id", middleware.RequirePermission("assets.read"), pt.Get)
	portsGroup.DELETE("/:id", middleware.RequirePermission("assets.write"), pt.Delete)

	groupGroup := api.Group("/asset-groups", userAuth)
	groupGroup.POST("", middleware.RequirePermission("assets.write"), gr.Create)
	groupGroup.PUT("", middleware.RequirePermission("assets.write"), gr.Update)
	groupGroup.GET("/search", middleware.RequirePermission("assets.read"), gr.Search)
	groupGroup.GET("/:id", middleware.RequirePermission("assets.read"), gr.Get)
	groupGroup.DELETE("/:id", middleware.RequirePermission("assets.write"), gr.Delete)

	typesGroup := api.Group("/asset-types", userAuth)
	typesGroup.GET("/all", middleware.RequirePermission("assets.read"), tp.List)
}
