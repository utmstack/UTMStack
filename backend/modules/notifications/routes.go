package notifications

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, m *Module, userAuth gin.HandlerFunc) {
	nh := m.GetNotificationHandler()

	g := api.Group("/utm-notifications", userAuth)
	g.POST("", nh.Create)
	g.GET("", nh.List)
	g.GET("/unread-count", nh.UnreadCount)
	g.PUT("/read-all", nh.MarkAllRead)
	g.GET("/:id", nh.GetByID)
	g.PUT("/:id/read", nh.UpdateRead)
	g.PUT("/:id/status", nh.UpdateStatus)
	g.DELETE("/:id", nh.Delete)
}
