package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type SpinRequest struct {
	Roulette string `json:"roulette"`
}

type SpinResponse struct {
	Event Event `json:"event"`
}

func roulettesListHandler(roulettes map[string]*Roulette) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := make([]string, 0, len(roulettes))
		for name := range roulettes {
			names = append(names, name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(names)
	}
}

func rouletteHandler(roulettes map[string]*Roulette) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/api/roulette/"):]
		roulette, ok := roulettes[name]
		if !ok {
			http.Error(w, "Roulette not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roulette.Events)
	}
}

func spinHandler(roulettes map[string]*Roulette, hub *Hub, sbClient *StreamerBotClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SpinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		roulette, ok := roulettes[req.Roulette]
		if !ok {
			http.Error(w, "Roulette not found", http.StatusNotFound)
			return
		}

		if len(roulette.Events) == 0 {
			http.Error(w, "No events in roulette", http.StatusInternalServerError)
			return
		}

		// Выбор события с весами
		totalWeight := 0
		for _, e := range roulette.Events {
			totalWeight += e.Weight
		}
		rnd := rand.Intn(totalWeight)
		var selected Event
		for _, e := range roulette.Events {
			rnd -= e.Weight
			if rnd < 0 {
				selected = e
				break
			}
		}

		// Отправляем событие в оверлей
		hub.Broadcast(selected, req.Roulette, roulette.Events)

		// --- ОТПРАВКА В ЧАТ через Streamer.bot ---
		if sbClient != nil {
			chatMessage := "🎲 Рулетка: " + selected.Name + " – " + selected.Description
			go func(msg string) {
				if err := sbClient.SendChatMessage(msg); err != nil {
					log.Printf("Failed to send chat message: %v", err)
				} else {
					log.Printf("Chat message sent: %s", msg)
				}
			}(chatMessage)
		}

		resp := SpinResponse{Event: selected}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
