package websocket

import (
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConn implements websocketConn interface for testing
type mockConn struct {
	writeMessageCalls []writeCall
	writeMessageErr   error
	closed            bool
	closeErr          error
	mu                sync.Mutex
	done              chan struct{}
}

type writeCall struct {
	messageType int
	data        []byte
}

func newMockConn() *mockConn {
	return &mockConn{
		done: make(chan struct{}),
	}
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeMessageCalls = append(m.writeMessageCalls, writeCall{messageType, data})
	return m.writeMessageErr
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	close(m.done)
	return m.closeErr
}

func (m *mockConn) ReadMessage() (int, []byte, error) {
	<-m.done
	return 0, nil, websocket.ErrCloseSent
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetPongHandler(h func(appData string) error) {
}

func (m *mockConn) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// TestHubBroadcast tests that Hub broadcast works correctly
func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create mock clients
	conn1 := newMockConn()
	conn2 := newMockConn()

	client1 := &Client{hub: hub, conn: conn1, send: make(chan []byte, 256), done: make(chan struct{})}
	client2 := &Client{hub: hub, conn: conn2, send: make(chan []byte, 256), done: make(chan struct{})}

	// Register clients
	hub.Register(client1)
	hub.Register(client2)

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Verify client count
	assert.Equal(t, 2, hub.ClientCount())

	// Broadcast a message
	testMsg := []byte(`{"type":"test"}`)
	hub.Broadcast(testMsg)

	// Wait for message to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify both clients received the message via their send channel
	select {
	case msg := <-client1.send:
		assert.Equal(t, testMsg, msg)
	default:
		t.Error("client1 did not receive message")
	}

	select {
	case msg := <-client2.send:
		assert.Equal(t, testMsg, msg)
	default:
		t.Error("client2 did not receive message")
	}
}

// TestHubRegisterUnregister tests Hub register and unregister functionality
func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Initially no clients
	assert.Equal(t, 0, hub.ClientCount())

	// Create and register a client
	conn := newMockConn()
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), done: make(chan struct{})}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.ClientCount())

	// Unregister the client
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.ClientCount())
}

// TestClientCloseIdempotent tests that Client.Close is idempotent
func TestClientCloseIdempotent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn := newMockConn()
	client := newClientWithConn(hub, conn)
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Close multiple times - should not panic
	client.Close()
	client.Close()
	client.Close()

	// Verify connection was closed only once
	conn.mu.Lock()
	assert.True(t, conn.closed)
	conn.mu.Unlock()

	// Verify done channel is closed
	select {
	case <-client.Done():
		// Expected - done channel is closed
	default:
		t.Error("done channel should be closed")
	}
}

// TestClientDoneChannelClosesGoroutines tests that done channel stops all goroutines
func TestClientDoneChannelClosesGoroutines(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn := newMockConn()
	client := newClientWithConn(hub, conn)
	hub.Register(client)

	// Start the client
	go client.Start()

	// Wait for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Close the client
	client.Close()

	// Wait for goroutines to finish
	time.Sleep(50 * time.Millisecond)

	// Verify done channel is closed
	select {
	case <-client.Done():
		// Expected
	default:
		t.Error("done channel should be closed")
	}
}

// TestHubStop tests that Hub.Stop stops the run goroutine
func TestHubStop(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Wait for run to start
	time.Sleep(10 * time.Millisecond)

	// Register a client
	conn := newMockConn()
	client := newClientWithConn(hub, conn)
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.ClientCount())

	// Stop the hub
	hub.Stop()
	time.Sleep(10 * time.Millisecond)

	// Verify that broadcast doesn't block after stop (it should return immediately)
	done := make(chan bool)
	go func() {
		hub.Broadcast([]byte("test"))
		done <- true
	}()

	select {
	case <-done:
		// Expected - broadcast returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast should not block after Stop")
	}
}

// TestNewClientCreatesDoneChannel tests that NewClient creates a done channel
func TestNewClientCreatesDoneChannel(t *testing.T) {
	hub := NewHub()
	conn := newMockConn()

	client := newClientWithConn(hub, conn)

	require.NotNil(t, client.done, "done channel should be created")
	require.NotNil(t, client.send, "send channel should be created")
	require.NotNil(t, client.once, "once should be initialized")

	// Verify done channel is not closed initially
	select {
	case <-client.Done():
		t.Error("done channel should not be closed initially")
	default:
		// Expected
	}
}

// TestClientUnregistersOnClose tests that client unregisters from hub on close
func TestClientUnregistersOnClose(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn := newMockConn()
	client := newClientWithConn(hub, conn)
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.ClientCount())

	// Unregister via hub (which will close the client)
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.ClientCount())

	// Verify done channel is closed
	select {
	case <-client.Done():
		// Expected
	default:
		t.Error("done channel should be closed")
	}
}

// TestMultipleClientsBroadcast tests broadcasting to multiple clients
func TestMultipleClientsBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create and register multiple clients
	clients := make([]*Client, 5)
	for i := 0; i < 5; i++ {
		conn := newMockConn()
		client := newClientWithConn(hub, conn)
		clients[i] = client
		hub.Register(client)
	}

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 5, hub.ClientCount())

	// Broadcast a message
	testMsg := []byte(`{"type":"broadcast_test"}`)
	hub.Broadcast(testMsg)

	// Wait for message propagation
	time.Sleep(50 * time.Millisecond)

	// Verify all clients received the message
	for i, client := range clients {
		select {
		case msg := <-client.send:
			assert.Equal(t, testMsg, msg, "client %d should receive message", i)
		default:
			t.Errorf("client %d did not receive message", i)
		}
	}
}
