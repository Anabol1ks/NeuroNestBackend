package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityLog struct {
	gorm.Model
	UserID      uint   `gorm:"not null"`
	Action      string `gorm:"not null"` // Например: "added_note", "edited_note", "deleted_note"
	Description string
	Timestamp   time.Time
}

type IntegrationLog struct {
	gorm.Model
	IntegrationID uint   `gorm:"not null"` // Связь с интеграцией
	Action        string `gorm:"not null"` // Действие (например, "imported_email", "synced_telegram")
	Details       string // Дополнительные данные о действии
}
