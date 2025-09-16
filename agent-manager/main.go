package main

import (
	"os"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/updates"
)

func main() {
	err := database.MigrateDatabase()
	if err != nil {
		_ = catcher.Error("failed to migrate database", err, map[string]any{})
		os.Exit(1)
	}

	go updates.InitUpdatesManager()
	agent.InitGrpcServer()
}
