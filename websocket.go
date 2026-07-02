package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			delete(h.clients, conn)
			conn.Close()
		case msg := <-h.broadcast:
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					conn.Close()
					delete(h.clients, conn)
				}
			}
		}
	}
}

// Broadcast отправляет событие всем клиентам
func (h *Hub) Broadcast(event Event, rouletteName string) {
	payload := map[string]interface{}{
		"action":   "spin",
		"roulette": rouletteName,
		"event":    event,
	}
	data, _ := json.Marshal(payload)
	h.broadcast <- data
}

func wsHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hub.register <- conn
		defer func() {
			hub.unregister <- conn
		}()
		// Ждём закрытия (read loop)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
