package database

import "github.com/utmstack/UTMStack/agent-manager/models"

func MigrateDatabase() error {
	db := GetDB()
	err := db.Migrate(&models.Agent{}, &models.AgentCommand{}, &models.LastSeen{}, &models.Collector{})
	if err != nil {
		return err
	}
	return nil
}
