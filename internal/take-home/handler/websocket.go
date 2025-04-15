package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Timeouts
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// in a production-ready system you would validate origins instead of setting this to true
	CheckOrigin: func(r *http.Request) bool { return true },
	// to prevent cross-origin attacks, you'd implement something along the lines of:
	/*
	 CheckOrigin: func(r *http.Request) bool {
	        origin := r.Header.Get("Origin")

	        // allow connections only from trusted domains
	        allowedOrigins := []string{
	            "https://oofone-frontend.com",
	            "https://app.oofone.com",
	        }

	        for _, allowed := range allowedOrigins {
	            if origin == allowed {
	                return true
	            }
	        }

	        log.Printf("Rejected WebSocket connection from origin: %s", origin)
	        return false
	    },
	*/
}

type WebSocketHubImpl struct {
	clients    map[*websocket.Conn]struct{}
	clientsMu  sync.RWMutex
	broadcast  chan *model.WeatherData
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	logger     *zap.Logger
}

func NewWebSocketHub(logger *zap.Logger) WebSocketHub {
	return &WebSocketHubImpl{
		broadcast:  make(chan *model.WeatherData, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]struct{}),
		logger:     logger.Named("websocket_hub"),
	}
}

func (h *WebSocketHubImpl) Run(ctx context.Context) {
	h.logger.Info("Starting WebSocket hub")
	defer h.logger.Info("WebSocket hub stopped")

	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			h.clients[client] = struct{}{}
			h.clientsMu.Unlock()
			h.logger.Debug("Client registered", zap.Int("count", len(h.clients)))

		case client := <-h.unregister:
			h.safeRemoveClient(client)

		case data := <-h.broadcast:
			h.broadcastToClients(data)

		case <-ctx.Done():
			h.cleanup()
			return
		}
	}
}

func (h *WebSocketHubImpl) safeRemoveClient(conn *websocket.Conn) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if _, exists := h.clients[conn]; exists {
		conn.Close()
		delete(h.clients, conn)
		h.logger.Debug("Client unregistered", zap.Int("count", len(h.clients)))
	}
}

func (h *WebSocketHubImpl) broadcastToClients(data *model.WeatherData) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	if len(h.clients) == 0 {
		return
	}

	for client := range h.clients {
		if err := h.writeData(client, data); err != nil {
			h.logger.Warn("Write failed", zap.Error(err))
			go func(c *websocket.Conn) { h.unregister <- c }(client)
		}
	}
}

func (h *WebSocketHubImpl) writeData(conn *websocket.Conn, data *model.WeatherData) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(data)
}

func (h *WebSocketHubImpl) cleanup() {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for client := range h.clients {
		client.Close()
	}
	h.logger.Info("Cleaned up all WebSocket connections")
}

func (h *WebSocketHubImpl) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Upgrade failed", zap.Error(err))
		return
	}

	// connection config
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// register client
	h.register <- conn
	defer func() { h.unregister <- conn }()

	// start heartbeat goroutine
	go h.heartbeat(conn)

	// blocking to keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *WebSocketHubImpl) heartbeat(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop() // immediately release resources rather than waiting for the GC to operate through the NewTicker() instance

	for range ticker.C {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

func (h *WebSocketHubImpl) Broadcast(data *model.WeatherData) {
	select {
	case h.broadcast <- data:
	default:
		h.logger.Warn("Broadcast channel full - dropping message")
	}
}
