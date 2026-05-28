package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	wslib "github.com/game-ops/ai-alert-system/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WSHandler struct {
	db             *gorm.DB
	jwtAuth        *auth.JWT
	allowedOrigins []string
	upgrader       websocket.Upgrader
	hub            *wslib.Hub
	mu             sync.RWMutex
}

func NewWSHandler(db *gorm.DB, jwtAuth *auth.JWT, allowedOrigins []string) *WSHandler {
	h := &WSHandler{
		db:             db,
		jwtAuth:        jwtAuth,
		allowedOrigins: allowedOrigins,
		hub:            wslib.NewHub(),
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	// Start the hub's run loop
	go h.hub.Run()
	return h
}

func (h *WSHandler) HandleAlerts(c *gin.Context) {
	origin := c.GetHeader("Origin")
	log.Printf("WS connection attempt: origin=%s, remote=%s", origin, c.Request.RemoteAddr)

	if !h.isAllowedOrigin(origin) {
		log.Printf("WS rejected: origin not allowed: %s", origin)
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}

	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		log.Printf("WS rejected: no token provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter required"})
		return
	}

	user, _, err := middleware.AuthenticateToken(h.jwtAuth, h.db, tokenString)
	if err != nil {
		log.Printf("WS rejected: auth failed: %v", err)
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

	// Create a new client using the websocket module
	client := wslib.NewClient(h.hub, conn)

	// Register the client with the hub
	h.hub.Register(client)

	log.Printf("WebSocket client connected")

	// Send current active alerts directly to this client
	var alerts []models.Alert
	h.db.Where("status = ?", "firing").Order("trigger_time DESC").Limit(50).Find(&alerts)
	if len(alerts) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":   "init",
			"alerts": alerts,
		})
		// Send directly to this client's send queue
		select {
		case client.SendQueue() <- msg:
		default:
			// Client buffer full, will be handled by the hub
		}
	}

	// Start the client's read and write pumps
	// This blocks until the connection is closed
	client.Start()

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
	h.hub.Broadcast(msg)
}
