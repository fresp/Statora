package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// dialTestWS upgrades a real WebSocket client against ServeWs on an httptest
// server and registers cleanup for both server and connection.
func dialTestWS(t *testing.T, hub *Hub) *websocket.Conn {
	t.Helper()

	router := gin.New()
	router.GET("/ws", ServeWs(hub))

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn
}

// TestWebSocketHubBroadcast verifies the gorilla/websocket v1.5.3 upgrade end
// to end: HTTP upgrade via Gin, client registration, and broadcast fanout.
func TestWebSocketHubBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	go hub.Run()

	conn := dialTestWS(t, hub)

	hub.broadcast <- []byte("ping_test")

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, "ping_test", string(message))
}

// TestWebSocketBroadcastEventEnvelope verifies the typed event helper still
// marshals and fans out WSEvent envelopes after the upgrade.
func TestWebSocketBroadcastEventEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	go hub.Run()

	conn := dialTestWS(t, hub)

	BroadcastEvent(hub, "monitor.updated", map[string]string{"id": "abc"})

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"monitor.updated","data":{"id":"abc"}}`, string(message))
}
