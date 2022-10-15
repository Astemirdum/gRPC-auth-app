package service

import (
	"context"
	"time"

	"github.com/Astemirdum/user-app/server/internal/repository"
	"github.com/Astemirdum/user-app/server/models"
)

const (
	tokenTTL = time.Minute * 15
	salt     = "aslkfaDDo39@2u21!!*"
	signKey  = "MySecretKey"
)

type Service struct {
	UserRepo
	AuthService
}

type AuthService interface {
	ParseToken(accessToken string) (string, error)
	GenerateToken(ctx context.Context, user *models.User) (string, error)
}

type UserRepo interface {
	GetAll(ctx context.Context) ([]*models.User, error)
	Delete(ctx context.Context, id int) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (int, error)
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		UserRepo:    repo,
		AuthService: NewAuthService(repo),
	}
}
