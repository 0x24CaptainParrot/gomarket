package repository

import (
	"context"
	"database/sql"

	"github.com/0x24CaptainParrot/gomarket/internal/models"
)

type Repository struct {
	AuthorizationRepository AuthorizationRepository
}

type AuthorizationRepository interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	GetUserByID(ctx context.Context, id int) (models.User, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		AuthorizationRepository: NewAuthPostgres(db),
	}
}
