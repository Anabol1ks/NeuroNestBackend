package models

import "gorm.io/gorm"

type Integration struct {
	gorm.Model
	UserID   uint   `gorm:"not null"` // Владелец интеграции
	Type     string `gorm:"not null"` // Тип интеграции (например, "telegram", "gmail")
	Settings string // JSON с настройками интеграции
}
