package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/asadlive84/shopper/user-svc/internal/application/core/domain"
	"github.com/asadlive84/shopper/user-svc/internal/ports"
	"github.com/google/uuid"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type APIService struct {
	db       ports.DBPort
	mongoDB  ports.MongoDBPort
	rabbitMQ ports.MessagingPort
	logger   *zap.Logger
}

func NewApplication(db ports.DBPort, mongoDB ports.MongoDBPort, rabbitMQ ports.MessagingPort, logger *zap.Logger) *APIService {
	return &APIService{
		db:       db,
		mongoDB:  mongoDB,
		rabbitMQ: rabbitMQ,
		logger:   logger,
	}
}
func (s *APIService) CreateUser(ctx context.Context, req *domain.User) (*domain.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create a new user object
	user := &domain.User{
		ID:       uuid.Must(uuid.NewRandom()),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Attempt to create the user in Postgres
	err = s.db.CreateUser(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			s.logger.Error("Duplicate email found", zap.String("email", user.Email))
			return nil, fmt.Errorf("%w: %s", domain.ErrUserAlreadyExists, user.Email)
		}
		s.logger.Error("Failed to create user in Postgres", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	// Publish to RabbitMQ
	msg := "User created: " + user.Email
	if err := s.rabbitMQ.Publish(ctx, "user_exchange", msg); err != nil {
		s.logger.Error("Failed to publish to RabbitMQ", zap.Error(err))
	}

	// Log in MongoDB
	if err := s.mongoDB.LogMessage(ctx, msg); err != nil {
		s.logger.Error("Failed to log message in MongoDB", zap.Error(err))
	}

	return user, nil
}

func (s *APIService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.db.GetUser(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user from Postgres", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (s *APIService) AuthenticateUser(ctx context.Context, req *domain.User) (*domain.UserAutenticate, error) {
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("User not found", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Invalid password", zap.String("email", req.Email))
		return nil, err
	}
	return &domain.UserAutenticate{
		User: domain.User{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		},
		Token: "dummy-token",
	}, nil
}
