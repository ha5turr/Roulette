package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type StreamerBotClient struct {
	baseURL string
}

func NewStreamerBotClient(host string, port int) *StreamerBotClient {
	return &StreamerBotClient{
		// Формируем URL для HTTP-сервера Streamer.bot
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
	}
}

// SendChatMessage отправляет сообщение в чат через вызов Action в Streamer.bot
func (s *StreamerBotClient) SendChatMessage(message string) error {
	// Формируем запрос на выполнение действия
	payload := map[string]interface{}{
		"action": map[string]interface{}{
			// Сюда вставляем скопированный ID твоего действия
			"id": "7b280d3a-6dd2-4dc4-82b6-66c637433dcc",
		},
		"args": map[string]string{
			// Это имя переменной (args), которую мы создали в Action
			// Её значение будет подставлено в поле message в Streamer.bot
			"message": message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Отправляем POST-запрос на эндпоинт /DoAction
	resp, err := http.Post(s.baseURL+"/DoAction", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Успешный ответ от Streamer.bot — это статус 204 No Content[reference:5]
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("Chat message sent: %s", message)
	return nil
}
