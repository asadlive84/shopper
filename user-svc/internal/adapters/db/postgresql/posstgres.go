package postgresql

import (
	"context"

	"github.com/asadlive84/shopper/user-svc/internal/application/core/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



func Adapter(dsn string) (*PostgresDB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{})
	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) CreateUser(ctx context.Context, user *domain.User) error {
	return p.db.WithContext(ctx).Create(user).Error
}

func (p *PostgresDB) GetUser(ctx context.Context, id string) (*domain.User, error) {
	var user User
	err := p.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &domain.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}, err
}

func (p *PostgresDB) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user User
	err := p.db.WithContext(ctx).First(&user, "email = ?", email).Error
	return &domain.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}, err
}
