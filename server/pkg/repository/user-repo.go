package repository

import (
	"context"
	"fmt"

	"github.com/Astemirdum/user-app/server/models"
	"github.com/jmoiron/sqlx"
)

const (
	userTable = "users"
)

type UserPostgres struct {
	db *sqlx.DB
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{db}
}

func (u *UserPostgres) GetAll(ctx context.Context) ([]*models.User, error) {
	users := make([]*models.User, 0)
	query := fmt.Sprintf("select * from %s", userTable)
	if err := u.db.SelectContext(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserPostgres) Delete(ctx context.Context, id int) (bool, error) {
	query := fmt.Sprintf("delete from %s where id=$1", userTable)
	if _, err := u.db.ExecContext(ctx, query, id); err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserPostgres) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	query := fmt.Sprintf("select * from %s where email=$1", userTable)
	if err := u.db.GetContext(ctx, user, query, email); err != nil {
		return nil, err
	}
	return user, nil
}
func (u *UserPostgres) Create(ctx context.Context, user *models.User) (int, error) {
	var id int
	query := fmt.Sprintf("insert into %s (email, password_hash) values ($1, $2) returning id", userTable)
	if err := u.db.GetContext(ctx, &id, query, user.Email, user.Password); err != nil {
		return 0, err
	}
	return id, nil
}
