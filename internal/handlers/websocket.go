package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type WSHandler struct {
	db        *gorm.DB
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mu        sync.RWMutex
}

func NewWSHandler(db *gorm.DB) *WSHandler {
	h := &WSHandler{
		db:        db,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte, 256),
	}
	// Start broadcast goroutine
	go h.run()
	return h
}

func (h *WSHandler) run() {
	for {
		msg := <-h.broadcast
		h.mu.RLock()
		for client := range h.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("WS write error: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mu.RUnlock()
	}
}

func (h *WSHandler) HandleAlerts(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	log.Printf("WebSocket client connected")

	// Send current active alerts
	var alerts []models.Alert
	h.db.Where("status = ?", "firing").Order("trigger_time DESC").Limit(50).Find(&alerts)
	if len(alerts) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":  "init",
			"alerts": alerts,
		})
		conn.WriteMessage(websocket.TextMessage, msg)
	}

	// Heartbeat
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				break
			}
		}
	}()

	// Read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
	log.Printf("WebSocket client disconnected")
}

// BroadcastAlert broadcasts an alert to all connected clients
func (h *WSHandler) BroadcastAlert(alert models.Alert) {
	msg, err := json.Marshal(map[string]interface{}{
		"type":  "new_alert",
		"alert": alert,
	})
	if err != nil {
		log.Printf("WS marshal error: %v", err)
		return
	}
	h.broadcast <- msg
}
