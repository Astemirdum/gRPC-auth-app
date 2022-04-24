package service

import (
	"context"
	"time"

	"github.com/Astemirdum/user-app/server/models"
	"github.com/Astemirdum/user-app/server/pkg/repository"
)

const (
	tokenTTL = time.Minute * 15
	salt     = "aslkfaDDo39@2u21!!*"
	signKey  = "MySecretKey"
)

type Service struct {
	repo *repository.Repository
	AuthService
}

type AuthService interface {
	ParseToken(accessToken string) (string, error)
	GenerateToken(ctx context.Context, user *models.User) (string, error)
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo:        repo,
		AuthService: NewAuthService(repo),
	}
}
