package main

import (
	"log"
	"net/http"
)

func main() {
	// Загружаем рулетки
	roulettes, err := LoadRoulettes("./configs")
	if err != nil {
		log.Fatal("Failed to load roulettes:", err)
	}
	log.Printf("Loaded %d roulettes", len(roulettes))

	// Инициализируем WebSocket-клиент для Streamer.bot
	// wsURL := os.Getenv("STREAMERBOT_WS_URL")
	// if wsURL == "" {
	// 	wsURL = "ws://localhost:8080" // стандартный порт Streamer.bot
	// 	log.Println("Using default Streamer.bot WebSocket URL:", wsURL)
	// }
	sbClient := NewStreamerBotClient("localhost", 7474) // Используй свой порт
	// if err := sbClient.Connect(); err != nil {
	// 	log.Printf("Warning: Could not connect to Streamer.bot: %v", err)
	// 	// не останавливаем сервер, просто логируем
	// }

	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/ws", wsHandler(hub))
	http.HandleFunc("/api/spin", spinHandler(roulettes, hub, sbClient))
	http.HandleFunc("/api/roulettes", roulettesListHandler(roulettes))
	http.HandleFunc("/api/roulette/", rouletteHandler(roulettes))
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server started on :3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
