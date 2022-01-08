package service

import (
	"authapp/server/entity"
	"authapp/server/pkg/repository"
	"time"

	"golang.org/x/net/context"
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
	GenerateToken(ctx context.Context, user *entity.User) (string, error)
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo:        repo,
		AuthService: NewAuthService(repo),
	}
}
