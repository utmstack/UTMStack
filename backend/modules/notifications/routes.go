package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	nh := m.GetNotificationHandler()

	g := api.Group("/notifications", userAuth)

	g.POST("", middleware.RequireInternal(), nh.Create)
	g.GET("", nh.List)
	g.GET("/unread-count", nh.UnreadCount)
	g.PUT("/read-all", nh.MarkAllRead)
	g.GET("/:id", nh.GetByID)
	g.PUT("/:id/read", nh.UpdateRead)
	g.PUT("/:id/status", nh.UpdateStatus)
	g.DELETE("/:id", nh.Delete)
}
