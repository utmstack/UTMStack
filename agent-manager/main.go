package main

import (
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/updates"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func main() {
	utils.InitLogger()
	utils.ALogger.Info("Starting Agent Manager...")

	err := database.MigrateDatabase()
	if err != nil {
		utils.ALogger.Fatal("failed to migrate database: %v", err)
	}

	updates.ManageUpdates()
	agent.InitGrpcServer()
}
