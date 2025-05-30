package db

import (
	"log"

	"NeuroNest/internal/models"
)

func AutoMigrateTables() {
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Note{},
		&models.Tag{},
		&models.Attachment{},
		&models.ChatHistory{},
		&models.ActivityLog{},
		&models.IntegrationLog{},
		&models.Integration{},
	); err != nil {
		log.Fatalf("Ошибка при миграции таблиц: %v", err)
	}
	log.Println("Автомиграция таблиц завершена успешно")
}
