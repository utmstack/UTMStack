package models

import "gorm.io/gorm"

type ConnectionKey struct {
	gorm.Model
	Key string `gorm:"not null"`
}
