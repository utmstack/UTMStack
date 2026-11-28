package servefile

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, module *Module, userAuth gin.HandlerFunc) {
	file := module.GetFileHandler()
	if file == nil {
		return
	}
	fileGroup := api.Group("/file", userAuth)
	fileGroup.POST("", file.Upload)
	fileGroup.GET("/:filename", file.Download)
}
