package repository

import (
	"context"

	"github.com/Astemirdum/user-app/server/models"
	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	GetAll(ctx context.Context) ([]*models.User, error)
	Delete(ctx context.Context, id int) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (int, error)
}

type Repository struct {
	UserRepo
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{NewUserPostgres(db)}
}
