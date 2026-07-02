package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SpinRequest от Streamer.bot
type SpinRequest struct {
	Roulette string `json:"roulette"` // имя рулетки: buffs, debuffs, etc.
}

// SpinResponse – выбранное событие
type SpinResponse struct {
	Event Event `json:"event"`
}

func spinHandler(roulettes map[string]*Roulette, hub *Hub) http.HandlerFunc {
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

		// Выбор с учётом весов (weighted random)
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

		// Отправляем результат оверлеям через WebSocket
		hub.Broadcast(selected, roulette.Name)

		// Возвращаем результат как подтверждение
		resp := SpinResponse{Event: selected}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func roulettesHandler(roulettes map[string]*Roulette) http.HandlerFunc {
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
		// Извлекаем имя из URL: /api/roulette/buffs
		path := strings.TrimPrefix(r.URL.Path, "/api/roulette/")
		if path == "" {
			http.Error(w, "Missing roulette name", http.StatusBadRequest)
			return
		}
		roulette, ok := roulettes[path]
		if !ok {
			http.Error(w, "Roulette not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roulette.Events)
	}
}
