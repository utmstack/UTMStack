package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
	"github.com/utmstack/UTMStack/plugins/modules-config/validations"
	"google.golang.org/grpc"
)

func startGRPCServer() error {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(GrpcUniMiddleware),
		grpc.ChainStreamInterceptor(GrpcStreamMiddleware),
	)

	listener, err := net.Listen("tcp", "0.0.0.0:9003")
	if err != nil {
		return catcher.Error("failed to listen on port 9003", err, nil)
	}

	config.RegisterConfigServiceServer(server, config.GetConfigServer())
	config.GetConfigServer().SyncConfigs(BackendService, InternalKey)

	if err := server.Serve(listener); err != nil {
		return catcher.Error("failed to serve grpc", err, nil)
	}

	return nil
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	modules := router.Group("/api/v1/modules-config")
	modules.POST("", HttpMiddleware(), UpdateModuleConfig)
	modules.POST("/validate", HttpMiddleware(), ValidateModuleConfig)

	router.GET("/api/v1/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	server := &http.Server{
		Addr:    ":9002",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		_ = catcher.Error("could not start http server", err, nil)
	}

}

func UpdateModuleConfig(c *gin.Context) {
	moduleName := c.Query("nameShort")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nameShort query parameter is required"})
		return
	}

	body := []config.ConfigurationSection{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(body) != 0 {
		config.GetConfigServer().NotifyUpdate(moduleName, &body[0])
	} else {
		fmt.Println("Received empty configuration body, no updates made")
	}

	c.JSON(http.StatusOK, gin.H{"status": "Module configuration updated successfully"})
}

func ValidateModuleConfig(c *gin.Context) {
	moduleName := c.Query("nameShort")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nameShort query parameter is required"})
		return
	}

	body := config.ModuleGroup{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := validations.ValidateModuleConfig(moduleName, &body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Module configuration is valid"})
}
