package main

import (
	"log"
	"net/http"
)

func main() {
	roulettes, err := LoadRoulettes("./configs")
	if err != nil {
		log.Fatal("Failed to load roulettes:", err)
	}
	log.Printf("Loaded %d roulettes", len(roulettes))

	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/ws", wsHandler(hub))
	http.HandleFunc("/api/spin", spinHandler(roulettes, hub))
	http.HandleFunc("/api/roulettes", roulettesListHandler(roulettes)) // список имён
	http.HandleFunc("/api/roulette/", rouletteHandler(roulettes))      // события по имени
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server started on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
