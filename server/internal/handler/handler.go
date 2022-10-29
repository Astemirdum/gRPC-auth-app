package handler

import (
	"context"

	"github.com/Astemirdum/user-app/server/internal/broker"
	"github.com/Astemirdum/user-app/server/internal/cache"
	"github.com/Astemirdum/user-app/server/internal/service"
	"github.com/Astemirdum/user-app/server/models"
	"github.com/Astemirdum/user-app/userpb"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	cacheKey = "lol"
)

type Handler struct {
	userpb.UnimplementedUserServiceServer

	s     *service.Service
	cache *cache.Cache
	prod  *broker.Producer
}

func NewHandler(
	srv *service.Service,
	cache *cache.Cache,
	producer *broker.Producer,
) *Handler {
	return &Handler{
		s:     srv,
		cache: cache,
		prod:  producer,
	}
}

func (h *Handler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user := pbToModelUser(req.GetUser())
	id, err := h.s.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateUser: %v", err)
	}
	user.Id = id

	if err = h.prod.Publish(user); err != nil {
		logrus.Warn("createUser- Publish fail:", err)
	}
	if err = h.cache.DeleteCache(ctx, cacheKey); err != nil {
		logrus.Warn("createUser- DeleteCache fail:", err)
	}
	return &userpb.CreateUserResponse{Id: int32(id)}, nil
}

func (h *Handler) GetAllUser(_ *userpb.GetAllUserRequest, stream userpb.UserService_GetAllUserServer) error {
	ctx := stream.Context()
	users, err := h.cache.GetCache(ctx, cacheKey)
	if err != nil && err != redis.Nil {
		return status.Errorf(codes.Unavailable, "GetCache: %v", err)
	}
	if err == redis.Nil {
		logrus.Debug("Empty cache")
		users, err = h.s.GetAllUser(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, "GetAllUser: %v", err)
		}
		if err = h.cache.SetCache(ctx, cacheKey, users); err != nil {
			logrus.Warn("SetCache", err)
		}
	}
	usrs := marshalUsers(users)
	for _, u := range usrs {
		if err = stream.Send(&userpb.GetAllUserResponse{User: u}); err != nil {
			return status.Errorf(codes.Internal, "stream.Send: %v", err)
		}
	}
	logrus.Debug("get users from cache")
	return nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	if _, err := h.s.DeleteUser(ctx, int(req.GetId())); err != nil {
		return nil, status.Errorf(codes.Internal, "deleteUser %v", err)
	}
	if err := h.cache.DeleteCache(ctx, cacheKey); err != nil {
		logrus.Warn("deleteUser- DeleteCache fail:", err)
	}
	return &userpb.DeleteUserResponse{}, nil
}

func (h *Handler) IssueToken(ctx context.Context, req *userpb.IssueTokenRequest) (*userpb.IssueTokenResponse, error) {
	token, err := h.s.GenerateToken(ctx, pbToModelUser(req.GetUser()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AuthUser %v", err)
	}

	return &userpb.IssueTokenResponse{Token: &userpb.Token{
		Token: token,
		Valid: true,
	}}, nil
}

func (h *Handler) AuthUser(_ context.Context, req *userpb.AuthUserRequest) (*userpb.AuthUserResponse, error) {
	token := req.GetToken().GetToken()
	if _, err := h.s.ParseToken(token); err != nil {
		return nil, status.Errorf(codes.Unavailable, "ValidateToken %v", err)
	}
	return &userpb.AuthUserResponse{}, nil
}

func pbToModelUser(user *userpb.User) *models.User {
	return &models.User{
		Email:    user.GetEmail(),
		Password: user.GetPassword(),
	}
}

func marshalUsers(usrs []*models.User) []*userpb.User {
	users := make([]*userpb.User, 0, len(usrs))
	for _, u := range usrs {
		users = append(users, &userpb.User{
			Id:       int32(u.Id),
			Email:    u.Email,
			Password: u.Password,
		})
	}
	return users
}
