package main

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
	"github.com/utmstack/UTMStack/plugins/modules-config/crypto"
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
		return catcher.Error("failed to listen on port 9003", err, map[string]any{"process": "plugin_com.utmstack.modules-config"})
	}
	// Remove after testing, before release to production
	catcher.Info("startGRPCServer: initializing decrypter", map[string]any{
		"process":        "plugin_com.utmstack.modules-config",
		"InternalKey":    InternalKey,
		"InternalKeyLen": len(InternalKey),
		"BackendService": BackendService,
	})

	config.GetConfigServer().SetDecrypter(func(section *config.ConfigurationSection) error {
		return crypto.DecryptConfigurationSection(section, InternalKey)
	})

	config.RegisterConfigServiceServer(server, config.GetConfigServer())
	config.GetConfigServer().SyncConfigs(BackendService, InternalKey)

	go config.GetConfigServer().StartPeriodicRetry(BackendService, InternalKey)

	if err := server.Serve(listener); err != nil {
		return catcher.Error("failed to serve grpc", err, map[string]any{"process": "plugin_com.utmstack.modules-config"})
	}

	return nil
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
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
		_ = catcher.Error("could not start http server", err, map[string]any{"process": "plugin_com.utmstack.modules-config"})
	}

}

func UpdateModuleConfig(c *gin.Context) {
	moduleName := c.Query("nameShort")
	if moduleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nameShort query parameter is required"})
		return
	}

	rawBody, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		_ = catcher.Error("failed to read request body", readErr, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	// Remove after testing, before release to production
	catcher.Info("UpdateModuleConfig: raw request body", map[string]any{
		"process": "plugin_com.utmstack.modules-config",
		"module":  moduleName,
		"bodyLen": len(rawBody),
		"rawBody": string(rawBody),
	})

	body := []config.ConfigurationSection{}
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = catcher.Error("failed to bind JSON in UpdateModuleConfig", err, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	catcher.Info("UpdateModuleConfig: parsed body", map[string]any{
		"process":      "plugin_com.utmstack.modules-config",
		"module":       moduleName,
		"sectionCount": len(body),
	})
	// Remove after testing, before release to production
	if len(body) != 0 {
		for _, g := range body[0].ModuleGroups {
			for _, cnf := range g.ModuleGroupConfigurations {
				catcher.Info("incoming Update field (pre-decrypt)", map[string]any{
					"process":      "plugin_com.utmstack.modules-config",
					"module":       moduleName,
					"groupId":      g.Id,
					"confKey":      cnf.ConfKey,
					"confDataType": cnf.ConfDataType,
					"valueLen":     len(cnf.ConfValue),
					"confValue":    cnf.ConfValue,
				})
			}
		}
		// Remove after testing, before release to production
		catcher.Info("UpdateModuleConfig: calling decrypter", map[string]any{
			"process":        "plugin_com.utmstack.modules-config",
			"module":         moduleName,
			"InternalKey":    InternalKey,
			"InternalKeyLen": len(InternalKey),
		})

		if err := crypto.DecryptConfigurationSection(&body[0], InternalKey); err != nil {
			_ = catcher.Error("failed to decrypt module config on update", err, map[string]any{
				"process": "plugin_com.utmstack.modules-config",
				"module":  moduleName,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt configuration"})
			return
		}
		// Remove after testing, before release to production
		for _, g := range body[0].ModuleGroups {
			for _, cnf := range g.ModuleGroupConfigurations {
				catcher.Info("Update field (post-decrypt)", map[string]any{
					"process":      "plugin_com.utmstack.modules-config",
					"module":       moduleName,
					"groupId":      g.Id,
					"confKey":      cnf.ConfKey,
					"confDataType": cnf.ConfDataType,
					"valueLen":     len(cnf.ConfValue),
					"confValue":    cnf.ConfValue,
				})
			}
		}

		config.GetConfigServer().NotifyUpdate(moduleName, &body[0])
	} else {
		catcher.Info("Received empty configuration body, no updates made", map[string]any{"process": "plugin_com.utmstack.modules-config"})
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
	// Remove after testing, before release to production
	if err := crypto.DecryptModuleGroup(moduleName, &body, InternalKey); err != nil {
		_ = catcher.Error("failed to decrypt module config on validate", err, map[string]any{
			"process": "plugin_com.utmstack.modules-config",
			"module":  moduleName,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt configuration"})
		return
	}

	err := validations.ValidateModuleConfig(moduleName, &body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Module configuration is valid"})
}
