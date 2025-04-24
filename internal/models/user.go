package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email      string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	YandexID   string `gorm:"unique"`
	TelegramID string `gorm:"unique"`
	FirstName  string
	LastName   string
	ProfilePic string // Ссылка на фото профиля
	Role       string // Например: user, admin
}
