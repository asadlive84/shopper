package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asadlive84/shopper/api-gateway/internal/adapters/grpc"
	"github.com/asadlive84/shopper/api-gateway/internal/rabbitmq"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Event defines the structure for incoming WebSocket messages
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// EventManager handles all events and manages WebSocket clients
type EventManager struct {
	logger       *zap.Logger
	rabbitClient *rabbitmq.Client
	clients      map[*websocket.Conn]bool
	grpcClient   *grpc.UserGRPCClient
}

// NewEventManager initializes a new EventManager
func NewEventManager(logger *zap.Logger, rabbit *rabbitmq.Client, grpcClient *grpc.UserGRPCClient) *EventManager {
	return &EventManager{
		logger:       logger,
		rabbitClient: rabbit,
		clients:      make(map[*websocket.Conn]bool),
		grpcClient:   grpcClient,
	}
}

// HandleWebSocketMessage processes incoming WebSocket messages
func (em *EventManager) HandleWebSocketMessage(ctx context.Context, conn *websocket.Conn, msg []byte) {
	var event Event
	if err := json.Unmarshal(msg, &event); err != nil {
		em.logger.Error("Failed to parse WebSocket message", zap.Error(err))
		return
	}

	em.logger.Info("Received WebSocket event", zap.String("type", event.Type))

	switch event.Type {
	case "logout":
		var data struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			em.logger.Error("Failed to parse logout data", zap.Error(err))
			return
		}
		em.handleLogout(ctx, data.UserID)
	case "message":
		var data struct {
			UserID  string `json:"user_id"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			em.logger.Error("Failed to parse message data", zap.Error(err))
			return
		}
		em.handleMessage(ctx, data.UserID, data.Content)
	default:
		em.logger.Warn("Unknown event type", zap.String("type", event.Type))
	}
}

// handleLogout processes logout events
func (em *EventManager) handleLogout(ctx context.Context, userID string) {
	em.logger.Info("Processing logout event", zap.String("user_id", userID))

	msg := []byte("logout:" + userID)
	if err := em.rabbitClient.Publish(ctx, "logout_notifications", string(msg)); err != nil {
		em.logger.Error("Failed to publish logout event to RabbitMQ", zap.Error(err))
	}

	em.broadcast(msg)
}

// handleMessage processes message events
func (em *EventManager) handleMessage(ctx context.Context, userID, content string) {
	em.logger.Info("Processing message event", zap.String("user_id", userID), zap.String("content", content))

	msg := []byte(fmt.Sprintf("message:%s:%s", userID, content))
	if err := em.rabbitClient.Publish(ctx, "message_notifications", string(msg)); err != nil {
		em.logger.Error("Failed to publish message to RabbitMQ", zap.Error(err))
	}

	em.broadcast(msg)
}

// broadcast sends a message to all connected WebSocket clients
func (em *EventManager) broadcast(msg []byte) {
	for client := range em.clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			em.logger.Error("Failed to send message to client", zap.Error(err))
			client.Close()
			delete(em.clients, client)
		}
	}
}

func (em *EventManager) ConsumeEvents(queue string) {
	err := em.rabbitClient.Consume(queue, func(ctx context.Context, msg string) {
		em.logger.Info("Received event from RabbitMQ", zap.String("message", msg))

		parts := strings.SplitN(msg, ":", 3)
		if len(parts) < 2 {
			em.logger.Warn("Invalid message format", zap.String("message", msg))
			return
		}

		eventType := parts[0]
		switch eventType {
		case "logout":
			userID := parts[1]
			fmt.Println("========================================")
			fmt.Printf("userID %+v\n", userID)
			fmt.Println("========================================")
		// case "message":
		// 	if len(parts) < 3 {
		// 		em.logger.Warn("Invalid message event format", zap.String("message", msg))
		// 		return
		// 	}
		// 	userID := parts[1]
		// 	content := parts[2]
		// 	// Action: মেসেজ প্রসেস করা (যেমন, ডাটাবেসে সেভ করা)
		// 	em.logger.Info("Message received", zap.String("user_id", userID), zap.String("content", content))
		// 	// এখানে ডাটাবেসে সেভ করার লজিক যোগ করতে পারো

		default:
			em.logger.Warn("Unknown RabbitMQ event type", zap.String("type", eventType))
		}

		// WebSocket-এ ব্রডকাস্ট করা
		em.broadcast([]byte(msg))
	})
	if err != nil {
		em.logger.Error("Failed to consume from RabbitMQ", zap.Error(err))
	}
}

// AddClient adds a new WebSocket client
func (em *EventManager) AddClient(conn *websocket.Conn) {
	em.clients[conn] = true
}

// RemoveClient removes a WebSocket client
func (em *EventManager) RemoveClient(conn *websocket.Conn) {
	delete(em.clients, conn)
}
