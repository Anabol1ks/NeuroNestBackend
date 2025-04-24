package db

import (
	"log"

	"NeuroNest/internal/models"
)

func AutoMigrateTables() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Note{},
		&models.Tag{},
		&models.Topic{},
		&models.Attachment{},
		&models.ChatHistory{},
		&models.ActivityLog{},
		&models.IntegrationLog{},
		&models.Integration{},
	)
	if err != nil {
		log.Fatalf("Ошибка при миграции таблиц: %v", err)
	}

	log.Println("Автомиграция таблиц завершена успешно")
}
