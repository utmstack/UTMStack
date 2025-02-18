package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/utmstack/UTMStack/correlation/api"
	"github.com/utmstack/UTMStack/correlation/cache"
	"github.com/utmstack/UTMStack/correlation/correlation"
	_ "github.com/utmstack/UTMStack/correlation/docs"
	"github.com/utmstack/UTMStack/correlation/geo"
	"github.com/utmstack/UTMStack/correlation/rules"
	"github.com/utmstack/UTMStack/correlation/search"
	"github.com/utmstack/UTMStack/correlation/sqldb"
	"github.com/utmstack/UTMStack/correlation/statistics"
	"github.com/utmstack/UTMStack/correlation/utils"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title UTMStack's Correlation Engine
// @version 1.0
// @description Rules based correlation engine for UTMStack.
// @contact.name Osmany Montero
// @contact.email osmany@quantfall.com
// @license.name Private
// @host localhost:8080
// @BasePath /v1

func main() {
	sqldb.Connect()

	ready := make(chan bool, 1)

	go geo.Update(ready)
	<-ready

	go rules.Update(ready)
	<-ready

	rulesL := rules.GetRules()
	for _, rule := range rulesL {
		go correlation.Finder(rule)
	}

	go cache.Status()
	go utils.Status()
	go cache.Clean()
	go cache.ProcessQueue()
	go search.ProcessQueue()
	go statistics.Update()

	go func() {
		gin.SetMode(gin.ReleaseMode)

		//r := gin.Default()
		r := gin.New()
		r.Use(gin.Recovery())
		r.Use(gin.ErrorLogger())

		v1 := r.Group("/v1")
		v1.POST("/newlog", api.NewLog)

		docURL := ginSwagger.URL("/swagger/doc.json")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docURL))

		r.Run()
	}()

	signals := make(chan os.Signal, 1)
	go rules.RulesChanges(signals)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
