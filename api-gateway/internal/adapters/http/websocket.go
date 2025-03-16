package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

func (h *Handler) WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	h.eventManager.AddClient(conn)
	defer h.eventManager.RemoveClient(conn)

	h.logger.Info("WebSocket client connected", zap.String("remote_addr", c.Request.RemoteAddr))

	// পিং-পং যোগ করা
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket closed unexpectedly", zap.Error(err))
			} else {
				h.logger.Error("Failed to read WebSocket message", zap.Error(err))
			}
			break
		}
		h.logger.Info("Received message", zap.Int("type", msgType), zap.String("message", string(msg)))
		h.eventManager.HandleWebSocketMessage(c, conn, msg)
	}
}
