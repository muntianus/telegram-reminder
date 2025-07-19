package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Проверяем переменные окружения
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ OPENAI_API_KEY не установлен")
		fmt.Println("Установите переменную: export OPENAI_API_KEY=your_key_here")
		return
	}

	// Создаем клиент OpenAI
	client := openai.NewClient(apiKey)

	// Тестируем разные модели
	models := []string{"gpt-4o", "gpt-4o-mini", "o3"}

	for _, model := range models {
		fmt.Printf("\n🧪 Тестируем модель: %s\n", model)
		fmt.Println(strings.Repeat("=", 50))

		// Тестируем крипто дайджест
		fmt.Printf("📊 Крипто дайджест (%s):\n", model)
		ctx, cancel := context.WithTimeout(context.Background(), 30)
		defer cancel()

		resp, err := botpkg.SystemCompletion(ctx, client, botpkg.CryptoDigestPrompt, model)
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
		} else {
			fmt.Printf("✅ Успех! Длина ответа: %d символов\n", len(resp))
			if len(resp) > 100 {
				fmt.Printf("📝 Начало ответа: %s...\n", resp[:100])
			}
		}

		fmt.Println()
	}
}
