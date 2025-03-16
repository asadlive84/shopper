package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asadlive84/shopper/api-gateway/internal/adapters/grpc"
	"github.com/asadlive84/shopper/api-gateway/internal/rabbitmq"

	// pb "github.com/asadlive84/shopper-proto/golang/user"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventManager struct {
	logger       *zap.Logger
	rabbitClient *rabbitmq.Client
	clients      map[*websocket.Conn]bool
	grpcClient   *grpc.UserGRPCClient
}

func NewEventManager(logger *zap.Logger, rabbit *rabbitmq.Client, grpcClient *grpc.UserGRPCClient) *EventManager {
	return &EventManager{
		logger:       logger,
		rabbitClient: rabbit,
		clients:      make(map[*websocket.Conn]bool),
		grpcClient:   grpcClient,
	}
}

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
			Email  string `json:"email"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			em.logger.Error("Failed to parse logout data", zap.Error(err))
			return
		}
		em.handleLogout(ctx, data.UserID, data.Email, msg)
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
	case "status_update":
		var data struct {
			UserID string `json:"user_id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			em.logger.Error("Failed to parse status data", zap.Error(err))
			return
		}
		em.handleStatusUpdate(ctx, data.UserID, data.Status)
	default:
		em.logger.Warn("Unknown event type", zap.String("type", event.Type))
	}
}

func (em *EventManager) handleLogout(ctx context.Context, userID, email string, originalMsg []byte) {
	em.logger.Info("Processing logout event", zap.String("user_id", userID), zap.String("email", email))
	// RabbitMQ-তে প্লেইন টেক্সট ফরম্যাটে পাঠানো
	rabbitMsg := []byte("logout:" + userID + ":" + email)
	if err := em.rabbitClient.Publish(ctx, "logout_notifications", string(rabbitMsg)); err != nil {
		em.logger.Error("Failed to publish logout to RabbitMQ", zap.Error(err))
	}
	// ক্লায়েন্টদের কাছে JSON ফরম্যাটে ব্রডকাস্ট
	em.broadcast(originalMsg)
}

// handleMessage, handleStatusUpdate, broadcast, ConsumeEvents, AddClient, RemoveClient আগের মতোই থাকবে
func (em *EventManager) handleMessage(ctx context.Context, userID, content string) {
	em.logger.Info("Processing message event", zap.String("user_id", userID), zap.String("content", content))
	msg := []byte(fmt.Sprintf("message:%s:%s", userID, content))
	if err := em.rabbitClient.Publish(ctx, "message_notifications", string(msg)); err != nil {
		em.logger.Error("Failed to publish message to RabbitMQ", zap.Error(err))
	}
	em.broadcast(msg)
}

func (em *EventManager) handleStatusUpdate(ctx context.Context, userID, status string) {
	em.logger.Info("Processing status update event", zap.String("user_id", userID), zap.String("status", status))
	msg := []byte(fmt.Sprintf("status:%s:%s", userID, status))
	if err := em.rabbitClient.Publish(ctx, "status_notifications", string(msg)); err != nil {
		em.logger.Error("Failed to publish status to RabbitMQ", zap.Error(err))
	}
	em.broadcast(msg)
}

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
	fmt.Println("===================ConsumeEvents=====================")
	fmt.Printf("queue %+v\n", queue)
	fmt.Println("========================================")
	err := em.rabbitClient.Consume(queue, func(ctx context.Context, msg string) {
		em.logger.Info("Received event from RabbitMQ", zap.String("message", msg))
		parts := strings.SplitN(msg, ":", 3)
		if len(parts) < 2 {
			em.logger.Warn("Invalid message format", zap.String("message", msg))
			return
		}

		eventType := parts[0]

		fmt.Println("================ConsumeEvents========================")
		fmt.Printf("eventType %+v\n\n", eventType)
		fmt.Println("========================================")

		switch eventType {
		case "logout":
			userID := parts[1]
			fmt.Println("================Logout========================")
			fmt.Printf("userID%+v\n", userID)
			fmt.Println("========================================")
			// _, err := em.grpcClient.CloseSession(ctx, &pb.CloseSessionRequest{UserId: userID})
			// if err != nil {
			// 	em.logger.Error("Failed to close session via gRPC", zap.Error(err))
			// } else {
			// 	em.logger.Info("Session closed successfully", zap.String("user_id", userID), zap.String("email", email))
			// }
		case "message":
			if len(parts) < 3 {
				em.logger.Warn("Invalid message event format", zap.String("message", msg))
				return
			}
			userID := parts[1]
			content := parts[2]
			em.logger.Info("Message received", zap.String("user_id", userID), zap.String("content", content))
		case "status":
			if len(parts) < 3 {
				em.logger.Warn("Invalid status event format", zap.String("message", msg))
				return
			}
			// userID := parts[1]
			// status := parts[2]
			// _, err := em.grpcClient.UpdateUserStatus(ctx, &pb.UpdateUserStatusRequest{
			// 	UserId: userID,
			// 	Status: status,
			// })
			// if err != nil {
			// 	em.logger.Error("Failed to update user status via gRPC", zap.Error(err))
			// } else {
			// 	em.logger.Info("User status updated", zap.String("user_id", userID), zap.String("status", status))
			// }
		default:
			em.logger.Warn("Unknown RabbitMQ event type", zap.String("type", eventType))
		}
		em.broadcast([]byte(msg))
	})
	if err != nil {
		em.logger.Error("Failed to consume from RabbitMQ", zap.Error(err))
	}
}

func (em *EventManager) AddClient(conn *websocket.Conn) {
	em.clients[conn] = true
}

func (em *EventManager) RemoveClient(conn *websocket.Conn) {
	delete(em.clients, conn)
}
