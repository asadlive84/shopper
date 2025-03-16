package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSData struct {
	UserID string
	Msg    map[string]interface{}
}

type WebSocketMessage struct {
	Type string `json:"type"`
	Data WSData `json:"data"`
}

func (h *Handler) WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket আপগ্রেডে সমস্যা", zap.Error(err))
		return
	}
	defer conn.Close()

	h.eventManager.AddClient(conn)
	defer h.eventManager.RemoveClient(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket বন্ধ হয়ে গেছে", zap.Error(err))
			}
			break
		}
		h.eventManager.HandleWebSocketMessage(c, conn, msg)
	}
}

// func (h *Handler) WebSocketHandler(c *gin.Context) {
// 	// Upgrade HTTP connection to WebSocket
// 	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		h.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
// 		return
// 	}
// 	defer conn.Close()

// 	// Add the new client to the clients map
// 	h.clients[conn] = true
// 	defer delete(h.clients, conn) // Remove the client when the connection closes

// 	// Handle incoming WebSocket messages
// 	for {
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			h.logger.Error("Failed to read WebSocket message", zap.Error(err))
// 			break
// 		}
// 		h.logger.Info("Received WebSocket message", zap.String("message", string(msg)))

// 		// Parse the message
// 		var message WebSocketMessage
// 		if err := json.Unmarshal(msg, &message); err != nil {
// 			h.logger.Error("Failed to parse WebSocket message", zap.Error(err))
// 			continue
// 		}

// 		// Handle the message based on its type
// 		switch message.Type {
// 		case "logout":
// 			userID := message.Data.UserID
// 			h.handleLogoutEvent(c, userID)
// 		default:
// 			h.logger.Warn("Unknown message type", zap.String("type", message.Type))
// 		}

// 		// Publish the message to RabbitMQ
// 		err = h.rabbitClient.Publish(c, "logout_notifications", string(msg))
// 		if err != nil {
// 			h.logger.Error("Failed to publish message to RabbitMQ", zap.Error(err))
// 		}

// 		// Broadcast the message to all connected clients
// 		for client := range h.clients {
// 			err := client.WriteMessage(websocket.TextMessage, msg)
// 			if err != nil {
// 				h.logger.Error("Failed to send message to client", zap.Error(err))
// 				client.Close()
// 				delete(h.clients, client) // Remove the client if there's an error
// 			}
// 		}
// 	}
// }
