package repository

import (
	"authapp/server/entity"
	"context"

	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	GetAll(ctx context.Context) ([]*entity.User, error)
	Delete(ctx context.Context, id int) (bool, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) (int, error)
}

type Repository struct {
	UserRepo
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{NewUserPostgres(db)}
}
