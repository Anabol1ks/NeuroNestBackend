package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatHistory struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Message   string `gorm:"not null"`
	Response  string `gorm:"not null"`
	Timestamp time.Time
	Embedding []byte `gorm:"type:bytea"` // Векторное представление сообщения
}
