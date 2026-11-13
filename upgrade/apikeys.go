package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

var apiKeyNamespace = uuid.MustParse("1f52d62b-6fdb-b2db-ca5e-fd519a49fafd")

type v11APIKey struct {
	ID          int64
	UserID      int64
	Name        string
	APIKey      string `gorm:"column:api_key"`
	AllowedIP   *string
	CreatedAt   time.Time
	GeneratedAt time.Time
	ExpiresAt   *time.Time
}

func apiKeyID(v11ID int64) string {
	return uuid.NewSHA1(apiKeyNamespace, []byte(strconv.FormatInt(v11ID, 10))).String()
}

func convertAPIKeys(db *gorm.DB) error {
	if !tableExists(db, "api_keys_v11") {
		return nil
	}

	var src []v11APIKey
	if err := db.Raw(`SELECT id, user_id, name, api_key, allowed_ip, created_at, generated_at, expires_at
		FROM api_keys_v11 ORDER BY id`).Scan(&src).Error; err != nil {
		return err
	}

	count := 0
	for _, k := range src {
		userID := v11UserUUID(k.UserID)
		if err := db.Exec(
			`INSERT INTO api_keys (id, user_id, tenant_id, name, key_hash, allowed_ip, created_at, generated_at, expires_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO NOTHING`,
			apiKeyID(k.ID), userID, defaultTenantID, k.Name,
			hashSHA256(k.APIKey), strFrom(k.AllowedIP),
			k.CreatedAt, k.GeneratedAt, k.ExpiresAt,
		).Error; err != nil {
			return err
		}
		count++
	}

	catcher.Info("v11 upgrade: api_keys converted", map[string]any{"keys": count})
	return nil
}

func hashSHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
