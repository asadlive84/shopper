package ports

import (
	"context"
	"github.com/asadlive84/shopper/user-svc/internal/application/core/domain"

)

type DBPort interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type MongoDBPort interface {
	LogMessage(ctx context.Context, msg string) error
}