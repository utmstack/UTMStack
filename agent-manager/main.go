package main

import (
	"os"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/updates"
)

func main() {
	catcher.Info("Starting Agent Manager v1.0.0 ...", nil)

	err := database.MigrateDatabase()
	if err != nil {
		catcher.Error("failed to migrate database", err, nil)
		os.Exit(1)
	}

	go updates.InitUpdatesManager()
	agent.InitGrpcServer()
}
