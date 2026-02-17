package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ldchengyi/linkflow/internal/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512
)

// Message WebSocket 推送消息格式
type Message struct {
	Type string      `json:"type"` // telemetry | device_status | alert | stats
	Data interface{} `json:"data"`
}

// Client 单个 WebSocket 连接
type Client struct {
	hub    *Hub
	userID string
	conn   *websocket.Conn
	send   chan []byte
}

// Hub 管理所有 WebSocket 连接，按 userID 分组
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]map[*Client]bool // userID → set of clients
	register   chan *Client
	unregister chan *Client
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动 Hub 事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			logger.Log.Infof("WS client connected: user_id=%s", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.userID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			logger.Log.Infof("WS client disconnected: user_id=%s", client.userID)
		}
	}
}

// SendToUser 向指定用户的所有连接推送消息
func (h *Hub) SendToUser(userID string, msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Errorf("WS marshal error: %v", err)
		return
	}

	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for client := range conns {
		select {
		case client.send <- data:
		default:
			// 缓冲区满，关闭连接
			h.unregister <- client
		}
	}
}

// ServeClient 创建客户端并启动读写协程
func ServeClient(hub *Hub, userID string, conn *websocket.Conn) {
	client := &Client{
		hub:    hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
	hub.register <- client
	go client.writePump()
	go client.readPump()
}

// readPump 读取客户端消息（主要处理 pong）
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump 向客户端写消息 + 定时 ping
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
