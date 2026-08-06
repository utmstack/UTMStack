package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	nh := m.GetNotificationHandler()

	read := middleware.RequirePermission("notifications.read")
	write := middleware.RequirePermission("notifications.write")

	g := api.Group("/notifications", userAuth)

	g.POST("", middleware.RequireInternal(), nh.Create)

	g.GET("", read, nh.List)
	g.GET("/grouped", read, nh.ListGrouped)
	g.GET("/unread-count", read, nh.UnreadCount)
	g.GET("/:id", read, nh.GetByID)

	g.PUT("/read-all", write, nh.MarkAllRead)
	g.PUT("/:id/read", write, nh.UpdateRead)
	g.PUT("/:id/status", write, nh.UpdateStatus)
	g.DELETE("/:id", write, nh.Delete)
}
