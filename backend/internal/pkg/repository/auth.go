package repository

import (
	"context"
	"database/sql"

	"github.com/0x24CaptainParrot/gomarket/internal/models"
)

type AuthPostgres struct {
	db *sql.DB
}

func NewAuthPostgres(db *sql.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

func (a *AuthPostgres) CreateUser(ctx context.Context, user models.User) (int, error) {
	return 0, nil
}

func (a *AuthPostgres) GetUserByID(ctx context.Context, id int) (models.User, error) {
	return models.User{}, nil
}

func (a *AuthPostgres) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return models.User{}, nil
}

