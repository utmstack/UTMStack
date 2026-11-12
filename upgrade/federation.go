package upgrade

import (
	"database/sql"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type connectionKeyRow struct {
	gorm.Model
	TenantID string `gorm:"not null;uniqueIndex"`
	Key      string `gorm:"not null;index"`
}

func (connectionKeyRow) TableName() string { return "connection_keys" }

// convertFederation ports v11 federation client tokens into the v12
// agentmanager DB. One row per distinct non-empty token; the default tenant
// is the only tenant a v11 install could have produced, so the token becomes
// that tenant's connection key.
func convertFederation(db *gorm.DB, amdb *sql.DB) error {
	if !tableExists(db, "utm_federation_service_client") {
		logSkipTable("utm_federation_service_client")
		return nil
	}

	var tokens []string
	if err := db.Raw(`SELECT fs_client_token FROM utm_federation_service_client
		WHERE fs_client_token IS NOT NULL AND fs_client_token <> '' ORDER BY id`).Scan(&tokens).Error; err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: amdb, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := gdb.AutoMigrate(&connectionKeyRow{}); err != nil {
		return err
	}

	tenantID := defaultTenantID.String()
	inserted := 0
	for _, token := range tokens {
		if err := gdb.Exec(`INSERT INTO connection_keys (tenant_id, key, created_at, updated_at)
			VALUES (?, ?, now(), now())
			ON CONFLICT (tenant_id) DO NOTHING`, tenantID, token).Error; err != nil {
			return err
		}
		inserted++
	}

	catcher.Info("v11 upgrade: federation tokens converted", map[string]any{"tokens": inserted})
	return nil
}
