package migration

import (
	"log"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
)

func Migrate() {
	db := config.GetDB()

	err := db.AutoMigrate(
		&models.AgentType{},
		&models.AgentGroup{},
		&models.AgentCommand{},
		&models.AgentConfiguration{},
		&models.AgentModule{},
		&models.AgentModuleConfiguration{},
		&models.AgentMalwareDetection{},
		&models.AgentMalwareHistory{},
		&models.AgentMalwareExclusion{},
		&models.AgentLastSeen{},
		&models.Agent{},
	)
	// Insert default data
	if err != nil {
		log.Fatal("unable to migrate", err)
	}

	err = seedAgentTypeTable(db)
	if err != nil {
		log.Fatal("unable to seed tables", err)
	}

}

func seedAgentTypeTable(db *gorm.DB) error {
	var count int64
	db.Model(&models.AgentType{}).Count(&count)
	if count > 0 {
		return nil
	}
	agentTypes := []models.AgentType{
		{TypeName: "DATABASE"},
		{TypeName: "FIREWALL"},
		{TypeName: "CLOUD"},
		{TypeName: "LAPTOP"},
		{TypeName: "SERVER"},
		{TypeName: "WORKSTATION"},
		{TypeName: "PC"},
		{TypeName: "BRIDGE"},
		{TypeName: "BROADBAND ROUTER"},
		{TypeName: "GENERAL PURPOSE"},
		{TypeName: "HUB"},
		{TypeName: "LOAD BALANCER"},
		{TypeName: "MEDIA DEVICE"},
		{TypeName: "PBX"},
		{TypeName: "PDA"},
		{TypeName: "POWER-DEVICE"},
		{TypeName: "PRINT SERVER"},
		{TypeName: "PROXY SERVER"},
		{TypeName: "REMOTE MANAGEMENT"},
		{TypeName: "ROUTER"},
		{TypeName: "SECURITY-MISC"},
		{TypeName: "SPECIALIZED"},
		{TypeName: "STORAGE-MISC"},
		{TypeName: "SWITCH"},
		{TypeName: "TELECOM-MISC"},
		{TypeName: "TERMINAL"},
		{TypeName: "TERMINAL SERVER"},
		{TypeName: "VOIP ADAPTER"},
		{TypeName: "VOIP PHONE"},
		{TypeName: "WAP"},
		{TypeName: "WEB SERVER"},
	}

	return db.Create(&agentTypes).Error

}
