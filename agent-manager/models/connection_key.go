package models

import "gorm.io/gorm"

// ConnectionKey is what an agent presents to enrol. There is one per tenant:
// the key is the tenant credential, so an agent cannot choose which tenant it
// joins — it can only prove which key it was given.
type ConnectionKey struct {
	gorm.Model
	TenantID string `gorm:"not null;uniqueIndex"`
	Key      string `gorm:"not null;index"`
}
