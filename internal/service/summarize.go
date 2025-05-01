package service

import (
	"NeuroNest/internal/config"
	"context"
	"fmt"
	"strings"

	"github.com/sheeiavellie/go-yandexgpt"
)

func SummarizeText(text string) (string, error) {
	if len(text) < 200 || wordCount(text) < 50 {
		return text, nil
	}

	ctx := context.Background()
	client := GetClient(ctx)

	modelURI := yandexgpt.MakeModelURI(config.CatalogID, yandexgpt.YandexGPT4Model)

	req := yandexgpt.YandexGPTRequest{
		ModelURI: modelURI,
		CompletionOptions: yandexgpt.YandexGPTCompletionOptions{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   1024,
		},
		Messages: []yandexgpt.YandexGPTMessage{
			{
				Role: yandexgpt.YandexGPTMessageRoleSystem,
				Text: "Вы — помощник, который кратко и точно суммирует предоставленный текст.",
			},
			{
				Role: yandexgpt.YandexGPTMessageRoleUser,
				Text: fmt.Sprintf("Пожалуйста, сделай краткое резюме этого текста (сжатие должно быть от 50%% до 80%%):\n\n%s", text),
			},
		},
	}

	resp, err := client.GetCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	summary := resp.Result.Alternatives[0].Message.Text
	return summary, nil
}

func wordCount(text string) int {
	return len(strings.Fields(text))
}
