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

	// Создаём WebSocket hub
	hub := NewHub()
	go hub.Run()

	// Роутинг
	http.HandleFunc("/ws", wsHandler(hub))
	http.HandleFunc("/api/spin", spinHandler(roulettes, hub))
	// Добавим в main.go роуты:
	http.HandleFunc("/api/roulettes", roulettesHandler(roulettes))
	http.HandleFunc("/api/roulette/", rouletteHandler(roulettes)) // обрати внимание на слеш в конце для пути с параметром
	http.Handle("/", http.FileServer(http.Dir("./static")))       // оверлей

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
