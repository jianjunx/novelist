package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type WSManager struct {
	clients map[uuid.UUID]map[*websocket.Conn]bool
	mu      sync.RWMutex
}

var WS = &WSManager{clients: make(map[uuid.UUID]map[*websocket.Conn]bool)}

func (m *WSManager) Add(userID uuid.UUID, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clients[userID] == nil {
		m.clients[userID] = make(map[*websocket.Conn]bool)
	}
	m.clients[userID][conn] = true
}

func (m *WSManager) Remove(userID uuid.UUID, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clients[userID] != nil {
		delete(m.clients[userID], conn)
	}
}

func (m *WSManager) Send(userID uuid.UUID, msg interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for conn := range m.clients[userID] {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("WS send error: %v", err)
		}
	}
}

func HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	uid := userID.(uuid.UUID)
	WS.Add(uid, conn)
	defer WS.Remove(uid, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
