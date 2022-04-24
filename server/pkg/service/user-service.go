package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/Astemirdum/user-app/server/models"
	"github.com/Astemirdum/user-app/server/pkg/repository"
	"github.com/dgrijalva/jwt-go"
)

type TokenService struct {
	*repository.Repository
}

func NewAuthService(repo *repository.Repository) AuthService {
	return &TokenService{repo}
}

func (u *Service) CreateUser(ctx context.Context, user *models.User) (int, error) {
	user.Password = genHashPassword(user.Password)
	return u.repo.Create(ctx, user)
}

func (u *Service) DeleteUser(ctx context.Context, id int) (bool, error) {
	return u.repo.Delete(ctx, id)
}

func (u *Service) GetAllUser(ctx context.Context) ([]*models.User, error) {
	return u.repo.GetAll(ctx)
}

type MyClaims struct {
	Email string
	jwt.StandardClaims
}

func (t *TokenService) GenerateToken(ctx context.Context, user *models.User) (string, error) {
	usr, err := t.Repository.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}
	if usr.Password != genHashPassword(user.Password) {
		return "", fmt.Errorf("wrong password")
	}
	claims := MyClaims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			Issuer:    "authapp.service.user",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	return token.SignedString([]byte(signKey))
}

func (t *TokenService) ParseToken(accessToken string) (string, error) {

	token, err := jwt.ParseWithClaims(accessToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signing method")
		}
		return []byte(signKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	myclaim, ok := token.Claims.(*MyClaims)
	if !ok {
		return "", fmt.Errorf("type assertion *MyClaims")
	}
	return myclaim.Email, nil
}

func genHashPassword(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
