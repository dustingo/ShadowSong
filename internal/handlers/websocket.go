package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WSHandler struct {
	db             *gorm.DB
	jwtAuth        *auth.JWT
	allowedOrigins []string
	upgrader       websocket.Upgrader
	clients        map[*websocket.Conn]bool
	broadcast      chan []byte
	mu             sync.RWMutex
}

func NewWSHandler(db *gorm.DB, jwtAuth *auth.JWT, allowedOrigins []string) *WSHandler {
	h := &WSHandler{
		db:             db,
		jwtAuth:        jwtAuth,
		allowedOrigins: allowedOrigins,
		clients:        make(map[*websocket.Conn]bool),
		broadcast:      make(chan []byte, 256),
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
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
	if !h.isAllowedOrigin(c.GetHeader("Origin")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}

	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter required"})
		return
	}

	user, _, err := middleware.AuthenticateToken(h.jwtAuth, h.db, tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if user.RequiresPasswordReset() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password reset required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
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
			"type":   "init",
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

func (h *WSHandler) checkOrigin(r *http.Request) bool {
	return h.isAllowedOrigin(r.Header.Get("Origin"))
}

func (h *WSHandler) isAllowedOrigin(origin string) bool {
	if strings.TrimSpace(origin) == "" {
		return false
	}

	for _, allowed := range h.allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
			continue
		}

		if origin == allowed {
			return true
		}
	}

	return false
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
