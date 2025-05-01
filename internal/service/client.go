package service

import (
	"NeuroNest/internal/config"
	"context"
	"log"

	"github.com/sheeiavellie/go-yandexgpt"
)

func GetClient(ctx context.Context) *yandexgpt.YandexGPTClient {
	if config.IAMtoken == "" {
		log.Fatal("IAM token is not set")
	}
	client := yandexgpt.NewYandexGPTClientWithIAMToken(config.IAMtoken)
	// // Обновляем токен из CLI-профиля или переменных окружения
	// if err := client.GetIAMToken(ctx); err != nil {
	// 	log.Fatalf("failed to refresh IAM token: %v", err)
	// }
	return client
}
