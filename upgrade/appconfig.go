package upgrade

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

var configUUIDNamespace = uuid.MustParse("04f5f62e-5ec0-4042-825b-2c7b8a38ec73")

type v11Config struct {
	ID         int64
	Key        *string `gorm:"column:conf_param_short"`
	Label      *string `gorm:"column:conf_param_large"`
	Desc       *string `gorm:"column:conf_param_description"`
	Value      *string `gorm:"column:conf_param_value"`
	Datatype   *string `gorm:"column:conf_param_datatype"`
	ModifiedAt *time.Time
	ModifiedBy *string
}

func convertAppConfig(db *gorm.DB) error {
	if !tableExists(db, "utm_configuration_parameter") {
		logSkipTable("utm_configuration_parameter")
		return nil
	}

	var src []v11Config
	if err := db.Raw(`SELECT id, conf_param_short, conf_param_large,
		conf_param_description, conf_param_value, conf_param_datatype,
		modification_time, modification_user
		FROM utm_configuration_parameter ORDER BY id`).Scan(&src).Error; err != nil {
		return err
	}

	seen := make(map[string]bool)
	inserted := 0
	for _, c := range src {
		key := strings.TrimSpace(strFrom(c.Key))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true

		value := strFrom(c.Value)
		isSecret := strings.EqualFold(strFrom(c.Datatype), "password")
		if isSecret && value == "*****" {
			value = ""
		}

		if err := db.Exec(
			`INSERT INTO app_config (id, tenant_id, key, label, description, value, is_secret, updated_at, updated_by)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (tenant_id, key) DO UPDATE SET
				value = EXCLUDED.value,
				is_secret = EXCLUDED.is_secret,
				label = EXCLUDED.label,
				description = EXCLUDED.description,
				updated_at = EXCLUDED.updated_at,
				updated_by = EXCLUDED.updated_by`,
			uuid.NewSHA1(configUUIDNamespace, []byte(key)).String(), defaultTenantID, key,
			strFrom(c.Label), strFrom(c.Desc), value, isSecret, c.ModifiedAt, strFrom(c.ModifiedBy),
		).Error; err != nil {
			return err
		}
		inserted++
	}

	catcher.Info("v11 upgrade: app_config converted", map[string]any{"rows": inserted})
	return nil
}
