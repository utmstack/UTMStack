package main

import (
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/updates"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func main() {
	utils.InitLogger()
	utils.ALogger.Info("Starting Agent Manager v1.0.0 ...")

	err := database.MigrateDatabase()
	if err != nil {
		utils.ALogger.Fatal("failed to migrate database: %v", err)
	}

	go updates.InitUpdatesManager()
	agent.InitGrpcServer()
}
