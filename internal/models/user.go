package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nickname     string  `gorm:"not null"`
	Email        string  `gorm:"unique;not null"`
	PasswordHASH string  `gorm:"not null"`
	YandexID     *string `gorm:"unique"`
	FirstName    string
	LastName     string
	ProfilePic   string // Ссылка на фото профиля
	Role         string `gorm:"not null;default:'user'"` // Например: user, admin
}
