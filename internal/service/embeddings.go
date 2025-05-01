package service

import (
	"NeuroNest/internal/config"
	"context"

	"github.com/sheeiavellie/go-yandexgpt"
)

func GenerateEmbedding(text string) ([]float64, error) {
	ctx := context.Background()
	client := GetClient(ctx)
	modelURI := yandexgpt.MakeEmbModelURI(config.CatalogID, yandexgpt.TextSearchQuery)

	// 5. Делаем запрос на embedding :contentReference[oaicite:4]{index=4}
	req := yandexgpt.YandexGPTEmbeddingsRequest{
		ModelURI: modelURI,
		Text:     "test",
	}
	resp, err := client.GetEmbedding(ctx, req)
	if err != nil {
		return nil, err
	}

	emb := resp.Embedding
	return emb, nil
}
