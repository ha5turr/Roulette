package main

import (
	"encoding/json"
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

// GET /api/roulettes – возвращает список имён рулеток
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

// GET /api/roulette/{name} – возвращает события конкретной рулетки
func rouletteHandler(roulettes map[string]*Roulette) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем имя из URL
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

// POST /api/spin – запуск рулетки (без изменений)
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

		hub.Broadcast(selected, req.Roulette, roulette.Events)

		resp := SpinResponse{Event: selected}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
