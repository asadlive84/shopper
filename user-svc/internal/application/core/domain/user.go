package domain

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}

type UserAutenticate struct {
	User  User
	Token string
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrDatabaseError     = errors.New("database error")
)
