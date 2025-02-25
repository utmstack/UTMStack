package main

import (
	"github.com/utmstack/UTMStack/correlation/ti"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
)

// @title UTMStack's Correlation Engine
// @version 1.0
// @description Rules-based correlation engine for UTMStack.
// @contact.name UTMStack LLC
// @contact.email contact@utmstack.com
// @license.name AGPLv3
// @host localhost:8080
// @BasePath /v1

func main() {
	sqldb.Connect()
	geo.Load()
	ti.Load()

	ready := make(chan bool, 1)
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
	go ti.IsBlocklisted()

	go func() {
		gin.SetMode(gin.ReleaseMode)

		r := gin.New()
		r.Use(gin.Recovery())
		r.Use(gin.ErrorLogger())

		v1 := r.Group("/v1")
		v1.POST("/newlog", api.NewLog)

		docURL := ginSwagger.URL("/swagger/doc.json")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docURL))

		err := r.Run()
		if err != nil {
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	go rules.Changes(signals)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
