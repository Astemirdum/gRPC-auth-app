package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	*UserPostgres
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{NewUserPostgres(db)}
}
