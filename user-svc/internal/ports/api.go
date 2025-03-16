package ports

import (
	"context"
	"github.com/asadlive84/shopper/user-svc/internal/application/core/domain"
)

type APIPort interface {
	CreateUser(ctx context.Context, req *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	AuthenticateUser(ctx context.Context, req *domain.User) (*domain.UserAutenticate, error)
}
