package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Astemirdum/user-app/server"
	"github.com/Astemirdum/user-app/server/models"
	"github.com/Astemirdum/user-app/server/pkg/service"
	"github.com/Astemirdum/user-app/userpb"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cacheKey = "lol"
)

type Handler struct {
	userpb.UnimplementedUserServiceServer

	s     *service.Service
	cache *server.Cache
}

func NewHandler(srv *service.Service, cache *server.Cache) *Handler {
	return &Handler{
		s:     srv,
		cache: cache,
	}
}

func (h *Handler) CreateUser(ctx context.Context, req *userpb.CreateRequest) (*userpb.CreateResponse, error) {
	id, err := h.s.CreateUser(ctx, unmarshalUser(req.GetUser()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateUser: %v", err)
	}

	if err = h.cache.DeleteCache(ctx, cacheKey); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateUser DeleteCache: %v", err)
	}
	return &userpb.CreateResponse{Id: int32(id)}, nil
}

func (h *Handler) GetAllUser(_ *userpb.GetAllRequest, stream userpb.UserService_GetAllUserServer) error {
	ctx := stream.Context()
	users, err := h.cache.GetCache(ctx, cacheKey)
	if err != nil && err != redis.Nil {
		return status.Errorf(codes.Unavailable, "GetCache: %v", err)
	}
	if err == redis.Nil {
		logrus.Info("Empty cache")
		users, err = h.s.GetAllUser(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, "GetAllUser: %v", err)
		}
		if err = h.cache.SetCache(ctx, cacheKey, users); err != nil {
			return status.Errorf(codes.Internal, "SetCache: %v", err)
		}
	} else {
		logrus.Info("get users from cache")
	}
	usrs := marshalUsers(users)
	for _, u := range usrs {
		if err = stream.Send(&userpb.GetAllResponse{User: u}); err != nil {
			return status.Errorf(codes.Unavailable, "stream.Send: %v", err)
		}
	}
	return nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *userpb.DeleteRequest) (*userpb.DeleteResponse, error) {
	if _, err := h.s.DeleteUser(ctx, int(req.GetId())); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteUser %v", err)
	}
	if err := h.cache.DeleteCache(ctx, cacheKey); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteUser DeleteCache: %v", err)
	}
	return &userpb.DeleteResponse{}, nil
}

func (h *Handler) IssueToken(ctx context.Context, req *userpb.TokenRequest) (*userpb.TokenResponse, error) {
	token, err := h.s.GenerateToken(ctx, unmarshalUser(req.GetUser()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AuthUser %v", err)
	}

	return &userpb.TokenResponse{Token: &userpb.Token{
		Token: token,
		Valid: true,
	}}, nil

}

func (h *Handler) AuthUser(_ context.Context, req *userpb.AuthRequest) (*userpb.AuthResponse, error) {
	token := req.GetToken().GetToken()
	if _, err := h.s.ParseToken(token); err != nil {
		return nil, status.Errorf(codes.Unavailable, "ValidateToken %v", err)
	}
	return &userpb.AuthResponse{}, nil
}

func (h *Handler) AuthInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	start := time.Now()
	if info.FullMethod == "/userpb.UserService/DeleteUser" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "retrieving metadata failed")
		}
		authMD, ok := md["authorization"]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no auth details supplied")
		}
		headerToken := authMD[0]
		if headerToken == "" {
			return nil, status.Errorf(codes.Unauthenticated, "empty authMD")
		}

		headerTokenParts := strings.Split(headerToken, " ")
		logrus.Infof("%v", headerTokenParts)
		if len(headerTokenParts) != 2 || headerTokenParts[0] != "Bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		token := headerTokenParts[1]
		if len(token) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "empty token")
		}
		email, err := h.s.ParseToken(token)

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "token not valid %v", err)
		}
		emailToValid, ok := md["email"]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no email details supplied")
		}
		if email != emailToValid[0] {
			return nil, status.Errorf(codes.InvalidArgument, "email not fit token")
		}
		logrus.Printf("Validate Token email:%s passed: %s", emailToValid[0], token)
	}
	reply, err := handler(ctx, req)

	logrus.Printf("request - Method:%s  Duration:%s	Error:%v",
		info.FullMethod,
		time.Since(start),
		err)
	return reply, err
}

func unmarshalUser(user *userpb.User) *models.User {
	return &models.User{
		Id:       int(user.GetId()),
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
