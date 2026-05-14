package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket client connections and message broadcasting.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	stop       chan struct{}
	once       sync.Once
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Run starts the Hub's main loop for handling client registration and broadcasting.
// It should be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop:
			// Close all clients before stopping
			h.mu.Lock()
			for client := range h.clients {
				client.removeFromHub()
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered, total: %d", h.ClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.removeFromHub()
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered, total: %d", h.ClientCount())

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// Client's send buffer is full, close it
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					client.removeFromHub()
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop stops the Hub and closes all client connections.
// It is safe to call multiple times.
func (h *Hub) Stop() {
	h.once.Do(func() {
		close(h.stop)
	})
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	case <-h.stop:
		// Hub is stopped, don't send
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
	case <-h.stop:
		// Hub is stopped
	}
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	case <-h.stop:
		// Hub is stopped
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// BroadcastChannel returns the broadcast channel for testing purposes.
func (h *Hub) BroadcastChannel() chan []byte {
	return h.broadcast
}

// websocketConn is an interface for websocket.Conn, used for testing.
type websocketConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (int, []byte, error)
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetPongHandler(h func(appData string) error)
}

// Client represents a WebSocket client connection.
type Client struct {
	hub  *Hub
	conn websocketConn
	send chan []byte
	done chan struct{}
	once sync.Once

	// Configuration
	writeWait      time.Duration
	pongWait       time.Duration
	pingPeriod     time.Duration
	maxMessageSize int64

	// closed tracks if client has been removed from hub
	closed bool
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		done:           make(chan struct{}),
		writeWait:      10 * time.Second,
		pongWait:       60 * time.Second,
		pingPeriod:     54 * time.Second,
		maxMessageSize: 512,
	}
}

// newClientWithConn creates a new Client with a websocketConn interface.
// This is used for testing with mock connections.
func newClientWithConn(hub *Hub, conn websocketConn) *Client {
	return &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		done:           make(chan struct{}),
		writeWait:      10 * time.Second,
		pongWait:       60 * time.Second,
		pingPeriod:     54 * time.Second,
		maxMessageSize: 512,
	}
}

// Start starts the client's read and write pumps.
// It should be called in a goroutine.
func (c *Client) Start() {
	// Start write pump in a separate goroutine
	go c.writePump()

	// Run read pump in current goroutine (blocks)
	c.readPump()
}

// Done returns the done channel for this client.
// The channel is closed when the client is closed.
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// readPump pumps messages from the WebSocket connection to the hub.
// It runs in a single goroutine per connection.
func (c *Client) readPump() {
	defer c.Close()

	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	for {
		select {
		case <-c.done:
			// Client is closing, exit read pump
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// It runs in a single goroutine per connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-c.done:
			// Client is closing, send close message and exit
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping for heartbeat
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket ping error: %v", err)
				return
			}
		}
	}
}

// Close closes the client connection.
// It is safe to call multiple times (idempotent).
func (c *Client) Close() {
	c.once.Do(func() {
		// Signal all goroutines to stop
		close(c.done)

		// Close the websocket connection
		if c.conn != nil {
			c.conn.Close()
		}
	})
}

// removeFromHub removes the client from the hub's client map.
// This is called by the hub when unregistering a client.
func (c *Client) removeFromHub() {
	c.once.Do(func() {
		close(c.done)
		if c.conn != nil {
			c.conn.Close()
		}
	})
}

// SendQueue returns the send channel for this client.
// This is useful for testing.
func (c *Client) SendQueue() chan []byte {
	return c.send
}
