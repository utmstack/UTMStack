package database

import "github.com/utmstack/UTMStack/agent-manager/models"

func MigrateDatabase() error {
	db := GetDB()
	err := db.Migrate(&models.Agent{}, &models.AgentCommand{}, &models.LastSeen{}, &models.Collector{}, &models.ConnectionKey{}, &models.Recovery{}, &models.RecoveryTarget{}, &models.CollectorIntegrationConfig{})
	if err != nil {
		return err
	}
	return nil
}
