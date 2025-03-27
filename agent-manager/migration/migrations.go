package migration

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"gorm.io/gorm"
)

type Changeset struct {
	ID         uint      `gorm:"primarykey"`
	Name       string    // Name of the migration change_date(dd/MM/yyyy)_number_of_the_change
	ExecutedBy string    // GitHub user
	RanAt      time.Time // When the migration was applied
}

func MigrateDatabase() {
	db := config.GetDB()
	err := db.AutoMigrate(&Changeset{})
	if err != nil {
		util.Logger.ErrorF("failed to auto-migrate MigrationRecord table: %v", err)
		return
	}
	performMigration(db, "performInitialMigrations_15022024_001", "jdieguez89", performInitialMigrations)
	performMigration(db, "renameLastSeenTableAndColumn_15022024_002", "jdieguez89", renameLastSeenTableAndColumnOrCreateTable)
	performMigration(db, "seedAgentTypeTable_15022024_003", "jdieguez89", seedAgentTypeTable)
	performMigration(db, "addLogCollectorTables_15022024_004", "jdieguez89", addLogCollectorTables)
	performMigration(db, "addDeletedByFieldToCollector", "jdieguez89", addDeletedByFieldToCollector)
	performMigration(db, "deleteColumnModuleIdFromCollectorGroupConfig", "jdieguez89", deleteColumnModuleIdFromCollectorGroupConfig)
	performMigration(db, "setConfigurationsPrimaryKey", "jdieguez89", setConfigurationsPrimaryKey)
}

// performMigration executes a given migration function if it has not been recorded yet.
// It records the migration upon successful completion to ensure it's not executed again.
// Parameters:
// - db *gorm.DB: The database connection object.
// - migrationName string: The unique name of the migration.
// - executedBy string: Identifier for who or what executed the migration.
// - migrationFunc func(*gorm.DB) error: The function to execute for the migration.
func performMigration(db *gorm.DB, migrationName string, executedBy string, migrationFunc func(*gorm.DB) error) {
	var migrationRecord Changeset
	result := db.Where("name = ?", migrationName).First(&migrationRecord)
	// If migration hasn't been recorded, perform it
	if result.RowsAffected == 0 {
		err := migrationFunc(db)
		if err != nil {
			util.Logger.ErrorF("Migration failed (%s): %v\n", migrationName, err)
			return
		}

		// Record successful migration
		db.Create(&Changeset{Name: migrationName, RanAt: time.Now(), ExecutedBy: executedBy})
		util.Logger.Info("Migration executed and recorded: %s\n", migrationName)
	}
}

// executeSQLCommands execute sql statements
func executeSQLCommands(db *gorm.DB, sqlCommands []string) error {
	for _, sql := range sqlCommands {
		if err := db.Exec(sql).Error; err != nil {
			util.Logger.ErrorF("Failed to execute SQL command: %v\n", err)
			return err
		}
	}
	return nil
}

// DeleteColumnFromTable deletes a column from a specified table in the database.
// tableName is the name of the table from which the column will be deleted.
// columnName is the name of the column to be deleted.
func deleteColumnFromTable(db *gorm.DB, table interface{}, columnName string) error {
	// Check if the column exists before trying to delete it
	if db.Migrator().HasColumn(table, columnName) {
		err := db.Migrator().DropColumn(table, columnName)
		if err != nil {
			util.Logger.ErrorF("Failed to delete column '%s' from table '%s': %v", columnName, table, err)
			return err
		}
	} else {
		util.Logger.ErrorF("Column '%s' does not exist in table '%s'.", columnName, table)
		return nil
	}
	return nil
}

func performInitialMigrations(db *gorm.DB) error {
	tables := []interface{}{
		&models.AgentType{},
		&models.AgentGroup{},
		&models.AgentCommand{},
		&models.AgentConfiguration{},
		&models.AgentModule{},
		&models.AgentModuleConfiguration{},
		&models.AgentMalwareDetection{},
		&models.AgentMalwareHistory{},
		&models.AgentMalwareExclusion{},
		&models.Agent{},
	}
	for _, table := range tables {
		modelName := fmt.Sprintf("%T", table) // Generates a unique name based on the model type
		if !db.Migrator().HasTable(table) {
			performMigration(db, modelName, "System", func(d *gorm.DB) error {
				return d.AutoMigrate(table)
			})
		}
	}
	return nil
}

// renameLastSeenTableAndColumn renaming `agent_last_seens` table to `last_seen`
// and the `agent_key` column to `key`. This table will be used by agents and collectors
func renameLastSeenTableAndColumnOrCreateTable(db *gorm.DB) error {
	// Rename the table from `agent_last_seens` to `last_seen`
	newName := "last_seens"
	oldName := "agent_last_seens"
	if db.Migrator().HasTable(oldName) {
		if err := db.Migrator().RenameTable("agent_last_seens", newName); err != nil {
			util.Logger.ErrorF("Failed to rename table: %v\n", err)
			return err
		}
		if err := db.Migrator().RenameColumn(&models.LastSeen{}, "agent_key", "key"); err != nil {
			util.Logger.ErrorF("Failed to rename column: %v\n", err)
			return err
		}
		sqlCommands := []string{
			`ALTER TABLE "last_seens" RENAME CONSTRAINT "agent_last_seens_pkey" TO "last_seens_pkey";`,
			`ALTER INDEX "agent_last_seens_pkey" RENAME TO "last_seens_pkey";`,
			`ALTER INDEX "idx_agent_last_seens_last_ping" RENAME TO "idx_last_seens_last_ping";`,
		}
		err := executeSQLCommands(db, sqlCommands)
		if err == nil {
			util.Logger.Info("Renamed table and column successfully.")
		}
		return err

	} else {
		performMigration(db, "*models.LastSeen", "System", func(d *gorm.DB) error {
			return d.AutoMigrate(&models.LastSeen{})
		})
	}

	return nil
}

// seedAgentTypeTable add agent types
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

// addCollectorTables Add collector and configurations tables
func addLogCollectorTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.Collector{},
		&models.CollectorConfigGroup{},
		&models.CollectorGroupConfigurations{},
	}
	for _, table := range tables {
		modelName := fmt.Sprintf("%T", table) // Generates a unique name based on the model type
		if !db.Migrator().HasTable(table) {
			performMigration(db, modelName, "System", func(d *gorm.DB) error {
				return d.AutoMigrate(table)
			})
		}
	}
	return nil
}

func addDeletedByFieldToCollector(db *gorm.DB) error {
	return db.AutoMigrate(&models.Collector{})
}

func deleteColumnModuleIdFromCollectorGroupConfig(db *gorm.DB) error {
	tableName := &models.CollectorConfigGroup{}
	columnName := "module_id"
	return deleteColumnFromTable(db, tableName, columnName)
}

// setConfigurationsPrimaryKey migrate the model to use the group id and the conf key as table primary key
// to avoid inconsistencies
func setConfigurationsPrimaryKey(db *gorm.DB) error {
	return db.AutoMigrate(&models.CollectorGroupConfigurations{})
}
